package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PCI struct {
	Class     string    `json:"Class,omitempty"`
	ClassID   string    `json:"Class ID,omitempty"`
	Device    string    `json:"Device Name,omitempty"`
	DeviceID  string    `json:"Device ID,omitempty"`
	Vendor    string    `json:"Vendor,omitempty"`
	VendorID  string    `json:"Vendor ID,omitempty"`
	SVendor   string    `json:"Sub Vendor,omitempty"`
	SVendorID string    `json:"Sub Vendor ID,omitempty"`
	SDevice   string    `json:"Sub Device Name,omitempty"`
	SDeviceID string    `json:"Sub Deive ID,omitempty"`
	PhySlot   string    `json:"Physical Slot,omitempty"`
	PCIID     string    `json:"PCI ID,omitempty"`
	PCIAddr   string    `json:"PCI Address,omitempty"`
	Driver    pciDriver `json:"Driver,omitempty"`
	Link      pciLink   `json:"Link Width,omitempty"`
}

type pciDriver struct {
	DriverName    string `json:"Driver Name,omitempty"`
	DriverVersion string `json:"Driver Version,omitempty"`
	DriverFile    string `json:"Driver File,omitempty"`
}
type pciLink struct {
	MaxSpeed  string `json:"Max Link Speed,omitempty"`
	MaxWidth  string `json:"Max Link Width,omitempty"`
	CurrSpeed string `json:"Current Link Speed,omitempty"`
	CurrWidth string `json:"Current Link Width,omitempty"`
}

func GetPCIe(slot string) *PCI {
	ret := new(PCI)
	msgByte, err := Run.Command("sh", "-c", fmt.Sprintf("lspci -Dnmv -s %s | grep -v %s", slot, slot))
	if err == nil {
		nameOrID(ret, strings.Split(string(msgByte), "\n"), true)
	}
	msgByte, err = Run.Command("sh", "-c", fmt.Sprintf("lspci -Dmkv -s %s | grep -v %s", slot, slot))
	if err == nil {
		nameOrID(ret, strings.Split(string(msgByte), "\n"), false)
	}
	ret.PCIAddr = slot
	ret.PCIID = fmt.Sprintf("%s:%s:%s:%s", ret.VendorID, ret.DeviceID, ret.SVendorID, ret.SDeviceID)
	ret.Link = ret.Link.link(slot)
	return ret
}

func nameOrID(ret *PCI, lines []string, isID bool) {
	for _, line := range lines {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		switch key {
		case "Class":
			if isID {
				ret.ClassID = value
			} else {
				ret.Class = value
			}
		case "Vendor":
			if isID {
				ret.VendorID = value
			} else {
				ret.Vendor = value
			}
		case "Device":
			if isID {
				ret.DeviceID = value
			} else {
				ret.Device = value
			}
		case "SVendor":
			if isID {
				ret.SVendorID = value
			} else {
				ret.SVendor = value
			}
		case "SDevice":
			if isID {
				ret.SDeviceID = value
			} else {
				ret.SDevice = value
			}
		case "PhySlot":
			ret.PhySlot = value
		case "Driver":
			ret.Driver = ret.Driver.Driver(value)
		}
	}
}

func (p *pciDriver) Driver(drv string) pciDriver {
	byteInfo, err := Run.Command("sh", "-c", fmt.Sprintf("modinfo %s | egrep '^(filename|version)'", drv))
	if err != nil {
		Log.Info("modinfo obtain driver failed")
		return pciDriver{}
	}
	lines := strings.Split(string(byteInfo), "\n")
	p.DriverName = drv
	for _, line := range lines {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		key, value := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
		switch key {
		case "filename":
			p.DriverFile = value
		case "version":
			p.DriverVersion = value
		}
	}
	return *p
}
func (p *pciLink) link(slot string) pciLink {
	linkPath := fmt.Sprintf("/sys/bus/pci/devices/%s/*_link_*", slot)
	file, err := filepath.Glob(linkPath)
	if err != nil {
		return pciLink{}
	}
	for _, fn := range file {
		info, err := os.ReadFile(fn)
		if err != nil {
			fmt.Printf("failed to read file %s", fn)
		}
		value := strings.TrimSpace(string(info))
		field := strings.Split(fn, "/")
		if len(field) > 0 {
			switch field[len(field)-1] {
			case "max_link_speed":
				p.MaxSpeed = value
			case "max_link_width":
				p.MaxWidth = value
			case "current_link_speed":
				p.CurrSpeed = value
			case "current_link_width":
				p.CurrWidth = value
			}
		}
	}
	return *p
}
