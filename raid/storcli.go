package raid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

type storcliRes struct {
	Controllers []controller
}

type controller struct {
	CommandStatus CommandStatus `json:"Command Status"`
	ResponseData  ResponseData  `json:"Response Data"`
}

type ResponseData map[string]interface{}

type CommandStatus struct {
	CLIVersion      string `json:"CLI Version"`
	OperatingSystem string `json:"Operating System"`
	Controller      int    `json:"Controller"`
	Status          string `json:"Status"`
	Description     string `json:"Description"`
}

func Storcli(pciAddr string, ctrNum int) RHController {
	beyteStorcli, _ := internal.Run.Command("sh", "-c", "file /usr/local/bin/storcli | awk '{print $5}'")
	storcli := string(bytes.TrimSpace(beyteStorcli))
	ret := RHController{}
	for i := 0; i <= ctrNum; i++ {
		hasPCI, _ := internal.Run.Command("sh", "-c", fmt.Sprintf("%s /c%s show | grep %s", storcli, strconv.Itoa(i), strings.Replace(pciAddr, `0000:`, "", 1)))
		if len(hasPCI) != 0 {
			ret.Cid = strconv.Itoa(i)
			break
		}
	}
	if len(ret.Cid) == 0 {
		log.Printf("RAID/HBA controller: %s not found.", pciAddr)
		return ret
	}
	cmdHeader := fmt.Sprintf("storcli /c%s", ret.Cid)
	ctrByte, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s show all J | grep -v ^$", cmdHeader))
	if err != nil {
		log.Println("failed to get RAID/HBA info in:", pciAddr)
		return ret
	}
	cardStruct := storcliRes{}
	if err := json.Unmarshal(ctrByte, &cardStruct); err != nil {
		log.Printf("failed to unmarshal json: %s", cmdHeader)
		return ret
	}
	diskMap := make(map[string]physicalDrive)
	for _, ctr := range cardStruct.Controllers {
		if ctr.CommandStatus.Status != "Success" {
			log.Println("failed to get RAID/HBA info:", cmdHeader)
			continue
		}
		for key, value := range ctr.ResponseData {
			switch key {
			case "Basics":
				if basicsMap, ok := value.(map[string]interface{}); ok {
					for subKey, subVal := range basicsMap {
						switch subKey {
						case "Model":
							ret.ProductName = internal.InterfaceToString(subVal)
						case "Serial Number":
							ret.SerialNumber = internal.InterfaceToString(subVal)
						case "Current Controller Date/Time":
							ret.ControllerTime = internal.InterfaceToString(subVal)
						case "SAS Address":
							ret.SasAddress = internal.InterfaceToString(subVal)
						}
					}
				}
			case "Version":
				if versionMap, ok := value.(map[string]interface{}); ok {
					for subKey, subVal := range versionMap {
						switch subKey {
						case "Firmware Package Build":
							ret.Firmware = internal.InterfaceToString(subVal)
						case "Firmware Version":
							ret.FwVersion = internal.InterfaceToString(subVal)
						case "Bios Version":
							ret.BiosVersion = internal.InterfaceToString(subVal)
						}
					}
				}
			case "Status":
				if statusMap, ok := value.(map[string]interface{}); ok {
					for subKey, subVal := range statusMap {
						switch subKey {
						case "Controller Status":
							ret.ControllerStatus = internal.InterfaceToString(subVal)
						case "Memory Correctable Errors":
							ret.MemoryCorrectableErrors = internal.InterfaceToString(subVal)
						case "Memory Uncorrectable Errors":
							ret.MemoryUncorrectableErrors = internal.InterfaceToString(subVal)
						}
					}
				}
			case "Supported Adapter Operations":
				if opsMap, ok := value.(map[string]interface{}); ok {
					for subKey, subVal := range opsMap {
						switch subKey {
						case "Rebuild Rate":
							ret.CurrentPersonality = internal.InterfaceToString(subVal)
						case "Foreign Config Import":
							ret.DegradedRaid = internal.InterfaceToString(subVal)
						case "Support JBOD":
						}
					}
				}
			case "HwCfg":
				if hwcfgMap, ok := value.(map[string]interface{}); ok {
					for subKey, subVal := range hwcfgMap {
						switch subKey {
						case "ChipRevision":
							ret.ChipRevision = internal.InterfaceToString(subVal)
						case "Front End Port Count":
							ret.FrontEndPortCount = internal.InterfaceToString(subVal)
						case "Backend Port Count":
							ret.BackendPortCount = internal.InterfaceToString(subVal)
						case "NVRAM Size":
							ret.NVRAMSize = internal.InterfaceToString(subVal)
						case "Flash Size":
							ret.FlashSize = internal.InterfaceToString(subVal)
						case "On Board Memory Size":
							ret.CacheSize = internal.InterfaceToString(subVal)
						case "CacheVault Flash Size":
						}
					}
				}
			case "Capabilities":
				if capMap, ok := value.(map[string]interface{}); ok {
					for subKey, subVal := range capMap {
						switch subKey {
						case "Supported Drives":
							ret.SupportedDrives = internal.InterfaceToString(subVal)
						case "RAID Level Supported":
							ret.RaidLevelSupported = internal.InterfaceToString(subVal)
						case "Enable JBOD":
							ret.EnableJBOD = internal.InterfaceToString(subVal)
						}
					}
				}
			case "Virtual Drives":
				ret.NumberOfRaid = internal.InterfaceToString(value)
			case "Physical Drives":
				ret.NumberOfDisk = internal.InterfaceToString(value)
			case "PD LIST":
				if pdSlice, ok := value.([]interface{}); ok {
					for _, pd := range pdSlice {
						if pdMap, ok := pd.(map[string]interface{}); ok {
							for subKey, subVal := range pdMap {
								if subKey != "EID:Slt" {
									continue
								}
								fields := strings.Split(subVal.(string), ":")
								pid := fmt.Sprintf("/c%s/e%s/s%s", ret.Cid, fields[0], fields[1])
								cmd := fmt.Sprintf("%s %s", storcli, pid)
								diskMap = physicalDrives(cmd, pid)
								for _, phyDrive := range diskMap {
									ret.PhysicalDriveList = append(ret.PhysicalDriveList, phyDrive)
								}
							}
						}
					}
				}
			case "Enclosures":
				ret.NumberOfBackplane = internal.InterfaceToString(value)
			case "Enclosure LIST":
				if enclSlice, ok := value.([]interface{}); ok {
					for _, backplane := range enclSlice {
						for subKey, subVal := range backplane.(map[string]interface{}) {
							if subKey != "EID" {
								continue
							}
							eid := fmt.Sprintf("/c%s/e%s", ret.Cid, strconv.FormatFloat(subVal.(float64), 'f', 0, 64))
							ret.BackPlanes = append(ret.BackPlanes, enclosure(storcli, eid))
						}
					}
				}
			case "Cachevault_Info":
				for _, bbu := range value.([]interface{}) {
					for subKey, subVal := range bbu.(map[string]interface{}) {
						switch subKey {
						case "Model":
							ret.Battery.Model = subVal.(string)
						case "State":
							ret.Battery.State = subVal.(string)
						case "RetentionTime":
							ret.Battery.RetentionTime = subVal.(string)
						case "Temp":
							ret.Battery.Temp = subVal.(string)
						case "Mode":
							ret.Battery.Mode = subVal.(string)
						case "MfgDate":
							ret.Battery.MfgDate = subVal.(string)
						}
					}
				}
			}
		}
	}
	ret.LogicalDriveList = logicalDrives(storcli, ret.Cid, diskMap)
	println("ret")
	return ret
}

