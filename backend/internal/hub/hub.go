// Package hub holds the web host's business logic: the state the Flutter app
// kept in Riverpod providers, the profile build pipeline from lib/common/task.dart,
// and the core control flow from lib/providers/actions/setup.dart.
package hub

import (
	"flclashweb/backend/internal/models"
	"encoding/json"
	"sync"

	"flclashweb/backend/internal/rpc"
	"flclashweb/backend/internal/store"
)

// Broadcast delivers an app event to every connected web client.
type Broadcast func(eventType string, data any)

type Hub struct {
	mu       sync.RWMutex
	cfg      Config
	profiles []*ProfileState
	client   *rpc.Client
	homeDir  string
	dataDir  string
	store    *store.Store
	emitter  Broadcast

	running     bool
	startTime   int64 // ms
	systemProxy bool

	lastConfigMd5 string
	lastCounts    Counts

	// proxies state
	proxiesData   *ProxiesData
	delayMap      map[string]map[string]int32 // test-url -> name -> delay
	groups        []Group
	totalTrafficCache Traffic

	// feeds
	traffics   []TrafficPoint
	logs       []LogLine
	requests   *RequestTracker
	dnsQueries []DnsQueryLine
	providers  map[string]ProviderView
	geoUpdating map[string]bool

	// web clients
	webClientsMu sync.Mutex
	webClients   map[int]func(eventType string, data any)
	webClientSeq int
	coreConnected bool
}

// New builds the hub. dataDir holds profiles and the running config; the core
// home dir is where mihomo keeps its own files.
func New(homeDir, dataDir string) *Hub {
	h := &Hub{
		homeDir:     homeDir,
		dataDir:     dataDir,
		cfg:         models.DefaultConfig(),
		delayMap:    map[string]map[string]int32{},
		traffics:    []TrafficPoint{},
		logs:        []LogLine{},
		dnsQueries:  []DnsQueryLine{},
		providers:   map[string]ProviderView{},
		geoUpdating: map[string]bool{},
		webClients:  map[int]func(string, any){},
	}
	h.requests = NewRequestTracker(h.onRequestCount)
	// The hub is the only producer of app events, so the hook fans out to the
	// browser sockets: traffic ticks, running-state changes, refreshed groups,
	// providers and errors all have to land there.
	h.emitter = h.emitWeb
	return h
}

func (h *Hub) SetClient(client *rpc.Client) {
	h.mu.Lock()
	h.client = client
	h.mu.Unlock()
	h.emitCoreStatus()
}

// StartEventPump subscribes to core events and maintains state from them.
func (h *Hub) StartEventPump() {
	h.mu.RLock()
	client := h.client
	h.mu.RUnlock()
	if client == nil {
		return
	}
	client.Subscribe(func(event rpc.Event) {
		h.handleEvent(event)
	})
}

func (h *Hub) handleEvent(event rpc.Event) {
	switch event.Type {
	case "log":
		var line LogLine
		if json.Unmarshal(event.Data, &line) == nil {
			h.mu.Lock()
			h.logs = append(h.logs, line)
			if len(h.logs) > maxLogs {
				h.logs = h.logs[len(h.logs)-maxLogs:]
			}
			h.mu.Unlock()
			h.emitWeb("log", line)
		}
	case "delay":
		var delay DelayEvent
		if json.Unmarshal(event.Data, &delay) == nil {
			h.mu.Lock()
			url := delay.Url
			if url == "" {
				url = h.cfg.PatchClashConfig.TestUrl
			}
			if h.delayMap[url] == nil {
				h.delayMap[url] = map[string]int32{}
			}
			h.delayMap[url][delay.Name] = delay.Value
			groups := h.computeGroupsLocked()
			h.groups = groups
			h.mu.Unlock()
			h.emitWeb("delay", delay)
			h.emitWeb("groups", groups)
		}
	case "request":
		var info TrackerInfo
		if json.Unmarshal(event.Data, &info) == nil {
			h.requests.HandleTracker(event.Data)
			h.emitWeb("request", info)
		}
	case "requestTraffic":
		var updates []RequestTraffic
		if json.Unmarshal(event.Data, &updates) == nil {
			h.requests.HandleTraffic(updates)
			h.emitWeb("requestTraffic", updates)
		}
	case "dns":
		var line DnsQueryLine
		if json.Unmarshal(event.Data, &line) == nil {
			h.mu.Lock()
			h.dnsQueries = append(h.dnsQueries, line)
			if len(h.dnsQueries) > maxDnsQueries {
				h.dnsQueries = h.dnsQueries[len(h.dnsQueries)-maxDnsQueries:]
			}
			h.mu.Unlock()
			h.emitWeb("dns", line)
		}
	case "loaded":
		var name string
		json.Unmarshal(event.Data, &name)
		h.refreshProxies()
		h.refreshProviders()
		h.emitWeb("loaded", name)
	case "geoUpdate":
		var status GeoUpdateStatus
		if json.Unmarshal(event.Data, &status) == nil {
			h.mu.Lock()
			if status.Updating {
				h.geoUpdating[status.Type] = true
			} else {
				delete(h.geoUpdating, status.Type)
			}
			h.mu.Unlock()
			h.emitWeb("geoUpdate", status)
		}
	case "dialerLoop":
		var name string
		json.Unmarshal(event.Data, &name)
		h.emitWeb("dialerLoop", name)
	case "routeChanged":
		h.emitWeb("routeChanged", event.Data)
	}
}

