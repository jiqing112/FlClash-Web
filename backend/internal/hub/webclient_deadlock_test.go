package hub

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// A connection opening while connections are being recorded used to deadlock:
// AddWebClient held the client registry lock while reading the tracker, and the
// tracker calls back into emitWeb (registry lock) while holding its own lock.
func TestAddWebClientWhileRecordingRequests(t *testing.T) {
	h := New("", "")

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			payload, _ := json.Marshal(TrackerInfo{ID: string(rune('a' + i%26)), Metadata: Metadata{Host: "example.com"}})
			h.requests.HandleTracker(payload)
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			id := h.AddWebClient(func(string, any) {})
			h.RemoveWebClient(id)
		}
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("AddWebClient deadlocked against the request tracker")
	}
	close(stop)
	wg.Wait()
}
