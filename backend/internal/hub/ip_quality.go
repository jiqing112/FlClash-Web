package hub

// Port of lib/common/ip_quality.dart: quality/type lookup for the exit IP,
// asked from the same sources in the same waves (the tightest quotas last).

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type IpQuality struct {
	IP           string `json:"ip"`
	Type         string `json:"type"` // residential | hosting | mobile | business | education | government | military
	Organization string `json:"organization,omitempty"`
	ASN          string `json:"asn,omitempty"`
	IsProxy      bool   `json:"isProxy"`
	IsVpn        bool   `json:"isVpn"`
	IsTor        bool   `json:"isTor"`
	IsAbuser     bool   `json:"isAbuser"`
}

// Level mirrors IpQualityExt.level: risky when tor/abuser, good for
// residential/mobile/business, normal for hosting.
func (q IpQuality) Level() string {
	if q.IsTor || q.IsAbuser {
		return "risky"
	}
	switch q.Type {
	case "residential", "mobile", "business":
		return "good"
	case "hosting":
		return "normal"
	default:
		return "normal"
	}
}

var qualityClient = &http.Client{Timeout: 8 * time.Second}

type qualityAnswer struct {
	quality IpQuality
	ok      bool
	at      time.Time
}

func declaredType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "hosting", "datacenter", "data center", "business", "isp", "residential",
		"mobile", "education", "government", "military":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func pickType(hosting, mobile bool) string {
	if hosting {
		return "hosting"
	}
	if mobile {
		return "mobile"
	}
	return ""
}

func lookupIdentMe(ip string) (IpQuality, bool) {
	data, err := fetchJSON("https://ident.me/json")
	if err != nil || data["ip"] != ip {
		return IpQuality{}, false
	}
	typeName := declaredType(text(data["type"]))
	if typeName == "" {
		return IpQuality{}, false
	}
	return IpQuality{
		IP:           ip,
		Type:         typeName,
		Organization: text(data["aso"]),
		ASN:          asnText(data["asn"]),
	}, true
}

func lookupIpApiCom(ip string) (IpQuality, bool) {
	data, err := fetchJSON(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,isp,org,as,mobile,proxy,hosting", ip))
	if err != nil || data["status"] != "success" {
		return IpQuality{}, false
	}
	hosting := flag(data["hosting"])
	mobile := flag(data["mobile"])
	typeName := pickType(hosting, mobile)
	if typeName == "" {
		typeName = "residential"
	}
	return IpQuality{
		IP:           ip,
		Type:         typeName,
		Organization: text(data["org"]),
		ASN:          asnText(data["as"]),
		IsProxy:      flag(data["proxy"]),
	}, true
}

func lookupIpQuery(ip string) (IpQuality, bool) {
	data, err := fetchJSON(fmt.Sprintf("https://api.ipquery.io/%s?format=json", ip))
	if err != nil {
		return IpQuality{}, false
	}
	risk, _ := data["risk"].(map[string]any)
	isp, _ := data["isp"].(map[string]any)
	hosting := flag(risk["is_datacenter"])
	mobile := flag(risk["is_mobile"])
	typeName := pickType(hosting, mobile)
	if typeName == "" {
		typeName = "residential"
	}
	organization := ""
	asnValue := ""
	if isp != nil {
		organization = text(isp["org"])
		if organization == "" {
			organization = text(isp["isp"])
		}
		asnValue = asnText(isp["asn"])
	}
	return IpQuality{
		IP:           ip,
		Type:         typeName,
		Organization: organization,
		ASN:          asnValue,
		IsProxy:      flag(risk["is_proxy"]),
		IsVpn:        flag(risk["is_vpn"]),
		IsTor:        flag(risk["is_tor"]),
	}, true
}

func fetchJSON(url string) (map[string]any, error) {
	resp, err := qualityClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func text(value any) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

func asnText(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("AS%d", int64(v))
	default:
		return ""
	}
}

func flag(value any) bool {
	checked, ok := value.(bool)
	return ok && checked
}

var (
	qualityCacheMu sync.Mutex
	qualityCache   = map[string]qualityAnswer{}
)

// IPQuality looks up quality for the exit IP, cached for an hour, from the
// same wave-one sources the app asks first.
func (h *Hub) IPQuality(ip string) (IpQuality, bool) {
	if ip == "" {
		return IpQuality{}, false
	}
	qualityCacheMu.Lock()
	if cached, ok := qualityCache[ip]; ok && time.Since(cached.at) < time.Hour {
		qualityCacheMu.Unlock()
		return cached.quality, cached.ok
	}
	qualityCacheMu.Unlock()

	lookup := func(source func(string) (IpQuality, bool)) (IpQuality, bool) {
		if quality, ok := source(ip); ok {
			return quality, true
		}
		return IpQuality{}, false
	}
	quality, ok := lookup(lookupIdentMe)
	if !ok {
		quality, ok = lookup(lookupIpQuery)
	}
	if !ok {
		quality, ok = lookup(lookupIpApiCom)
	}

	qualityCacheMu.Lock()
	qualityCache[ip] = qualityAnswer{quality: quality, ok: ok, at: time.Now()}
	if len(qualityCache) > 64 {
		// drop the oldest quarter
		for key := range qualityCache {
			delete(qualityCache, key)
			if len(qualityCache) <= 48 {
				break
			}
		}
	}
	qualityCacheMu.Unlock()
	return quality, ok
}
