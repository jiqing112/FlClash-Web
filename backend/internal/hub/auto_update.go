package hub

// Auto-update loop: URL profiles with autoUpdate on are re-synced on their
// interval, mirroring the original's auto-update task.

import (
	"time"
)

func (h *Hub) StartAutoUpdater() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			h.autoUpdateTick()
		}
	}()
}

func (h *Hub) autoUpdateTick() {
	now := time.Now().UnixMilli()
	for _, profile := range h.Profiles() {
		if profile.Type != "url" || !profile.AutoUpdate || profile.Updating {
			continue
		}
		intervalMs := int64(profile.AutoUpdateMins) * 60_000
		if intervalMs <= 0 {
			intervalMs = 120 * 60_000
		}
		if profile.LastUpdateDate > 0 && now-profile.LastUpdateDate < intervalMs {
			continue
		}
		// mark due work; UpdateProfile clears the flag when done
		h.mu.Lock()
		for _, p := range h.profiles {
			if p.ID == profile.ID {
				p.Updating = true
			}
		}
		h.mu.Unlock()
		id := profile.ID
		go func() {
			defer h.markUpdating(id, false)
			if _, err := h.UpdateProfile(id); err != nil {
				h.emitWeb("error", map[string]any{"message": "auto update: " + err.Error()})
			}
		}()
	}
}
