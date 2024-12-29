package memory

import (
	"log"
	"strconv"
	"strings"
)

func memoryDiagnose(ret *MEMORY) {

	detail := make([]string, 0)

	// 物理内存总容量 与 系统内存总容量的差值大于8GB，则认为存在故障内存
	virMemorySize, err1 := strconv.ParseFloat(strings.Fields(ret.MemTotal)[0], 64)
	phyMemorySize, err2 := strconv.ParseFloat(strings.Fields(ret.PhyMem)[0], 64)
	if err1 == nil && err2 == nil {
		if phyMemorySize-virMemorySize > 8 {
			detail = append(detail, "Memory Overcommitment")
		}
	} else {
		log.Println("memory size convert to float failed.")
	}

	// 物理内存条容量需要保持一致
	// 每路CPU上内存数量需一样
	sizeMap := make(map[string]int)
	cpuMap := make(map[string]int)
	for _, mem := range ret.MemEntities {
		sizeMap[mem.Size] += 1
		cpuMap[mem.Locator] += 1
	}
	if len(sizeMap) != 1 {
		detail = append(detail, "Memory Size Mismatch")
	}
	var temp int
	for _, v := range cpuMap {
		if temp == 0 {
			temp = v
		} else if v != temp {
			detail = append(detail, "Memory Channel Mismatch")
		}
	}

	// 内存数量需为偶数
	if len(ret.MemEntities)%2 != 0 {
		detail = append(detail, "Memory Count Mismatch")
	}

	if len(detail) != 0 {
		ret.Diagnose = "Unhealthy"
		ret.DiagnoseDetail = strings.Join(detail, "; ")
	} else {
		ret.Diagnose = "Healthy"
	}

}
