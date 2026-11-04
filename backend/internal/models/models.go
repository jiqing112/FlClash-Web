// Package models ports the FlClash Dart models (lib/models) with their exact
// JSON keys, so the persisted config and the web client share one format.
package models

import "encoding/json"

// ---- enums (values match lib/enum/enum.dart) ----

const (
	DefaultTestURL          = "https://www.gstatic.com/generate_204"
	DefaultMixedPort        = 7890
	DefaultKeepAlive        = 30
	DefaultTunMTU           = 9000
	DefaultDelayTestTimeout = 8000
	DefaultPrimaryColor     = 0xFF007AFF
)

var DefaultPrimaryColors = []int64{
	0xFF007AFF, 0xFFCB30E0, 0xFFFF2D55, 0xFFFF383C,
	0xFFFF8D28, 0xFFFFCC00, 0xFF34C759, 0xFF8E8E93,
}

var DefaultGeoXUrl = map[string]string{
	"mmdb":    "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/country-lite.mmdb",
	"asn":     "https://github.com/xishang0128/geoip/releases/download/latest/GeoLite2-ASN.mmdb",
	"geoip":   "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip-lite.dat",
	"geosite": "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat",
}

var DefaultBypassDomain = []string{
	"*zhihu.com", "*zhimg.com", "*jd.com", "100ime-iat-api.xfyun.cn", "*360buyimg.com",
	"localhost", "*.local", "127.*", "10.*",
	"172.16.*", "172.17.*", "172.18.*", "172.19.*", "172.2*", "172.30.*", "172.31.*",
	"192.168.*",
}

var DefaultUserAgents = []string{"clash-verge/v2.4.2", "ClashforWindows/0.19.23"}

type Filter struct {
	Label string `json:"label"`
	Regex string `json:"regex"`
}

var DefaultFilters = []Filter{
	{Label: "Hong Kong", Regex: `(?i)🇭🇰|港|(?<![a-z])hk(?![a-z])|hong\s*kong`},
	{Label: "Taiwan", Regex: `(?i)🇹🇼|台湾|台灣|台北|新北|彰化|(?<![a-z])tw(?![a-z])|taiwan`},
	{Label: "Japan", Regex: `(?i)🇯🇵|日本|东京|東京|大阪|埼玉|(?<![a-z])jp(?![a-z])|japan|tokyo|osaka`},
	{Label: "Singapore", Regex: `(?i)🇸🇬|新加坡|狮城|獅城|(?<![a-z])sg(?![a-z])|singapore`},
	{Label: "United States", Regex: `(?i)🇺🇸|美国|美國|美西|美东|美東|洛杉矶|圣何塞|西雅图|硅谷|纽约|(?<![a-z])usa?(?![a-z])|united\s*states`},
	{Label: "South Korea", Regex: `(?i)🇰🇷|韩国|韓國|首尔|首爾|(?<![a-z])kr(?![a-z])|korea|seoul`},
	{Label: "United Kingdom", Regex: `(?i)🇬🇧|英国|英國|伦敦|倫敦|(?<![a-z])(uk|gb)(?![a-z])|united\s*kingdom|britain|london`},
	{Label: "Remaining Traffic", Regex: `(?i)剩余|剩餘|已用|流量\s*[:：]|traffic|remaining`},
	{Label: "Expiry Date", Regex: `(?i)到期|过期|過期|有效期|expir`},
	{Label: "Traffic Reset", Regex: `(?i)重置|reset`},
	{Label: "Website & Notices", Regex: `(?i)官网|官網|网址|網址|域名|公告|套餐|客服|群组|群組|频道|頻道|website|telegram|(?<![a-z])tg(?![a-z])`},
}

// DashboardWidget names the editable dashboard cards.
var DefaultDashboardWidgets = []string{
	"currentOutbound", "networkSpeed", "outboundMode", "networkTakeover",
	"intranetIp", "serviceStatus", "profiles", "proxyGroups", "connectionStats",
}

// ---- Profile ----

type SubscriptionInfo struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
	Expire   int64 `json:"expire"`
}

