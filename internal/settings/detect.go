package settings

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// PathValidation describes read/write access for a filesystem path.
type PathValidation struct {
	Path     string `json:"path"`
	Readable bool   `json:"readable"`
	Writable bool   `json:"writable"`
}

// SystemProfile is the hardware + storage snapshot returned by detection.
type SystemProfile struct {
	CPUCores    int              `json:"cpuCores"`
	MemoryBytes int64            `json:"memoryBytes"`
	GPUPresent  bool             `json:"gpuPresent"`
	Paths       []PathValidation `json:"paths"`
}

// DetectSystem gathers CPU, memory, GPU presence, and validates paths.
func DetectSystem(paths []string) SystemProfile {
	profile := SystemProfile{
		CPUCores:    runtime.NumCPU(),
		MemoryBytes: detectMemoryBytes(),
		GPUPresent:  detectGPU(),
	}
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		profile.Paths = append(profile.Paths, validatePath(p))
	}
	return profile
}

// ValidatePaths checks read/write access for each path.
func ValidatePaths(paths []string) []PathValidation {
	out := make([]PathValidation, 0, len(paths))
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		out = append(out, validatePath(p))
	}
	return out
}

func validatePath(path string) PathValidation {
	info, err := os.Stat(path)
	readable := err == nil
	writable := false
	if readable {
		writable = info.IsDir() && isWritableDir(path)
	} else {
		// Parent may be writable for mkdir-on-first-use paths.
		writable = isWritableDir(path) || isWritableParent(path)
	}
	return PathValidation{Path: path, Readable: readable, Writable: writable}
}

func isWritableDir(path string) bool {
	f, err := os.CreateTemp(path, ".somra-write-test-*")
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return true
}

func isWritableParent(path string) bool {
	parent := path
	for parent != "" && parent != "/" && parent != "." {
		if info, err := os.Stat(parent); err == nil && info.IsDir() {
			return isWritableDir(parent)
		}
		parent = strings.TrimSuffix(parent, "/")
		if idx := strings.LastIndex(parent, "/"); idx >= 0 {
			parent = parent[:idx]
		} else {
			break
		}
	}
	return false
}

func detectMemoryBytes() int64 {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/meminfo")
		if err != nil {
			return 0
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if kb, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						return kb * 1024
					}
				}
			}
		}
	case "darwin":
		out, err := exec.CommandContext(context.Background(), "sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			if bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil {
				return bytes
			}
		}
	}
	return 0
}

func detectGPU() bool {
	switch runtime.GOOS {
	case "linux":
		for _, p := range []string{"/dev/dri/renderD128", "/dev/nvidia0"} {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	case "darwin":
		out, err := exec.CommandContext(context.Background(), "system_profiler", "SPDisplaysDataType").Output()
		if err == nil && !strings.Contains(string(out), "Chipset Model: Apple") {
			// Discrete GPU or external GPU likely present when not only Apple Silicon iGPU.
			if strings.Contains(strings.ToLower(string(out)), "nvidia") ||
				strings.Contains(strings.ToLower(string(out)), "amd") ||
				strings.Contains(strings.ToLower(string(out)), "radeon") {
				return true
			}
		}
	}
	return false
}
