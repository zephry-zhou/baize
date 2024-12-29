package network

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/zephry-zhou/baize/internal"
)

type NETWORK struct {
	Port []netPort `json:"Net Entry"`
	Bond []bond    `json:"Bond Entry"`
}

type netPort struct {
	PortName string `json:"Device Name,omitempty"`
	MACAddr  string `json:"MAC Address,omitempty"`
	ethPort
	LLDP ethLLDP `json:"LLDP,omitempty"`
	IPv4 ipv4    `json:"IPv4 Address,omitempty"`
}

type ethLLDP struct {
	RID    string `json:"RID,omitempty"`
	MAC    string `json:"TOR MAC,omitempty"`
	Name   string `json:"TOR Name,omitempty"`
	IfName string `json:"Interface Name,omitempty"`
	MgmtIP string `json:"MGMT IP,omitempty"`
	TTL    string `json:"TTL,omitempty"`
	MFS    string `json:"Maximum Frame Size,omitempty"`
	Vlan   string `json:"Vlan,omitempty"`
	PPVID  string `json:"Port Protocol Vlan ID,omitempty"`
}

type ipv4 struct {
	IPAddr  string `json:"IP Address,omitempty"`
	Netmask string `json:"Netmask,omitempty"`
	Gateway string `json:"Gateway,omitempty"`
}

type bond struct {
	Name             string    `json:"Bond Name,omitempty"`
	Mode             string    `json:"Bonding Mode,omitempty"`
	HashPolicy       string    `json:"Hash Policy,omitempty"`
	Status           string    `json:"MII Status,omitempty"`
	LACP             string    `json:"LACP Status,omitempty"`
	LACPRate         string    `json:"LACP Rate,omitempty"`
	MACAddr          string    `json:"MAC Address,omitempty"`
	Aggregator       string    `json:"Aggregator ID,omitempty"`
	AggregatorPolicy string    `json:"Aggregator Policy,omitempty"`
	Ports            string    `json:"Number of Ports,omitempty"`
	Slave            []slaveIf `json:"Slave Interfaces,omitempty"`
}

type slaveIf struct {
	Interface  string `json:"Interface,omitempty"`
	Status     string `json:"MII Status,omitempty"`
	Speed      string `json:"Speed,omitempty"`
	Duplex     string `json:"Duplex,omitempty"`
	LinkF      string `json:"Link Failure Count,omitempty"`
	MACAddr    string `json:"MAC Address,omitempty"`
	Aggregator string `json:"Aggregator ID,omitempty"`
}

func (n *NETWORK) Result() {
	netDIR := "/sys/class/net"
	dirEntry, err := os.ReadDir(netDIR)
	if err != nil {
		log.Printf("The network port directory was not found: %v", err)
		return
	}
	for _, dir := range dirEntry {
		if dir.Name() == "lo" || dir.Name() == "bonding_masters" {
			continue
		}
		n.Port = append(n.Port, parsePort(dir.Name()))
	}
	getBond(n)
}

func (n *NETWORK) BriefFormat() {
	println("[NETWORK INFO]")
	bondField := []string{"Name", "Mode", "Status", "LACP", "MACAddr", "Speed", "LinkState", "Aggregator", "IPv4"}
	portField := []string{"PortName", "Status", "MACAddr", "PCIID", "PCIAddr", "Speed", "LinkState", "LinkF", "LLDP"}
	if internal.IsEmptyValue(reflect.ValueOf(n.Bond)) {
		for _, port := range n.Port {
			println()
			internal.StructSelectFieldOutput(port, portField, 1)
		}
	} else {
		for _, bond := range n.Bond {
			println()
			internal.StructSelectFieldOutput(bond, bondField, 1)
			for _, slave := range bond.Slave {
				println()
				internal.StructSelectFieldOutput(slave, portField, 2)
			}
		}
	}
}

func (n *NETWORK) Format() {

}