type ProfileOverrides struct {
	Dns                 json.RawMessage   `json:"dns,omitempty"`
	DnsOverrideKeys     []string          `json:"dns-override-keys,omitempty"`
	Ntp                 json.RawMessage   `json:"ntp,omitempty"`
	NtpOverrideKeys     []string          `json:"ntp-override-keys,omitempty"`
	Sniffer             json.RawMessage   `json:"sniffer,omitempty"`
	SnifferOverrideKeys []string          `json:"sniffer-override-keys,omitempty"`
	Tun                 json.RawMessage   `json:"tun,omitempty"`
	TunOverrideKeys     []string          `json:"tun-override-keys,omitempty"`
	ProxyProviders      map[string]any    `json:"proxy-providers,omitempty"`
	RuleSets            map[string]any    `json:"rule-sets,omitempty"`
	Empty               map[string]any    `json:"-"`
	Extra               map[string]string `json:"-"`
}

type Profile struct {
	ID               int64             `json:"id"`
	Type             string            `json:"type"` // file | url | custom
	Label            string            `json:"label"`
	CurrentGroupName string            `json:"currentGroupName,omitempty"`
	URL              string            `json:"url"`
	AgeSecretKey     string            `json:"ageSecretKey"`
	LastUpdateDate   int64             `json:"lastUpdateDate,omitempty"` // ms
	AutoUpdate       bool              `json:"autoUpdate"`
	AutoUpdateMins   int               `json:"autoUpdateDuration"` // minutes
	SubscriptionInfo *SubscriptionInfo `json:"subscriptionInfo,omitempty"`
	SelectedMap      map[string]string `json:"selectedMap"`
	UnfoldSet        []string          `json:"unfoldSet"`
	ExtendType       string            `json:"extendType"` // standard | script
	ScriptID         *int64            `json:"scriptId,omitempty"`
	MatchTarget      *string           `json:"matchTarget,omitempty"`
	Order            string            `json:"order,omitempty"`
	Overrides        ProfileOverrides  `json:"overrides"`
}

// ---- Config (lib/models/config.dart) ----

type TextScale struct {
	Enable bool    `json:"enable"`
	Scale  float64 `json:"scale"`
}

type AppSettingProps struct {
	Locale                 *string          `json:"locale,omitempty"`
	DashboardWidgets       []string         `json:"dashboardWidgets"`
	OnlyStatisticsProxy    bool             `json:"onlyStatisticsProxy"`
	ShowNotificationStop   bool             `json:"showNotificationStopAction"`
	AutoLaunch             bool             `json:"autoLaunch"`
	SilentLaunch           bool             `json:"silentLaunch"`
	AutoRun                bool             `json:"autoRun"`
	OpenLogs               bool             `json:"openLogs"`
	CloseConnections       bool             `json:"closeConnections"`
	RequestFeed            string           `json:"requestFeed"` // live | recent
	DelayTestTimeout       int              `json:"delayTestTimeout"`
	TabAnimation           string           `json:"tabAnimation"` // slide | fade
	FloatingNavigationBar  bool             `json:"floatingNavigationBar"`
	AutoCheckUpdate        bool             `json:"autoCheckUpdate"`
	SidebarExpanded        bool             `json:"sidebarExpanded"`
	AcceptedDisclaimer     int              `json:"acceptedDisclaimerVersion"`
	MinimizeOnExit         bool             `json:"minimizeOnExit"`
	Hidden                 bool             `json:"hidden"`
	DeveloperMode          bool             `json:"developerMode"`
	RestoreStrategy        string           `json:"restoreStrategy"`
	ShowTrayTitle          bool             `json:"showTrayTitle"`
	MonochromeTrayIcon     bool             `json:"monochromeTrayIcon"`
	CheckCertificate       bool             `json:"checkCertificate"`
	UserAgents             []string         `json:"userAgents"`
	Filters                []Filter         `json:"filters"`
	HideIP                 bool             `json:"hideIp"`
	EditorLineWrap         bool             `json:"editorLineWrap"`
	EditorFontSize         string           `json:"editorFontSize"`
	ServiceOrder           []string         `json:"serviceOrder"`
	DisabledServices       []string         `json:"disabledServices"`
	TableColumnWidths      map[string]any   `json:"tableColumnWidths"`
	RecordGroupings        map[string]any   `json:"recordGroupings"`
	ResourceUpdateInterval int              `json:"resourceUpdateInterval"`
}

