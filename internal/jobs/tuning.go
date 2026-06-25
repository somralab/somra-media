package jobs

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"
)

// QueueTuning holds background worker pool sizing for home-server hardware.
type QueueTuning struct {
	Workers int
	Buffer  int
}

const (
	defaultWorkers = 2
	defaultBuffer  = 16
)

// RecommendQueueTuning derives conservative worker counts from CPU cores.
// Targets 4c/8GB home hardware: 2 workers, moderate buffer; scales down on
// low-core hosts and up slightly on larger machines.
func RecommendQueueTuning(cpuCores int) QueueTuning {
	if cpuCores <= 0 {
		cpuCores = runtime.NumCPU()
	}
	t := QueueTuning{Workers: defaultWorkers, Buffer: defaultBuffer}
	switch {
	case cpuCores <= 2:
		t.Workers = 1
		t.Buffer = 8
	case cpuCores <= 4:
		t.Workers = 2
		t.Buffer = 16
	default:
		t.Workers = 3
		t.Buffer = 32
	}
	return t
}

// MemoryQueueConfigFromEnv returns queue sizing with optional SOMRA_JOB_* overrides.
func MemoryQueueConfigFromEnv(logger *slog.Logger) MemoryQueueConfig {
	tuning := RecommendQueueTuning(runtime.NumCPU())
	if v := os.Getenv("SOMRA_JOB_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			tuning.Workers = n
		}
	}
	if v := os.Getenv("SOMRA_JOB_BUFFER"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			tuning.Buffer = n
		}
	}
	cfg := MemoryQueueConfig{
		Workers: tuning.Workers,
		Buffer:  tuning.Buffer,
		Logger:  logger,
	}
	return cfg
}
