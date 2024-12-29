package raid

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

func parseSMART(ret *physicalDrive, s []byte) {
	v := make(map[string]interface{})
	err := json.Unmarshal(s, &v)
	if err != nil {
		log.Fatalln(err)
	}
	for key, value := range v {
		switch key {
		case "device":
			vm := value.(map[string]interface{})
			if vm["protocol"] == "ATA" {
				ret.Interface = "SATA"
			} else if vm["protocol"] == "SCSI" {
				ret.Interface = "SAS"
			} else {
				ret.Interface = vm["protocol"].(string)
			}
		case "model_name":
			ret.Vendor, ret.Product = diskManufacturer(value.(string))
		case "nvme_total_capacity":
			cap, unit, err := internal.Unit2Human(value.(float64), "B", true)
			if err != nil {
				log.Fatalln("converion disk size failed: ", err)
			}
			ret.Capacity = strings.Join([]string{fmt.Sprintf("%.2f", cap), unit}, " ")
		case "user_capacity":
			cap, unit, err := internal.Unit2Human(value.(map[string]interface{})["bytes"].(float64), "B", true)
			if err != nil {
				log.Fatalln("converion disk size failed: ", err)
			}
			ret.Capacity = strings.Join([]string{fmt.Sprintf("%.2f", cap), unit}, " ")
		case "serial_number":
			ret.SN = value.(string)
		case "firmware_version", "revision", "firmware version":
			ret.Firmware = value.(string)
		case "rotation_rate":
			if strconv.FormatFloat(value.(float64), 'f', 0, 64) == "0" {
				ret.RotationRate = "Solid State Device"
			} else {
				ret.RotationRate = strconv.FormatFloat(value.(float64), 'f', 0, 64)
			}
		case "form_factor":
			ret.FormFactor = value.(map[string]interface{})["name"].(string)
		case "smart_status":
			ret.SmartHealthStatus = strconv.FormatBool(value.(map[string]interface{})["passed"].(bool))
		case "temperature":
			ret.Temperature = strconv.FormatFloat(value.(map[string]interface{})["current"].(float64), 'f', 0, 64)
		case "power_on_time":
			ret.PowerOnTime = strconv.FormatFloat(value.(map[string]interface{})["hours"].(float64), 'f', 0, 64)
		case "scsi_grown_defect_list":
			ret.SmartAttribute = append(ret.SmartAttribute, map[string]interface{}{"scsi_grown_defect_list": strconv.FormatFloat(value.(float64), 'f', 0, 64)})
		case "scsi_error_counter_log":
			vm := value.(map[string]interface{})
			for _, i := range []string{"read", "write", "verify"} {
				if _, ok := vm[i]; ok {
					v := vm[i].(map[string]interface{})["total_uncorrected_errors"]
					ret.SmartAttribute = append(ret.SmartAttribute, map[string]interface{}{i: v})
				}
			}
		case "ata_smart_attributes":
			vm := value.(map[string]interface{})["table"]
			for _, attr := range vm.([]interface{}) {
				ret.SmartAttribute = append(ret.SmartAttribute, attr.(map[string]interface{}))
			}
		case "nvme_smart_health_information_log":
			ret.SmartAttribute = append(ret.SmartAttribute, value.(map[string]interface{}))
		}
	}
}

func diskManufacturer(s string) (string, string) {
	retVendor, retProduct := "Unkown", "Unkown"
	strReplace := []string{"IBM-ESXS", "HP", "LENOVO-X", "ATA",
		"-", "_", "SAMSUNG", "INTEL", "SEAGATE", "TOSHIBA", "HGST",
		"Micron", "KIOXIA"}
	for _, i := range strReplace {
		s = strings.ReplaceAll(strings.TrimSpace(s), i, " ")
	}
	dir, _ := os.Getwd()
	filePath := fmt.Sprintf("%s/config/devmap.json", dir)
	devmap := internal.ReadJSONFile(filePath)
	sl := strings.Split(s, " ")
	for key, value := range devmap["disk"].(map[string]interface{}) {
		for _, i := range value.([]interface{}) {
			rule := i.(map[string]interface{})["regular"].(string)
			name := i.(map[string]interface{})["stdName"].(string)
			mat, err := regexp.MatchString(rule, sl[len(sl)-1])
			if mat && err == nil {
				if key == "manufaceturer" {
					retProduct = name
				} else {
					retVendor = name
				}
			}
		}
	}
	return retVendor, retProduct
}
