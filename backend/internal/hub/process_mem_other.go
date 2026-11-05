//go:build !windows

package hub

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// processResidentBytes reads the resident set on Linux from /proc; elsewhere
// it falls back to the runtime's own accounting.
func processResidentBytes() uint64 {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/self/statm")
		if err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 2 {
				if pages, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					return pages * uint64(os.Getpagesize())
				}
			}
		}
	}
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Sys
}
