//go:build !windows

package rpc

import (
	"net"
	"strconv"
)

// listenEndpoint creates the host-side endpoint for the core: a loopback TCP
// port, matching dial_socket.go.
func listenEndpoint() (net.Listener, string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}
	return ln, strconv.Itoa(ln.Addr().(*net.TCPAddr).Port), nil
}
