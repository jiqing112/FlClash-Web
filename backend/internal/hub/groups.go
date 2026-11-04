package hub

// Port of lib/common/task.dart buildGroups and lib/common/compute.dart.

import (
	"strings"

	"flclashweb/backend/internal/models"
)

const DefaultTestURL = models.DefaultTestURL

// buildGroups turns the core's ProxiesData into client-facing groups whose
// `all` carries full proxy views.
func buildGroups(data *ProxiesData) []Group {
	if data == nil || len(data.Proxies) == 0 {
		return []Group{}
	}
	groups := []Group{}
	for _, groupName := range data.All {
		raw, ok := data.Proxies[groupName]
		if !ok {
			continue
		}
		if !groupTypes[normalizeGroupType(raw.Type)] {
			continue
		}
		group := Group{
			Name:    raw.Name,
			Type:    normalizeGroupType(raw.Type),
			Now:     raw.Now,
			Hidden:  raw.Hidden,
			TestUrl: raw.TestUrl,
			Icon:    raw.Icon,
			All:     []Proxy{},
		}
		for _, memberName := range raw.All {
			if member, ok := data.Proxies[memberName]; ok {
				group.All = append(group.All, member)
			}
		}
		groups = append(groups, group)
	}
	return groups
}

func normalizeGroupType(t string) string {
	switch strings.ToLower(t) {
	case "urltest", "url-test":
		return "url-test"
	case "selector", "select":
		return "select"
	case "fallback":
		return "fallback"
	case "loadbalance", "load-balance":
		return "load-balance"
	case "relay":
		return "relay"
	default:
		return strings.ToLower(t)
	}
}

// selectedProxyState follows a group pick down to the real node.
type selectedProxyState struct {
	proxyName string
	group     bool
	testUrl   string
}

func realSelectedProxyState(name string, groups []Group, selectedMap map[string]string) selectedProxyState {
	state := selectedProxyState{proxyName: name}
	for depth := 0; depth < 16 && state.proxyName != ""; depth++ {
		var group *Group
		for i := range groups {
			if groups[i].Name == state.proxyName {
				group = &groups[i]
				break
			}
		}
		state.group = true
		if group == nil {
			return state
		}
		pick := selectedMap[state.proxyName]
		if pick == "" {
			pick = group.Now
		}
		if pick == "" {
			return state
		}
		state.proxyName = pick
		if group.TestUrl != "" {
			state.testUrl = group.TestUrl
		}
	}
	return state
}

// computeProxyDelayState mirrors Dart's computeProxyDelayState: 0 means unknown.
func computeProxyDelayState(name, testUrl string, groups []Group, selectedMap map[string]string, delayMap map[string]map[string]int32) int32 {
	state := realSelectedProxyState(name, groups, selectedMap)
	url := state.testUrl
	if url == "" {
		url = testUrl
	}
	if url == "" {
		url = DefaultTestURL
	}
	if current := delayMap[url]; current != nil {
		if delay, ok := current[state.proxyName]; ok {
			return delay
		}
	}
	return 0
}

func sortProxyNamesByDelay(proxies []Proxy, groups []Group, selectedMap map[string]string, testUrl string, delayMap map[string]map[string]int32) []Proxy {
	states := map[string]int32{}
	for _, proxy := range proxies {
		states[proxy.Name] = computeProxyDelayState(proxy.Name, testUrl, groups, selectedMap, delayMap)
	}
	out := append([]Proxy{}, proxies...)
	// stable insertion sort on the delay state
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && states[out[j].Name] < states[out[j-1].Name]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func sortProxiesByName(proxies []Proxy) []Proxy {
	out := append([]Proxy{}, proxies...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Name < out[j-1].Name; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// computeSort applies the proxies style sort to every group.
func computeSort(groups []Group, sortType string, delayMap map[string]map[string]int32, selectedMap map[string]string, defaultTestUrl string) []Group {
	out := make([]Group, 0, len(groups))
	for _, group := range groups {
		proxies := group.All
		testUrl := group.TestUrl
		if testUrl == "" {
			testUrl = defaultTestUrl
		}
		switch sortType {
		case "delay":
			proxies = sortProxyNamesByDelay(proxies, groups, selectedMap, testUrl, delayMap)
		case "name":
			proxies = sortProxiesByName(proxies)
		}
		group.All = proxies
		out = append(out, group)
	}
	return out
}

// computeHideTimeout mirrors Dart's computeHideTimeout.
var unprobeableProxyTypes = map[string]bool{
	"Reject": true, "RejectDrop": true, "Pass": true, "PassRule": true,
	"Rematch": true, "Compatible": true, "Dns": true,
}

func computeHideTimeout(groups, allGroups []Group, delayMap map[string]map[string]int32, selectedMap map[string]string, defaultTestUrl string) []Group {
	realStates := map[string]selectedProxyState{}
	proxyTypes := map[string]string{}
	for _, group := range allGroups {
		for _, proxy := range group.All {
			proxyTypes[proxy.Name] = proxy.Type
		}
	}
	out := make([]Group, 0, len(groups))
	for _, group := range groups {
		groupTestUrl := group.TestUrl
		if groupTestUrl == "" {
			groupTestUrl = defaultTestUrl
		}
		var withNow *Group
		for i := range allGroups {
			if allGroups[i].Name == group.Name {
				withNow = &allGroups[i]
				break
			}
		}
		if withNow == nil {
			withNow = &group
		}
		selectedName := withNow.Now
		if pick := selectedMap[group.Name]; pick != "" {
			selectedName = pick
		}
		visible := []Proxy{}
		for _, proxy := range group.All {
			if proxy.Name == selectedName {
				visible = append(visible, proxy)
				continue
			}
			isGroup := false
			for _, candidate := range allGroups {
				if candidate.Name == proxy.Name {
					isGroup = true
					break
				}
			}
			if isGroup {
				visible = append(visible, proxy)
				continue
			}
			state, ok := realStates[proxy.Name]
			if !ok {
				state = realSelectedProxyState(proxy.Name, allGroups, selectedMap)
				realStates[proxy.Name] = state
			}
			if unprobeableProxyTypes[proxyTypes[state.proxyName]] {
				visible = append(visible, proxy)
				continue
			}
			testUrl := state.testUrl
			if testUrl == "" {
				testUrl = groupTestUrl
			}
			delay, seen := delayMap[testUrl][state.proxyName]
			if !seen || delay > 0 {
				visible = append(visible, proxy)
			}
		}
		if len(visible) == 0 {
			visible = group.All
		}
		group.All = visible
		out = append(out, group)
	}
	return out
}

// getCurrentSelectedName mirrors Group.getCurrentSelectedName.
func getCurrentSelectedName(group Group, selectedMap map[string]string) string {
	now := group.Now
	if pick := selectedMap[group.Name]; pick != "" {
		return pick
	}
	return now
}
