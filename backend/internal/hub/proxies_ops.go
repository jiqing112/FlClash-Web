package hub

// Proxy operations: refresh, select, delay tests. Mirrors the Dart
// proxies providers' state flow.

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"flclashweb/backend/internal/rpc"
)

func (h *Hub) refreshProxies() {
	if h.client == nil {
		return
	}
	raw, err := h.client.Call("proxy.list", nil, 0)
	if err != nil {
		return
	}
	var data ProxiesData
	if json.Unmarshal(raw, &data) != nil {
		return
	}
	h.mu.Lock()
	h.proxiesData = &data
	groups := h.computeGroupsLocked()
	h.groups = groups
	h.mu.Unlock()
	h.emitter("groups", groups)
}

// computeGroupsLocked builds and sorts groups; caller holds h.mu.
func (h *Hub) computeGroupsLocked() []Group {
	groups := buildGroups(h.proxiesData)
	sortType := h.cfg.ProxiesStyle.SortType
	defaultTestUrl := h.cfg.PatchClashConfig.TestUrl
	selectedMap := h.currentSelectedMapLocked()
	groups = computeSort(groups, sortType, h.delayMap, selectedMap, defaultTestUrl)
	if h.cfg.ProxiesStyle.HideTimeoutProxys {
		groups = computeHideTimeout(groups, buildGroups(h.proxiesData), h.delayMap, selectedMap, defaultTestUrl)
	}
	return groups
}

// currentSelectedMapLocked merges each profile's stored selections with the
// running groups' current picks.
func (h *Hub) currentSelectedMapLocked() map[string]string {
	selectedMap := map[string]string{}
	if profile := h.currentProfileLocked(); profile != nil {
		for k, v := range profile.SelectedMap {
			selectedMap[k] = v
		}
	}
	return selectedMap
}

// Groups returns the computed groups snapshot.
func (h *Hub) Groups() []Group {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.groups
}

// ProxiesStyleUpdate recomputes groups after a style change.
func (h *Hub) ProxiesStyleUpdate() {
	h.mu.Lock()
	h.groups = h.computeGroupsLocked()
	groups := h.groups
	h.mu.Unlock()
	h.emitter("groups", groups)
}

// SelectProxy picks a node in a group and persists the selection.
func (h *Hub) SelectProxy(groupName, proxyName string) error {
	raw, err := h.client.Call("proxy.select", rpc.ChangeProxyParams{
		GroupName: groupName,
		ProxyName: proxyName,
	}, 0)
	if err != nil {
		return err
	}
	var result rpc.ChangeProxyResult
	json.Unmarshal(raw, &result)

	h.mu.Lock()
	// record the selection on the current profile, like selectedMapProvider
	if profile := h.currentProfileLocked(); profile != nil {
		if profile.SelectedMap == nil {
			profile.SelectedMap = map[string]string{}
		}
		profile.SelectedMap[groupName] = proxyName
		h.saveProfile(profile)
	}
	// keep the running view in sync
	if h.proxiesData != nil {
		if group, ok := h.proxiesData.Proxies[groupName]; ok {
			group.Now = proxyName
			if group.Fixed != "" {
				group.Fixed = proxyName
			}
			h.proxiesData.Proxies[groupName] = group
		}
	}
	h.groups = h.computeGroupsLocked()
	groups := h.groups
	h.mu.Unlock()

	h.emitWeb("groups", groups)
	h.emitWeb("profiles", h.Profiles())
	return nil
}

// TestDelay tests one proxy; the delay arrives through the delay event.
func (h *Hub) TestDelay(proxyName string, timeoutMs int64) error {
	cfg := h.Config()
	timeout := timeoutMs
	if timeout <= 0 {
		timeout = normalizeDelayTimeout(cfg.AppSetting.DelayTestTimeout)
	}
	_, err := h.client.Call("proxy.testDelay", rpc.TestDelayParams{
		ProxyName: proxyName,
		TestUrl:   cfg.PatchClashConfig.TestUrl,
		Timeout:   timeout,
	}, callLong)
	return err
}

// TestGroupDelay tests every node in a group, in parallel like the app.
func (h *Hub) TestGroupDelay(groupName string) error {
	h.mu.RLock()
	var group *Group
	for i := range h.groups {
		if h.groups[i].Name == groupName {
			group = &h.groups[i]
			break
		}
	}
	cfg := h.cfg
	h.mu.RUnlock()
	if group == nil {
		return nil
	}
	cfg2 := cfg
	timeout := normalizeDelayTimeout(cfg2.AppSetting.DelayTestTimeout)
	type job struct{ name string }
	jobs := make([]job, 0, len(group.All))
	for _, proxy := range group.All {
		jobs = append(jobs, job{proxy.Name})
	}
	sem := make(chan struct{}, 50)
	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(name string) {
			defer wg.Done()
			defer func() { <-sem }()
			_, _ = h.client.Call("proxy.testDelay", rpc.TestDelayParams{
				ProxyName: name,
				TestUrl:   cfg2.PatchClashConfig.TestUrl,
				Timeout:   timeout,
			}, callLong)
		}(j.name)
	}
	wg.Wait()
	return nil
}

// UnfoldSet tracks expanded groups in list mode.
func (h *Hub) ToggleUnfold(name string, unfold bool) {
	h.mu.Lock()
	if profile := h.currentProfileLocked(); profile != nil {
		profile.UnfoldSet = uniq(append(filterStr(profile.UnfoldSet, name), boolKey(name, unfold)))
		h.saveProfile(profile)
	}
	h.mu.Unlock()
}

func filterStr(list []string, needle string) []string {
	out := []string{}
	for _, item := range list {
		if item != needle {
			out = append(out, item)
		}
	}
	return out
}

func boolKey(name string, unfold bool) string {
	if unfold {
		return name
	}
	return ""
}

func uniq(list []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range list {
		if item != "" && !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}

func sortStrings(list []string) []string {
	out := append([]string{}, list...)
	sort.Strings(out)
	return out
}

// currentProfileLocked returns the active profile; caller holds h.mu.
func (h *Hub) currentProfileLocked() *ProfileState {
	if h.cfg.CurrentProfileID == nil {
		return nil
	}
	for _, p := range h.profiles {
		if p.ID == *h.cfg.CurrentProfileID {
			return p
		}
	}
	return nil
}

func (h *Hub) currentProfile() *ProfileState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.cfg.CurrentProfileID == nil {
		return nil
	}
	for _, p := range h.profiles {
		if p.ID == *h.cfg.CurrentProfileID {
			cp := *p
			return &cp
		}
	}
	return nil
}

var _ = time.Now