func enclosure(cmd string, eid string) backplate {
	bpByte, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s %s show all J | grep -v ^$", cmd, eid))
	ret := backplate{}
	if err != nil {
		log.Printf("failed to get eid info: %v", err)
		return ret
	}
	bpStruct := storcliRes{}
	json.Unmarshal(bpByte, &bpStruct)
	for _, enclo := range bpStruct.Controllers {
		for key, value := range enclo.ResponseData[fmt.Sprintf("Enclosure %s ", eid)].(map[string]interface{}) {
			switch key {
			case "Information":
				for subKey, subVal := range value.(map[string]interface{}) {
					switch subKey {
					case "Device ID":
						ret.Eid = strconv.FormatFloat(subVal.(float64), 'f', 0, 64)
					case "Connector Name":
						ret.ConnectorName = subVal.(string)
					case "Enclosure Type":
						ret.EnclosureType = subVal.(string)
					case "Status":
						ret.State = subVal.(string)
					case "Enclosure Serial Number":
						ret.EnclosureSerialNumber = subVal.(string)
					case "Device Type":
						ret.DeviceType = subVal.(string)
					}
				}
			case "Inquiry Data":
				for subKey, subVal := range value.(map[string]interface{}) {
					switch subKey {
					case "Vendor Identification":
						ret.Vendor = subVal.(string)
					case "Product Identification":
						ret.ProductIdentification = subVal.(string)
					case "Product Revision Level":
						ret.ProductRevisionLevel = subVal.(string)
					}
				}
			case "Properties":
				for _, property := range value.([]interface{}) {
					for subKey, subVal := range property.(map[string]interface{}) {
						switch subKey {
						case "Slots":
							ret.Slots = strconv.FormatFloat(subVal.(float64), 'f', 0, 64)
						case "PD":
							ret.PhysicalDriveCount = strconv.FormatFloat(subVal.(float64), 'f', 0, 64)
						}
					}
				}
			}
		}
	}
	return ret
}

