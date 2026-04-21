//go:build linux

package sysinfo

import (
	"os"
	"runtime"
	"strings"
)

// Get returns system info read from the Linux /proc filesystem.
func Get() Info {
	return Info{
		OS:      "Linux",
		Arch:    runtime.GOARCH,
		CPUInfo: readProcField("/proc/cpuinfo", "model name"),
		MemInfo: readProcField("/proc/meminfo", "MemTotal"),
	}
}

func readProcField(path, field string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "unavailable"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, field) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}
