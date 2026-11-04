package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"flclashweb/backend/internal/models"
	"flclashweb/backend/internal/rpc"
)

// ---- core lifecycle (port of setup.dart's control flow) ----

// InitCore initializes the core process for the given home dir.
func (h *Hub) InitCore() error {
	if h.client == nil {
		return fmt.Errorf("core client missing")
	}
	_, err := h.client.Call("core.init", rpc.InitParams{
		HomeDir: h.homeDir,
		Version: 1,
	}, 0)
	if err != nil {
		return err
	}
	// The log pump only runs while started; without it the Logs page starves.
	_, _ = h.client.Call("log.start", nil, 0)
	return nil
}

// ApplyProfile builds the current profile and pushes it to the core
// (config.setup). Runs on profile switch, override edits and startup.
func (h *Hub) ApplyProfile(force bool) error {
	cfg := h.Config()
	profile := h.currentProfile()

	// no profile: push an empty config so the core stops serving the old one
	baseYaml, md5, baseParams, counts, buildErr := h.buildProfileYaml(profile, cfg)
	if buildErr != nil {
		baseYaml, md5, baseParams, counts = "", "", nil, Counts{}
	}

	h.mu.Lock()
	if !force && md5 != "" && md5 == h.lastConfigMd5 {
		h.mu.Unlock()
		return nil
	}
	h.lastConfigMd5 = ""
	h.mu.Unlock()

	params := rpc.SetupParams{
		SelectedMap: map[string]string{},
		Options:     h.coreOptions(cfg),
		Base:        baseParams,
	}

	if baseParams == nil {
		configPath := h.configPath()
		if err := os.WriteFile(configPath, []byte(baseYaml), 0o644); err != nil {
			return err
		}
	}

	_, err := h.client.Call("config.setup", params, callLong)
	if err != nil {
		// config_rejected still leaves the default config running; report up
		h.emitter("error", map[string]any{"message": err.Error()})
		return err
	}

	h.mu.Lock()
	h.lastConfigMd5 = md5
	h.lastCounts = counts
	h.delayMap = map[string]map[string]int32{}
	h.mu.Unlock()

	h.refreshProxies()
	h.refreshProviders()
	return nil
}

// UpdatePatch pushes live-patchable settings (ports, mode, ...) via config.update.
func (h *Hub) UpdatePatch() error {
	cfg := h.Config()
	patch := cfg.PatchClashConfig
	network := cfg.NetworkSetting
	params := rpc.UpdateParams{
		Tun: &rpc.TunSchema{
			Enable:    patch.Tun.Enable,
			Device:    strPtr(patch.Tun.Device),
			Stack:     strPtr(patch.Tun.Stack),
			AutoRoute: boolPtr(patch.Tun.AutoRoute),
		},
		AllowLan:        &patch.AllowLan,
		MixedPort:       &patch.MixedPort,
		FindProcessMode: strPtr(patch.FindProcessMode),
		Mode:            strPtr(patch.Mode),
		LogLevel:        strPtr(patch.LogLevel),
		IPv6:            boolPtr(pick(patch.GeneralOverrideKeys, "ipv6", patch.IPv6)),
		TCPConcurrent:   boolPtr(pick(patch.GeneralOverrideKeys, "tcp-concurrent", patch.TCPConcurrent)),
		UnifiedDelay:    boolPtr(pick(patch.GeneralOverrideKeys, "unified-delay", patch.UnifiedDelay)),
		Authentication:  &[]string{},
		GeoAutoUpdate:   &patch.GeoAutoUpdate,
		GeoUpdateInterval: &patch.GeoUpdateInterval,
		GeoXUrl:         patch.GeoXUrl,
		Options:         h.coreOptionsPtr(cfg),
	}
	if creds := network.Authentication.Credentials(); creds != nil {
		params.Authentication = &creds
	}
	_, err := h.client.Call("config.update", params, callLong)
	return err
}