func physicalDrives(cmd string, pid string) map[string]physicalDrive {
	pdByte, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s show all J | grep -v ^$", cmd))
	if err != nil {
		log.Println("failed to get RAID/HBA disk info:", cmd)
		return map[string]physicalDrive{}
	}
	pdStruct := storcliRes{}
	if err := json.Unmarshal(pdByte, &pdStruct); err != nil {
		log.Println("json unmarshal failed:", err)
		return map[string]physicalDrive{}
	}
	ret := physicalDrive{}
	ret.Location = pid
	for _, data := range pdStruct.Controllers {
		for key, value := range data.ResponseData {
			switch key {
			case fmt.Sprintf("Drive %s", pid):
				if valSlice, ok := value.([]interface{}); ok {
					parseDrive(&ret, valSlice)
				}
			case fmt.Sprintf("Drive %s - Detailed Information", pid):
				if valMap, ok := value.(map[string]interface{}); ok {
					parseDriveInfo(&ret, valMap)
				}
			}
		}
	}

	if ret.State != "UBad" && ret.State != "Offln" {
		byteSMART, _ := internal.Run.Command("sh", "-c", fmt.Sprintf("smartctl /dev/bus/%s -d megaraid,%s -a -j | grep -v ^$", internal.InterfaceToString(pdStruct.Controllers[0].CommandStatus.Controller), ret.DeviceId))
		parseSMART(&ret, byteSMART)
	}
	ret.RebuildInfo = diskRebuild(cmd)
	if ret.Product == "Unkown" {
		ret.Product = strings.Join([]string{ret.OemVendor, strings.TrimSpace(ret.Model), ret.Capacity}, " ")
	}
	diskDiagnose(&ret)
	return map[string]physicalDrive{
		ret.DG: ret,
	}
}

func parseDrive(ret *physicalDrive, pdSlice []interface{}) {
	for _, pd := range pdSlice {
		if pdMap, ok := pd.(map[string]interface{}); ok {
			for subKey, subVal := range pdMap {
				switch subKey {
				case "EID:Slt":
					fields := strings.SplitN(subVal.(string), ":", 2)
					ret.EnclosureId, ret.SlotId = fields[0], fields[1]
				case "DID":
					ret.DeviceId = internal.InterfaceToString(subVal)
				case "State":
					ret.State = internal.InterfaceToString(subVal)
				case "DG":
					ret.DG = internal.InterfaceToString(subVal)
				case "Intf":
					ret.Interface = internal.InterfaceToString(subVal)
				case "Med":
					ret.MediumType = internal.InterfaceToString(subVal)
				case "SeSz":
					ret.PhysicalSectorSize = internal.InterfaceToString(subVal)
				case "Model":
					ret.Model = internal.InterfaceToString(subVal)
				case "Type":
					ret.Type = internal.InterfaceToString(subVal)
				}
			}
		}
	}
}

func parseDriveInfo(ret *physicalDrive, infoMap map[string]interface{}) {
	for subKey, subVal := range infoMap {
		valMap, ok := subVal.(map[string]interface{})
		if !ok {
			continue
		}
		if strings.HasSuffix(subKey, "State") {
			for k, v := range valMap {
				switch k {
				case "Media Error Count":
					ret.MediaErrorCount = internal.InterfaceToString(v)
				case "Other Error Count":
					ret.OtherErrorCount = internal.InterfaceToString(v)
				case "Drive Temperature":
					ret.Temperature = internal.InterfaceToString(v)
				case "Predictive Failure Count":
					ret.PredictiveFailureCount = internal.InterfaceToString(v)
				case "S.M.A.R.T alert flagged by drive":
					ret.SmartAlert = internal.InterfaceToString(v)
				}
			}
		} else if strings.HasSuffix(subKey, "attributes") {
			for k, v := range valMap {
				switch k {
				case "SN":
					ret.SN = internal.InterfaceToString(v)
				case "Manufacturer Id":
					ret.OemVendor = internal.InterfaceToString(v)
				case "Model Number":
					ret.Model = internal.InterfaceToString(v)
				case "WWN":
					ret.WWN = internal.InterfaceToString(v)
				case "Firmware Revision":
					ret.Firmware = internal.InterfaceToString(v)
				case "Device Speed":
					ret.DeviceSpeed = internal.InterfaceToString(v)
				case "Link Speed":
					ret.LinkSpeed = internal.InterfaceToString(v)
				case "Write Cache":
					ret.WriteCache = internal.InterfaceToString(v)
				case "Logical Sector Size":
					ret.LogicalSectorSize = internal.InterfaceToString(v)
				case "Physical Sector Size":
					ret.PhysicalSectorSize = internal.InterfaceToString(v)
				}
			}
		}
	}
}

