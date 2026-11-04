package hub

// Port of the Dart requests provider: live connections from `request`
// events, per-connection totals from `requestTraffic`, and the recent feed.

import (
	"encoding/json"
	"sort"
	"sync"
	"time"
)

// TrackerInfo is one connection as the core announces it.
type TrackerInfo struct {
	ID        string          `json:"id"`
	Metadata  Metadata        `json:"metadata"`
	Upload    int64           `json:"upload"`
	Download  int64           `json:"download"`
	Start     string          `json:"start"`
	Chains    []string        `json:"chains"`
	Rule      string          `json:"rule"`
	RulePayload string        `json:"rulePayload"`
}

type Metadata struct {
	Network         string   `json:"network"`
	Type            string   `json:"type"`
	SourceIP        string   `json:"sourceIP"`
	DestinationIP   string   `json:"destinationIP"`
	SourcePort      string   `json:"sourcePort"`
	DestinationPort string   `json:"destinationPort"`
	Host            string   `json:"host"`
	Process         string   `json:"process"`
	ProcessPath     string   `json:"processPath"`
	RemoteDestination string `json:"remoteDestination"`
	SourceGeoIP     []string `json:"sourceGeoIP"`
	DestinationGeoIP []string `json:"destinationGeoIP"`
	DestinationIPASN string  `json:"destinationIPASN"`
	UID             uint32   `json:"uid"`
}

// RequestView is the client-facing request row: totals plus speed.
type RequestView struct {
	TrackerInfo
	UploadSpeed   int64 `json:"uploadSpeed,omitempty"`
	DownloadSpeed int64 `json:"downloadSpeed,omitempty"`
}

type liveRequest struct {
	info      TrackerInfo
	lastUp    int64
	lastDown  int64
	speedUp   int64
	speedDown int64
	missing   int
	lastSeen  time.Time
}

// RequestTracker maintains both feeds. The live feed mirrors what the core
// still holds; a request leaves it after several missed traffic reports.
type RequestTracker struct {
	mu       sync.Mutex
	live     map[string]*liveRequest
	recent   []RequestView
	order    []string
	countFn  func(int)
	maxTotal int
}

func NewRequestTracker(countFn func(int)) *RequestTracker {
	return &RequestTracker{
		live:     map[string]*liveRequest{},
		recent:   []RequestView{},
		maxTotal: maxRequests,
		countFn:  countFn,
	}
}

const maxRequests = 3000

// HandleTracker ingests a `request` event (the raw tracker JSON).
func (t *RequestTracker) HandleTracker(data json.RawMessage) {
	var info TrackerInfo
	if json.Unmarshal(data, &info) != nil || info.ID == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if existing, ok := t.live[info.ID]; ok {
		existing.info = info
		existing.missing = 0
		existing.lastSeen = time.Now()
	} else {
		t.live[info.ID] = &liveRequest{
			info:     info,
			lastSeen: time.Now(),
		}
		t.order = append(t.order, info.ID)
		t.pushRecentLocked(info)
		t.notifyLocked()
	}
}

// HandleTraffic ingests a `requestTraffic` batch, updating totals and speeds.
func (t *RequestTracker) HandleTraffic(updates []RequestTraffic) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, update := range updates {
		req, ok := t.live[update.ID]
		if !ok {
			continue
		}
		now := time.Now()
		elapsed := now.Sub(req.lastSeen).Seconds()
		if elapsed > 0.2 {
			req.speedUp = int64(float64(update.Upload-req.lastUp) / elapsed)
			req.speedDown = int64(float64(update.Download-req.lastDown) / elapsed)
		}
		req.lastUp = update.Upload
		req.lastDown = update.Download
		req.info.Upload = update.Upload
		req.info.Download = update.Download
		req.lastSeen = now
		for i := range t.recent {
			if t.recent[i].ID == update.ID {
				t.recent[i].Upload = update.Upload
				t.recent[i].Download = update.Download
				break
			}
		}
	}
}

func (t *RequestTracker) pushRecentLocked(info TrackerInfo) {
	view := RequestView{TrackerInfo: info}
	t.recent = append([]RequestView{view}, t.recent...)
	if len(t.recent) > t.maxTotal {
		t.recent = t.recent[:t.maxTotal]
	}
}

func (t *RequestTracker) notifyLocked() {
	if t.countFn != nil {
		t.countFn(len(t.live))
	}
}

// Live returns the live connections, oldest first.
func (t *RequestTracker) Live() []RequestView {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]RequestView, 0, len(t.live))
	for _, id := range t.order {
		if req, ok := t.live[id]; ok {
			view := RequestView{TrackerInfo: req.info}
			view.UploadSpeed = req.speedUp
			view.DownloadSpeed = req.speedDown
			out = append(out, view)
		}
	}
	return out
}

// Recent returns the recent feed, newest first.
func (t *RequestTracker) Recent() []RequestView {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]RequestView{}, t.recent...)
}

// Count returns the live connection count.
func (t *RequestTracker) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.live)
}

// Close marks a connection closed client-side; the core confirms it gone.
func (t *RequestTracker) Close(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.live, id)
	for i, orderID := range t.order {
		if orderID == id {
			t.order = append(t.order[:i], t.order[i+1:]...)
			break
		}
	}
	t.notifyLocked()
}

// CloseAll clears both feeds.
func (t *RequestTracker) CloseAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.live = map[string]*liveRequest{}
	t.order = nil
	t.recent = []RequestView{}
	t.notifyLocked()
}

// Reset clears state after a stop.
func (t *RequestTracker) Reset() {
	t.CloseAll()
}

// GC drops stale live entries the traffic pump no longer reports.
func (t *RequestTracker) GC() {
	t.mu.Lock()
	defer t.mu.Unlock()
	deadline := time.Now().Add(-30 * time.Second)
	for id, req := range t.live {
		if req.lastSeen.Before(deadline) {
			delete(t.live, id)
		}
	}
	// rebuild order
	order := make([]string, 0, len(t.live))
	for _, id := range t.order {
		if _, ok := t.live[id]; ok {
			order = append(order, id)
		}
	}
	t.order = order
	t.notifyLocked()
}

// LiveSorted sorts by start time for the table.
func LiveSorted(views []RequestView) []RequestView {
	out := append([]RequestView{}, views...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Start < out[j].Start
	})
	return out
}