func parsePort(port string) netPort {
	ret := netPort{}
	dir := filepath.Join("/sys/class/net", port)
	ret.PortName = port
	subDir, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("failed to get %s port info: %v", port, err)
		return ret
	}
	ret.ethPort = *ethtoolPort(port)
	for _, sub := range subDir {
		if sub.Name() == "device" {
			ret.LLDP = lldpctl(port)
			if internal.IsEmptyValue(reflect.ValueOf(ret.LLDP)) && ret.Driver.DriverName == "i40e" {
				checkI40e(port, &ret)
				ret.LLDP = lldpctl(port)
			}
		} else if sub.Name() == "address" {
			addr, err := internal.ReadFile(filepath.Join(dir, "address"))
			if err != nil {
				log.Printf("failed to get %s mac address: %v", port, err)
				continue
			}
			ret.MACAddr = strings.TrimSpace(addr)
		}
	}
	ipv4Info := getIPv4(port)
	if !internal.IsEmptyValue(reflect.ValueOf(ipv4Info)) {
		ret.IPv4 = ipv4Info
	}
	return ret
}

func checkI40e(port string, ret *netPort) {
	byteFW, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool --show-priv-flags %s|awk -F: '/disable-fw-lldp/{print $2}'", port))
	if err != nil {
		log.Printf("failed to get %s fw-lldp info: %v", port, err)
		return
	}
	if strings.TrimSpace(string(byteFW)) == "off" {
		kernel, err := internal.Run.Command("uname", "-r")
		if err != nil {
			log.Printf("failed to get kernel version: %v", err)
			return
		}
		if strings.HasPrefix(string(kernel), "3.16") {
			internal.Run.Command(fmt.Sprintf("echo 'lldp stop' > /sys/kernel/debug/i40e/%s/command", ret.PCIAddr))
		}
		internal.Run.Command("sh", "-c", fmt.Sprintf("ethtool --set-priv-flags %s disable-fw-lldp on 2>/dev/null", port))
		time.Sleep(time.Second * 30)
	}
}

func lldpctl(port string) ethLLDP {
	byteLLDP, err := internal.Run.Command("sh", "-c", fmt.Sprintf("lldpctl %s -f keyvalue", port))
	if err != nil {
		log.Printf("failed to get %s lldp info.", port)
		return ethLLDP{}
	}
	ret := ethLLDP{}
	lines := internal.SplitAndTrim(string(byteLLDP), "\n")
	for _, line := range lines {
		fields := internal.SplitAndTrim(line, "=")
		if len(fields) != 2 {
			continue
		}
		key, value := fields[0], fields[1]
		key = strings.Replace(key, fmt.Sprintf("lldp.%s.", port), "", 1)
		switch key {
		case "rid":
			ret.RID = value
		case "chassis.mac":
			ret.MAC = value
		case "chassis.name":
			ret.Name = value
		case "chassis.mgmt-ip":
			ret.MgmtIP = value
		case "port.ifname":
			ret.IfName = value
		case "port.ttl":
			ret.TTL = value
		case "port.mfs":
			ret.MFS = value
		case "vlan.vlan-id":
			ret.Vlan += value
		case "vlan.pvid":
			ret.Vlan += fmt.Sprintf(" pvid:%s", value)
		case "ppvid.supported":
			ret.PPVID += fmt.Sprintf("%s supported", value)
		case "ppvid.enabled":
			ret.PPVID += fmt.Sprintf(" %s enabled", value)
		}
	}
	return ret
}

func getIPv4(port string) ipv4 {
	ret := ipv4{}
	byteIP4, err := internal.Run.Command("sh", "-c", fmt.Sprintf("ip addr show %s | awk '/inet /{print $2}'", port))
	if err != nil {
		println("failed to get ipv4 info: ", err)
	}
	lines := internal.SplitAndTrim(string(byteIP4), "\n")
	if len(lines) == 0 {
		return ret
	}
	for _, line := range lines {
		fields := internal.SplitAndTrim(line, "/")
		if len(fields) != 2 {
			continue
		}
		ret.IPAddr = fields[0]
		cid, _ := strconv.Atoi(fields[1])
		mask := cidrMask(cid, 32)
		ret.Netmask = v4String(mask)
		addr, _ := parseAddr(ret.IPAddr)
		ret.Gateway = v4String(gateway(addr, mask))
	}
	return ret
}

const (
	IPv4len = 4
	IPv6len = 16
)

func cidrMask(ones, bits int) []byte {
	if bits != 8*IPv4len && bits != 8*IPv6len {
		return nil
	}
	if ones < 0 || ones > bits {
		return nil
	}
	l := bits / 8
	m := make([]byte, l)
	n := uint(ones)
	for i := 0; i < l; i++ {
		if n > 8 {
			m[i] = 0xff
			n -= 8
			continue
		}
		m[i] = ^byte(0xff >> n)
		n = 0
	}
	return m
}