type ThemeProps struct {
	PrimaryColor  *int64    `json:"primaryColor,omitempty"`
	PrimaryColors []int64   `json:"primaryColors"`
	ThemeMode     string    `json:"themeMode"` // system | light | dark
	ColorStyle    string    `json:"colorStyle"`
	PureBlack     bool      `json:"pureBlack"`
	SidebarBlur   bool      `json:"sidebarBlur"`
	LiquidGlass   string    `json:"liquidGlass"`
	TextScale     TextScale `json:"textScale"`
	SystemFont    *string   `json:"systemFontFamily,omitempty"`
}

type ProxiesStyleProps struct {
	Type              string `json:"type"` // tab | list
	SortType          string `json:"sortType"`
	Layout            string `json:"layout"`
	IconStyle         string `json:"iconStyle"`
	CardType          string `json:"cardType"`
	HideTimeoutProxys bool   `json:"hideTimeoutProxies"`
}

type AuthenticationProps struct {
	Enable   bool   `json:"enable"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a AuthenticationProps) Credentials() []string {
	if a.Enable && a.Username != "" {
		return []string{a.Username + ":" + a.Password}
	}
	return nil
}

type NetworkProps struct {
	SystemProxy        bool                `json:"systemProxy"`
	BypassDomain       []string            `json:"bypassDomain"`
	BypassPrivateRoute bool                `json:"bypassPrivateRoute"`
	AutoSetSystemDNS   bool                `json:"autoSetSystemDns"`
	AppendSystemDNS    bool                `json:"appendSystemDns"`
	Authentication     AuthenticationProps `json:"authentication"`
}

type VpnProps struct {
	Enable     bool `json:"enable"`
	SystemProxy bool `json:"systemProxy"`
	IPv6       bool `json:"ipv6"`
	AllowBypass bool `json:"allowBypass"`
	DNSHijacking bool `json:"dnsHijacking"`
}

// ---- clash config patch (lib/models/clash_config.dart) ----

type Tun struct {
	Enable        bool     `json:"enable"`
	Device        string   `json:"device"`
	Stack         string   `json:"stack"`
	RouteAddress  []string `json:"route-address"`
	MTU           int      `json:"mtu"`
	DNSHijack     []string `json:"dns-hijack"`
	AutoRoute     bool     `json:"auto-route"`
	AutoDetectInterface     bool `json:"auto-detect-interface"`
	StrictRoute   bool     `json:"strict-route"`
}

type Dns struct {
	Enable            bool              `json:"enable"`
	Listen            string            `json:"listen"`
	ListenRoutingMark int               `json:"listen-routing-mark"`
	PreferH3          bool              `json:"prefer-h3"`
	UseHosts          bool              `json:"use-hosts"`
	UseSystemHosts    bool              `json:"use-system-hosts"`
	RespectRules      bool              `json:"respect-rules"`
	IPv6              bool              `json:"ipv6"`
	IPv6Timeout       int               `json:"ipv6-timeout"`
	CacheAlgorithm    string            `json:"cache-algorithm"`
	CacheMaxSize      int               `json:"cache-max-size"`
	DefaultNameserver []string          `json:"default-nameserver"`
	EnhancedMode      string            `json:"enhanced-mode"`
	FakeIPRange       string            `json:"fake-ip-range"`
	FakeIPFilter      []string          `json:"fake-ip-filter"`
	FakeIPFilterMode  string            `json:"fake-ip-filter-mode"`
	FakeIPTTL         int               `json:"fake-ip-ttl"`
	NameserverPolicy  map[string]string `json:"nameserver-policy"`
	Nameserver        []string          `json:"nameserver"`
	Fallback          []string          `json:"fallback"`
	FallbackLazyQuery bool              `json:"fallback-lazy-query"`
	ProxyServerNameserver []string     `json:"proxy-server-nameserver"`
	ProxyServerNameserverPolicy map[string]string `json:"proxy-server-nameserver-policy"`
	DirectNameserver  []string          `json:"direct-nameserver"`
	DirectNameserverFollowPolicy bool   `json:"direct-nameserver-follow-policy"`
	FallbackFilter    FallbackFilter    `json:"fallback-filter"`
}

type FallbackFilter struct {
	GeoIP     bool     `json:"geoip"`
	GeoIPCode string   `json:"geoip-code"`
	Geosite   []string `json:"geosite"`
	IPCIDR    []string `json:"ipcidr"`
	Domain    []string `json:"domain"`
}

type ProfileTun struct {
	MTU                 int      `json:"mtu"`
	CongestionController string  `json:"congestion-controller"`
	StrictRoute         bool     `json:"strict-route"`
	RouteAddress        []string `json:"route-address"`
	RouteExcludeAddress []string `json:"route-exclude-address"`
	DNSHijack           []string `json:"dns-hijack"`
	UDPTimeout          int      `json:"udp-timeout"`
}

type PatchClashConfig struct {
	MixedPort           int               `json:"mixed-port"`
	SocksPort           int               `json:"socks-port"`
	Port                int               `json:"port"`
	RedirPort           int               `json:"redir-port"`
	TproxyPort          int               `json:"tproxy-port"`
	Mode                string            `json:"mode"` // rule | global | direct
	AllowLan            bool              `json:"allow-lan"`
	LogLevel            string            `json:"log-level"`
	IPv6                bool              `json:"ipv6"`
	FindProcessMode     string            `json:"find-process-mode"`
	InterfaceNameMode   string            `json:"interface-name-mode"`
	InterfaceName       string            `json:"interface-name"`
	KeepAliveInterval   int               `json:"keep-alive-interval"`
	KeepAliveIdle       int               `json:"keep-alive-idle"`
	DisableKeepAlive    bool              `json:"disable-keep-alive"`
	RoutingMark         int               `json:"routing-mark"`
	UnifiedDelay        bool              `json:"unified-delay"`
	TCPConcurrent       bool              `json:"tcp-concurrent"`
	Tun                 Tun               `json:"tun"`
	Dns                 Dns               `json:"dns"`
	DnsOverrideKeys     []string          `json:"dns-override-keys,omitempty"`
	TunOverride         ProfileTun        `json:"tun-override"`
	TunOverrideKeys     []string          `json:"tun-override-keys,omitempty"`
	GeoXUrl             map[string]string `json:"geox-url"`
	GeodataLoader       string            `json:"geodata-loader"`
	GlobalUa            *string           `json:"global-ua,omitempty"`
	ExternalController  string            `json:"external-controller"`
	Secret              string            `json:"secret"`
	Hosts               map[string]string `json:"hosts"`
	GeoAutoUpdate       bool              `json:"geo-auto-update"`
	GeoUpdateInterval   int               `json:"geo-update-interval"`
	TestUrl             string            `json:"test-url"`
	GeneralOverrideKeys []string          `json:"general-override-keys,omitempty"`
}

type Config struct {
	CurrentProfileID  *int64            `json:"currentProfileId"`
	HotKeyActions     []any             `json:"hotKeyActions"`
	AppSetting        AppSettingProps   `json:"appSettingProps"`
	DAVProps          *any              `json:"davProps,omitempty"`
	NetworkSetting    NetworkProps      `json:"networkProps"`
	VpnSetting        VpnProps          `json:"vpnProps"`
	ThemeSetting      ThemeProps        `json:"themeProps"`
	ProxiesStyle      ProxiesStyleProps `json:"proxiesStyleProps"`
	WindowGeometry    map[string]any    `json:"windowGeometry"`
	PatchClashConfig  PatchClashConfig  `json:"patchClashConfig"`
	ExcludeSSIDs      []string          `json:"excludeSSIDs"`
}

// DefaultConfig mirrors Config.realFromJson for a first run.
func DefaultConfig() Config {
	locale := ""
	return Config{
		HotKeyActions: []any{},
		AppSetting: AppSettingProps{
			DashboardWidgets:     DefaultDashboardWidgets,
			ShowNotificationStop: true,
			CloseConnections:     true,
			RequestFeed:          "live",
			DelayTestTimeout:     DefaultDelayTestTimeout,
			TabAnimation:         "fade",
			FloatingNavigationBar: true,
			AutoCheckUpdate:      true,
			SidebarExpanded:      true,
			MinimizeOnExit:       true,
			RestoreStrategy:      "compatible",
			ShowTrayTitle:        true,
			MonochromeTrayIcon:   true,
			CheckCertificate:     true,
			UserAgents:           DefaultUserAgents,
			Filters:              DefaultFilters,
			Locale:               &locale,
		},
		NetworkSetting: NetworkProps{
			SystemProxy:        true,
			BypassDomain:       DefaultBypassDomain,
			BypassPrivateRoute: true,
			AutoSetSystemDNS:   true,
			Authentication:     AuthenticationProps{},
		},
		VpnSetting:     VpnProps{Enable: true, SystemProxy: true, AllowBypass: true},
		ThemeSetting: ThemeProps{
			PrimaryColors: DefaultPrimaryColors,
			ThemeMode:     "light",
			ColorStyle:    "vivid",
			LiquidGlass:   "none",
			TextScale:     TextScale{},
		},
		ProxiesStyle:     ProxiesStyleProps{Type: "tab", SortType: "none", Layout: "standard", IconStyle: "filled", CardType: "standard"},
		WindowGeometry:   map[string]any{},
		PatchClashConfig: DefaultPatchClashConfig(),
		ExcludeSSIDs:     []string{},
	}
}

var defaultDNSOverrideKeys = []string{
	"enable", "listen", "use-hosts", "use-system-hosts", "ipv6", "respect-rules",
	"prefer-h3", "enhanced-mode", "fake-ip-range", "fake-ip-filter",
	"default-nameserver", "nameserver-policy", "nameserver", "fallback",
	"proxy-server-nameserver", "fallback-filter.geoip", "fallback-filter.geoip-code",
	"fallback-filter.geosite", "fallback-filter.ipcidr", "fallback-filter.domain",
}

func DefaultPatchClashConfig() PatchClashConfig {
	return PatchClashConfig{
		MixedPort:         DefaultMixedPort,
		Mode:              "rule",
		LogLevel:          "error",
		FindProcessMode:   "off",
		InterfaceNameMode: "clear",
		KeepAliveInterval: DefaultKeepAlive,
		KeepAliveIdle:     15,
		UnifiedDelay:      true,
		TCPConcurrent:     true,
		Tun: Tun{
			Device:              "FlClash",
			Stack:               "mips",
			MTU:                 DefaultTunMTU,
			DNSHijack:           []string{"any:53"},
			AutoRoute:           true,
			AutoDetectInterface: true,
		},
		Dns: Dns{
			Listen:            "",
			UseHosts:          true,
			UseSystemHosts:    true,
			IPv6Timeout:       100,
			CacheAlgorithm:    "lru",
			CacheMaxSize:      4096,
			DefaultNameserver: []string{"114.114.114.114", "223.5.5.5", "8.8.8.8", "1.0.0.1"},
			EnhancedMode:      "redir-host",
			FakeIPRange:       "198.18.0.1/16",
			FakeIPFilter:      []string{"dns.msftnsci.com", "www.msftnsci.com", "www.msftconnecttest.com"},
			FakeIPFilterMode:  "blacklist",
			FakeIPTTL:         1,
			NameserverPolicy:  map[string]string{},
			Nameserver:        []string{"https://doh.pub/dns-query", "tls://223.5.5.5:853"},
			FallbackLazyQuery: false,
			ProxyServerNameserverPolicy: map[string]string{},
			FallbackFilter: FallbackFilter{GeoIP: true, GeoIPCode: "CN"},
		},
		DnsOverrideKeys:   defaultDNSOverrideKeys,
		GeoXUrl:           DefaultGeoXUrl,
		GeodataLoader:     "memconservative",
		ExternalController: "close",
		Hosts:             map[string]string{},
		GeoUpdateInterval: 24,
		TestUrl:           DefaultTestURL,
	}
}
