package hub

// System proxy per desktop platform, matching the app's proxy plugin:
// Windows registry, GNOME gsettings, macOS networksetup.

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// applySystemProxy sets the OS proxy to 127.0.0.1:port with bypass domains.
func (h *Hub) syncSystemProxy() {
	cfg := h.Config()
	if !cfg.NetworkSetting.SystemProxy || !cfg.VpnSetting.SystemProxy {
		h.clearSystemProxy()
		return
	}
	port := cfg.PatchClashConfig.MixedPort
	bypass := strings.Join(cfg.NetworkSetting.BypassDomain, bypassSeparator())
	var err error
	switch runtime.GOOS {
	case "windows":
		err = setSystemProxyWindows(port, bypass)
	case "darwin":
		err = setSystemProxyDarwin(port, bypass)
	case "linux":
		err = setSystemProxyLinux(port, bypass)
	}
	h.mu.Lock()
	h.systemProxy = err == nil
	h.mu.Unlock()
	if err != nil {
		h.emitter("error", map[string]any{"message": "system proxy: " + err.Error()})
	}
}

func (h *Hub) clearSystemProxy() {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = clearSystemProxyWindows()
	case "darwin":
		err = clearSystemProxyDarwin()
	case "linux":
		err = clearSystemProxyLinux()
	}
	h.mu.Lock()
	h.systemProxy = false
	h.mu.Unlock()
	_ = err
}

func bypassSeparator() string {
	switch runtime.GOOS {
	case "windows":
		return ";"
	default:
		return ", "
	}
}

// ---- macOS ----

func networkServices() []string {
	out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return nil
	}
	var services []string
	for i, line := range strings.Split(string(out), "\n") {
		if i == 0 || strings.TrimSpace(line) == "" || strings.HasPrefix(line, "*") {
			continue
		}
		services = append(services, strings.TrimSpace(line))
	}
	return services
}

func setSystemProxyDarwin(port int, bypass string) error {
	for _, service := range networkServices() {
		for _, args := range [][]string{
			{"-setwebproxy", service, "127.0.0.1", fmt.Sprintf("%d", port)},
			{"-setsecurewebproxy", service, "127.0.0.1", fmt.Sprintf("%d", port)},
			{"-setsocksfirewallproxy", service, "127.0.0.1", fmt.Sprintf("%d", port)},
		} {
			if err := exec.Command("networksetup", args...).Run(); err != nil {
				return err
			}
		}
		bypassArgs := append([]string{"-setproxybypassdomains", service}, strings.Split(bypass, ", ")...)
		if err := exec.Command("networksetup", bypassArgs...).Run(); err != nil {
			return err
		}
	}
	return nil
}

func clearSystemProxyDarwin() error {
	for _, service := range networkServices() {
		for _, args := range [][]string{
			{"-setwebproxystate", service, "off"},
			{"-setsecurewebproxystate", service, "off"},
			{"-setsocksfirewallproxystate", service, "off"},
		} {
			_ = exec.Command("networksetup", args...).Run()
		}
	}
	return nil
}

// ---- Linux (GNOME) ----

func setSystemProxyLinux(port int, bypass string) error {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return fmt.Errorf("gsettings not available")
	}
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", "127.0.0.1").Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "port", fmt.Sprintf("%d", port)).Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "host", "127.0.0.1").Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "port", fmt.Sprintf("%d", port)).Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "host", "127.0.0.1").Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "port", fmt.Sprintf("%d", port)).Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy", "ignore-hosts",
		fmt.Sprintf("['%s']", strings.Join(strings.Split(bypass, ", "), "','"))).Run()
	return nil
}

func clearSystemProxyLinux() error {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return fmt.Errorf("gsettings not available")
	}
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
	return nil
}

