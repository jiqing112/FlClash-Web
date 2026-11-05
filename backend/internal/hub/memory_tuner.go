package hub

// Keeps the process working set tight. Go only returns heap to the OS at
// staggered intervals; with the core's own allocations on top the resident
// size creeps up, so the tuner nudges it back on idle.

import (
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

// gcPercent mirrors GOGC; the default 100 keeps a heap headroom the size of
// the live set, which on a busy proxy doubles the resident cost. A GOGC env
// var set before launch keeps precedence.
const defaultGCPercent = 50

// freeInterval is how often idle heap goes back to the OS.
const freeInterval = 90 * time.Second

// StartMemoryTuner tightens GC and schedules periodic memory release.
func StartMemoryTuner() {
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(defaultGCPercent)
	}
	go func() {
		ticker := time.NewTicker(freeInterval)
		defer ticker.Stop()
		for range ticker.C {
			debug.FreeOSMemory()
			runtime.GC()
		}
	}()
}