func parseAddr(addr string) ([]byte, error) {
	for i := 0; i < (len(addr)); i++ {
		switch addr[i] {
		case '.':
			return parseIPV4(addr)
		case ':':
			return parseIPV6(addr)
		case '%':
			return []byte{}, fmt.Errorf("missing IPV6 address")
		}
	}
	return []byte{}, fmt.Errorf("unable to parse IP")
}

func parseIPV4(addr string) ([]byte, error) {
	var pos, val, diglen int
	byteIP := make([]byte, 4)
	for i := 0; i < len(addr); i++ {
		if addr[i] >= '0' && addr[i] <= '9' {
			if diglen == 1 && val == 0 {
				println("addrv4 field has octet with leading zero")
				return byteIP, fmt.Errorf("IPv4 field has octet with leading zero")
			}
			val = val*10 + int(addr[i]) - '0'
			diglen++
			if val > 255 {
				return byteIP, fmt.Errorf("IPv4 field has value >255")
			}
		} else if addr[i] == '.' {
			if i == 0 || i == len(addr)-1 || addr[i-1] == '.' {
				return []byte{}, fmt.Errorf("addrv4 field must have at least one digit")
			}
			if pos == 3 {
				return byteIP, fmt.Errorf("IPv4 address too long")
			}
			byteIP[pos] = uint8(val)
			pos++
			val = 0
			diglen = 0
		} else {
			println("unexpected character")
			return byteIP, fmt.Errorf("unexpected character")
		}
	}
	if pos < 3 {
		return byteIP, fmt.Errorf("IPV4 address too short")
	}
	byteIP[3] = uint8(val)
	return byteIP, nil
}

func parseIPV6(addr string) ([]byte, error) {
	return []byte(addr), nil
}

func v4String(addr []byte) string {
	ret := make([]string, 0, 8)
	for _, seg := range addr {
		ret = append(ret, strconv.Itoa(int(seg)))
	}
	return strings.Join(ret, ".")
}

func gateway(ip, mask []byte) []byte {
	ret := make([]byte, len(ip))
	for i, seg := range ip {
		ret[i] = seg & mask[i]
	}
	ret[len(ret)-1] = ret[len(ret)-1] + 1
	return ret
}

func getBond(ret *NETWORK) {
	bondDir := `/proc/net/bonding/`
	if !internal.PathExists(bondDir) {
		println("No bond found")
		return
	}
	fileList, err := os.ReadDir(bondDir)
	if err != nil {
		println(fmt.Sprintf("not found file in %s", bondDir))
		return
	}
	for _, file := range fileList {
		res := bond{}
		res.Name = file.Name()
		info, err := internal.ReadLines(fmt.Sprintf("%s/%s", bondDir, res.Name))
		if err != nil {
			println(fmt.Sprintf("failed to open %s", file))
			continue
		}
		sFlag := false
		sRes := slaveIf{}
		res.Slave = []slaveIf{}
		for _, line := range info {
			fields := strings.SplitN(line, ":", 2)
			if len(fields) != 2 {
				continue
			}
			key, value := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
			if sFlag {
				switch key {
				case "Slave Interface":
					if !internal.IsEmptyValue(reflect.ValueOf(sRes)) {
						res.Slave = append(res.Slave, sRes)
						sRes = slaveIf{}
					}
					sRes.Interface = value
				case "MII Status":
					sRes.Status = value
				case "Speed":
					sRes.Speed = value
				case "Duplex":
					sRes.Duplex = value
				case "Link Failure Count":
					sRes.LinkF = value
				case "Permanent HW addr":
					sRes.MACAddr = value
				case "Aggregator ID":
					sRes.Aggregator = value
				}
			} else {
				switch key {
				case "Bonding Mode":
					res.Mode = value
				case "MII Status":
					res.Status = value
				case "LACP active":
					res.LACP = value
				case "System MAC address":
					res.MACAddr = value
				case "Aggregator ID":
					res.Aggregator = value
				case "Number of ports":
					res.Ports = value
				case "Slave Interface":
					sFlag = true
					sRes.Interface = value
				}
			}
		}
		res.Slave = append(res.Slave, sRes)
		ret.Bond = append(ret.Bond, res)
	}
}
