package raid

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

var arc string = `/usr/local/hwtool/tool/arcconf`

func arcconf(pci string, ctrNum int) RHController {
	snFile := fmt.Sprintf("/sys/bus/pci/devices/%s/host0/scsi_host/host0/serial_number", pci)
	ret := RHController{}
	if !internal.PathExists(snFile) {
		internal.Log.Error(pci, " sn file not exists")
		return ret
	}
	sn, err := internal.ReadFile(snFile)
	if err != nil {
		internal.Log.Error("failed to read sn file: ", err)
		return ret
	}
	ret.SerialNumber = strings.TrimSpace(sn)
	for i := 0; i <= ctrNum; i++ {
		out, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s GETCONFIG %d AD | grep %s", arc, i, ret.SerialNumber))
		if err != nil {
			continue
		}
		if len(out) != 0 {
			ret.Cid = strconv.Itoa(i)
		}
	}
	if ret.Cid == "" {
		internal.Log.Error("failed to get controller id")
		return ret
	}
	arcCtr(&ret, ret.Cid)
	diskMap := arcPD(ret.Cid)
	for _, pd := range diskMap {
		ret.PhysicalDriveList = append(ret.PhysicalDriveList, pd)
	}
	ret.LogicalDriveList = arcLD(ret.Cid, diskMap)
	return ret
}

func arcCtr(ret *RHController, ctrNum string) {
	byteCtr, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s GETCONFIG %s AD", arc, ctrNum))
	if err != nil {
		internal.Log.Error("failed to get controller info: ", err)
		return
	}
	lines := strings.Split(string(byteCtr), "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		fields := internal.SplitAndTrim(line, ":")
		switch fields[0] {
		case "Controller Status":
			ret.ControllerStatus = fields[1]
		case "Controller Mode":
			ret.CurrentPersonality = fields[1]
		case "Controller Model":
			ret.ProductName = fields[1]
		case "Controller Serial Number":
			ret.SerialNumber = fields[1]
		case "Installed memory":
			ret.CacheSize = fields[1]
		case "Logical devices/Failed/Degraded":
			val := internal.SplitAndTrim(fields[1], "/")
			ret.NumberOfRaid = val[0]
			ret.FailedRaid = val[1]
			ret.DegradedRaid = val[2]
		case "BIOS":
			ret.BiosVersion = fields[1]
		case "Firmware":
			ret.FwVersion = fields[1]
		}
	}
}

func arcPD(ctrNum string) map[string]physicalDrive {
	ret := map[string]physicalDrive{}
	bytePD, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s GETCONFIG %s PD", arc, ctrNum))
	if err != nil {
		internal.Log.Error("failed to get physical drive info: ", err)
		return ret
	}
	pds := strings.Split(string(bytePD), "\n\n")
	smartHead := "/usr/sbin/smartctl -d aacraid,"
	byteDev, err := internal.Run.Command("sh", "-c", `lsscsi -b | awk '/\/dev/ {print $2; exit}'`)
	if err != nil {
		internal.Log.Error("failed to get device info: ", err)
		return ret
	}
	dev := strings.TrimSpace(string(byteDev))
	h, err := strconv.Atoi(ctrNum)
	if err != nil {
		internal.Log.Error("failed to convert ctrNum to int: ", err)
		return ret
	}
	for _, pd := range pds {
		res := physicalDrive{}
		if !strings.Contains(pd, "Device is a Hard drive") {
			continue
		}
		lines := strings.Split(pd, "\n")
		for _, line := range lines {
			if !strings.Contains(line, ":") {
				continue
			}
			fields := internal.SplitAndTrim(line, ":")
			switch fields[0] {
			case "State":
				res.State = fields[1]
			case "Block Size":
				res.PhysicalSectorSize = fields[1]
			case "Transfer Speed":
				res.LinkSpeed = fields[1]
			case "Reported Location":
				val := strings.Split(fields[1], ",")
				res.EnclosureId = strings.Fields(val[0])[1]
				res.SlotId = strings.Fields(val[1])[1]
				res.Location = fmt.Sprintf("/c%s/e%s/s%s", ctrNum, res.EnclosureId, res.SlotId)
				hlid := fmt.Sprintf("%d,%s,%s", h-1, res.EnclosureId, res.SlotId)
				byteSMART, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s%s %s -a -j | grep -v ^$", smartHead, hlid, dev))
				if err != nil {
					internal.Log.Error("failed to get disk smart info: ", err)
					continue
				}
				parseSMART(&res, byteSMART)
			case "Verdor":
				res.Vendor = fields[1]
			case "Model":
				res.Model = fields[1]
			case "Firmware":
				res.Firmware = fields[1]
			case "Serial Number":
				res.SN = fields[1]
			case "World-wide name":
				res.WWN = fields[1]
			case "Write cache":
				res.WriteCache = fields[1]
			case "S.M.A.R.T.":
				res.SmartAlert = fields[1]
			}
		}
		if _, ok := ret[res.SN]; !ok {
			ret[res.SN] = res
		}
	}
	return ret
}

func arcLD(ctrNum string, diskMap map[string]physicalDrive) []logicalDrive {
	ret := []logicalDrive{}
	byteLD, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s GETCONFIG %s LD", arc, ctrNum))
	if err != nil {
		internal.Log.Error("failed to get logical drive info: ", err)
		return ret
	}
	lds := strings.Split(string(byteLD), "\n\n")
	for _, ld := range lds {
		if !strings.Contains(ld, "Logical Device number") {
			continue
		}
		res := logicalDrive{}
		lines := strings.Split(ld, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Logical Device number") {
				ldn := strings.Fields(line)
				res.VD = ldn[len(ldn)-1]
				continue
			}
			if !strings.Contains(line, ":") {
				continue
			}
			fields := strings.SplitAfterN(line, ":", 2)
			key := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])
			switch key {
			case "Logical Device name":
				res.Location = value
			case "RAID Level":
				res.Type = value
			case "Unique Identifier":
				res.ScsiNaaId = value
			case "State of Logical Drive":
				res.State = value
			case "Size":
				val := strings.Fields(value)
				if len(val) != 2 {
					res.Capacity = value
				}
				s, err := strconv.ParseFloat(val[0], 64)
				if err != nil {
					internal.Log.Error("failed to convert size to float64: ", err)
					continue
				}
				size, unit, err := internal.Unit2Human(s, val[1], true)
				if err != nil {
					internal.Log.Error("failed to convert size to human: ", err)
					continue
				}
				res.Capacity = fmt.Sprintf("%f %s", size, unit)
			}
			if strings.HasPrefix(key, "Segment ") {
				val := strings.Fields(value)
				println(val[len(val)-1])
				res.PhysicalDrives = append(res.PhysicalDrives, diskMap[val[len(val)-1]])
			}
		}
		ret = append(ret, res)
	}
	return ret
}
