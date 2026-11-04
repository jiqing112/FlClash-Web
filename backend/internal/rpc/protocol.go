package rpc

// Protocol payload shapes, mirroring core/protocol.go in the FlClash core.

type InitParams struct {
	HomeDir string `json:"home-dir"`
	Version int    `json:"version"`
}

type CoreOptions struct {
	TestURL          string `json:"test-url"`
	SkipCertVerify   bool   `json:"skip-cert-verify"`
	DelayTestTimeout int64  `json:"delay-test-timeout"`
}

type SetupBase struct {
	Path    string `json:"path"`
	Overlay string `json:"overlay"`
}

type SetupParams struct {
	Base        *SetupBase        `json:"base,omitempty"`
	SelectedMap map[string]string `json:"selected-map"`
	Options     CoreOptions       `json:"options"`
}

type TunSchema struct {
	Enable    bool    `json:"enable"`
	Device    *string `json:"device"`
	Stack     *string `json:"stack"`
	AutoRoute *bool   `json:"auto-route"`
}

type UpdateParams struct {
	Tun               *TunSchema          `json:"tun,omitempty"`
	AllowLan          *bool               `json:"allow-lan,omitempty"`
	MixedPort         *int                `json:"mixed-port,omitempty"`
	FindProcessMode   *string             `json:"find-process-mode,omitempty"`
	Mode              *string             `json:"mode,omitempty"`
	LogLevel          *string             `json:"log-level,omitempty"`
	IPv6              *bool               `json:"ipv6,omitempty"`
	TCPConcurrent     *bool               `json:"tcp-concurrent,omitempty"`
	ExternalController *string            `json:"external-controller,omitempty"`
	Secret            *string             `json:"secret,omitempty"`
	UnifiedDelay      *bool               `json:"unified-delay,omitempty"`
	Authentication    *[]string           `json:"authentication,omitempty"`
	GeoAutoUpdate     *bool               `json:"geo-auto-update,omitempty"`
	GeoUpdateInterval *int                `json:"geo-update-interval,omitempty"`
	GeoXUrl           map[string]string   `json:"geox-url,omitempty"`
	Options           *CoreOptions        `json:"options,omitempty"`
}

type ChangeProxyParams struct {
	GroupName string `json:"group-name"`
	ProxyName string `json:"proxy-name"`
}

type ChangeProxyResult struct {
	Message string `json:"message"`
	Changed bool   `json:"changed"`
}

type TestDelayParams struct {
	ProxyName string `json:"proxy-name"`
	TestUrl   string `json:"test-url"`
	Timeout   int64  `json:"timeout"`
}

type TrafficParams struct {
	OnlyProxy bool `json:"only-proxy"`
}

type Traffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

type CloseConnectionParams struct {
	ID string `json:"id"`
}

type GeoUpdateParams struct {
	Type string `json:"type"`
}

type ProviderSideLoadParams struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// Close closes the transport; the core process shuts down with it.
func (c *Client) Close() {
	c.mu.Lock()
	conn := c.conn
	c.closed = true
	c.conn = nil
	c.mu.Unlock()
	if conn != nil {
		conn.Close()
	}
}
