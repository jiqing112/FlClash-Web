// Package api serves the web client: REST endpoints over the hub and a
// WebSocket event stream, plus the embedded frontend bundle.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flclashweb/backend/internal/hub"
	"flclashweb/backend/internal/rpc"

	"github.com/gorilla/websocket"
)

// Server wires HTTP handlers to the hub.
type Server struct {
	hub    *hub.Hub
	client *rpc.Client
	mux    *http.ServeMux
	static http.Handler
}

func NewServer(h *hub.Hub, client *rpc.Client, static http.Handler) *Server {
	s := &Server{hub: h, client: client, mux: http.NewServeMux(), static: static}
	s.routes()
	return s
}

func (s *Server) routes() {
	m := s.mux
	m.HandleFunc("GET /api/bootstrap", s.bootstrap)
	m.HandleFunc("PUT /api/config", s.putConfig)
	m.HandleFunc("POST /api/config/apply", s.applyConfig)

	m.HandleFunc("GET /api/profiles", s.listProfiles)
	m.HandleFunc("POST /api/profiles", s.addProfile)
	m.HandleFunc("PUT /api/profiles/{id}", s.putProfile)
	m.HandleFunc("DELETE /api/profiles/{id}", s.deleteProfile)
	m.HandleFunc("POST /api/profiles/{id}/update", s.updateProfile)
	m.HandleFunc("POST /api/profiles/{id}/select", s.selectProfile)
	m.HandleFunc("GET /api/profiles/{id}/file", s.getProfileFile)
	m.HandleFunc("PUT /api/profiles/{id}/file", s.putProfileFile)

	m.HandleFunc("GET /api/proxies", s.getProxies)
	m.HandleFunc("POST /api/proxies/select", s.selectProxy)
	m.HandleFunc("POST /api/proxies/delay", s.testDelay)

	m.HandleFunc("GET /api/requests", s.getRequests)
	m.HandleFunc("DELETE /api/requests/{id}", s.closeRequest)
	m.HandleFunc("DELETE /api/requests", s.closeAllRequests)
	m.HandleFunc("GET /api/dns", s.getDns)
	m.HandleFunc("POST /api/dns/clear", s.clearDns)
	m.HandleFunc("GET /api/intranet", s.intranet)
	m.HandleFunc("GET /api/ip-quality/{ip}", s.ipQuality)
	m.HandleFunc("GET /api/logs", s.getLogs)
	m.HandleFunc("POST /api/traffic/reset", s.resetTraffic)

	m.HandleFunc("POST /api/core/start", s.coreStart)
	m.HandleFunc("POST /api/core/stop", s.coreStop)
	m.HandleFunc("POST /api/core/restart", s.coreRestart)
	m.HandleFunc("GET /api/core/memory", s.coreMemory)
	m.HandleFunc("POST /api/probe/ip", s.probeIP)
	m.HandleFunc("POST /api/probe/services", s.probeServices)
	m.HandleFunc("POST /api/geo/update", s.geoUpdate)

	m.HandleFunc("GET /api/providers", s.listProviders)
	m.HandleFunc("POST /api/providers/{name}/update", s.updateProvider)
	m.HandleFunc("POST /api/providers/{name}/sideload", s.sideloadProvider)

	m.HandleFunc("POST /api/tools/share-decode", s.shareDecode)
	m.HandleFunc("POST /api/tools/share-encode", s.shareEncode)

	m.HandleFunc("GET /api/debug/raw", s.debugRaw)
	m.HandleFunc("GET /api/ws", s.websocket)
	// An unmatched API path must answer JSON: falling through to the SPA shell
	// makes a typo look like a 200 with an HTML body the client cannot parse.
	m.HandleFunc("/api/", s.apiNotFound)
	m.Handle("/", s.static)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) apiNotFound(w http.ResponseWriter, r *http.Request) {
	writeErr(w, http.StatusNotFound, fmt.Errorf("unknown api endpoint: %s %s", r.Method, r.URL.Path))
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func decodeBody(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// ---- state ----

func (s *Server) bootstrap(w http.ResponseWriter, r *http.Request) {
	state := s.hub.Bootstrap()
	writeJSON(w, 200, state)
}

type configPatch struct {
	AppSetting      *map[string]any `json:"appSettingProps,omitempty"`
	Theme           *map[string]any `json:"themeProps,omitempty"`
	Network         *map[string]any `json:"networkProps,omitempty"`
	Vpn             *map[string]any `json:"vpnProps,omitempty"`
	ProxiesStyle    *map[string]any `json:"proxiesStyleProps,omitempty"`
	PatchClash      *map[string]any `json:"patchClashConfig,omitempty"`
	CurrentProfileID *int64         `json:"currentProfileId,omitempty"`
}

// putConfig merges a partial config section, persists it, and triggers the
// downstream core updates, like the Dart configProvider listener.
func (s *Server) putConfig(w http.ResponseWriter, r *http.Request) {
	var patch configPatch
	if err := decodeBody(r, &patch); err != nil {
		writeErr(w, 400, err)
		return
	}
	cfg := s.hub.UpdateConfig(func(cfg *hub.Config) {
		if patch.AppSetting != nil {
			mergeInto(&cfg.AppSetting, *patch.AppSetting)
		}
		if patch.Theme != nil {
			mergeInto(&cfg.ThemeSetting, *patch.Theme)
		}
		if patch.Network != nil {
			mergeInto(&cfg.NetworkSetting, *patch.Network)
		}
		if patch.Vpn != nil {
			mergeInto(&cfg.VpnSetting, *patch.Vpn)
		}
		if patch.ProxiesStyle != nil {
			mergeInto(&cfg.ProxiesStyle, *patch.ProxiesStyle)
		}
		if patch.PatchClash != nil {
			mergeInto(&cfg.PatchClashConfig, *patch.PatchClash)
		}
		if patch.CurrentProfileID != nil {
			cfg.CurrentProfileID = patch.CurrentProfileID
		}
	})

	// downstream effects
	if patch.ProxiesStyle != nil {
		s.hub.ProxiesStyleUpdate()
	}
	if patch.PatchClash != nil {
		if err := s.hub.UpdatePatch(); err != nil {
			writeJSON(w, 200, map[string]any{"config": cfg, "warning": err.Error()})
			return
		}
	}
	if patch.Network != nil || patch.Vpn != nil {
		s.hub.SyncSystemProxySetting()
	}
	writeJSON(w, 200, map[string]any{"config": cfg})
}

// applyConfig rebuilds the profile and pushes it (profile switch, overrides).
func (s *Server) applyConfig(w http.ResponseWriter, r *http.Request) {
	err := s.hub.ApplyProfile(true)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func mergeInto(target any, patch map[string]any) {
	raw, err := json.Marshal(target)
	if err != nil {
		return
	}
	var asMap map[string]any
	if json.Unmarshal(raw, &asMap) != nil {
		return
	}
	for key, value := range patch {
		asMap[key] = value
	}
	merged, err := json.Marshal(asMap)
	if err != nil {
		return
	}
	_ = json.Unmarshal(merged, target)
}

// ---- profiles ----

func (s *Server) listProfiles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.hub.Profiles())
}

func (s *Server) addProfile(w http.ResponseWriter, r *http.Request) {
	var input hub.AddProfileInput
	if err := decodeBody(r, &input); err != nil {
		writeErr(w, 400, err)
		return
	}
	state, err := s.hub.AddProfile(input)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, state)
}

