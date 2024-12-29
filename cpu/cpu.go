package cpu

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

func (c *CPU) Result() {
	//	ret := new(CPU)
	lscpuByte, err := internal.Run.Command("lscpu")
	if err != nil {
		log.Printf("lscpu running failed. %v", err)
	}
	lines := strings.Split(string(lscpuByte), "\n")
	lscpu(c, lines)
	dmiCPU(c)
}

func lscpu(ret *CPU, lines []string) {
	for _, line := range lines {
		fields := internal.SplitAndTrim(line, ":")
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
	cpuMap := internal.DMI["4"]
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
				val := internal.InterfaceToString(value)
				if socketRegex0.MatchString(val) {
					res.SocketID = "0"
				} else if socketRegex1.MatchString(val) {
					res.SocketID = "1"
				}
			case "Manufacturer":
				res.Manufacturer = internal.InterfaceToString(value)
			case "Signature":
				res.Signature = internal.InterfaceToString(value)
			case "Version":
				res.Version = internal.InterfaceToString(value)
			case "Voltage":
				res.Voltage = internal.InterfaceToString(value)
			case "External Clock":
				res.ExClock = internal.InterfaceToString(value)
			case "Max Speed":
				res.MaxSpeed = internal.InterfaceToString(value)
			case "Current Speed":
				res.CurSpeed = internal.InterfaceToString(value)
				baseFreq, _ = strconv.ParseFloat(strings.Fields(res.CurSpeed)[0], 64)
			case "Core Count":
				res.Cores = internal.InterfaceToString(value)
			case "Core Enabled":
				res.CoreEnable = internal.InterfaceToString(value)
			case "Thread Count":
				res.Threads = internal.InterfaceToString(value)
				threads, _ = strconv.ParseFloat(res.Threads, 64)
			}
		}
		res.ThreadList = procMap[res.SocketID]
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

func (c *CPU) BriefFormat() {
	fmt.Println("[CPU INFO]")
	selectFields := []string{"ModelName", "Vendor", "Power", "Hyper", "MinFreq", "MaxFreq"}
	internal.StructSelectFieldOutput(*c, selectFields, 1)
}

func (c *CPU) Format() {
	fmt.Println("[CPU INFO]")
	selectFields := []string{"ModelName", "Architecture", "Vendor", "Socket", "CorePerSocket", "ThrPerCore", "Threads", "MaxFreq", "MinFreq", "Power", "Hyper"}
	phy := []string{"Version", "Manufacturer", "SocketID", "MaxSpeed", "CurSpeed", "Cores", "CoreEnable", "Threads"}
	sliProc := []string{"Processor", "CoreID", "Temp", "Freq"}
	internal.StructSelectFieldOutput(*c, selectFields, 1)
	for _, phycpu := range c.PhyCPU {
		fmt.Println()
		internal.StructSelectFieldOutput(phycpu, phy, 2)
		for _, proc := range phycpu.ThreadList {
			fmt.Println()
			internal.StructSelectFieldOutput(proc, sliProc, 3)
		}
	}
}
