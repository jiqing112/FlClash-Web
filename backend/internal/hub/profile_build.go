package hub

// Port of lib/common/task.dart `_makeRealProfileTask` and the SetupBase
// optimization from setup.dart's getProfile, against the same core methods.

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"flclashweb/backend/internal/rpc"
)

const defaultUA = "FlClash/0.9.3"

var baselineDnsOverrideKeys = []string{
	"enable", "listen", "enhanced-mode", "nameserver",
}

var baselineDns = map[string]any{
	"enable":            true,
	"listen":            ":1053",
	"enhanced-mode":     "redir-host",
	"default-nameserver": []any{"223.5.5.5", "8.8.8.4", "1.0.0.1"},
	"nameserver":        []any{"223.5.5.5", "8.8.8.8", "1.0.0.1"},
}

var privateRouteAddress = []any{
	"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	"169.254.0.0/16", "100.64.0.0/10", "224.0.0.0/4", "fc00::/7", "fe80::/5", "ff00::/8",
}

var defaultDNSHijack = []any{"any:53"}

const providersDirectoryName = "providers"

// buildProfileYaml builds the overlay yaml for the current profile and returns
// (yaml, md5, setup base, counts). For a standard profile the core caches the
// big base file and the overlay carries only the patches, exactly like the
// Flutter app.
func (h *Hub) buildProfileYaml(profile *ProfileState, cfg Config) (string, string, *rpc.SetupBase, Counts, error) {
	emptyCounts := Counts{}
	if profile == nil {
		return "", "", nil, emptyCounts, nil
	}
	profilePath := h.profilePath(profile.ID)
	rawBuf, err := osReadFile(profilePath)
	if err != nil {
		return "", "", nil, emptyCounts, fmt.Errorf("read profile: %w", err)
	}
	if len(strings.TrimSpace(string(rawBuf))) == 0 {
		return "", "", nil, emptyCounts, fmt.Errorf("profile file is empty")
	}

	network := cfg.NetworkSetting
	patch := cfg.PatchClashConfig

	// The effective patch config: a secret is required for an open controller.
	externalController := ""
	if patch.Secret != "" {
		externalController = patch.ExternalController
	}
	secret := patch.Secret

	// getProfileBase through the core: strips proxies/groups/rules into the
	// cached base and returns the remaining config as JSON.
	var baseResult struct {
		Config      json.RawMessage `json:"config"`
		Digest      string          `json:"digest"`
		MatchRules  []string        `json:"match-rules"`
		Proxies     int             `json:"proxies"`
		ProxyGroups int             `json:"proxy-groups"`
		Rules       int             `json:"rules"`
	}
	baseResp, err := h.client.Call("profile.base", rpcPathParams{Path: profilePath}, callLong)
	if err == nil && json.Unmarshal(baseResp, &baseResult) == nil && baseResult.Config != nil {
		// base path available
	} else {
		baseResult.Config = nil
	}

	var rawConfig map[string]any
	if baseResult.Config != nil {
		if err := json.Unmarshal(baseResult.Config, &rawConfig); err != nil {
			return "", "", nil, emptyCounts, fmt.Errorf("decode base config: %w", err)
		}
	} else {
		// no base path: read the whole file as raw yaml
		if rawConfig, err = yamlUnmarshal(rawBuf); err != nil {
			return "", "", nil, emptyCounts, fmt.Errorf("parse profile: %w", err)
		}
		if rawConfig == nil {
			rawConfig = map[string]any{}
		}
	}

	matchTarget := profile.MatchTarget
	if matchTarget == nil || strings.TrimSpace(*matchTarget) == "" {
		if len(baseResult.MatchRules) > 0 {
			target := lastMatchTarget(baseResult.MatchRules)
			matchTarget = &target
		}
	}

	authentication := network.Authentication.Credentials()
	if authentication == nil {
		authentication = []string{}
	}

	state := makeRealProfileState{
		rawConfig:           rawConfig,
		patch:               &patch,
		overrides:           &profile.Overrides,
		externalController:  externalController,
		secret:              secret,
		authentication:      authentication,
		appendSystemDns:     network.AppendSystemDNS,
		bypassPrivateRoute:  network.BypassPrivateRoute,
		keepEmptyDns:        true,
		profilesPath:        h.profilesDir(),
		profileID:           profile.ID,
		defaultUA:           defaultUA,
		addedRules:          []string{},
		matchTarget:         matchTarget,
	}

	counts, err := state.build()
	if err != nil {
		return "", "", nil, emptyCounts, err
	}
	yamlText, err := yamlMarshal(state.rawConfig)
	if err != nil {
		return "", "", nil, emptyCounts, err
	}
	sum := md5.Sum([]byte(yamlText))
	md5hex := hex.EncodeToString(sum[:])

	var setupBase *rpc.SetupBase
	if baseResult.Config != nil {
		basePath := filepath.Join(h.profilesDir(), fmt.Sprintf(".base-%d.yaml", profile.ID))
		// The core caches the base from the profile file itself; hand it the
		// original path and the overlay we built.
		setupBase = &rpc.SetupBase{Path: profilePath, Overlay: yamlText}
		counts.Proxies += baseResult.Proxies
		counts.Groups += baseResult.ProxyGroups
		counts.Rules += baseResult.Rules
		md5hex = md5Hex(md5hex + baseResult.Digest)
		_ = basePath
	}
	return yamlText, md5hex, setupBase, counts, nil
}

