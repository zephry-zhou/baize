package cpu

import (
	"baize/internal/utils"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type CPU struct {
	ModelName      string   `json:"Product Name,omitempty"`
	Vendor         string   `json:"Vendor,omitempty"`
	Architecture   string   `json:"Artchitrcture,omitempty"`
	Hyper          string   `json:"Hyper Threading,omitempty"`
	Power          string   `json:"Power State,omitempty"`
	OpMode         string   `json:"CPU op-mode,omitempty"`
	ByteOrder      string   `json:"Byte Order,omitempty"`
	AdrSize        string   `json:"Address Size,omitempty"`
	Threads        string   `json:"Number Of Threads,omitempty"`
	OnLineThreads  string   `json:"Online CPU,omitempty"`
	ThrPerCore     string   `json:"Threads per Core,omitempty"`
	CorePerSocket  string   `json:"Cores per Socket,omitempty"`
	Socket         string   `json:"Sockets,omitempty"`
	NUMANode       string   `json:"NUMA Node,omitempty"`
	Family         string   `json:"Family,omitempty"`
	Model          string   `json:"Model,omitempty"`
	Step           string   `json:"Stepping,omitempty"`
	BogoMIPS       string   `json:"BogoMIPS,omitempty"`
	Virtualization string   `json:"Virtualization,omitempty"`
	MinFreq        string   `json:"Minimum Frequency,omitempty"`
	MaxFreq        string   `json:"Maximum Frequency,omitempty"`
	L1d            string   `json:"L1d Cache,omitempty"`
	L1i            string   `json:"L1i Cache,omitempty"`
	L2             string   `json:"L2 Cache,omitempty"`
	L3             string   `json:"L3 Cache,omitempty"`
	Flags          []string `json:"Flags,omitempty"`
	PhyCPU         []phyCPU `json:"Physical CPU Entities,omitempty"`
}

type phyCPU struct {
	SocketID     string      `json:"Socket ID,omitempty"`
	Family       string      `json:"Family,omitempty"`
	Manufacturer string      `json:"Vendor,omitempty"`
	Signature    string      `json:"Signature,omitempty"`
	Version      string      `json:"Prodcut Name,omitempty"`
	Voltage      string      `json:"Voltage,omitempty"`
	ExClock      string      `json:"External Speed,omitempty"`
	MaxSpeed     string      `json:"Max Speed,omitempty"`
	CurSpeed     string      `json:"Based Speed,omitempty"`
	Status       string      `json:"State,omitempty"`
	Cores        string      `json:"Cores,omitempty"`
	CoreEnable   string      `json:"Core Enabled,omitempty"`
	Threads      string      `json:"Threads,omitempty"`
	ProcEntities []processor `json:"Processor Entities"`
}

type processor struct {
	Processor string `json:"Processor,omitempty"`
	Freq      string `json:"Core Frequency,omitempty"`
	PhyID     string `json:"Physical ID,omitempty"`
	CoreID    string `json:"Core ID,omitempty"`
	Temp      string `json:"Temperature,omitempty"`
}

var run utils.RunSheller = &utils.RunShell{}

func GetCPU() *CPU {

	ret := new(CPU)
	lscpuByte, err := run.Command("lscpu")
	if err != nil {
		log.Printf("lscpu running failed. %v", err)
	}
	lines := strings.Split(string(lscpuByte), "\n")
	lscpu(ret, lines)
	dmiCPU(ret)
	return ret
}

func lscpu(ret *CPU, lines []string) {
	for _, line := range lines {
		fields := utils.SplitAndTrim(line, ":")
		if len(fields) != 2 {
			continue
		}
		key, value := fields[0], fields[1]
		switch key {
		case "Architecture":
			ret.Architecture = value
		case "Byte Order":
			ret.ByteOrder = value
		case "CPU(s)", "CPU[s]":
			ret.Threads = value
		case "Thread(s) per core", "Thread[s] per core":
			ret.ThrPerCore = value
		case "Core(s) per socket", "Core[s] per socket":
			ret.CorePerSocket = value
		case "Socket(s)", "Socket[s]":
			ret.Socket = value
		case "NUMA node(s)", "NUMA node[s]":
			ret.NUMANode = value
		case "Vendor ID":
			ret.Vendor = value
		case "CPU family":
			ret.Family = value
		case "Model":
			ret.Model = value
		case "Model name":
			ret.ModelName = value
		case "Stepping":
			ret.Step = value
		case "Virtualization":
			ret.Virtualization = value
		case "L1d cache":
			ret.L1d = value
		case "L1i cache":
			ret.L1i = value
		case "L2 cache":
			ret.L2 = value
		case "L3 cache":
			ret.L3 = value
		case "Flags":
			ret.Flags = strings.Split(value, " ")
		}
	}
}

func dmiCPU(ret *CPU) {
	cpuMap := utils.Processor.Dmidecode()
	socketRegex0 := regexp.MustCompile(`^(P0|Proc 1|CPU 1|CPU01|CPU1)$`)
	socketRegex1 := regexp.MustCompile(`^(P1|Proc 2|CPU 2|CPU02|CPU2)$`)

	procMap, freqMap := cpupower()
	if len(procMap) == 0 {
		procMap, freqMap = cpuinfo()
	}

	var threads, baseFreq float64

	for _, cpu := range cpuMap {
		res := phyCPU{}
		for key, value := range cpu {
			switch key {
			case "Socket Designation":
				val := utils.InterfaceToString(value)
				if socketRegex0.MatchString(val) {
					res.SocketID = "0"
				} else if socketRegex1.MatchString(val) {
					res.SocketID = "1"
				}
			case "Manufacturer":
				res.Manufacturer = utils.InterfaceToString(value)
			case "Signature":
				res.Signature = utils.InterfaceToString(value)
			case "Version":
				res.Version = utils.InterfaceToString(value)
			case "Voltage":
				res.Voltage = utils.InterfaceToString(value)
			case "External Clock":
				res.ExClock = utils.InterfaceToString(value)
			case "Max Speed":
				res.MaxSpeed = utils.InterfaceToString(value)
			case "Current Speed":
				res.CurSpeed = utils.InterfaceToString(value)
				baseFreq, _ = strconv.ParseFloat(strings.Fields(res.CurSpeed)[0], 64)
			case "Core Count":
				res.Cores = utils.InterfaceToString(value)
			case "Core Enabled":
				res.CoreEnable = utils.InterfaceToString(value)
			case "Thread Count":
				res.Threads = utils.InterfaceToString(value)
				threads, _ = strconv.ParseFloat(res.Threads, 64)
			}
		}
		res.ProcEntities = procMap[res.SocketID]
		ret.PhyCPU = append(ret.PhyCPU, res)
	}

	if freqMap["FreqNums"] < threads && ret.Vendor == "GenuineIntel" {
		freqMap["MinFreq"], freqMap["MaxFreq"] = i7zFreq()
	}

	if strings.HasPrefix(ret.Architecture, "x86") {
		if freqMap["MinFreq"] >= baseFreq {
			ret.Power = "Performance"
		} else {
			ret.Power = "PowerSave"
		}
		if ret.ThrPerCore == "1" {
			ret.Hyper = "Support Disabled"
		} else {
			ret.Hyper = "Support Enabled"
		}

	} else if strings.HasPrefix(ret.Architecture, "aarch") {
		ret.Power = "Performance"
		ret.Hyper = "Not Support"
	}

	ret.MinFreq = fmt.Sprintf("%.2f", freqMap["MinFreq"])
	ret.MaxFreq = fmt.Sprintf("%.2f", freqMap["MaxFreq"])
}
