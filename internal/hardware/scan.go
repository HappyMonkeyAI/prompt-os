package hardware

import (
	"os/exec"
	"runtime"
	"strings"
)

type HardwareInfo struct {
	CPU   string
	GPU   string
	RAM   string
	Disk  string
	Arch  string
	Cores int
}

func Scan() HardwareInfo {
	info := HardwareInfo{
		Arch:  runtime.GOARCH,
		Cores: runtime.NumCPU(),
	}

	// CPU (lscpu or fallback)
	if out, err := exec.Command("lscpu").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "Model name:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.CPU = strings.TrimSpace(parts[1])
				}
				break
			}
		}
	}
	if info.CPU == "" {
		info.CPU = "unknown"
	}

	// GPU (lspci stub)
	if out, err := exec.Command("lspci").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(strings.ToLower(line), "vga") || strings.Contains(strings.ToLower(line), "3d") {
				info.GPU = strings.TrimSpace(line)
				break
			}
		}
	}
	if info.GPU == "" {
		info.GPU = "none/unknown"
	}

	// RAM (free -h)
	if out, err := exec.Command("free", "-h").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 1 {
				info.RAM = fields[1]
			}
		}
	}
	if info.RAM == "" {
		info.RAM = "unknown"
	}

	// Disk (df -h /)
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 1 {
				info.Disk = fields[1] + " used"
			}
		}
	}
	if info.Disk == "" {
		info.Disk = "unknown"
	}

	return info
}