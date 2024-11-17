package product

import (
	"baize/internal/utils"
	"log"
	"os"
	"strings"
)

type PRODUCT struct {
	DistribID          string `json:"Distributor ID,omitempty"`
	DistribRelease     string `json:"Distributor Release,omitempty"`
	DistribCodename    string `json:"Distributor Codename,omitempty"`
	DistribDescription string `json:"Distributor Description,omitempty"`
	IDLike             string `json:"ID Like,omitempty"`
	KernelName         string `json:"Kernel,omitempty"`
	KernelRelease      string `json:"Kernel Release,omitempty"`
	KernelVersion      string `json:"Kernel Version,omitempty"`
	HostName           string `json:"Host Name,omitempty"`
	Type               string `json:"Chassis Type,omitempty"`
	Asset              string `json:"Asset Tag,omitempty"`
	Height             string `json:"Chassis Height,omitempty"`
	PowerCords         string `json:"Numbers Of Power Cords,omitempty"`
	BiosVersion        string `json:"BIOS Version,omitempty"`
	Date               string `json:"BIOS Release Date,omitempty"`
	Revision           string `json:"BIOS Revison,omitempty"`
	Firmware           string `json:"BMC Version,omitempty"`
	Manufacturer       string `json:"Vendor,omitempty"`
	Product            string `json:"Product Name,omitempty"`
	Version            string `json:",omitempty"`
	SN                 string `json:"SN,omitempty"`
	UUID               string `json:"UUID,omitempty"`
	SKU                string `json:"SKU,omitempty"`
	Family             string `json:"Family,omitempty"`
}

var run utils.RunSheller = &utils.RunShell{}

func GetProduct() *PRODUCT {
	ret := new(PRODUCT)

	// 获取操作系统相关信息
	osFile := "/etc/os-release"
	if utils.PathExists(osFile) {
		lines, err := utils.ReadLines(osFile)
		if err != nil {
			log.Printf("Error reading lines from file %s: %v", osFile, err)
		} else {
			parseOS(ret, lines)
		}
	}

	// 获取BIOS相关信息
	biosSlice := utils.BIOS.Dmidecode()
	for _, biosMap := range biosSlice {
		parseBIOS(ret, biosMap)
	}
	// 获取机箱相关信息
	chassisSlice := utils.Chassis.Dmidecode()
	for _, chassisMap := range chassisSlice {
		parseChassis(ret, chassisMap)
	}
	// 获取机器型号等信息
	SystemSlice := utils.System.Dmidecode()
	for _, systemMap := range SystemSlice {
		parseSystem(ret, systemMap)
	}
	return ret
}

func parseOS(ret *PRODUCT, lines []string) {
	for _, line := range lines {
		fields := utils.SplitAndTrim(line, "=")
		if len(fields) != 2 {
			continue
		}
		key, value := fields[0], strings.ReplaceAll(fields[1], "\"", "")
		switch key {
		case "PRETTY_NAME":
			ret.DistribDescription = value
		case "NAME":
			ret.DistribID = value
		case "VERSION_ID":
			ret.DistribRelease = value
		case "VERSION_CODENAME":
			ret.DistribCodename = value
		case "ID_LIKE":
			ret.IDLike = value
		}
	}

	byteUname, err := run.Command("uname", "-a")
	if err != nil {
		log.Printf("uname -a running failed: %v\n", err)
	}
	fields := strings.Fields(string(byteUname))
	if len(fields) > 3 {
		ret.KernelName = fields[0]
		ret.HostName = fields[1]
		ret.KernelRelease = fields[2]
	}

	byteVersion, err := run.Command("uname", "-v")
	if err != nil {
		log.Printf("uname -v running failed: %v\n", err)
	}
	ret.KernelVersion = strings.TrimSpace(string(byteVersion))

	if strings.Contains(ret.DistribID, "Debian") {
		byteVersion, err := os.ReadFile("/etc/debian_version")
		if err != nil {
			log.Printf("Failed to read /etc/debian_version: %v\n", err)
		}
		ret.DistribRelease = strings.TrimSpace(string(byteVersion))
	}
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
