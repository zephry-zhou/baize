package memory

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

type MEMORY struct {
	MemTotal       string     `json:"System Memory,omitempty"`
	MemFree        string     `json:"SysMem Free,omitempty"`
	MemAvailable   string     `json:"SysMem Available,omitempty"`
	Buffer         string     `json:"Buffer,omitempty"`
	Cached         string     `json:"Cached,omitempty"`
	SwapCached     string     `json:"Swap Cached,omitempty"`
	SwapTotal      string     `json:"Swap Total,omitempty"`
	SwapFree       string     `json:"Swap Free,omitempty"`
	VmallocTotal   string     `json:"Vmalloc Total,omitempty"`
	VmallocUsed    string     `json:"Vmalloc Used,omitempty"`
	VmallocChunk   string     `json:"Vmalloc Chunk,omitempty"`
	Hugepagesize   string     `json:"Huge Pape Size,omitempty"`
	DirectMap4k    string     `json:"Direct Map4K,omitempty"`
	DirectMap2M    string     `json:"Direct Map2M,omitempty"`
	DirectMap1G    string     `json:"Direct Map1G,omitempty"`
	SlotMax        string     `json:"Slot Max,omitempty"`
	SlotUsed       string     `json:"Slot Used,omitempty"`
	PhyMem         string     `json:"Physical Memory,omitempty"`
	Diagnose       string     `json:"Diagnose,omitempty"`
	DiagnoseDetail string     `json:"Diagnose Detail,omitempty"`
	MemEntities    []dmiMem   `json:"Memory Entities,omitempty"`
	EdacInfo       []edacInfo `json:"EDAC Info,omitempty"`
}

type dmiMem struct {
	TotalWidth       string `json:"Total Width,omitempty"`
	DataWidth        string `json:"Data Width,omitempty"`
	Size             string `json:"Size,omitempty"`
	FormFactor       string `json:"Form Factor,omitempty"`
	BankLocator      string `json:"Bank Locator,omitempty"`
	Type             string `json:"Type,omitempty"`
	TypeDetail       string `json:"Type Detail,omitempty"`
	MaxSpeed         string `json:"Max Speed,omitempty"`
	Manufacturer     string `json:"Vendor,omitempty"`
	SN               string `json:"SN,omitempty"`
	PartNumber       string `json:"Part Number,omitempty"`
	RunningSpeed     string `json:"Running Speed,omitempty"`
	Rank             string `json:"Rank,omitempty"`
	Voltage          string `json:"Voltage,omitempty"`
	Locator          string `json:"Locator,omitempty"`
	MemoryTechnology string `json:"Memory Medium,omitempty"`
}

type edacInfo struct {
	CE       string `json:"Correctable Errors,omitempty"`
	UE       string `json:"Uncorrectable Errors,omitempty"`
	Dev      string `json:"Device Type,omitempty"`
	Edac     string `json:"EDAC Mode,omitempty"`
	Location string `json:"Memory Location,omitempty"`
	Type     string `json:"Memory Type,omitempty"`
	Soc      string `json:"Socket ID,omitempty"`
	MC       string `json:"Memory Contoller ID,omitempty"`
	Channel  string `json:"Memory Channel ID,omitempty"`
	DIMM     string `json:"DIMM ID,omitempty"`
}

func (m *MEMORY) Result() {

	dmiMem := internal.DMI["17"]
	m.SlotMax = strconv.Itoa(len(dmiMem))
	phyMem(m, dmiMem)
	procMem, err := internal.ReadLines("/proc/meminfo")
	if err == nil {
		meminfo(m, procMem)
	}
	m.EdacInfo = edac()
	memoryDiagnose(m)
}

func (m *MEMORY) BriefFormat() {
	fmt.Println("[MEMORY INFO]")
	selectFields := []string{"PhyMem", "SlotMax", "SlotUsed", "MemTotal", "MemAvailable", "Diagnose", "DiagnoseDetail"}
	internal.StructSelectFieldOutput(*m, selectFields, 1)
}

func (m *MEMORY) Format() {
	fmt.Println("[MEMORY INFO]")
	selectFields := []string{"PhyMem", "SlotMax", "SlotUsed", "MemTotal", "MemAvailable", "Diagnose", "DiagnoseDetail"}
	internal.StructSelectFieldOutput(*m, selectFields, 1)
	phymem := []string{"Locator", "BankLocator", "Manufacturer", "SN", "Size", "MaxSpeed", "RunningSpeed", "TotalWidth", "DataWidth", "Type"}
	for _, mem := range m.MemEntities {
		fmt.Println()
		internal.StructSelectFieldOutput(mem, phymem, 2)
	}
}

