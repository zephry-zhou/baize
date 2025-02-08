package gpu

import (
	"reflect"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

func (g *GPU) fromLspci() {
	pci, err := internal.Run.Command("lspci", "-Dnn", "| egrep", `'\[030[0-2]\]'`)
	if err != nil {
		internal.Log.Error("failed to get gpu pci by lspci: ", err)
		return
	}
	lines := strings.Split(string(pci), "\n")
	if len(lines) == 0 {
		return
	}
	for _, line := range lines {
		res := &graphicsCard{}
		addr := strings.Fields(line)[0]
		res.PCI = internal.GetPCIe(addr)
		if internal.IsEmptyValue(reflect.ValueOf(res.PCI.Link)) {
			res.IsOnBoard = true
		} else {
			res.IsOnBoard = false
		}
		g.GraphicsCards = append(g.GraphicsCards, res)
	}
}
