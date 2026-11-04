package hub

// Bootstrap state snapshot for the web client plus the polling loops.

import (
	"encoding/json"
	"time"
)

// BootstrapState is what the web client loads first.
type BootstrapState struct {
	Config    Config          `json:"config"`
	Profiles  []ProfileState  `json:"profiles"`
	Running   RunningState    `json:"running"`
	Groups    []Group         `json:"groups"`
	DelayMap  map[string]map[string]int32 `json:"delayMap"`
	Providers map[string]ProviderView     `json:"providers"`
	LiveCount int             `json:"liveCount"`
	Traffics  []TrafficPoint  `json:"traffics"`
	TotalTraffic Traffic      `json:"totalTraffic"`
	CoreVersion string        `json:"coreVersion"`
	CoreConnected bool        `json:"coreConnected"`
}

type Traffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

func (h *Hub) Bootstrap() BootstrapState {
	// Both of these take their own locks, so they are read before the hub lock
	// is held: a nested RLock deadlocks as soon as a writer is queued between
	// the two acquisitions, which hangs the bootstrap request for good.
	connected := h.CoreConnected()
	liveCount := h.requests.Count()
	h.mu.RLock()
	defer h.mu.RUnlock()
	return BootstrapState{
		Config:        h.cfg,
		Profiles:      h.profilesSnapshotLocked(),
		Running:       h.runningStateLocked(),
		Groups:        h.groups,
		DelayMap:      h.delayMap,
		Providers:     h.providers,
		LiveCount:     liveCount,
		Traffics:      h.traffics,
		CoreConnected: connected,
	}
}

func (h *Hub) profilesSnapshotLocked() []ProfileState {
	out := make([]ProfileState, 0, len(h.profiles))
	for _, p := range h.profiles {
		out = append(out, *p)
	}
	return out
}

func (h *Hub) runningStateLocked() RunningState {
	return RunningState{
		Running:     h.running,
		StartTime:   h.startTime,
		SystemProxy: h.systemProxy,
		Port:        h.cfg.PatchClashConfig.MixedPort,
		Mode:        h.cfg.PatchClashConfig.Mode,
	}
}

// StartTrafficPoller samples traffic once a second while clients listen.
func (h *Hub) StartTrafficPoller() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if !h.HasWebClients() {
				continue
			}
			h.requests.GC()
			h.mu.RLock()
			client := h.client
			onlyProxy := h.cfg.AppSetting.OnlyStatisticsProxy
			h.mu.RUnlock()
			var traffic Traffic
			var total Traffic
			totalOK := false
			if client != nil {
				raw, err := client.Call("traffic.current", map[string]any{"only-proxy": onlyProxy}, 0)
				if err == nil && json.Unmarshal(raw, &traffic) == nil {
					rawTotal, errTotal := client.Call("traffic.total", map[string]any{"only-proxy": onlyProxy}, 0)
					if errTotal == nil && json.Unmarshal(rawTotal, &total) == nil {
						totalOK = true
					}
				}
			}
			h.mu.Lock()
			if totalOK {
				h.totalTrafficCache = total
			} else {
				total = h.totalTrafficCache
			}
			h.traffics = append(h.traffics, TrafficPoint{Up: traffic.Up, Down: traffic.Down})
			if len(h.traffics) > 180 {
				h.traffics = h.traffics[len(h.traffics)-180:]
			}
			h.mu.Unlock()
			h.emitter("traffic", map[string]any{
				"up": traffic.Up, "down": traffic.Down,
				"totalUp": total.Up, "totalDown": total.Down,
			})
		}
	}()
}

// Memory returns the core's memory stats.
func (h *Hub) Memory() (MemoryStats, error) {
	var stats MemoryStats
	if h.client == nil {
		return stats, nil
	}
	raw, err := h.client.Call("core.memory", nil, 0)
	if err != nil {
		return stats, err
	}
	json.Unmarshal(raw, &stats)
	return stats, nil
}

// outboundIpSources mirrors lib/common/service_probe.dart ipInfoSources.
var outboundIpSources = []string{
	"https://www.cloudflare.com/cdn-cgi/trace",
	"https://api.ip.sb/geoip",
	"https://ipinfo.io/json",
	"https://ipwho.is",
	"https://get.geojs.io/v1/ip/geo.json",
	"https://api.country.is",
	"https://api.ipquery.io/?format=json",
	"https://ident.me/json",
}

// OutboundIP probes the exit IP via the core's probe.outboundIp.
func (h *Hub) OutboundIP() (json.RawMessage, error) {
	if h.client == nil {
		return nil, nil
	}
	return h.client.Call("probe.outboundIp", map[string]any{
		"proxy-name": "",
		"urls":       outboundIpSources,
		"timeout":    int64(15000),
	}, callLong)
}

// ServiceCheck probes service reachability.
func (h *Hub) ServiceCheck(names []string) (json.RawMessage, error) {
	if h.client == nil {
		return nil, nil
	}
	return h.client.Call("probe.services", map[string]any{"names": names}, callLong)
}

// refreshProviders syncs the external provider list.
func (h *Hub) refreshProviders() {
	if h.client == nil {
		return
	}
	raw, err := h.client.Call("provider.list", nil, 0)
	if err != nil {
		return
	}
	var list map[string]ProviderView
	if json.Unmarshal(raw, &list) == nil {
		h.mu.Lock()
		h.providers = list
		h.mu.Unlock()
		h.emitter("providers", list)
	}
}

func (h *Hub) Providers() map[string]ProviderView {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.providers
}

// UpdateProvider updates one external provider.
func (h *Hub) UpdateProvider(name string) error {
	if h.client == nil {
		return nil
	}
	_, err := h.client.Call("provider.update", map[string]any{"name": name}, callLong)
	h.refreshProviders()
	return err
}

// SideLoadProvider writes a local file into a provider.
func (h *Hub) SideLoadProvider(name, data string) error {
	if h.client == nil {
		return nil
	}
	_, err := h.client.Call("provider.sideLoad", map[string]any{"name": name, "data": data}, callLong)
	h.refreshProviders()
	return err
}

// UpdateGeo triggers a geo resource update.
func (h *Hub) UpdateGeo(geoType string) error {
	if h.client == nil {
		return nil
	}
	_, err := h.client.Call("geo.update", map[string]any{"type": geoType}, callLong)
	return err
}

// DnsQueries returns the cached DNS query feed.
func (h *Hub) DnsQueries() []DnsQueryLine {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]DnsQueryLine{}, h.dnsQueries...)
}

// Logs returns the cached log feed.
func (h *Hub) Logs() []LogLine {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]LogLine{}, h.logs...)
}

// ClearDnsCache clears the core DNS cache.
func (h *Hub) ClearDnsCache() error {
	if h.client == nil {
		return nil
	}
	_, err := h.client.Call("dns.clearCache", nil, 0)
	return err
}

// ClearTraffic resets the traffic statistics.
func (h *Hub) ClearTraffic() error {
	if h.client == nil {
		return nil
	}
	_, err := h.client.Call("traffic.reset", nil, 0)
	h.mu.Lock()
	h.traffics = []TrafficPoint{}
	h.totalTrafficCache = Traffic{}
	h.mu.Unlock()
	return err
}
