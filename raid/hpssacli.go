package raid

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

func Hpssacli(pci string, ctrNum int) RHController {
	hpasscli := `/usr/local/baize/tool/hpssacli`
	if !internal.PathExistsWithContent(hpasscli) {
		log.Println("hpssacli not exists")
		return RHController{}
	}
	ret := RHController{}
	for i := 0; i <= ctrNum; i++ {
		pciSlot, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s ctrl slot=%s show  | grep %s", hpasscli, strconv.Itoa(i), pci))
		if err == nil && len(pciSlot) != 0 {
			ret.Cid = strconv.Itoa(i)
			break
		}
	}
	cmdHeader := fmt.Sprintf("%s ctrl slot=%s ", hpasscli, ret.Cid)
	parseCtr(&ret, cmdHeader)
	diskMap := parseDrv(cmdHeader)
	for _, pd := range diskMap {
		ret.PhysicalDriveList = append(ret.PhysicalDriveList, pd...)
	}
	ret.LogicalDriveList = parseLD(cmdHeader, ret.Cid, diskMap)
	ret.BackPlanes = parseEnclosure(cmdHeader)
	return ret
}

func parseCtr(ret *RHController, cmdHeader string) {
	ctrByte, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s show | grep -v ^$", cmdHeader))
	if err != nil {
		log.Println("failed to get RAID/HBA info in:", cmdHeader)
		return
	}
	lines := internal.SplitAndTrim(string(ctrByte), "\n")
	ret.ProductName = lines[0]
	for _, line := range lines[1:] {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		switch key {
		case "Serial Number":
			ret.SerialNumber = value
		case "Controller Status":
			ret.ControllerStatus = value
		case "Firmware Version":
			ret.Firmware = value
		case "Total Cache Size":
			ret.CacheSize = value + " GB"
		case "Battery/Capacitor Status":
			ret.Battery.State = value
		case "Controller Mode":
			ret.ControllerStatus = value
		}
	}
}

func parseEnclosure(cmdHeader string) []backplate {
	byteEnc, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s enclosure all show | grep 'at Port'", cmdHeader))
	if err != nil {
		log.Printf("failed to get enclosure info: %v \n", err)
		return []backplate{}
	}
	ret := []backplate{}
	re := regexp.MustCompile(`Internal Drive Cage at Port (\d+I), Box (\d+), ([A-Za-z]+)`)
	lines := internal.SplitAndTrim(string(byteEnc), "\n")
	for _, line := range lines {
		res := backplate{}
		matches := re.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		res.Eid = fmt.Sprintf("%s:%s", matches[1], matches[2])
		res.State = matches[3]
		byteEid, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s enclosure %s show | awk -F: '/Drive Bays/{print $2}'", cmdHeader, res.Eid))
		if err != nil {
			log.Printf("failed to get enclosure %s detail info: %v \n", res.Eid, err)
			continue
		}
		if len(byteEid) != 0 {
			res.PhysicalDriveCount = strings.TrimSpace(string(byteEid))
		}
		ret = append(ret, res)
	}
	return ret
}

func parseDrv(cmdHeader string) map[string][]physicalDrive {
	byteDrv, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s pd all show | awk '/physicaldrive/{print $2}'", cmdHeader))
	if err != nil {
		log.Println("failed to get RAID/HBA disk info:", cmdHeader)
		return map[string][]physicalDrive{}
	}
	drvList := internal.SplitAndTrim(string(byteDrv), "\n")
	ret := map[string][]physicalDrive{}
	for _, drv := range drvList {
		if len(drv) == 0 {
			continue
		}
		pd := physicalDrive{}
		pd.Location = drv
		byteDrvInfo, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s pd %s show | grep -v ^$", cmdHeader, drv))
		if err != nil {
			log.Println("failed to get RAID/HBA disk detail info:", cmdHeader)
			continue
		}
		lines := internal.SplitAndTrim(string(byteDrvInfo), "\n")
		key := lines[1]
		parseDrvInfo(&pd, lines[2:])
		ret[key] = append(ret[key], pd)
	}
	return ret
}

