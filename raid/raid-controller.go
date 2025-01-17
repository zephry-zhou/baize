package raid

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

func (r *Controller) Result() {
	byteCard, err := internal.Run.Command("sh", "-c", `lspci -Dn | egrep '\010[47]' | grep -v ^$`)
	if err != nil || len(byteCard) == 0 {
		log.Printf("failed to search raid controller through lspci: %v", err)
		return
	}
	ret := []RHController{}
	lines := internal.SplitAndTrim(string(byteCard), "\n")
	vrocFlag := false
	ctrNum := len(lines)

	for i := 0; i < ctrNum; i++ {
		fields := strings.Fields(lines[i])
		if strings.HasPrefix(fields[2], "1000:") {
			r.Ctr = append(r.Ctr, Storcli(fields[0], ctrNum))
		} else if strings.HasPrefix(fields[2], "103c:") {
			r.Ctr = append(r.Ctr, Hpssacli(fields[0], ctrNum))
		} else if strings.HasPrefix(fields[2], "8086:") {
			if !vrocFlag {
				r.Ctr = append(ret, Mdadm(fields[0], i))
				vrocFlag = true
			}
		} else if strings.HasPrefix(fields[2], "9005:") {
			r.Ctr = append(r.Ctr, arcconf(fields[0], ctrNum))
		} else {
			log.Println("not support raid controller vendor: ", fields[2])
		}
	}
}

func (r *Controller) JsonFormat() {
	byteRet, err := json.MarshalIndent(r, "", "")
	if err != nil {
		internal.Log.Error("json marshal failed: ", err)
		os.Exit(1)
	}
	fmt.Println(string(byteRet))
}

func (r *Controller) BriefFormat() {
}

func (r *Controller) Format() {
}

func rawValue(raw interface{}) float64 {
	var ret float64
	for key, value := range raw.(map[string]interface{}) {
		if key == "value" {
			ret = value.(float64)
		}
	}
	return ret
}

func diskDiagnose(ret *physicalDrive) {
	UnhealthyReason := make([]string, 0)
	if ret.MediaErrorCount != "0" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Media Error Count: %s", ret.MediaErrorCount))
	}
	if ret.PredictiveFailureCount != "0" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Predictive Failure Count: %s", ret.PredictiveFailureCount))
	}
	if ret.OtherErrorCount != "0" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Other Error Count: %s", ret.OtherErrorCount))
	}
	if powerTime, err := strconv.Atoi(ret.PowerOnTime); err == nil && powerTime > 61320 {
		UnhealthyReason = append(UnhealthyReason, "Power On Time more than 7 years")
	}
	if ret.SmartAlert != "No" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Alert: %s", ret.SmartAlert))
	}
	if ret.SmartHealthStatus != "true" {
		UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Health Status: %s", ret.SmartHealthStatus))
	}
	checkElements := map[float64]int{1: 0, 5: 0, 184: 0, 197: 0, 198: 0, 199: 0}
	for _, attr := range ret.SmartAttribute {
		if id, ok := attr["id"].(float64); ok {
			if _, ok := checkElements[id]; ok && rawValue(attr["raw"]) != 0 {
				UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Attribute %v: %s", id, attr["value"]))
			}
		} else {
			for key, value := range attr {
				if val, ok := value.(float64); ok && val != 0 {
					UnhealthyReason = append(UnhealthyReason, fmt.Sprintf("Smart Attribute %s: %s", key, value))
				}
			}
		}
	}
	if len(UnhealthyReason) != 0 {
		ret.DiagnoseDetail = strings.Join(UnhealthyReason, "; ")
		ret.Diagnose = "Unhealthy"
	} else {
		ret.Diagnose = "Healthy"
	}
}
