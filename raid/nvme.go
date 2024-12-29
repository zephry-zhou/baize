package raid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/zephry-zhou/baize/internal"
)

type nvme struct {
	internal.PCI
	NameSpace []string `json:"Namespace,omitempty"`
	physicalDrive
}

func GetNVMeDevices() []nvme {
	byteNvmes, err := internal.Run.Command("sh", "-c", `lspci -Dnn | egrep '\[0108\]' | awk '{print $1}'`)
	if err != nil {
		log.Printf("failed to search NVMe information through lspci: %v", err)
		return []nvme{}
	}
	if len(byteNvmes) == 0 {
		log.Printf("no NVMe device found")
		return []nvme{}
	}
	lines := internal.SplitAndTrim(string(byteNvmes), "\n")
	ret := []nvme{}
	for _, line := range lines {
		res := nvme{}
		res.PCI = *internal.GetPCIe(line)
		parseNVMePCIDirectory(line, &res)
		ret = append(ret, res)
	}
	return ret
}

func parseNVMePCIDirectory(pciAddr string, ret *nvme) {
	path := fmt.Sprintf("/sys/bus/pci/devices/%s/nvme", pciAddr)
	nameSlice, err := os.ReadDir(path)
	if err != nil {
		log.Printf("failed to read nvme directory %s: %v", path, err)
		return
	}
	for _, name := range nameSlice {
		ret.MappingFile = fmt.Sprintf("/dev/%s", name.Name())
		byteSmart, err := internal.Run.Command("sh", "-c", fmt.Sprintf("/usr/sbin/smartctl %s -d nvme -a -j | grep -v ^$", ret.MappingFile))
		if err != nil {
			log.Printf("failed to get %s smart info: %v", ret.MappingFile, err)
			continue
		}
		parseSMART(&ret.physicalDrive, byteSmart)
		nameSpace, err := filepath.Glob(fmt.Sprintf("%s/%s/%sn*", path, name.Name(), name.Name()))
		if err != nil {
			log.Printf("failed to get %s namespace: %v", ret.MappingFile, err)
			continue
		}
		ret.NameSpace = append(ret.NameSpace, nameSpace...)
	}
}
