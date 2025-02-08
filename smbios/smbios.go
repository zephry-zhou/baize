package smbios

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

const (
	smbios3HeaderSize = 0x18
	smbios2HeaderSize = 0x1f
)

var (
	anchor32 = []byte("_SM_")
	anchor64 = []byte("_SM3_")
)

func entryBase() (int64, int64, error) {
	base, size, err := EntryFromEFI()
	if err != nil {
		base, size, err = EntryFromLegacy()
		if err != nil {
			return 0, 0, err
		}
	}
	return base, size, err
}

func EntryFromEFI() (int64, int64, error) {
	return entryFromEFI()
}

func entryFromEFI() (int64, int64, error) {
	file, err := os.Open("/sys/firmware/efi/systable")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	const (
		smbios3 = "SMBIOS3="
		smbios  = "SMBIOS="
	)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		start := ""
		size := int64(0)
		if strings.HasPrefix(line, smbios3) {
			start = strings.TrimPrefix(line, smbios3)
			size = smbios3HeaderSize
		}
		if strings.HasPrefix(line, smbios) {
			start = strings.TrimPrefix(line, smbios)
			size = smbios2HeaderSize
		}
		if start == "" {
			continue
		}

		base, err := strconv.ParseInt(start, 0, 63)
		if err != nil {
			continue
		}
		return base, size, nil
	}
	if err := scanner.Err(); err != nil {
		internal.Log.Error("error while reading EFI systab:  ", err)
	}
	return 0, 0, fmt.Errorf("invalid /sys/firmware/efi/systab file")
}

func EntryFromLegacy() (int64, int64, error) {
	file, err := os.Open("/dev/mem")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	return getMemBase(file, 0xf0000, 0x100000)
}

func getMemBase(file io.ReaderAt, start, end int64) (int64, int64, error) {
	b := make([]byte, 5)
	for base := start; base < end-5; base++ {
		if _, err := io.ReadFull(io.NewSectionReader(file, base, 5), b); err != nil {
			return 0, 0, err
		}
		if bytes.Equal(b[:4], anchor32) {

		}
	}
}
