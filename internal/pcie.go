package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PCI struct {
	PCIID   string `json:"PCI ID,omitempty"`
	PCIAddr string `json:"PCI Address,omitempty"`
	*modaliasInfo
	Driver   string   `json:"Driver,omitempty"`
	Numa     string   `json:"NUMA Node,omitempty"`
	Revision string   `json:"Revision,omitempty"`
	Link     *pciLink `json:"PCI Link,omitempty"`
}

type pciLink struct {
	MaxSpeed  string `json:"Max Link Speed,omitempty"`
	MaxWidth  string `json:"Max Link Width,omitempty"`
	CurrSpeed string `json:"Current Link Speed,omitempty"`
	CurrWidth string `json:"Current Link Width,omitempty"`
}

type modaliasInfo struct {
	VendorID   string `json:"Vendor ID,omitempty"`
	Vendor     string `json:"Vendor,omitempty"`
	DeviceID   string `json:"Device ID,omitempty"`
	Device     string `json:"Device,omitempty"`
	SVendorID  string `json:"SubVendor ID,omitempty"`
	SVendor    string `json:"SubVendor,omitempty"`
	SDeviceID  string `json:"SubDevice ID,omitempty"`
	SDevice    string `json:"SubDevice,omitempty"`
	ClassID    string `json:"Class ID,omitempty"`
	Class      string `json:"Class,omitempty"`
	SubClassID string `json:"SubClass ID,omitempty"`
	ProgIfID   string `json:"Programming Interface ID,omitempty"`
}

var (
	modaliasExceptLen        = 54
	busDir            string = "/sys/bus/pci/devices"
	pcidb                    = New()
)

func modalias(pciAddr string) *modaliasInfo {
	f := filepath.Join(busDir, pciAddr, "modalias")
	if _, err := os.Stat(f); err != nil {
		Log.Error("modalias file not found for ", pciAddr)
		return nil
	}
	data, err := os.ReadFile(f)
	if err != nil {
		Log.Error("failed to read modalias file for ", pciAddr)
		return nil
	}

	return parseModalias(data)
}

func parseModalias(data []byte) *modaliasInfo {
	if len(data) < modaliasExceptLen {
		Log.Error("modalias file too short")
		return nil
	}
	vendorID := strings.ToLower(string(data[9:13]))
	deviceID := strings.ToLower(string(data[18:22]))
	sVendorID := strings.ToLower(string(data[28:32]))
	sDeviceID := strings.ToLower(string(data[38:42]))
	classID := strings.ToLower(string(data[44:46]))
	subClassID := strings.ToLower(string(data[48:50]))
	progIfID := strings.ToLower(string(data[51:53]))
	vendor := getVendor(vendorID)
	sVendor := getVendor(sVendorID)
	device := getDevice(fmt.Sprintf("%s %s", vendorID, deviceID))
	sDevice := getSDevice(fmt.Sprintf("%s %s", vendorID, deviceID), fmt.Sprintf("%s %s", sVendorID, sDeviceID))
	if sDevice == "Unknown" {
		sDevice = device
	}
	class := getClass(classID, subClassID)
	return &modaliasInfo{
		Vendor:     vendor,
		VendorID:   vendorID,
		Device:     device,
		DeviceID:   deviceID,
		SVendorID:  sVendorID,
		SVendor:    sVendor,
		SDeviceID:  sDeviceID,
		SDevice:    sDevice,
		Class:      class,
		ClassID:    classID,
		SubClassID: subClassID,
		ProgIfID:   progIfID,
	}
}

func link(pciAddr string) *pciLink {
	linkPath := filepath.Join(busDir, pciAddr, "*_link_*")
	f, err := filepath.Glob(linkPath)
	if err != nil {
		Log.Error("link file not found for ", pciAddr)
		return nil
	}
	if len(f) == 0 {
		return nil
	}
	ret := pciLink{}
	for _, fn := range f {
		info, err := os.ReadFile(fn)
		if err != nil {
			Log.Error("failed to read file ", fn)
			continue
		}
		value := strings.TrimSpace(string(info))
		field := filepath.Base(fn)
		switch field {
		case "max_link_speed":
			ret.MaxSpeed = value
		case "max_link_width":
			ret.MaxWidth = value
		case "current_link_speed":
			ret.CurrSpeed = value
		case "current_link_width":
			ret.CurrWidth = value
		}
	}
	return &ret
}

func pciDriver(pciAddr string) string {
	file := filepath.Join(busDir, pciAddr, "driver")
	if _, err := os.Stat(file); err != nil {
		Log.Error("driver file not found for ", pciAddr)
		return ""
	}
	dest, err := os.Readlink(file)
	if err != nil {
		Log.Error("failed to read driver file for ", pciAddr)
		return ""
	}
	return filepath.Base(dest)
}

func pciNUMA(pciAddr string) string {
	file := filepath.Join(busDir, pciAddr, "numa_node")
	if _, err := os.Stat(file); err != nil {
		Log.Error("numa file not found for ", pciAddr)
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		Log.Error("failed to read numa file for", pciAddr)
	}
	return strings.TrimSpace(string(data))
}

func pciRevision(pciAddr string) string {
	file := filepath.Join(busDir, pciAddr, "revision")
	if _, err := os.Stat(file); err != nil {
		Log.Error("revision file not found for ", pciAddr)
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		Log.Error("failed to read revision file for ", pciAddr)
	}
	return strings.TrimSpace(string(data))
}

func getVendor(vendorID string) string {
	vendorMap := pcidb.Vendor
	if vendor, ok := vendorMap[vendorID]; ok {
		return vendor.Name
	}
	return "Unknown"
}

func getClass(classID, subClassId string) string {
	classMap := pcidb.Class
	if class, ok := classMap[classID]; ok {
		for _, subClass := range class.SubClass {
			if subClass.ID == subClassId {
				return subClass.Name
			}
		}
	}
	return "Unknown"
}

func getDevice(id string) string {
	deviceMap := pcidb.Device
	if device, ok := deviceMap[id]; ok {
		return device.Name
	}
	return "Unknown"
}

func getSDevice(id, subID string) string {
	deviceMap := pcidb.Device
	if device, ok := deviceMap[id]; ok {
		for _, subDevice := range device.Subsystem {
			if subDevice.ID == subID {
				return subDevice.Name
			}
		}
	}
	return "Unknown"
}

func GetPCIe(pciAddr string) *PCI {
	ret := PCI{}
	ret.PCIAddr = pciAddr
	ret.Driver = pciDriver(pciAddr)
	ret.Link = link(pciAddr)
	ret.modaliasInfo = modalias(pciAddr)
	ret.Numa = pciNUMA(pciAddr)
	ret.Revision = pciRevision(pciAddr)
	ret.PCIID = fmt.Sprintf("%s:%s:%s:%s", ret.VendorID, ret.DeviceID, ret.SVendorID, ret.SDeviceID)
	return &ret
}
