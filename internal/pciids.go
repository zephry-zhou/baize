package internal

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type PCIids struct {
	Class  map[string]*class  `json:"Class,omitempty"`
	Vendor map[string]*vendor `json:"Vendor,omitempty"`
	Device map[string]*device `json:"Device,omitempty"`
}

type class struct {
	ID       string      `json:"ID,omitempty"`
	Name     string      `json:"Name,omitempty"`
	SubClass []*subClass `json:"SubClass,omitempty"`
}

type subClass struct {
	ID               string              `json:"ID,omitempty"`
	Name             string              `json:"Name,omitempty"`
	ProgramInterface []*programInterface `json:"ProgramInterface,omitempty"`
}

type programInterface struct {
	ID   string `json:"ID,omitempty"`
	Name string `json:"Name,omitempty"`
}

type vendor struct {
	ID      string    `json:"ID,omitempty"`
	Name    string    `json:"Name,omitempty"`
	Devices []*device `json:"Devices,omitempty"`
}

type device struct {
	ID        string       `json:"ID,omitempty"`
	Name      string       `json:"Name,omitempty"`
	VendorID  string       `json:"VendorID,omitempty"`
	Subsystem []*subsystem `json:"Subsystem,omitempty"`
}

type subsystem struct {
	ID          string `json:"ID,omitempty"`
	SubVendorID string `json:"SubVendorID,omitempty"`
	SubDeviceID string `json:"SubDeviceID,omitempty"`
	Name        string `json:"Name,omitempty"`
}

var (
	filePath []string = []string{
		"/usr/share/misc/pci.ids",
		"/usr/share/misc/pci.ids.gz",
		"/usr/share/hwdata/pci.ids",
		"/usr/share/hwdata/pci.ids.gz",
	}
	fileURL string = "https://raw.githubusercontent.com/pciutils/pciids/master/pci.ids"
)

func getFile() string {
	tmpFile := "/tmp/pci.ids"
	for _, path := range filePath {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	resp, err := http.Get(fileURL)
	if err != nil {
		Log.Error("failed to get pci.ids")
		return ""
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		Log.Error("failed to read pci.ids")
		return ""
	}
	err = os.WriteFile(tmpFile, respBody, 0644)
	if err != nil {
		Log.Error("failed to write pci.ids")
		return ""
	}
	return tmpFile
}

func getPCIContent(file string) (io.ReadCloser, error) {
	f, err := os.Open(file)
	if err != nil {
		Log.Error("failed to open pci.ids")
		return nil, err
	}
	if strings.HasSuffix(file, ".gz") {
		zipF, err := gzip.NewReader(f)
		if err != nil {
			Log.Error("failed to open pci.ids.gz")
			return nil, err
		}
		defer zipF.Close()
		return zipF, nil
	}
	return f, nil
}

func parsePCIContent(r io.ReadCloser) *PCIids {
	defer r.Close()
	scan := bufio.NewScanner(r)
	isClass := false
	var (
		curClasses          *class
		curVendor           *vendor
		curDevice           *device
		curSubClass         *subClass
		curProgramInterface *programInterface
		curSubsystem        *subsystem

		classes           = make(map[string]*class, 20)
		vendors           = make(map[string]*vendor, 200)
		devices           = make(map[string]*device, 1000)
		vendorDevice      = make([]*device, 0)
		subsystems        = make([]*subsystem, 0)
		subClasses        = make([]*subClass, 0)
		programInterfaces = make([]*programInterface, 0)
	)

	for scan.Scan() {
		line := scan.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lineByte := []rune(line)
		if lineByte[0] == 'C' {
			if curClasses != nil {
				curClasses.SubClass = subClasses
				subClasses = make([]*subClass, 0)
			}
			isClass = true
			curClasses = &class{
				ID:       string(lineByte[2:4]),
				Name:     string(lineByte[6:]),
				SubClass: subClasses,
			}
			classes[curClasses.ID] = curClasses
			continue
		}

		if lineByte[0] != '\t' {
			if curVendor != nil {
				curVendor.Devices = vendorDevice
				vendorDevice = make([]*device, 0)
			}
			isClass = false
			curVendor = &vendor{
				ID:      string(lineByte[0:4]),
				Name:    string(lineByte[6:]),
				Devices: vendorDevice,
			}
			vendors[curVendor.ID] = curVendor
			continue
		}

		if len(lineByte) != 0 && lineByte[1] != '\t' {
			if isClass {
				if curSubClass != nil {
					curSubClass.ProgramInterface = programInterfaces
					programInterfaces = make([]*programInterface, 0)
				}
				curSubClass = &subClass{
					ID:               string(lineByte[1:3]),
					Name:             string(lineByte[5:]),
					ProgramInterface: programInterfaces,
				}
				subClasses = append(subClasses, curSubClass)
			} else {
				if curDevice != nil {
					curDevice.Subsystem = subsystems
					subsystems = make([]*subsystem, 0)
				}
				curDevice = &device{
					ID:        fmt.Sprintf("%s %s", curVendor.ID, string(lineByte[1:5])),
					Name:      string(lineByte[7:]),
					VendorID:  curVendor.ID,
					Subsystem: subsystems,
				}
				vendorDevice = append(vendorDevice, curDevice)
				devices[curDevice.ID] = curDevice
			}
		} else {
			if isClass {
				curProgramInterface = &programInterface{
					ID:   string(lineByte[2:4]),
					Name: string(lineByte[6:]),
				}
				programInterfaces = append(programInterfaces, curProgramInterface)
			} else {
				curSubsystem = &subsystem{
					ID:          string(lineByte[2:11]),
					SubVendorID: string(lineByte[2:6]),
					SubDeviceID: string(lineByte[7:11]),
					Name:        string(lineByte[13:]),
				}
				subsystems = append(subsystems, curSubsystem)
			}
		}
	}
	return &PCIids{
		Class:  classes,
		Device: devices,
		Vendor: vendors,
	}
}

func New() *PCIids {
	file := getFile()
	f, err := getPCIContent(file)
	if err != nil {
		Log.Error("failed to get pci.ids")
		return nil
	}
	return parsePCIContent(f)
}

func (p *PCIids) JsonFormat() {
	data, err := json.MarshalIndent(p, "", "   ")
	if err != nil {
		Log.Error("failed to marshal pci.ids")
		return
	}
	fmt.Println(string(data))
}
