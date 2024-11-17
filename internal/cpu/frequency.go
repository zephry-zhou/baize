package cpu

import (
	"baize/internal/utils"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func cpupower() (map[string][]processor, map[string]float64) {

	msgByte, err := run.Command("sh", "-c", `cpupower monitor -m Mperf | egrep -v 'Mperf|PKG'`)

	if err != nil {
		return nil, nil
	}

	ret := make(map[string][]processor)
	lines := utils.SplitAndTrim(string(msgByte), "\n")

	temp, err := hwmonTemp()
	if err != nil || len(temp) == 0 {
		return nil, nil // todo: obtain temperature from other methods.
	}

	var freqSlice []float64

	for _, line := range lines {

		res := processor{}

		fields := utils.SplitAndTrim(line, "|")
		if len(fields) < 4 {
			continue
		}

		res.PhyID, res.CoreID, res.Processor, res.Freq = fields[0], fields[1], fields[2], fields[len(fields)-1]
		freq, _ := strconv.ParseFloat(res.Freq, 64)
		freqSlice = append(freqSlice, freq)
		res.Temp = temp[res.PhyID][res.CoreID]
		ret[res.PhyID] = append(ret[res.PhyID], res)
	}
	minfreq, maxfreq := utils.FindMinAndMax(freqSlice)

	return ret, map[string]float64{"FreqNums": float64(len(utils.UniqueSlice(freqSlice))), "MinFreq": minfreq, "MaxFreq": maxfreq}
}

func cpuinfo() (map[string][]processor, map[string]float64) {

	ret := make(map[string][]processor)
	strCPU, err := utils.ReadFile("/proc/cpuinfo")
	if err != nil {
		return ret, nil
	}

	dicTemp, err := hwmonTemp()
	if err != nil || len(dicTemp) == 0 {
		return nil, nil // todo: obtain temperature from other methods.
	}

	procs := strings.Split(strCPU, "\n\n")
	var freqSlice []float64
	for _, proc := range procs {
		res := processor{}
		lines := strings.Split(proc, "\n")
		for _, line := range lines {
			fields := utils.SplitAndTrim(line, ":")
			if len(fields) != 2 {
				continue
			}
			key := fields[0]
			value := fields[1]
			switch key {
			case "processor":
				res.Processor = value
			case "cpu MHz":
				res.Freq = value
			case "physical id":
				res.PhyID = value
			case "core id":
				res.CoreID = value
			}
			freq, _ := strconv.ParseFloat(res.Freq, 64)
			freqSlice = append(freqSlice, freq)
		}
		res.Temp = dicTemp[res.PhyID][res.CoreID]
		ret[res.PhyID] = append(ret[res.PhyID], res)
	}
	minfreq, maxfreq := utils.FindMinAndMax(freqSlice)

	return ret, map[string]float64{"FreqNums": float64(len(utils.UniqueSlice(freqSlice))), "MinFreq": minfreq, "MaxFreq": maxfreq}
}

func i7zFreq() (min, max float64) {
	// 确定 i7z 工具路径
	i7zTool := `/usr/local/baize/tool/i7z`

	// 清除临时文件
	if utils.PathExistsWithContent("/tmp/i7z_0") {
		os.Remove("/tmp/i7z_0")
	}
	if utils.PathExistsWithContent("/tmp/i7z_1") {
		os.Remove("/tmp/i7z_1")
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 启动 i7z 命令
	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("%s -w l -l /tmp/i7z --nogui", i7zTool))
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		println("i7z running failed:", err)
		return 0, 0
	}
	var cpu0Freq, cpu1Freq []string
	// 等待临时文件生成
	startTime := time.Now().Unix()
	for {
		time.Sleep(10 * time.Millisecond)
		nextTime := time.Now().Unix()
		if (nextTime - startTime) > 20 {
			break
		}
		if utils.PathExistsWithContent("/tmp/i7z_0") && utils.PathExistsWithContent("/tmp/i7z_1") {
			cpu0Freq, _ = utils.ReadLines("/tmp/i7z_0")
			cpu1Freq, _ = utils.ReadLines("/tmp/i7z_0")
			if len(cpu0Freq) == len(cpu1Freq) && len(cpu0Freq) > 4 {
				break
			}
		}
	}

	// 解析频率数据
	ret := make([]float64, 0, 64) // 预分配数组，减少内存分配次数
	for _, freq := range cpu0Freq[1:] {
		fre, err := strconv.ParseFloat(strings.TrimSpace(freq), 64)
		if err != nil {
			println("Failed to parse frequency:", err)
			continue
		}
		ret = append(ret, fre)
	}
	for _, freq := range cpu1Freq[1:] {
		fre, err := strconv.ParseFloat(strings.TrimSpace(freq), 64)
		if err != nil {
			println("Failed to parse frequency:", err)
			continue
		}
		ret = append(ret, fre)
	}

	minfreq, maxfreq := utils.FindMinAndMax(ret)

	return minfreq, maxfreq
}
