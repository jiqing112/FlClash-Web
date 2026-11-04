//go:build windows

package rpc

import (
	"fmt"
	"net"
	"os"

	"github.com/Microsoft/go-winio"
)

// listenEndpoint creates the host-side endpoint for the core: a named pipe on
// Windows, matching dial_pipe.go.
func listenEndpoint() (net.Listener, string, error) {
	path := fmt.Sprintf(`\\.\pipe\flclash-web-%d`, os.Getpid())
	ln, err := winio.ListenPipe(path, nil)
	if err != nil {
		return nil, "", err
	}
	return ln, path, nil
}
