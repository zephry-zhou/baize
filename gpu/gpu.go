package gpu

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zephry-zhou/baize/internal"
)

type GPU struct {
	GraphicsCards []*graphicsCard `json:"GPU List,omitempty"`
}

type graphicsCard struct {
	IsOnBoard     bool `json:"On Board,omitempty"`
	*internal.PCI `json:"PCIe Info,omitempty"`
}

func (g *GPU) Result() {
	g.fromDrm()
	if g == nil {
		g.fromLspci()
	}
}

func (g *GPU) BriefFormat() {
	println("[GPU INFO]")
	for _, gpu := range g.GraphicsCards {
		println()
		internal.StructSelectFieldOutput(*gpu, []string{"IsOnBoard", "PCIID"}, 1)
	}
}

func (g *GPU) Format() {
	g.BriefFormat()
}

func (g *GPU) JsonFormat() {
	byteRet, err := json.MarshalIndent(g, "", "")
	if err != nil {
		internal.Log.Error("json marshal failed: ", err)
		os.Exit(1)
	}
	fmt.Println(string(byteRet))
}
