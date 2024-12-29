package gpu

import (
	"log"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

type GPU struct {
	Vendor      string `json:"vendor"`
	Device      string `json:"device"`
	DeviceClass string `json:"device_class"`
	PCI         struct {
		Bus      string `json:"bus"`
		Device   string `json:"device"`
		Function string `json:"function"`
		Slot     string `json:"slot"`
		Vendor   string `json:"vendor"`
	} `json:"pci"`
}

func GetGPU() ([]map[string]interface{}, error) {
	var gpus []map[string]interface{}
	byteGPU, err := internal.Run.Command("sh", "-c", `lspci -Dnn | egrep '\[030[0-2]\]'`)
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
		pciInfo := internal.GetPCIe(strings.TrimSpace(fields[0]))
		pciMap, err := internal.StructToMap(pciInfo)
		if err != nil {
			log.Printf("failed to convert PCI info to map: %v", err)
			continue
		}
		gpus = append(gpus, pciMap)

	}
	return gpus, nil
}