type rpcPathParams struct {
	Path string `json:"path"`
}

type makeRealProfileState struct {
	rawConfig          map[string]any
	patch              *PatchClashConfigAlias
	overrides          *ProfileOverridesAlias
	externalController string
	secret             string
	authentication     []string
	appendSystemDns    bool
	bypassPrivateRoute bool
	keepEmptyDns       bool
	profilesPath       string
	profileID          int64
	defaultUA          string
	addedRules         []string
	matchTarget        *string
}

// aliases keep the build function readable without renaming the models.
type PatchClashConfigAlias = modelsPatchClashConfig
type ProfileOverridesAlias = modelsProfileOverrides

func (s *makeRealProfileState) build() (Counts, error) {
	raw := s.rawConfig
	patch := s.patch
	overrides := s.overrides

	raw["external-controller"] = s.externalController
	raw["secret"] = s.secret
	raw["external-controller-tls"] = ""
	raw["external-controller-unix"] = ""
	raw["external-controller-pipe"] = ""
	if iptables, ok := raw["iptables"].(map[string]any); ok {
		if iface, ok := iptables["inbound-interface"].(string); ok && strings.Contains(iface, " ") {
			raw["iptables"] = map[string]any{"enable": false}
		}
	}
	raw["external-ui"] = ""
	switch patch.InterfaceNameMode {
	case "clear":
		raw["interface-name"] = ""
	case "custom":
		raw["interface-name"] = patch.InterfaceName
	}
	raw["external-ui-url"] = ""

	// general overrides picked by key
	for key, value := range generalOverrideFragment(patch) {
		raw[key] = value
	}
	raw["log-level"] = patch.LogLevel
	raw["keep-alive-interval"] = patch.KeepAliveInterval
	raw["keep-alive-idle"] = patch.KeepAliveIdle
	raw["disable-keep-alive"] = patch.DisableKeepAlive
	raw["routing-mark"] = patch.RoutingMark
	raw["mixed-port"] = patch.MixedPort
	raw["port"] = patch.Port
	raw["socks-port"] = patch.SocksPort
	raw["redir-port"] = patch.RedirPort
	raw["tproxy-port"] = patch.TproxyPort
	raw["find-process-mode"] = patch.FindProcessMode
	raw["allow-lan"] = patch.AllowLan
	raw["authentication"] = toAnySlice(s.authentication)
	raw["skip-auth-prefixes"] = []any{}
	raw["mode"] = patch.Mode

	tun := map[string]any{}
	if t, ok := raw["tun"].(map[string]any); ok {
		tun = t
	}
	tun["dns-hijack"] = defaultDNSHijack
	for key, value := range overrideFragment(tunOverrideJSON(patch.TunOverride), patch.TunOverrideKeys) {
		tun[key] = value
	}
	tun["enable"] = patch.Tun.Enable
	tun["device"] = patch.Tun.Device
	tun["stack"] = patch.Tun.Stack
	tun["route-address"] = toAnySlice(patch.Tun.RouteAddress)
	tun["mtu"] = patch.Tun.MTU
	tun["auto-route"] = patch.Tun.AutoRoute
	tun["auto-detect-interface"] = true
	if s.bypassPrivateRoute {
		existing := toAnySlice(tun["route-exclude-address"])
		existing = append(existing, privateRouteAddress...)
		tun["route-exclude-address"] = existing
	}
	raw["tun"] = tun

	raw["geo-auto-update"] = patch.GeoAutoUpdate
	raw["geo-update-interval"] = patch.GeoUpdateInterval

	// sniffer override
	if snifferJSON := overrideJSONFromRaw(overrides.Sniffer); snifferJSON != nil {
		frag := overrideFragment(snifferJSON, overrides.SnifferOverrideKeys)
		if len(frag) > 0 {
			existing := map[string]any{}
			if s, ok := raw["sniffer"].(map[string]any); ok {
				existing = s
			}
			raw["sniffer"] = mergeMaps(existing, frag)
		}
	}

	// profile.store-selected = false
	profileSection := map[string]any{}
	if p, ok := raw["profile"].(map[string]any); ok {
		profileSection = p
	}
	profileSection["store-selected"] = false
	raw["profile"] = profileSection

	raw["geox-url"] = toAnyMap(patch.GeoXUrl)
	globalUa := s.defaultUA
	if patch.GlobalUa != nil && *patch.GlobalUa != "" {
		globalUa = *patch.GlobalUa
	}
	raw["global-ua"] = globalUa

	hosts := map[string]any{}
	if m, ok := raw["hosts"].(map[string]any); ok {
		hosts = m
	}
	for host, value := range patch.Hosts {
		hosts[host] = value
	}
	raw["hosts"] = hosts

	// DNS: baseline when disabled, then the profile override fragment
	rawDns := map[string]any{}
	if d, ok := raw["dns"].(map[string]any); ok {
		rawDns = d
	}
	enabled := rawDns["enable"] == true
	if !enabled {
		rawDns = mergeMaps(rawDns, overrideFragment(baselineDns, baselineDnsOverrideKeys))
	}
	if dnsJSON := overrideJSONFromRaw(overrides.Dns); dnsJSON != nil {
		frag := overrideFragmentKeepEmpty(dnsJSON, overrides.DnsOverrideKeys, s.keepEmptyDns)
		rawDns = mergeMaps(rawDns, frag)
	}
	raw["dns"] = rawDns
	if s.appendSystemDns {
		nameserver := toAnySlice(rawDns["nameserver"])
		has := false
		for _, item := range nameserver {
			if str, ok := item.(string); ok && str == "system://" {
				has = true
			}
		}
		if !has {
			nameserver = append(nameserver, "system://")
			rawDns["nameserver"] = nameserver
		}
	}

	// ntp override
	if ntpJSON := overrideJSONFromRaw(overrides.Ntp); ntpJSON != nil {
		frag := overrideFragment(ntpJSON, overrides.NtpOverrideKeys)
		if len(frag) > 0 {
			existing := map[string]any{}
			if n, ok := raw["ntp"].(map[string]any); ok {
				existing = n
			}
			raw["ntp"] = mergeMaps(existing, frag)
		}
	}

	// providers confinement
	confineProviders(raw, "proxy-providers", "proxies", s.profilesPath, s.profileID)
	confineProviders(raw, "rule-providers", "rules", s.profilesPath, s.profileID)

	// rules
	rules := []string{}
	if existing, ok := raw["rules"].([]any); ok {
		for _, item := range existing {
			if str, ok := item.(string); ok {
				rules = append(rules, str)
			}
		}
	}
	added := s.addedRules
	if len(added) > 0 {
		replacement := ""
		if s.matchTarget != nil {
			replacement = strings.TrimSpace(*s.matchTarget)
		}
		if replacement == "" {
			replacement = lastMatchTarget(rules)
		}
		if replacement != "" {
			fixed := make([]string, 0, len(added))
			for _, rule := range added {
				fixed = append(fixed, retargetMatchRule(rule, replacement))
			}
			added = fixed
		}
		rules = append(added, rules...)
	}
	raw["rules"] = toAnySlice(rules)

	return Counts{
		Groups:  listLen(raw["proxy-groups"]),
		Proxies: listLen(raw["proxies"]),
		Rules:   listLen(raw["rules"]),
	}, nil
}