func diskRebuild(cmd string) string {
	ret := map[string]string{
		"Drive-ID":            "",
		"Progress":            "",
		"Status":              "",
		"Estimated Time Left": "",
	}
	byteRBLD, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s show rebuild J | grep -v ^$", cmd))
	if err != nil {
		log.Fatalln("got rebuild info failed: ", err)
		return ""
	}
	rbMap := make(map[string]interface{})
	if err := json.Unmarshal(byteRBLD, &rbMap); err != nil {
		log.Fatalln("json unmarshal failed: ", err)
	}
	if ctr, ok := rbMap["Controllers"].([]interface{}); ok {
		for _, data := range ctr {
			if data, ok := data.(map[string]interface{}); ok {
				if response, ok := data["Response Data"]; ok {
					if val, ok := response.(map[string]interface{}); ok {
						for k, v := range val {
							switch k {
							case "Drive ID":
								ret["Drive-ID"] = internal.InterfaceToString(v)
							case "Progress%":
								ret["Progress"] = internal.InterfaceToString(v)
							case "Status":
								ret["Status"] = internal.InterfaceToString(v)
							case "Estimated Time Left":
								ret["Estimated Time Left"] = internal.InterfaceToString(v)
							}
						}
					}
				}
			}
		}
	}
	if ret["Status"] == "Not in progress" {
		return ""
	} else {
		return fmt.Sprintf("Rebuilding %s %s %s", ret["Progress"], ret["Status"], ret["Estimated Time Left"])
	}
}

func logicalDrives(cmd string, cid string, diskMap map[string]physicalDrive) []logicalDrive {
	byteLD, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s /c%s/vall show | egrep -o '[0-9]+/[0-9]+' | grep -v ^$", cmd, cid))
	if err != nil {
		log.Printf("failed to get logical drive info.")
		return []logicalDrive{}
	}
	ret := []logicalDrive{}
	lines := strings.Split(string(byteLD), "\n")
	for _, line := range lines {
		res := logicalDrive{}
		fields := internal.SplitAndTrim(line, "/")
		if len(fields) != 2 {
			continue
		}
		res.DG, res.VD = fields[0], fields[1]
		res.PhysicalDrives = append(res.PhysicalDrives, diskMap[res.DG])
		res.Location = fmt.Sprintf("/c%s/v%s", cid, fields[1])
		byteld, err := internal.Run.Command("sh", "-c", fmt.Sprintf("%s %s show all J | grep -v ^$", cmd, res.Location))
		if err != nil {
			log.Printf("failed to get logical drive info.")
			ret = append(ret, res)
			continue
		}
		ldStruct := storcliRes{}
		if err := json.Unmarshal(byteld, &ldStruct); err == nil {
			parseVD(ldStruct, &res)
		} else {
			log.Printf("failed to parse json:%v", err)
		}
		ret = append(ret, res)
	}
	return ret
}

func parseVD(vd storcliRes, ret *logicalDrive) {
	for _, vd := range vd.Controllers {
		for key, value := range vd.ResponseData {
			switch key {
			case ret.Location:
				if vdSlice, ok := value.([]interface{}); ok {
					for _, vdInfo := range vdSlice {
						if vdMap, ok := vdInfo.(map[string]interface{}); ok {
							for subKey, subVal := range vdMap {
								switch subKey {
								case "TYPE":
									ret.Type = internal.InterfaceToString(subVal)
								case "State":
									ret.State = internal.InterfaceToString(subVal)
								case "Access":
									ret.Access = internal.InterfaceToString(subVal)
								case "Consist":
									ret.Consist = internal.InterfaceToString(subVal)
								case "Cache":
									ret.Cache = internal.InterfaceToString(subVal)
								case "Size":
									ret.Capacity = internal.InterfaceToString(subVal)
								}
							}
						}
					}
				}
			case fmt.Sprintf("VD%s Properties", ret.VD):
				if ppMap, ok := value.(map[string]interface{}); ok {
					for subKey, subVal := range ppMap {
						switch subKey {
						case "Strip Size":
							ret.StripSize = internal.InterfaceToString(subVal)
						case "Number of Blocks":
							ret.NumberOfBlocks = internal.InterfaceToString(subVal)
						case "Span Depth":
							ret.SpanDepth = internal.InterfaceToString(subVal)
						case "Number of Drives Per Span":
							ret.NumberOfDrivesPerSpan = internal.InterfaceToString(subVal)
						case "OS Drive Name":
							ret.MappingFile = internal.InterfaceToString(subVal)
						case "Creation Date":
							ret.CreateTime += internal.InterfaceToString(subVal)
						case "Creation Time":
							ret.CreateTime += internal.InterfaceToString(subVal)
						case "SCSI NAA Id":
							ret.ScsiNaaId = internal.InterfaceToString(subVal)
						}
					}
				}
			}
		}
	}
}