func (h *Hub) onRequestCount(count int) {
	h.emitWeb("requestCount", count)
}

// EmitCoreDisconnected pushes a disconnected core status.
func (h *Hub) EmitCoreDisconnected() { h.emitCoreStatus() }

// CoreConnected reports whether the core transport is up.
func (h *Hub) CoreConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.client == nil {
		return false
	}
	return h.client.Connected()
}

// emitCoreStatus pushes the core transport state to web clients.
func (h *Hub) emitCoreStatus() {
	// CoreConnected takes the read lock itself; taking it here as well nests
	// two RLocks and parks this goroutine forever once a writer queues between
	// them, which wedges every later request on the same mutex.
	connected := h.CoreConnected()
	h.mu.Lock()
	h.coreConnected = connected
	h.mu.Unlock()
	h.emitWeb("coreStatus", map[string]bool{"connected": connected})
}

// emitWeb fans an app event out to every connected web client.
func (h *Hub) emitWeb(eventType string, data any) {
	h.webClientsMu.Lock()
	subs := make([]func(string, any), 0, len(h.webClients))
	for _, sub := range h.webClients {
		subs = append(subs, sub)
	}
	h.webClientsMu.Unlock()
	for _, sub := range subs {
		sub(eventType, data)
	}
}

// AddWebClient registers a client sink and immediately replays the existing
// feeds, so a freshly opened page continues from the buffers instead of
// starting blank; returns the unsubscribe id.
func (h *Hub) AddWebClient(sub func(eventType string, data any)) int {
	h.webClientsMu.Lock()
	h.webClientSeq++
	h.webClients[h.webClientSeq] = sub
	id := h.webClientSeq
	h.webClientsMu.Unlock()

	// Snapshot the feeds outside the client registry: the tracker calls back
	// into emitWeb (and so takes webClientsMu) while it holds its own lock, so
	// reading the tracker with webClientsMu held inverts that order and wedges
	// both goroutines the first time a connection opens mid-traffic.
	h.mu.RLock()
	logs := append([]LogLine{}, h.logs...)
	dns := append([]DnsQueryLine{}, h.dnsQueries...)
	h.mu.RUnlock()
	snapshot := map[string]any{
		"logs":      logs,
		"dns":       dns,
		"live":      h.requests.Live(),
		"recent":    h.requests.Recent(),
		"liveCount": h.requests.Count(),
	}
	go sub("init", snapshot)
	return id
}

func (h *Hub) RemoveWebClient(id int) {
	h.webClientsMu.Lock()
	defer h.webClientsMu.Unlock()
	delete(h.webClients, id)
}

func (h *Hub) HasWebClients() bool {
	h.webClientsMu.Lock()
	defer h.webClientsMu.Unlock()
	return len(h.webClients) > 0
}

// Config returns the current config snapshot.
func (h *Hub) Config() Config {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cfg
}

// UpdateConfig applies a mutation to the config, persists and broadcasts it.
func (h *Hub) UpdateConfig(mutate func(*Config)) Config {
	h.mu.Lock()
	h.cfg = deepCopyConfig(h.cfg)
	mutate(&h.cfg)
	cfg := h.cfg
	h.mu.Unlock()
	h.persistConfig()
	h.emitWeb("config", cfg)
	return cfg
}

func (h *Hub) persistConfig() {
	h.mu.RLock()
	cfg := h.cfg
	h.mu.RUnlock()
	if h.store != nil {
		if data, err := json.Marshal(cfg); err == nil {
			_ = h.store.Set("config", data)
		}
	}
}