func pick(keys []string, key string, value bool) bool {
	for _, k := range keys {
		if k == key {
			return value
		}
	}
	return false
}

func (h *Hub) coreOptions(cfg Config) rpc.CoreOptions { return coreOptionsOf(cfg) }

func coreOptionsOf(cfg Config) rpc.CoreOptions {
	return rpc.CoreOptions{
		TestURL:          cfg.PatchClashConfig.TestUrl,
		SkipCertVerify:   !cfg.AppSetting.CheckCertificate,
		// stored in milliseconds like the original; older web builds saved
		// whole seconds, so scale those up
		DelayTestTimeout: normalizeDelayTimeout(cfg.AppSetting.DelayTestTimeout),
	}
}

func (h *Hub) coreOptionsPtr(cfg Config) *rpc.CoreOptions {
	o := coreOptionsOf(cfg)
	return &o
}

// StartListener turns the run toggle on: listeners up, system proxy applied.
// A start pressed before the boot apply finished would start listeners over a
// nil config and answer success while nothing listens, so apply first when
// the core has never taken a config.
func (h *Hub) StartListener() error {
	h.mu.RLock()
	neverApplied := h.lastConfigMd5 == ""
	h.mu.RUnlock()
	if neverApplied {
		if err := h.ApplyProfile(true); err != nil {
			return err
		}
	}
	if _, err := h.client.Call("listener.start", nil, 0); err != nil {
		return err
	}
	h.mu.Lock()
	if !h.running {
		h.running = true
		h.startTime = time.Now().UnixMilli()
	}
	h.mu.Unlock()
	h.syncSystemProxy()
	h.emitter("running", h.RunningState())
	return nil
}

// StopListener turns the run toggle off.
func (h *Hub) StopListener() error {
	if _, err := h.client.Call("listener.stop", nil, 0); err != nil {
		return err
	}
	h.mu.Lock()
	h.running = false
	h.startTime = 0
	h.mu.Unlock()
	h.clearSystemProxy()
	h.requests.Reset()
	h.emitter("running", h.RunningState())
	return nil
}

func (h *Hub) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

type RunningState struct {
	Running     bool   `json:"running"`
	StartTime   int64  `json:"startTime"`
	SystemProxy bool   `json:"systemProxy"`
	Port        int    `json:"port"`
	Mode        string `json:"mode"`
}

func (h *Hub) RunningState() RunningState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return RunningState{
		Running:     h.running,
		StartTime:   h.startTime,
		SystemProxy: h.systemProxy,
		Port:        h.cfg.PatchClashConfig.MixedPort,
		Mode:        h.cfg.PatchClashConfig.Mode,
	}
}

// RestartCore tears down and re-inits the core, then reapplies the profile.
func (h *Hub) RestartCore() error {
	h.mu.RLock()
	old := h.client
	h.mu.RUnlock()
	if old != nil {
		old.Close()
	}
	time.Sleep(200 * time.Millisecond)
	client, err := rpc.Start(h.homeDir)
	if err != nil {
		return err
	}
	h.mu.Lock()
	h.client = client
	h.mu.Unlock()
	h.StartEventPump()
	h.emitCoreStatus()
	if err := h.InitCore(); err != nil {
		return err
	}
	if h.IsRunning() {
		if err := h.ApplyProfile(true); err != nil {
			return err
		}
		return h.StartListener()
	}
	return h.ApplyProfile(true)
}

func (h *Hub) configPath() string {
	return filepath.Join(h.homeDir, "config.yaml")
}

func strPtr(s string) *string    { return &s }
func boolPtr(b bool) *bool       { return &b }
func intPtr(i int) *int          { return &i }

const callLong = 10 * time.Minute

// counts of the built profile, like Dart's ConfigCounts.
type Counts struct {
	Groups  int `json:"groups"`
	Proxies int `json:"proxies"`
	Rules   int `json:"rules"`
}

var _ = models.DefaultConfig // keep import
