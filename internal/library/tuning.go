package library

import (
	"os"
	"runtime"
	"strconv"
)

const defaultScanProgressBatch = 25

// RecommendScanProgressBatch returns how often scan progress is flushed to DB/SSE.
// Larger batches reduce SQLite write pressure on big libraries.
func RecommendScanProgressBatch(cpuCores int) int {
	if cpuCores <= 0 {
		cpuCores = runtime.NumCPU()
	}
	switch {
	case cpuCores <= 2:
		return 10
	case cpuCores <= 4:
		return 25
	default:
		return 50
	}
}

// ScanProgressBatchFromEnv resolves batch size with optional SOMRA_SCAN_PROGRESS_BATCH override.
func ScanProgressBatchFromEnv() int {
	if v := os.Getenv("SOMRA_SCAN_PROGRESS_BATCH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return RecommendScanProgressBatch(runtime.NumCPU())
}
