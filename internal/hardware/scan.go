package hardware

import (
	"os/exec"
	"runtime"
	"strings"
)

type BlockDevice struct {
	Name      string
	Path      string
	Size      string
	Model     string
	Type      string
	Removable bool
	ReadOnly  bool
	Transport string
}

type HardwareInfo struct {
	CPU   string
	GPU   string
	RAM   string
	Disk  string
	Disks []BlockDevice
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

	// Disk (df -h /) — correct column for "used"
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 2 {
				info.Disk = fields[2] + " used"
			}
		}
	}
	if info.Disk == "" {
		info.Disk = "unknown"
	}
	info.Disks = parseBlockDevices(commandOutput("lsblk", "-dnP", "-o", "NAME,SIZE,MODEL,TYPE,RM,RO,TRAN"))

	return info
}

func commandOutput(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func parseBlockDevices(out string) []BlockDevice {
	var devices []BlockDevice
	for _, line := range strings.Split(out, "\n") {
		attrs := parseKeyValueLine(line)
		if attrs["TYPE"] != "disk" || attrs["NAME"] == "" {
			continue
		}
		name := attrs["NAME"]
		devices = append(devices, BlockDevice{
			Name:      name,
			Path:      "/dev/" + name,
			Size:      attrs["SIZE"],
			Model:     attrs["MODEL"],
			Type:      attrs["TYPE"],
			Removable: attrs["RM"] == "1",
			ReadOnly:  attrs["RO"] == "1",
			Transport: attrs["TRAN"],
		})
	}
	return devices
}

func parseKeyValueLine(line string) map[string]string {
	attrs := make(map[string]string)
	for i := 0; i < len(line); {
		for i < len(line) && line[i] == ' ' {
			i++
		}
		start := i
		for i < len(line) && line[i] != '=' {
			i++
		}
		if i >= len(line) {
			continue
		}
		key := line[start:i]
		i++
		if i >= len(line) || line[i] != '"' {
			continue
		}
		i++
		valueStart := i
		for i < len(line) && line[i] != '"' {
			i++
		}
		attrs[key] = line[valueStart:i]
		if i < len(line) {
			i++
		}
	}
	return attrs
}