func confineProviders(raw map[string]any, section, kind, profilesPath string, profileID int64) {
	providers, ok := raw[section].(map[string]any)
	if !ok {
		return
	}
	for name, value := range providers {
		provider, ok := value.(map[string]any)
		if !ok || provider["type"] == "inline" {
			continue
		}
		url, hasURL := provider["url"].(string)
		key := section + "/" + name
		if hasURL && url != "" {
			key = name + "@" + url
		}
		path := filepath.Join(profilesPath, providersDirectoryName,
			fmt.Sprintf("%d", profileID), kind, md5Hex(key))
		provider["path"] = path
	}
}

func listLen(value any) int {
	if list, ok := value.([]any); ok {
		return len(list)
	}
	return 0
}

func toAnySlice(value any) []any {
	switch v := value.(type) {
	case nil:
		return []any{}
	case []any:
		return v
	case []string:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, item)
		}
		return out
	default:
		return []any{value}
	}
}

func toAnyMap(value map[string]string) map[string]any {
	out := map[string]any{}
	for k, v := range value {
		out[k] = v
	}
	return out
}

func mergeMaps(base, override map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		if k == "fallback-filter" {
			if existingMap, ok := out[k].(map[string]any); ok {
				if overrideMap, ok := v.(map[string]any); ok {
					out[k] = mergeMaps(existingMap, overrideMap)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// lastMatchTarget mirrors Dart's lastMatchTarget.
func lastMatchTarget(rules []string) string {
	for i := len(rules) - 1; i >= 0; i-- {
		fields := strings.Split(rules[i], ",")
		if strings.EqualFold(strings.TrimSpace(fields[0]), "MATCH") && len(fields) >= 2 {
			target := strings.TrimSpace(fields[1])
			if target != "" {
				return target
			}
		}
	}
	return ""
}

// retargetMatchRule rewrites a MATCH placeholder rule to the real target.
func retargetMatchRule(rule, target string) string {
	fields := strings.Split(rule, ",")
	if len(fields) >= 2 && strings.EqualFold(strings.TrimSpace(fields[0]), "MATCH") {
		fields[1] = target
		return strings.Join(fields, ",")
	}
	return rule
}

func md5Hex(text string) string {
	sum := md5.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}
