package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zephry-zhou/baize/cpu"
	"github.com/zephry-zhou/baize/memory"
	"github.com/zephry-zhou/baize/network"
)

type InfoGetter interface {
	Result()
	BriefFormat()
	Format()
}

var surpportedModes = map[string]bool{
	"all":     true,
	"product": true,
	"cpu":     true,
	"memory":  true,
	"raid":    true,
	"network": true,
	"gpu":     true,
	"power":   true,
	"system":  true,
}

func main() {

	mode := flag.String("m", "all", fmt.Sprintf("Query mode infomation,surpported value: %v", getSurpportedModes()))
	detail := flag.Bool("d", false, "Show detail information.")
	js := flag.Bool("j", false, "Output json format.")
	flag.Parse()

	if _, ok := surpportedModes[*mode]; !ok {
		fmt.Println("Unsupported mode:", *mode)
		flag.Usage()
		os.Exit(1)
	}
	if *js {
		Printjson()
	} else {
		Printdetail(*detail, *mode)
	}
}

func Printdetail(d bool, m string) {

	exeMap := map[string]InfoGetter{
		"cpu":     &cpu.CPU{},
		"memory":  &memory.MEMORY{},
		"network": &network.NETWORK{},
	}

	var process func(InfoGetter)

	if d {
		process = func(i InfoGetter) {
			i.Result()
			i.Format()
		}
	} else {
		process = func(i InfoGetter) {
			i.Result()
			i.BriefFormat()
		}
	}

	if m == "all" {
		for _, v := range exeMap {
			process(v)
		}
	} else {
		process(exeMap[m])
	}
}

func Printjson() {

	exeMap := map[string]InfoGetter{
		"cpu":     &cpu.CPU{},
		"memory":  &memory.MEMORY{},
		"network": &network.NETWORK{},
	}

	for _, v := range exeMap {
		v.Result()
	}
}

func getSurpportedModes() []string {
	var modes []string
	for k := range surpportedModes {
		modes = append(modes, k)
	}
	return modes
}
