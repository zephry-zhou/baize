package raid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

func Mdadm(pciAddr string, num int) RHController {
	_, err := internal.Run.Command("sh", "-c", fmt.Sprintf("/usr/sbin/mdadm --detail-platform | grep %s", pciAddr))
	if err != nil {
		log.Printf("failed to get mdadm  %s info: %v \n", pciAddr, err)
		return RHController{}
	}
	byteCTRL, err := internal.Run.Command("sh", "-c", "/usr/sbin/mdadm --detail-platform | grep -v ^$")
	if err != nil || len(byteCTRL) == 0 {
		log.Printf("failed to get raid card %s info (mdadm): %v \n", pciAddr, err)
		return RHController{}
	}

	ret := RHController{}
	diskmap := make(map[string]physicalDrive)
	ret.Cid = fmt.Sprintf("c%v", num)
	parseVROC(byteCTRL, &ret, &diskmap)
	ret.LogicalDriveList = vrocLogicalDrive(strconv.Itoa(num), diskmap)
	return ret
}

func parseVROC(data []byte, ret *RHController, diskMap *map[string]physicalDrive) {
	lines := internal.SplitAndTrim(string(data), "\n")
	for _, line := range lines {
		fileds := strings.SplitN(line, ":", 2)
		if len(fileds) != 2 {
			continue
		}
		key := strings.TrimSpace(fileds[0])
		value := strings.TrimSpace(fileds[1])
		switch key {
		case "Platform":
			ret.ProductName = value
		case "Version":
			ret.Firmware = value
		case "RAID Levels":
			ret.RaidLevelSupported = value
		case `Port[0-9]+`:
			val := internal.SplitAndTrim(value, " ")
			if strings.HasPrefix(val[0], "/dev/") {
				pd := parsePhysicalDrive(val[0])
				pd.Location = key
				pd.SlotId = key
				ret.PhysicalDriveList = append(ret.PhysicalDriveList, pd)
				(*diskMap)[val[0]] = pd
			}
		case "NVMe under VMD":
			pdMap := parseVROCNVMe(value)
			for key, pd := range pdMap {
				(*diskMap)[key] = pd
				ret.PhysicalDriveList = append(ret.PhysicalDriveList, pd)
			}
		}
	}
	ret.CacheSize = "0 MB"
	ret.CurrentPersonality = "RAID Mode"
}

func vrocLogicalDrive(cid string, diskMap map[string]physicalDrive) []logicalDrive {
	file := `/proc/mdstat`
	byteLD, err := internal.Run.Command("sh", "-c", fmt.Sprintf("awk '/: active/{print $1}' %s", file))
	if err != nil {
		log.Printf("failed to get logical drive info: %v \n", err)
		return []logicalDrive{}
	}
	if len(byteLD) == 0 {
		return []logicalDrive{}
	}
	sliceLD := internal.SplitAndTrim(string(byteLD), "\n")
	ret := []logicalDrive{}
	for i, ld := range sliceLD {
		res := logicalDrive{}
		res.Location = fmt.Sprintf("/c%s/v%v", cid, i)
		res.MappingFile = fmt.Sprintf("/dev/%s", ld)
		byteLDInfo, err := internal.Run.Command("sh", "-c", fmt.Sprintf("/usr/sbin/mdadm --detail %s | grep -v ^$", res.MappingFile))
		if err != nil {
			log.Printf("failed to get %s info: %v", res.MappingFile, err)
			continue
		}
		parseLogicalDrive(&res, byteLDInfo, diskMap)
	}
	return ret
}

func parseLogicalDrive(ld *logicalDrive, byteLDInfo []byte, diskMap map[string]physicalDrive) {
	lines := internal.SplitAndTrim(string(byteLDInfo), "\n")
	re := regexp.MustCompile(`(\d+\.\d+ GB)`)
	for _, line := range lines {
		if strings.Contains(line, " : ") {
			fields := strings.SplitN(line, " : ", 2)
			key := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])
			switch key {
			case "Container":
			case "RAID Level":
				ld.Type = value
			case "Array Size":
				match := re.FindStringSubmatch(value)
				if len(match) > 0 {
					ld.Capacity = match[1]
				}
			case "Raid Devices":
				ld.NumberOfDrivesPerSpan = value
			case "State":
				ld.State = value
			case "Consistency Policy":
				ld.Consist = value
			}
		} else if strings.Contains(line, "/dev/") {
			fields := internal.SplitAndTrim(line, " ")
			ld.PhysicalDrives = append(ld.PhysicalDrives, diskMap[fields[len(fields)-1]])
		}
	}
}

func parsePhysicalDrive(device string) physicalDrive {
	ret := physicalDrive{}
	byteSMART, err := internal.Run.Command("sh", "-c", fmt.Sprintf("/usr/sbin/smartctl -d %s sat -a -j | grep -v '^$|#'", device))
	if err != nil {
		log.Printf("failed to get %s smart info: %v \n", device, err)
		return physicalDrive{}
	}
	parseSMART(&ret, byteSMART)
	return ret
}

func parseVROCNVMe(nvmeDir string) map[string]physicalDrive {
	nameSlice, err := os.ReadDir(filepath.Join(nvmeDir, "nvme"))
	if err != nil {
		log.Printf("failed to read nvme directory %s: %v", nvmeDir, err)
		return map[string]physicalDrive{}
	}
	ret := physicalDrive{}
	for _, name := range nameSlice {
		deviceName := fmt.Sprintf("/dev/%s", name.Name())
		byteSmart, err := internal.Run.Command("sh", "-c", fmt.Sprintf("/usr/sbin/smartctl %s -d nvme -a -j | grep -v ^$", deviceName))
		if err != nil {
			log.Printf("failed to get %s smart info: %v", deviceName, err)
			continue
		}
		parseSMART(&ret, byteSmart)

		nameSpace, err := filepath.Glob(fmt.Sprintf("%s/%s/%sn*", filepath.Join(nvmeDir, "nvme"), name.Name(), name.Name()))
		if err != nil {
			log.Printf("failed to get %s namespace: %v", ret.MappingFile, err)
			continue
		}
		if len(nameSpace) == 1 {
			ret.MappingFile = nameSpace[0]
		} else if len(nameSpace) > 1 {
			log.Printf("%s namespace more than one,not surpport.", deviceName)
			continue
		}
	}
	return map[string]physicalDrive{ret.MappingFile: ret}
}
