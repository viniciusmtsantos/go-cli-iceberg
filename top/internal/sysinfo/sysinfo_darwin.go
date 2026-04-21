//go:build darwin

package sysinfo

import (
	"os/exec"
	"runtime"
	"strings"
)

// Get returns system info using macOS sysctl.
func Get() Info {
	return Info{
		OS:      "macOS (Darwin)",
		Arch:    runtime.GOARCH,
		CPUInfo: sysctl("machdep.cpu.brand_string"),
		MemInfo: sysctl("hw.memsize"),
	}
}

func sysctl(key string) string {
	out, err := exec.Command("sysctl", "-n", key).Output()
	if err != nil {
		return "unavailable"
	}
	return strings.TrimSpace(string(out))
}