func phyMem(ret *MEMORY, memSlice []map[string]interface{}) {
	var (
		totalSize float64
		memUnit   string
	)
	for _, memMap := range memSlice {
		res := dmiMem{}
		for key, value := range memMap {
			if _, ok := memMap["Speed"]; !ok {
				continue
			}
			if memMap["Speed"] == "Unknown" || memMap["Data Width"] == "Unknown" {
				continue
			}
			val := internal.InterfaceToString(value)
			switch key {
			case "Total Width":
				res.TotalWidth = val
			case "Data Width":
				res.DataWidth = val
			case "Size":
				res.Size = val
				fields := internal.SplitAndTrim(res.Size, " ")
				numStr := fields[0]
				unitStr := fields[1]
				sf, err := strconv.ParseFloat(numStr, 64)
				if err != nil {
					fmt.Println("Memory size convert to int failed.")
					break
				}
				cap, unit, err := internal.Unit2Human(sf, unitStr, false)
				if err != nil {
					fmt.Println("memory capacity unit conversion failed.")
				}
				totalSize += cap
				if len(memUnit) == 0 {
					memUnit = unit
				} else if memUnit != unit {
					fmt.Println("memory capacity unit not match.")
				}
			case "Form Factor":
				res.FormFactor = val
			case "Locator":
				res.Locator = val
			case "Bank Locator":
				res.BankLocator = val
			case "Type":
				res.Type = val
			case "Type Detail":
				res.TypeDetail = val
			case "Speed":
				res.MaxSpeed = val
			case "Manufacturer":
				res.Manufacturer = val
			case "Serial Number":
				res.SN = val
			case "Part Number":
				res.PartNumber = val
			case "Rank":
				res.Rank = val
			case "Configured Memory Speed", "Configured Clock Speed":
				res.RunningSpeed = val
			case "Configured Voltage":
				res.Voltage = val
			case "Memory Technology":
				res.MemoryTechnology = val
			}
		}

		if internal.IsEmptyValue(reflect.ValueOf(res)) {
			continue
		}
		ret.MemEntities = append(ret.MemEntities, res)

	}
	ret.PhyMem = fmt.Sprintf("%.2f %s", totalSize, memUnit)
	ret.SlotUsed = strconv.Itoa(len(ret.MemEntities))
}

func meminfo(ret *MEMORY, memSlice []string) {
	for _, mem := range memSlice {
		fields := internal.SplitAndTrim(mem, ":")
		if len(fields) != 2 {
			continue
		}
		key := fields[0]
		value := fields[1]
		value = strings.ReplaceAll(value, " kB", "")
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			println(fmt.Sprintf("%s conversion to float failed.", key))
			continue
		}
		res, unit, err := internal.Unit2Human(v, "KB", false)
		if err != nil {
			println(fmt.Sprintf("%s conversion to unit failed.", key))
			continue
		}
		switch key {
		case "MemTotal":
			ret.MemTotal = fmt.Sprintf("%.2f %s", res, unit)
		case "MemFree":
			ret.MemFree = fmt.Sprintf("%.2f %s", res, unit)
		case "MemAvailable":
			ret.MemAvailable = fmt.Sprintf("%.2f %s", res, unit)
		case "Buffers":
			ret.Buffer = fmt.Sprintf("%.2f %s", res, unit)
		case "Cached":
			ret.Cached = fmt.Sprintf("%.2f %s", res, unit)
		case "SwapCached":
			ret.SwapCached = fmt.Sprintf("%.2f %s", res, unit)
		case "SwapTotal":
			ret.SwapTotal = fmt.Sprintf("%.2f %s", res, unit)
		case "SwapFree":
			ret.SwapFree = fmt.Sprintf("%.2f %s", res, unit)
		case "VmallocTotal":
			ret.VmallocTotal = fmt.Sprintf("%.2f %s", res, unit)
		case "VmallocUsed":
			ret.VmallocUsed = fmt.Sprintf("%.2f %s", res, unit)
		case "VmallocChunk":
			ret.VmallocChunk = fmt.Sprintf("%.2f %s", res, unit)
		case "Hugepagesize":
			ret.Hugepagesize = fmt.Sprintf("%.2f %s", res, unit)
		case "DirectMap4k":
			ret.DirectMap4k = fmt.Sprintf("%.2f %s", res, unit)
		case "DirectMap2M":
			ret.DirectMap2M = fmt.Sprintf("%.2f %s", res, unit)
		case "DirectMap1G":
			ret.DirectMap1G = fmt.Sprintf("%.2f %s", res, unit)
		}
	}
}

func edac() []edacInfo {
	ret := []edacInfo{}
	mcPath := `/sys/devices/system/edac/mc`
	if !internal.PathExists(mcPath) {
		println("mc not found in /sys/devices/system/edac")
		return ret
	}
	dimmDir, err := filepath.Glob(fmt.Sprintf("%s/mc*/dimm*", mcPath))
	if err != nil {
		println("dimm not found in mc.")
		return ret
	}
	for _, mc := range dimmDir {
		files, err := filepath.Glob(fmt.Sprintf("%s/dimm_*", mc))
		if err != nil {
			println(fmt.Sprintf("no dimm file found in %s", mc))
			continue
		}
		res := edacInfo{}
		for _, file := range files {
			fields := strings.Split(file, "/")
			fileName := fields[len(fields)-1]
			value, _ := internal.ReadFile(file)
			value = strings.TrimSpace(value)
			switch fileName {
			case "dimm_ce_count":
				res.CE = value
			case "dimm_dev_type":
				res.Dev = value
			case "dimm_edac_mode":
				res.Edac = value
			case "dimm_label":
				labels := strings.Split(value, "_")
				res.Soc = strings.Split(labels[1], "#")[1]
				res.MC = strings.Split(labels[2], "#")[1]
				res.Channel = strings.Split(labels[3], "#")[1]
				res.DIMM = strings.Split(labels[4], "#")[1]
			case "dimm_location":
				res.Location = value
			case "dimm_mem_type":
				res.Type = value
			case "dimm_ue_count":
				res.UE = value
			}
		}
		ret = append(ret, res)
	}
	return ret
}
