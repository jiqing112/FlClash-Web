//go:build !windows

package hub

import "errors"

// openInternetSettings is only meaningful on Windows.
func openInternetSettings() (int, error) { return 0, errors.New("unsupported") }

// The Windows registry path is unreachable on these platforms (the runtime
// switch never picks it), but the symbols must exist to compile.
func setSystemProxyWindows(port int, bypass string) error { return errors.New("unsupported") }
func clearSystemProxyWindows() error                      { return errors.New("unsupported") }
