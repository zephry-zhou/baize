package network

import (
	"fmt"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

type ringBuffer struct {
	MaxRxRing string `json:"Ringbuffer Pre-max RX,omitempty"`
	MaxTxRing string `json:"Ringbuffer Pre-max TX,omitempty"`
	CurRxRing string `json:"Ringbuffer Pre-curr RX,omitempty"`
	CurTxRing string `json:"Ringbuffer Pre-curr Tx,omitempty"`
}

type channel struct {
	MaxRxChan  string `json:"Channel Pre-max RX,omitempty"`
	MaxTxChan  string `json:"Channel Pre-max TX,omitempty"`
	MaxComchan string `json:"Channel Pre-max Combined,omitempty"`
	CurRxChan  string `json:"Channel Pre-curr RX,omitempty"`
	CurTxChan  string `json:"Channel Pre-curr TX,omitempty"`
	CurComChan string `json:"Channel Pre-curr Combined,omitempty"`
}

type ethPort struct {
	Speed     string `json:"Speed,omitempty"`
	Duplex    string `json:"Duplex,omitempty"`
	LinkState string `json:"Link Detected,omitempty"`
	Port      string `json:"Port Type,omitempty"`
	Firmware  string `json:"Firmware Version,omitempty"`
	internal.PCI
	Ring ringBuffer `json:"Ring Buffer,omitempty"`
	Chan channel    `json:"Channel,omitempty"`
}

func ethtoolPort(port string) *ethPort {
	ret := new(ethPort)

	// 获取端口队列长度
	byteRingPre, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool -g %s|sed -n '/Pre-set/,/Current/p'|egrep -v 'Pre-set|Current'", port))
	if err == nil {
		lines := strings.Split(string(byteRingPre), "\n")
		parseRingBuffer(ret, lines, true)
	}
	byteRingCur, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool -g %s|sed -n '/Current/,//p'|egrep -v 'Current'", port))
	if err == nil {
		lines := strings.Split(string(byteRingCur), "\n")
		parseRingBuffer(ret, lines, false)
	}

	// 获取端口队列数
	byteChannelPre, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool -l %s|sed -n '/Pre-set/,/Current/p'|egrep -v 'Pre-set|Current'", port))
	if err == nil {
		lines := strings.Split(string(byteChannelPre), "\n")
		parseChannel(ret, lines, true)
	}
	byteChannelCur, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool -l %s|sed -n '/Current/,//p'|egrep -v 'Pre-set|Current'", port))
	if err == nil {
		lines := strings.Split(string(byteChannelCur), "\n")
		parseChannel(ret, lines, false)
	}

	// 获取端口pcie 和 驱动信息
	byteDrive, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool -i %s", port))
	if err == nil {
		lines := strings.Split(string(byteDrive), "\n")
		ethDriver(ret, lines)
	}

	// 获取端口设置
	byteSetting, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool %s", port))
	if err == nil {
		lines := strings.Split(string(byteSetting), "\n")
		ethSetting(ret, lines)
	}
	return ret
}

func parseRingBuffer(ret *ethPort, msgSlice []string, isPre bool) {
	for _, line := range msgSlice {
		fields := internal.SplitAndTrim(line, ":")
		if len(fields) != 2 {
			continue
		}
		key, value := fields[0], fields[1]
		switch key {
		case "RX":
			if isPre {
				ret.Ring.MaxRxRing = value
			} else {
				ret.Ring.CurRxRing = value
			}
		case "TX":
			if isPre {
				ret.Ring.MaxTxRing = value
			} else {
				ret.Ring.CurTxRing = value
			}
		}
	}
}

func parseChannel(ret *ethPort, msgSlice []string, isPre bool) {
	for _, line := range msgSlice {
		fields := internal.SplitAndTrim(line, ":")
		if len(fields) != 2 {
			continue
		}
		key, value := fields[0], fields[1]
		switch key {
		case "RX":
			if isPre {
				ret.Chan.MaxRxChan = value
			} else {
				ret.Chan.CurRxChan = value
			}
		case "TX":
			if isPre {
				ret.Chan.MaxTxChan = value
			} else {
				ret.Chan.CurTxChan = value
			}
		case "Combined":
			if isPre {
				ret.Chan.MaxComchan = value
			} else {
				ret.Chan.CurComChan = value
			}
		}
	}
}

func ethDriver(ret *ethPort, msgSlice []string) {
	for _, line := range msgSlice {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		key, value := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
		switch key {
		case "bus-info":
			if len(value) == 0 {
				ret.PCIAddr = "0000:00:00.0"
			} else {
				ret.PCIAddr = value
			}
		case "driver":
			ret.Driver.DriverName = value
		case "version":
			ret.Driver.DriverVersion = value
		case "firmware-version":
			ret.Firmware = value
		}
	}
	if ret.PCIAddr == "0000:00:00.0" {
		ret.Driver = ret.Driver.Driver(ret.Driver.DriverName)
	} else {
		ret.PCI = *internal.GetPCIe(ret.PCIAddr)
	}
}

func ethSetting(ret *ethPort, msgSlice []string) {
	for _, line := range msgSlice {
		fields := internal.SplitAndTrim(line, ":")
		if len(fields) != 2 {
			continue
		}
		key, value := fields[0], fields[1]
		switch key {
		case "Speed":
			ret.Speed = value
		case "Duplex":
			ret.Duplex = value
		case "Port":
			ret.Port = value
		case "Link detected":
			ret.LinkState = value
		}
	}
}