func parseDrvInfo(pd *physicalDrive, lines []string) {
	enclosureId := make([]string, 2)
	dev := "sda"
	byteDev, err := internal.Run.Command("sh", "-c", `lsblk -d -o NAME | grep -v NAME`)
	if err != nil {
		log.Printf("failed to get system device name:%v \n", err)
	}
	sliceDev := strings.Split(string(byteDev), "\n")
	if len(sliceDev) > 0 {
		dev = sliceDev[0]
	}
	for _, line := range lines {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		switch key {
		case "Port":
			enclosureId = append(enclosureId, value)
		case "Box":
			enclosureId = append(enclosureId, value)
		case "Bay":
			pd.SlotId = value
		case "Status":
			pd.State = value
		case "Interface Type":
			pd.Type = value
		case "Size":
			pd.Capacity = value
		case "Logical/Physical Block Size":
			block := internal.SplitAndTrim(value, "/")
			pd.PhysicalSectorSize = block[0]
			pd.LogicalSectorSize = block[1]
		case "Rotational Speed":
			pd.RotationRate = value
		case "Firmware Revision":
			pd.Firmware = value
		case "Serial Number":
			pd.SN = value
		case "WWID":
			pd.WWN = value
		case "Model":
			pd.Model = value
		case "Current Temperature (C)":
			pd.Temperature = value
		case "PHY Transfer Rate":
			pd.DeviceSpeed = value
		}
	}
	pd.DG = strings.Join(enclosureId, ":")
	if strings.ToUpper(pd.State) == "OK" {
		slot, err := strconv.Atoi(pd.SlotId)
		if err != nil {
			log.Println("conversion bay to int failed.")
			return
		}
		cmd := fmt.Sprintf("/usr/sbin/smartctl -d cciss,%v /dev/%s -a -j | grep -v ^$", slot-1, dev)
		byteSMART, err := internal.Run.Command("sh", "-c", cmd)
		if err != nil {
			log.Println("failed to get disk smart info:", cmd)
			return
		}
		parseSMART(pd, byteSMART)
	}
}

func parseLD(cmdHeader string, cid string, diskMap map[string][]physicalDrive) []logicalDrive {
	byteLD, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s ld all show | awk '/logicaldrive/{print $2}'", cmdHeader))
	if err != nil {
		log.Printf("failed to get logical drive info: %v \n", err)
		return []logicalDrive{}
	}
	sliceLD := internal.SplitAndTrim(string(byteLD), "\n")
	ret := []logicalDrive{}
	for _, ld := range sliceLD {
		if len(ld) == 0 {
			continue
		}
		res := logicalDrive{}
		res.Location = fmt.Sprintf("/c%s/v%s", cid, ld)
		byteLDInfo, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s ld %s show detail| grep -v ^$", cmdHeader, ld))
		if err != nil {
			log.Printf("failed to get logical drive detail info: %v \n", err)
			continue
		}
		lines := internal.SplitAndTrim(string(byteLDInfo), "\n")
		res.DG = lines[1]
		res.PhysicalDrives = append(res.PhysicalDrives, diskMap[lines[1]]...)
		parseLDInfo(&res, lines[2:])
		ret = append(ret, res)
	}
	return ret
}

func parseLDInfo(ld *logicalDrive, lines []string) {
	for _, line := range lines {
		fileds := internal.SplitAndTrim(line, ":")
		if len(fileds) != 2 {
			continue
		}
		key := fileds[0]
		val := fileds[1]
		switch key {
		case "Logical Drive":
			ld.VD = val
		case "Size":
			ld.Capacity = val
		case "Fault Tolerance":
			ld.Type = fmt.Sprintf("RAID%s", val)
		case "Strip Size":
			ld.StripSize = val
		case "Status":
			ld.State = val
		case "Caching":
			ld.Cache = val
		case "Unique Identifier":
			ld.ScsiNaaId = val
		case "Disk Name":
			ld.MappingFile = val
		}
	}
}