func (s *Server) putProfile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	var patch map[string]any
	if err := decodeBody(r, &patch); err != nil {
		writeErr(w, 400, err)
		return
	}
	state, err := s.hub.EditProfile(id, patch)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, state)
}

func (s *Server) deleteProfile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.hub.RemoveProfile(id); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	state, err := s.hub.UpdateProfile(id)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, state)
}

func (s *Server) selectProfile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.hub.SelectProfile(id); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) getProfileFile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	content, err := s.hub.ProfileFile(id)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]string{"content": content})
}

func (s *Server) putProfileFile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	var body struct{ Content string `json:"content"` }
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.hub.SaveProfileFile(id, body.Content); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- proxies ----

func (s *Server) getProxies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"groups":   s.hub.Groups(),
		"delayMap": s.hub.DelayMap(),
	})
}

func (s *Server) selectProxy(w http.ResponseWriter, r *http.Request) {
	var body rpc.ChangeProxyParams
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.hub.SelectProxy(body.GroupName, body.ProxyName); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) testDelay(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Group string `json:"group"`
		Proxy string `json:"proxy"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, 400, err)
		return
	}
	var err error
	if body.Group != "" {
		err = s.hub.TestGroupDelay(body.Group)
	} else if body.Proxy != "" {
		err = s.hub.TestDelay(body.Proxy, 0)
	}
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- feeds ----

func (s *Server) getRequests(w http.ResponseWriter, r *http.Request) {
	feed := r.URL.Query().Get("feed")
	if feed == "recent" {
		writeJSON(w, 200, s.hub.RecentRequests())
		return
	}
	writeJSON(w, 200, s.hub.LiveRequests())
}

func (s *Server) closeRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.hub.CloseRequest(id); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) closeAllRequests(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.CloseAllRequests(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) getDns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.hub.DnsQueries())
}

func (s *Server) clearDns(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.ClearDnsCache(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) ipQuality(w http.ResponseWriter, r *http.Request) {
	ip := r.PathValue("ip")
	quality, ok := s.hub.IPQuality(ip)
	if !ok {
		writeErr(w, 404, fmt.Errorf("no quality answer"))
		return
	}
	writeJSON(w, 200, map[string]any{
		"ip":           quality.IP,
		"type":         quality.Type,
		"organization": quality.Organization,
		"asn":          quality.ASN,
		"isProxy":      quality.IsProxy,
		"isVpn":        quality.IsVpn,
		"isTor":        quality.IsTor,
		"isAbuser":     quality.IsAbuser,
		"level":        quality.Level(),
	})
}

func (s *Server) intranet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"ip": s.hub.IntranetIP()})
}

func (s *Server) getLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.hub.Logs())
}

func (s *Server) resetTraffic(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.ClearTraffic(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- core control ----

func (s *Server) coreStart(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.StartListener(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, s.hub.RunningState())
}

func (s *Server) coreStop(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.StopListener(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, s.hub.RunningState())
}

func (s *Server) coreRestart(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.RestartCore(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, s.hub.RunningState())
}

func (s *Server) coreMemory(w http.ResponseWriter, r *http.Request) {
	stats, err := s.hub.Memory()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, stats)
}

func (s *Server) probeIP(w http.ResponseWriter, r *http.Request) {
	raw, err := s.hub.OutboundIP()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (s *Server) probeServices(w http.ResponseWriter, r *http.Request) {
	var body struct{ Names []string `json:"names"` }
	_ = decodeBody(r, &body)
	raw, err := s.hub.ServiceCheck(body.Names)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (s *Server) geoUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct{ Type string `json:"type"` }
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.hub.UpdateGeo(body.Type); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- providers ----

func (s *Server) listProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.hub.Providers())
}

func (s *Server) updateProvider(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.hub.UpdateProvider(name); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) sideloadProvider(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var body struct{ Data string `json:"data"` }
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.hub.SideLoadProvider(name, body.Data); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- tools ----

func (s *Server) shareDecode(w http.ResponseWriter, r *http.Request) {
	var body struct{ Lines []string `json:"lines"` }
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, 400, err)
		return
	}
	raw, err := s.hub.DecodeShareLinks(body.Lines)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (s *Server) shareEncode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Proxies []map[string]any `json:"proxies"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, 400, err)
		return
	}
	raw, err := s.hub.EncodeShareLinks(body.Proxies)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

// ---- websocket ----

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 1 << 16,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) debugRaw(w http.ResponseWriter, r *http.Request) {
	raw, err := s.client.Call("proxy.list", nil, 0)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (s *Server) websocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(1 << 20)
	_ = conn.SetReadDeadline(time.Now().Add(pingInterval + 30*time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pingInterval + 30*time.Second))
	})

	events := make(chan hubEvent, 512)
	id := s.hub.AddWebClient(func(eventType string, data any) {
		select {
		case events <- hubEvent{Type: eventType, Data: data}:
		default: // drop on a slow client
		}
	})
	defer s.hub.RemoveWebClient(id)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	writeEvent := func(event hubEvent) bool {
		data, err := json.Marshal(event)
		if err != nil {
			return true
		}
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteMessage(websocket.TextMessage, data) == nil
	}

	for {
		select {
		case <-done:
			conn.Close()
			return
		case event := <-events:
			if !writeEvent(event) {
				conn.Close()
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				conn.Close()
				return
			}
		}
	}
}

const pingInterval = 25 * time.Second

type hubEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

var _ = fmt.Sprintf
var _ = strings.TrimSpace
