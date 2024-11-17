package utils

import (
	"log/slog"
	"regexp"
	"strings"
)

type DmiOption string

const (
	All          DmiOption = "all"
	BIOS         DmiOption = "0"
	System       DmiOption = "1"
	BaseBoard    DmiOption = "2"
	Chassis      DmiOption = "3"
	Processor    DmiOption = "4"
	Cache        DmiOption = "7"
	MemoryArray  DmiOption = "16"
	MemoryDevice DmiOption = "17"
	PowerSupply  DmiOption = "39"
)

var (
	handleRe  = regexp.MustCompile(`^Handle\s+(.+),\s+DMI\s+type\s+(\d+),\s+\d+\s+bytes`)
	recordRe  = regexp.MustCompile(`\t(.+):\s+(.+)$`)
	recordRe2 = regexp.MustCompile(`\t(.+):$`)
	eleRe     = regexp.MustCompile(`\t\t(.+)`)
)

func (d DmiOption) Dmidecode() []map[string]interface{} {
	var (
		byteDmi []byte
		err     error
		ret     = make([]map[string]interface{}, 0)
	)

	if d == "all" {
		byteDmi, err = run.Command("dmidecode")
	} else {
		byteDmi, err = run.Command("sudo", "dmidecode", "-t", string(d))
	}
	if err != nil {
		slog.Info("No dmidecode information found.", "Error", err)
		return ret
	}

	sliceDmi := SplitAndTrim(string(byteDmi), "\n\n")

	for _, strDmi := range sliceDmi {
		handle := handleRe.FindAllStringSubmatch(strDmi, -1)
		if handle == nil {
			continue
		}
		line := strings.Split(strDmi, "\n")
		record := parseRecord(line)
		if record != nil {
			ret = append(ret, record)
		}
	}
	return ret
}

func parseRecord(lines []string) map[string]interface{} {
	if len(lines) < 2 {
		return nil
	}

	handle := handleRe.FindAllStringSubmatch(lines[0], -1)
	if handle == nil {
		return nil
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

	return record
}
