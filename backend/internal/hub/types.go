package hub

import (
	"encoding/json"

	"flclashweb/backend/internal/models"
)

// Types here mirror the JSON shapes the Dart providers build from core
// results, so the web client can consume them unchanged.

type Config = models.Config

// ProfileState pairs a profile with its live info (counts, updating flag).
type ProfileState struct {
	models.Profile
	Updating bool `json:"updating"`
}

// ---- proxy data shapes ----

// Proxy is a node or group view from proxy.list.
type Proxy struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	ProviderName string `json:"provider-name,omitempty"`
	// group-only fields (from mihomo's MarshalJSON)
	Now     string          `json:"now,omitempty"`
	All     []string        `json:"all,omitempty"`
	Hidden  bool            `json:"hidden,omitempty"`
	TestUrl string          `json:"testUrl,omitempty"`
	Icon    string          `json:"icon,omitempty"`
	Udp     bool            `json:"udp,omitempty"`
	History json.RawMessage `json:"history,omitempty"`
	Extra   map[string]any  `json:"extra,omitempty"`
	Fixed   string          `json:"fixed,omitempty"`
	Order   json.RawMessage `json:"order,omitempty"`
}

type ProxiesData struct {
	Proxies map[string]Proxy   `json:"proxies"`
	All     []string           `json:"all"`
}

// Group is the client-facing group built server-side, matching the Dart Group
// model: `all` holds full proxy views, not names.
type Group struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	All     []Proxy  `json:"all"`
	Now     string   `json:"now,omitempty"`
	Hidden  bool     `json:"hidden,omitempty"`
	TestUrl string   `json:"testUrl,omitempty"`
	Icon    string   `json:"icon,omitempty"`
}

type GroupTypeValue struct{ value string }

var groupTypes = map[string]bool{
	"select": true, "selector": true,
	"url-test": true, "urltest": true,
	"fallback": true,
	"load-balance": true, "loadbalance": true,
	"relay": true,
}

// ---- feeds ----

type TrafficPoint struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

// LogLine matches mihomo's log.Event field names (capitalized, no tags).
type LogLine struct {
	LogLevel string `json:"LogLevel"`
	Payload  string `json:"Payload"`
}

type DnsQueryLine struct {
	Domain    string   `json:"domain"`
	Type      string   `json:"type"`
	Initiator string   `json:"initiator"`
	Upstream  string   `json:"upstream,omitempty"`
	Cached    bool     `json:"cached,omitempty"`
	Answers   []string `json:"answers"`
	Rcode     string   `json:"rcode,omitempty"`
	Error     string   `json:"error,omitempty"`
	Delay     int64    `json:"delay"`
	Time      string   `json:"time"`
}

type DelayEvent struct {
	Url   string `json:"url"`
	Name  string `json:"name"`
	Value int32  `json:"value"`
}

type RequestTraffic struct {
	ID       string `json:"id"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

type ProviderView struct {
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	VehicleType      string          `json:"vehicle-type"`
	Count            int             `json:"count"`
	Path             string          `json:"path"`
	UpdateAt         string          `json:"update-at"`
	SubscriptionInfo json.RawMessage `json:"subscription-info"`
}

type GeoUpdateStatus struct {
	Type     string `json:"type"`
	Updating bool   `json:"updating"`
	Skipped  bool   `json:"skipped,omitempty"`
	Error    string `json:"error,omitempty"`
}

// MemoryStats mirrors the core's core.memory result, plus the host process
// resident size so the UI can split 应用/内核 the way the desktop does.
type MemoryStats struct {
	RSS          uint64 `json:"rss"`
	HeapInuse    uint64 `json:"heap-inuse"`
	HeapIdle     uint64 `json:"heap-idle"`
	StackInuse   uint64 `json:"stack-inuse"`
	RuntimeOther uint64 `json:"runtime-other"`
	App          uint64 `json:"app,omitempty"`
}

const (
	maxLogs       = 600
	maxDnsQueries = 600
)

func deepCopyConfig(c Config) Config {
	raw, _ := json.Marshal(c)
	var out Config
	_ = json.Unmarshal(raw, &out)
	return out
}
