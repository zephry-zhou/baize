package server

import (
	"os"
	"path/filepath"
	"strings"
)

type Server struct {
	operatingSystem
	kernel
	product
}

func (s *Server) Result() {
	s.operatingSystem = osRelease()
	s.kernel = kernelInfo()
}

func readDMI(value string) string {
	dmiFlie := "/sys/devices/virtual/dmi/id/"
	data, err := os.ReadFile(filepath.Join(dmiFlie, value))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
