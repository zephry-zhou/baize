package internal

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
)

var (
	handleRe                                      = regexp.MustCompile(`^Handle\s+(.+),\s+DMI\s+type\s+(\d+),\s+\d+\s+bytes`)
	recordRe                                      = regexp.MustCompile(`\t(.+):\s+(.+)$`)
	recordRe2                                     = regexp.MustCompile(`\t(.+):$`)
	eleRe                                         = regexp.MustCompile(`\t\t(.+)`)
	dmiCache  map[string][]map[string]interface{} = map[string][]map[string]interface{}{}
	once      sync.Once
)

// 通过sync.Once保证只执行一次,避免重复执行
// dmi信息将缓存在变量DMI中
func getDmiInfo() map[string][]map[string]interface{} {
	once.Do(func() {
		byteDmi, err := Run.Command("dmidecode")
		if err != nil {
			slog.Info("No dmidecode information found.", "Error", err)
			dmiCache = map[string][]map[string]interface{}{}
		}

		sliceDmi := SplitAndTrim(string(byteDmi), "\n\n")

		for _, strDmi := range sliceDmi {
			handle := handleRe.FindAllStringSubmatch(strDmi, -1)
			if handle == nil {
				continue
			}
			line := strings.Split(strDmi, "\n")
			hid, record := parseRecord(line)
			if record != nil {
				dmiCache[hid] = append(dmiCache[hid], record)
			}
		}
	})
	return dmiCache
}

func parseRecord(lines []string) (string, map[string]interface{}) {
	if len(lines) < 2 {
		return "", nil
	}

	handle := handleRe.FindAllStringSubmatch(lines[0], -1)
	if handle == nil {
		return "", nil
	}

	record := make(map[string]interface{})
	record["DMI Type"] = handle[0][2]
	record["DMI Name"] = lines[1]
	record["Handle ID"] = handle[0][1]

	var cha string
	record2 := []string{}

	for i := 2; i < len(lines); i++ {
		switch {
		case recordRe.MatchString(lines[i]):
			sliceRecord := recordRe.FindAllStringSubmatch(lines[i], -1)
			if len(sliceRecord) > 0 {
				key := strings.TrimSpace(sliceRecord[0][1])
				value := strings.TrimSpace(sliceRecord[0][2])
				record[key] = value
			}
		case recordRe2.MatchString(lines[i]):
			sliceRecord2 := recordRe2.FindAllStringSubmatch(lines[i], -1)
			if len(sliceRecord2) > 0 {
				cha = strings.TrimSpace(sliceRecord2[0][1])
			}
		case eleRe.MatchString(lines[i]):
			sliceEle := eleRe.FindAllStringSubmatch(lines[i], -1)
			if len(sliceEle) > 0 {
				record2 = append(record2, strings.TrimSpace(sliceEle[0][0]))
			}
		}
	}

	if len(cha) != 0 || len(record2) != 0 {
		record[cha] = record2
	}
	return handle[0][2], record
}

func format() {
	for k, v := range dmiCache {
		fmt.Printf("Handle ID: %s\n", k)
		for _, v2 := range v {
			for k1, v3 := range v2 {
				if v3 == nil {
					continue
				}
				switch v3.(type) {
				case string:
					fmt.Printf("%s: %s\n", k1, v3)
				case []string:
					for _, v4 := range v3.([]string) {
						fmt.Printf("%s: %s\n", k1, v4)
					}
				}
			}
		}
	}
}
