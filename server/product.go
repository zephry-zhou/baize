package server

import (
	"github.com/zephry-zhou/baize/internal"
)

type product struct {
	Family string `json:"Family,omitempty"`
}

func GetProduct() *PRODUCT {
	ret := new(PRODUCT)

	// 获取BIOS相关信息
	biosSlice := internal.DMI["0"]
	for _, biosMap := range biosSlice {
		parseBIOS(ret, biosMap)
	}
	// 获取机箱相关信息
	chassisSlice := internal.DMI["4"]
	for _, chassisMap := range chassisSlice {
		parseChassis(ret, chassisMap)
	}
	// 获取机器型号等信息
	SystemSlice := internal.DMI["1"]
	for _, systemMap := range SystemSlice {
		parseSystem(ret, systemMap)
	}
	return ret
}

func parseBIOS(ret *PRODUCT, biosMap map[string]interface{}) {

	for key, value := range biosMap {
		var v string
		if str, ok := value.(string); ok {
			v = str
		}
		switch key {
		case "Version":
			ret.Version = v
		case "Release Date":
			ret.Date = v
		case "BIOS Revision":
			ret.Revision = v
		case "Firmware Revision":
			ret.Firmware = v
		}
	}
}

func parseChassis(ret *PRODUCT, chassisMap map[string]interface{}) {
	for key, value := range chassisMap {
		var v string
		if str, ok := value.(string); ok {
			v = str
		}
		switch key {
		case "Type":
			ret.Type = v
		case "Asset Tag":
			ret.Asset = v
		case "Height":
			ret.Height = v
		case "Number Of Power Cords":
			ret.PowerCords = v
		}
	}
}

func parseSystem(ret *PRODUCT, systemMap map[string]interface{}) {
	for key, value := range systemMap {
		var v string
		if str, ok := value.(string); ok {
			v = str
		}
		switch key {
		case "Manufacturer":
			ret.Manufacturer = v
		case "Product Name":
			ret.Product = v
		case "Serial Number":
			ret.SN = v
		case "UUID":
			ret.UUID = v
		case "SKU Number":
			ret.SKU = v
		case "Family":
			ret.Family = v
		}
	}
}
