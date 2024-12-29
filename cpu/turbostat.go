package cpu

import (
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

type turboStat struct {
	avgFreq  int
	minFreq  int
	maxFreq  int
	baseFreq int
	isTurbo  bool
	temp     string
	watt     string
	pkgInfo  map[string]map[string]string
	thrList  []thread
}

func turbostat() (turboStat, error) {
	byteTurbostat, err := internal.Run.Command("sh", "-c", `turbostat -q -s topology,Bzy_MHz,TSC_MHz,CoreTmp,PkgTmp,PkgWatt sleep 1`)
	if err != nil {
		internal.Log.Error("failed to get turbostat information: %v", err)
		return turboStat{}, err
	}
	lines := strings.Split(string(byteTurbostat), "\n")
	if len(lines) < 2 {
		internal.Log.Error("failed to get turbostat information: %v", err)
		return turboStat{}, err
	}
	coreTemp := make(map[string]string)
	ret := turboStat{
		minFreq: 0,
	}
	for _, line := range lines[2:] {

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if strings.Contains(line, "-") {
			if tsc_mhz, err := strconv.Atoi(fields[4]); err == nil {
				ret.baseFreq = tsc_mhz
			}
			ret.temp = fields[6]
			ret.watt = fields[7]
			continue
		}
		thr := thread{
			PhyID:     fields[0],
			CoreID:    fields[1],
			Processor: fields[2],
			Freq:      fields[3],
		}
		tag := strings.Join(fields[:2], "_")

		if len(fields) >= 6 {
			if _, ok := coreTemp[tag]; !ok {
				coreTemp[tag] = fields[5]
			}
		}
		if len(fields) == 8 {
			ret.pkgInfo[fields[0]] = map[string]string{
				"temp": fields[6],
				"watt": fields[7],
			}
		}

		thr.Temp = coreTemp[tag]
		if bzy_mhz, err := strconv.Atoi(fields[3]); err == nil {
			if bzy_mhz > ret.maxFreq {
				ret.maxFreq = bzy_mhz
			}
			if bzy_mhz < ret.minFreq {
				ret.minFreq = bzy_mhz
			}
		}
		ret.thrList = append(ret.thrList, thr)

	}
	ret.isTurbo = ret.minFreq > ret.baseFreq
	return ret, nil
}
