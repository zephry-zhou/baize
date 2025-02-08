package server

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zephry-zhou/baize/internal"
)

type operatingSystem struct {
	PrettyName   string `json:"Pretty Name,omitempty"`
	Distr        string `json:"Distribution,omitempty"`
	Releases     string `json:"Releases,omitempty"`
	CodeName     string `json:"Codename,omitempty"`
	MinorVersion string `json:"Minor Version,omitempty"`
	IDLike       string `json:"ID Like,omitempty"`
}

type kernel struct {
	KernelName    string `json:"Kernel,omitempty"`
	KernelRelease string `json:"Kernel Release,omitempty"`
	KernelVersion string `json:"Kernel Version,omitempty"`
	HostName      string `json:"Host Name,omitempty"`
}

func kernelInfo() kernel {
	procKern := "/proc/sys/kernel"
	ret := kernel{}
	if release, err := internal.ReadFile(filepath.Join(procKern, "osrelease")); err == nil {
		ret.KernelRelease = strings.TrimSpace(release)
	}
	if version, err := internal.ReadFile(filepath.Join(procKern, "version")); err == nil {
		ret.KernelVersion = strings.TrimSpace(version)
	}
	if name, err := internal.ReadFile(filepath.Join(procKern, "hostname")); err == nil {
		ret.HostName = strings.TrimSpace(name)
	}
	if ostype, err := internal.ReadFile(filepath.Join(procKern, "ostype")); err == nil {
		ret.KernelName = strings.TrimSpace(ostype)
	}
	return ret
}

func osRelease() operatingSystem {
	ret := operatingSystem{}
	lines, err := internal.ReadLines("/etc/os-release")
	if err != nil {
		internal.Log.Error("read file /etc/os-release error:", err)
		return ret
	}
	for _, line := range lines {
		fields := internal.SplitAndTrim(line, "=")
		if len(fields) != 2 {
			continue
		}
		key, value := fields[0], strings.ReplaceAll(fields[1], "\"", "")
		switch key {
		case "PRETTY_NAME":
			ret.PrettyName = value
		case "NAME":
			ret.Distr = value
		case "VERSION_ID":
			ret.Releases = value
		case "VERSION_CODENAME":
			ret.CodeName = value
		case "ID_LIKE":
			ret.IDLike = value
		}
	}
	ret.MinorVersion = minorVersion(ret.PrettyName)
	return ret
}

func minorVersion(distr string) string {
	ret := "Unknown"
	distr = strings.ToLower(distr)
	var (
		reUbuntu = regexp.MustCompile(`[\( ]([\d\.]+)`)
		reCentOS = regexp.MustCompile(`^CentOS( Linux)? release ([\d\.]+)`)
		reRocky  = regexp.MustCompile(`^Rocky Linux release ([\d\.]+)`)
		reRedHat = regexp.MustCompile(`[\( ]([\d\.]+)`)
	)
	switch {
	case strings.HasPrefix(distr, "debian"):
		if data, err := internal.ReadFile("/etc/debian_version"); err == nil {
			ret = strings.TrimSpace(string(data))
		}
	case strings.HasPrefix(distr, "ubuntu"):
		if m := reUbuntu.FindStringSubmatch(distr); m != nil {
			ret = m[1]
		}
	case strings.HasPrefix(distr, "centos"):
		if data, err := internal.ReadFile("/etc/centos-release"); err == nil && data != "" {
			if m := reCentOS.FindStringSubmatch(string(data)); m != nil {
				ret = m[2]
			}
		}
	case strings.HasPrefix(distr, "rhel"):
		if data, err := internal.ReadFile("/etc/redhat-release"); err == nil && data != "" {
			if m := reRedHat.FindStringSubmatch(string(data)); m != nil {
				ret = m[1]
			}
		}
	case strings.HasPrefix(distr, "rocky"):
		if data, err := internal.ReadFile("/etc/rocky-release"); err == nil && data != "" {
			if m := reRocky.FindStringSubmatch(string(data)); m != nil {
				ret = m[1]
			}
		}
	}
	return ret
}
