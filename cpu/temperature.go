package cpu

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

// hwmonTemp reads the CPU temperature from the hardware monitor diretory and returns a map.
// returns a map of maps,likes: {cpu_pid:{core_id:temperature}}
func hwmonTemp() (map[string]map[string]string, error) {

	ret := make(map[string]map[string]string)
	hwmonPath := `/sys/class/hwmon`
	if !internal.PathExists(hwmonPath) {
		return ret, fmt.Errorf("%s not exist", hwmonPath)
	}
	dirEntry, err := os.ReadDir(hwmonPath)
	if err != nil || len(dirEntry) == 0 {
		return ret, fmt.Errorf("ReadDir failed or %s is empty", hwmonPath)
	}

	for _, hwmon := range dirEntry {
		hwmonPath := path.Join(hwmonPath, hwmon.Name())
		if hwmon.Type().IsDir() {
			continue
		}
		symLink, err := os.Readlink(hwmonPath)
		if err != nil {
			log.Printf("Failed to read hwmon symlink: %v", err)
			continue
		}
		if strings.Contains(symLink, "coretemp") {
			pid, res, err := parseCoreTemp(hwmonPath)
			if err == nil {
				ret[pid] = res
			}
		}
	}
	return ret, nil
}

// parseCoreTemp parses the core temperature files and returns a map.
// path: /sys/class/hwmon/hwmonX (Linux).
// returns the package id,a map of core temperature, and an error.
// read temperature from the input_file ,core id and package id from the label_file.
func parseCoreTemp(path string) (string, map[string]string, error) {

	tempFiles, err := filepath.Glob(path + "/temp*_input")
	res := make(map[string]string)
	if err != nil || len(tempFiles) == 0 {
		log.Printf("The temperature file was not found in %s", path)
		return "", res, fmt.Errorf("temperature file was not found in %s", path)
	}
	var pid []byte
	for _, file := range tempFiles {

		temp, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Read %s failed: %v", file, err)
			continue
		}

		cid, err := os.ReadFile(strings.Replace(file, "input", "label", -1))
		if err != nil {
			log.Printf("ReadFile failed: %v", err)
			continue
		}

		temp, cid = bytes.TrimSpace(temp), bytes.TrimSpace(cid)

		if bytes.Contains(cid, []byte("Package id")) {
			pid = bytes.Fields(cid)[2]
		} else {
			cid = bytes.Fields(cid)[1]
		}
		t, err := strconv.Atoi(string(temp))
		if err != nil {
			log.Printf("failed to convert temperature int: %v", err)
			continue
		}
		res[string(cid)] = strconv.Itoa(t / 1000)
	}
	return string(pid), res, nil
}
