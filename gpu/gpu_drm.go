package gpu

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

var drmDIr = "/sys/class/drm"

func (g *GPU) fromDrm() {
	dirEnt, err := os.ReadDir(drmDIr)
	if err != nil {
		internal.Log.Warn("/sys/class/drm does not exist on this system")
		return
	}
	for _, dir := range dirEnt {
		dirName := dir.Name()
		if !strings.HasPrefix(dirName, "card") {
			continue
		}
		if strings.ContainsRune(dirName, '-') {
			continue
		}
		uevent := filepath.Join(drmDIr, dirName, "device", "uevent")
		data, err := os.ReadFile(uevent)
		if err != nil {
			internal.Log.Warn("no uevent file in ", dirName, ", skip")
			continue
		}
		lines := strings.Split(string(data), "\n")
		g.GraphicsCards = append(g.GraphicsCards, parseUevent(lines))
	}
}

func parseUevent(lines []string) *graphicsCard {
	res := graphicsCard{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "PCI_SLOT_NAME=") {
			continue
		}
		fields := strings.Split(line, "=")
		res.PCI = internal.GetPCIe(fields[1])
		if internal.IsEmptyValue(reflect.ValueOf(res.PCI.Link)) {
			res.IsOnBoard = true
		} else {
			res.IsOnBoard = false
		}
	}
	return &res
}
