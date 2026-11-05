//go:build windows

package hub

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var psapi = windows.NewLazySystemDLL("psapi.dll")

var getProcessMemoryInfo = psapi.NewProc("GetProcessMemoryInfo")

type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

// processResidentBytes reads the current process working set on Windows.
func processResidentBytes() uint64 {
	var counters processMemoryCounters
	counters.CB = uint32(unsafe.Sizeof(counters))
	handle := windows.CurrentProcess()
	r1, _, _ := getProcessMemoryInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&counters)),
		uintptr(counters.CB),
	)
	if r1 == 0 {
		return 0
	}
	return uint64(counters.WorkingSetSize)
}
