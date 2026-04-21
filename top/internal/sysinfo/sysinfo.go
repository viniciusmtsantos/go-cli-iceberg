package sysinfo

// Info holds platform-specific system information.
type Info struct {
	OS      string
	Arch    string
	CPUInfo string
	MemInfo string
}
