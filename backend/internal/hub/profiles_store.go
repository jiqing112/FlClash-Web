package hub

import (
	"encoding/json"
	"flclashweb/backend/internal/store"
)

func (h *Hub) SetStore(s *store.Store) { h.store = s }

// LoadState restores config and profiles from SQLite.
func (h *Hub) LoadState() error {
	if h.store == nil {
		return nil
	}
	raw, err := h.store.Get("config")
	if err != nil {
		return err
	}
	if raw != nil {
		var cfg Config
		if json.Unmarshal(raw, &cfg) == nil && cfg.AppSetting.UserAgents != nil {
			h.mu.Lock()
			h.cfg = cfg
			h.mu.Unlock()
		}
	}
	rawProfiles, err := h.store.Profiles()
	if err != nil {
		return err
	}
	h.mu.Lock()
	h.profiles = nil
	for _, raw := range rawProfiles {
		var p ProfileState
		if json.Unmarshal(raw, &p) == nil {
			h.profiles = append(h.profiles, &p)
		}
	}
	h.sortProfilesLocked()
	h.mu.Unlock()
	return nil
}

// byOrder keeps profiles in the stored order field.
func (h *Hub) sortProfilesLocked() {
	// stable insertion by Order (fractional indexing); empty orders keep position
	for i := 1; i < len(h.profiles); i++ {
		for j := i; j > 0 && profileLess(h.profiles[j], h.profiles[j-1]); j-- {
			h.profiles[j], h.profiles[j-1] = h.profiles[j-1], h.profiles[j]
		}
	}
}

func profileLess(a, b *ProfileState) bool {
	if a.Order == b.Order {
		return a.ID < b.ID
	}
	if a.Order == "" {
		return false
	}
	if b.Order == "" {
		return true
	}
	return a.Order < b.Order
}

func (h *Hub) saveProfile(p *ProfileState) {
	if h.store != nil {
		if data, err := json.Marshal(p.Profile); err == nil {
			_ = h.store.SaveProfile(p.ID, data)
		}
	}
}

func (h *Hub) removeProfile(id int64) {
	if h.store != nil {
		_ = h.store.DeleteProfile(id)
	}
}

// Profiles returns the profile list snapshot.
func (h *Hub) Profiles() []ProfileState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]ProfileState, 0, len(h.profiles))
	for _, p := range h.profiles {
		out = append(out, *p)
	}
	return out
}

func (h *Hub) CurrentProfileID() *int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cfg.CurrentProfileID
}

func (h *Hub) Profile(id int64) *ProfileState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range h.profiles {
		if p.ID == id {
			cp := *p
			return &cp
		}
	}
	return nil
}

// PutProfile inserts or replaces a profile.
func (h *Hub) PutProfile(p ProfileState) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, existing := range h.profiles {
		if existing.ID == p.ID {
			h.profiles[i] = &p
			h.sortProfilesLocked()
			h.saveProfile(&p)
			return
		}
	}
	h.profiles = append(h.profiles, &p)
	h.sortProfilesLocked()
	h.saveProfile(&p)
}

// DeleteProfile removes a profile from the store and memory.
func (h *Hub) DeleteProfile(id int64) {
	h.mu.Lock()
	for i, p := range h.profiles {
		if p.ID == id {
			h.profiles = append(h.profiles[:i], h.profiles[i+1:]...)
			break
		}
	}
	h.mu.Unlock()
	h.removeProfile(id)
}
