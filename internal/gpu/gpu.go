package gpu

import (
	"baize/internal/utils"
	"log"
	"strings"
)

var run utils.RunSheller = &utils.RunShell{}

func GetGPU() ([]map[string]interface{}, error) {
	var gpus []map[string]interface{}
	byteGPU, err := run.Command("sh", "-c", `lspci -Dnn | egrep '\[030[0-2]\]'`)
	if err != nil {
		log.Printf("failed to search GPU information through lspci: %v", err)
		return gpus, err
	}
	lines := strings.Split(string(byteGPU), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pciInfo := utils.GetPCIe(strings.TrimSpace(fields[0]))
		pciMap, err := utils.StructToMap(pciInfo)
		if err != nil {
			log.Printf("failed to convert PCI info to map: %v", err)
			continue
		}
		gpus = append(gpus, pciMap)

	}
	return gpus, nil
}
