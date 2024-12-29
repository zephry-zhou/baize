package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	TimeOut    = 20 * time.Second
	ErrTimeOut = fmt.Errorf("command time out")
	Run        RunSheller
	DMI        map[string][]map[string]interface{}
	Log        *Logger
)

func init() {
	Run = &RunShell{}
	DMI = getDmiInfo()
	Log = NewStreamLogger(slog.LevelInfo)
}

type RunShell struct {
}

type RunSheller interface {
	Command(string, ...string) ([]byte, error)
	CommandWithContext(context.Context, string, ...string) ([]byte, error)
}

func (r *RunShell) Command(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeOut)
	defer cancel()
	return r.CommandWithContext(ctx, name, args...)
}

func (r *RunShell) CommandWithContext(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return stdout.Bytes(), fmt.Errorf("failed to start command: %v", err)
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if err := cmd.Process.Kill(); err != nil {
				log.Printf("kill process failed: %v", err)
			}
		case <-done:
		}
	}()
	defer close(done)
	if err := cmd.Wait(); err != nil {
		return stderr.Bytes(), fmt.Errorf("Command failed: %v", err)
	}

	return stdout.Bytes(), nil
}

func Unit2Human(f float64, u string, isHuman bool) (float64, string, error) {
	if f < 0 {
		return f, u, fmt.Errorf("negative value not surpported")
	}
	unitMap := map[string]int{"B": 0, "KB": 1, "MB": 2, "GB": 3, "TB": 4, "PB": 5, "EB": 6, "ZB": 7, "YB": 8}
	index, ok := unitMap[strings.ToUpper(u)]
	if !ok {
		return f, u, fmt.Errorf("unkown unit: %s", u)
	}
	base := 1024
	if isHuman {
		base = 1000
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	for f >= float64(base) && index < len(units)-1 {
		f /= float64(base)
		index += 1
	}
	return f, units[index], nil
}

func SplitAndTrim(s string, sep string) []string {
	if len(s) == 0 || len(sep) == 0 {
		return []string{}
	}
	ret := []string{}
	ss := strings.Split(s, sep)
	for _, v := range ss {
		temp := strings.TrimSpace(v)
		if len(temp) != 0 {
			ret = append(ret, temp)
		}
	}
	return ret
}

func StructToMap(s interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct")
	}
	t := v.Type()
	ret := make(map[string]interface{})
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			tag = field.Name
		}
		ret[tag] = v.Field(i).Interface()
	}
	return ret, nil
}

// InterfaceToString 安全地将 interface{} 转换为字符串
func InterfaceToString(value interface{}) string {
	if val, ok := value.(string); ok {
		return val
	}
	return fmt.Sprintf("%v", value)
}

func ReadFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return string(content), nil
}

func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// ReadLinesOffsetN reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
// n >= 0: at most n lines
// n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				ret = append(ret, strings.Trim(line, "\n"))
			}
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

func UniqueSlice[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return []T{}
	}
	seen := sync.Map{}
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if _, ok := seen.Load(item); !ok {
			seen.Store(item, struct{}{})
			result = append(result, item)
		}
	}
	return result
}

func PathExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func PathExistsWithContent(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return info.Size() > 4 && !info.IsDir()
}

func IsEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !IsEmptyValue(v.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

func ReadJSONFile(file string) map[string]interface{} {
	byteJSON, err := os.ReadFile(file)
	if err != nil {
		log.Fatalln("Read json file failed: ", err)
	}
	ret := JSONToMap(byteJSON)
	return ret
}

func JSONToMap(src []byte) map[string]interface{} {
	ret := make(map[string]interface{})
	err := json.Unmarshal(src, &ret)
	if err != nil {
		log.Fatalln(err)
	}
	return ret
}

type OrderedType interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | string
}

func FindMinAndMax[T OrderedType](nums []T) (minValue, maxValue T) {
	if len(nums) == 0 {
		return minValue, maxValue
	}
	minValue = nums[0]
	maxValue = nums[0]
	for _, num := range nums[1:] {
		if num < minValue {
			minValue = num
		}
		if num > maxValue {
			maxValue = num
		}
	}
	return minValue, maxValue
}

func StructSelectFieldOutput(v interface{}, selectFields []string, indent int) {
	separator := strings.Repeat("    ", indent)
	nums := 40 - indent*4
	val := reflect.ValueOf(v)
	key := reflect.TypeOf(v)
	for _, field := range selectFields {
		vv := val.FieldByName(field)
		kk, ok := key.FieldByName(field)
		if !ok {
			continue
		}
		tag := kk.Tag.Get("json")
		if IsEmptyValue(vv) {
			continue
		}
		switch vv.Kind() {
		case reflect.Struct:
			StructSelectFieldOutput(vv.Interface(), selectFields, indent+1)
		case reflect.Slice:
			sLen := vv.Len()
			for i := 0; i < sLen; i++ {
				elem := vv.Index(i).Interface()
				if reflect.TypeOf(elem).Kind() == reflect.Struct {
					StructSelectFieldOutput(elem, selectFields, indent+1)
				}
			}
		}
		fmt.Printf("%s%-*s: %v\n", separator, nums, strings.Split(tag, ",")[0], vv.Interface())
	}
}
