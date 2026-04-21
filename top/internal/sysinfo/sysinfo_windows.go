//go:build windows

package sysinfo

import (
	"os/exec"
	"runtime"
	"strings"
)

// Get returns system info using Windows Management Instrumentation (WMIC).
func Get() Info {
	return Info{
		OS:      "Windows",
		Arch:    runtime.GOARCH,
		CPUInfo: wmicQuery("cpu", "Name"),
		MemInfo: wmicQuery("memorychip", "Capacity"),
	}
}

func wmicQuery(class, field string) string {
	out, err := exec.Command("wmic", class, "get", field).Output()
	if err != nil {
		return "unavailable"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 2 {
		return strings.TrimSpace(lines[1])
	}
	return "unknown"
}
