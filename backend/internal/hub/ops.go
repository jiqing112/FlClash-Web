package hub

// Remaining hub operations the REST layer calls.

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"flclashweb/backend/internal/models"
)

// EditProfile applies a partial profile patch (JSON keys of the Profile model).
func (h *Hub) EditProfile(id int64, patch map[string]any) (*ProfileState, error) {
	h.mu.Lock()
	var target *ProfileState
	for _, p := range h.profiles {
		if p.ID == id {
			target = p
			break
		}
	}
	if target == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("profile not found")
	}
	raw, err := json.Marshal(target.Profile)
	if err != nil {
		h.mu.Unlock()
		return nil, err
	}
	var asMap map[string]any
	if err := json.Unmarshal(raw, &asMap); err != nil {
		h.mu.Unlock()
		return nil, err
	}
	for key, value := range patch {
		asMap[key] = value
	}
	merged, err := json.Marshal(asMap)
	if err != nil {
		h.mu.Unlock()
		return nil, err
	}
	var updated models.Profile
	if err := json.Unmarshal(merged, &updated); err != nil {
		h.mu.Unlock()
		return nil, err
	}
	*target = ProfileState{Profile: updated}
	h.saveProfile(target)
	state := *target
	h.mu.Unlock()
	h.emitWeb("profiles", h.Profiles())
	return &state, nil
}

// RemoveProfile deletes a profile and clears it if current. The profile
// file and its provider cache go with it, matching the original's orphan
// cleanup (shakingProfileTask).
func (h *Hub) RemoveProfile(id int64) error {
	h.UpdateConfig(func(cfg *Config) {
		if cfg.CurrentProfileID != nil && *cfg.CurrentProfileID == id {
			cfg.CurrentProfileID = nil
		}
	})
	h.DeleteProfile(id)
	h.removeProfileFiles(id)
	h.emitWeb("profiles", h.Profiles())
	// the core keeps serving the deleted profile until an apply pushes the
	// empty one, exactly like the original's applyProfile with no profile
	if h.Config().CurrentProfileID == nil {
		return h.ApplyProfile(true)
	}
	return nil
}

// removeProfileFiles deletes the profile yaml and the per-profile provider
// directories; a missing file is fine.
func (h *Hub) removeProfileFiles(id int64) {
	_ = os.Remove(h.profilePath(id))
	_ = os.RemoveAll(filepath.Join(h.profilesDir(), providersDirectoryName, fmt.Sprintf("%d", id)))
	_ = os.Remove(filepath.Join(h.profilesDir(), fmt.Sprintf(".base-%d.yaml", id)))
}

// SelectProfile switches the current profile and reapplies.
func (h *Hub) SelectProfile(id int64) error {
	if h.Profile(id) == nil {
		return fmt.Errorf("profile not found")
	}
	h.UpdateConfig(func(cfg *Config) {
		cfg.CurrentProfileID = &id
	})
	return h.ApplyProfile(true)
}

// ProfileFile reads the profile yaml.
func (h *Hub) ProfileFile(id int64) (string, error) {
	data, err := os.ReadFile(h.profilePath(id))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveProfileFile validates and stores profile yaml, then reapplies.
func (h *Hub) SaveProfileFile(id int64, content string) error {
	profile := h.Profile(id)
	if profile == nil {
		return fmt.Errorf("profile not found")
	}
	temp := h.profilePath(id) + ".tmp"
	if err := os.WriteFile(temp, []byte(content), 0o644); err != nil {
		return err
	}
	params := map[string]any{"path": temp}
	raw, err := h.client.Call("config.validate", params, 0)
	if err != nil {
		os.Remove(temp)
		return err
	}
	if raw != nil && string(raw) != "null" && len(strings.TrimSpace(string(raw))) > 2 {
		os.Remove(temp)
		var message string
		if json.Unmarshal(raw, &message) == nil && message != "" {
			return fmt.Errorf("%s", message)
		}
	}
	if err := os.Rename(temp, h.profilePath(id)); err != nil {
		return err
	}
	profile.LastUpdateDate = timeNowMillis()
	h.PutProfile(*profile)
	return h.ApplyProfile(true)
}

// DelayMap returns the delay state snapshot.
func (h *Hub) DelayMap() map[string]map[string]int32 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := map[string]map[string]int32{}
	for url, entries := range h.delayMap {
		copied := map[string]int32{}
		for name, delay := range entries {
			copied[name] = delay
		}
		out[url] = copied
	}
	return out
}

// LiveRequests returns the live feed.
func (h *Hub) LiveRequests() []RequestView { return h.requests.Live() }

// RecentRequests returns the recent feed.
func (h *Hub) RecentRequests() []RequestView { return h.requests.Recent() }

// CloseRequest closes one connection.
func (h *Hub) CloseRequest(id string) error {
	h.requests.Close(id)
	if h.client != nil {
		_, err := h.client.Call("connection.close", map[string]any{"id": id}, 0)
		return err
	}
	return nil
}

// CloseAllRequests closes every connection.
func (h *Hub) CloseAllRequests() error {
	h.requests.CloseAll()
	if h.client != nil {
		_, err := h.client.Call("connection.closeAll", nil, 0)
		return err
	}
	return nil
}

// SyncSystemProxySetting applies or clears the system proxy after a change.
func (h *Hub) SyncSystemProxySetting() {
	h.mu.RLock()
	running := h.running
	h.mu.RUnlock()
	if running {
		h.syncSystemProxy()
	} else {
		h.clearSystemProxy()
	}
	h.emitWeb("running", h.RunningState())
}

// DecodeShareLinks decodes share links through the core.
func (h *Hub) DecodeShareLinks(lines []string) (json.RawMessage, error) {
	if h.client == nil {
		return nil, fmt.Errorf("core unavailable")
	}
	data, err := json.Marshal(map[string]any{"lines": lines})
	if err != nil {
		return nil, err
	}
	return h.client.Call("shareLink.decode", json.RawMessage(data), callLong)
}

// EncodeShareLinks encodes proxies to share links through the core.
func (h *Hub) EncodeShareLinks(proxies []map[string]any) (json.RawMessage, error) {
	if h.client == nil {
		return nil, fmt.Errorf("core unavailable")
	}
	data, err := json.Marshal(map[string]any{"proxies": proxies})
	if err != nil {
		return nil, err
	}
	return h.client.Call("shareLink.encode", json.RawMessage(data), callLong)
}

func timeNowMillis() int64 { return timeNow() }

// IntranetIP returns the primary outbound interface address.
func (h *Hub) IntranetIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	address, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}
	return address.IP.String()
}
