//go:build windows

package hub

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

func openInternetSettings() (registry.Key, error) {
	return registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.SET_VALUE|registry.QUERY_VALUE)
}
// ---- Windows ----

const internetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

func setSystemProxyWindows(port int, bypass string) error {
	key, err := openInternetSettings()
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.SetDWordValue("ProxyEnable", 1); err != nil {
		return err
	}
	if err := key.SetStringValue("ProxyServer", fmt.Sprintf("127.0.0.1:%d", port)); err != nil {
		return err
	}
	if bypass != "" {
		return key.SetStringValue("ProxyOverride", bypass)
	}
	return key.DeleteValue("ProxyOverride")
}

func clearSystemProxyWindows() error {
	key, err := openInternetSettings()
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.SetDWordValue("ProxyEnable", 0); err != nil {
		return err
	}
	return key.DeleteValue("ProxyServer")
}

