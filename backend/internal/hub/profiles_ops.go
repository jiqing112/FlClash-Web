package hub

// Profile operations: add/edit/delete, subscription download, validation.
// Mirrors the Dart Profile model's update/saveFile flow.

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"flclashweb/backend/internal/models"
	"flclashweb/backend/internal/rpc"
)

// snowflake-ish id generation: ms timestamp shifted, like the Dart side's ids.
func nextID() int64 {
	return time.Now().UnixMilli()<<10 | int64(nextIDSeq)%1024
}

var nextIDSeq int64

func init() { nextIDSeq = time.Now().UnixNano() & 1023 }

func (h *Hub) profilesDir() string {
	return filepath.Join(h.dataDir, "profiles")
}

func (h *Hub) profilePath(id int64) string {
	return filepath.Join(h.profilesDir(), fmt.Sprintf("%d.yaml", id))
}

// AddProfileInput matches the AddProfileSheet options.
type AddProfileInput struct {
	URL          string          `json:"url"`
	Label        string          `json:"label"`
	Type         string          `json:"type"` // url | file | custom
	FileContent  json.RawMessage `json:"fileContent"` // base64 or text
	AutoUpdate   *bool           `json:"autoUpdate"`
	AutoUpdateMins int           `json:"autoUpdateDuration"`
}

// AddProfile creates a profile; for a URL it downloads and validates.
func (h *Hub) AddProfile(input AddProfileInput) (*ProfileState, error) {
	profile := models.Profile{
		ID:             nextID(),
		AutoUpdate:     true,
		AutoUpdateMins: orDefault(input.AutoUpdateMins, 120),
		SelectedMap:    map[string]string{},
		UnfoldSet:      []string{},
		ExtendType:     "standard",
		Label:          input.Label,
		URL:            input.URL,
	}
	if input.AutoUpdate != nil {
		profile.AutoUpdate = *input.AutoUpdate
	}
	switch {
	case input.Type == "file" || (input.Type == "" && input.URL == "" && len(input.FileContent) > 0):
		profile.Type = "file"
		content := decodeFileContent(input.FileContent)
		if err := h.validateProfileContent(content, profile.AgeSecretKey); err != nil {
			return nil, err
		}
		if profile.Label == "" {
			profile.Label = "Profile"
		}
		profile.LastUpdateDate = time.Now().UnixMilli()
		if err := os.WriteFile(h.profilePath(profile.ID), content, 0o644); err != nil {
			return nil, err
		}
	case input.Type == "custom":
		profile.Type = "custom"
		profile.LastUpdateDate = time.Now().UnixMilli()
	default:
		profile.Type = "url"
	state, err := h.downloadProfile(&profile)
	if err != nil {
		return nil, err
	}
	profile = state.Profile
	}
	state := &ProfileState{Profile: profile}
	h.PutProfile(*state)
	return state, nil
}

func decodeFileContent(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return []byte(text)
	}
	// base64 payload
	var b64 string
	if json.Unmarshal(raw, &b64) == nil {
		if data, err := base64.StdEncoding.DecodeString(b64); err == nil {
			return data
		}
	}
	return raw
}

// downloadProfile fetches the subscription, honoring the userinfo header and
// content-disposition naming, like Profile.update in Dart.
func (h *Hub) downloadProfile(profile *models.Profile) (*ProfileState, error) {
	if profile.URL == "" {
		return nil, fmt.Errorf("empty url")
	}
	content, disposition, userinfo, err := fetchSubscription(profile.URL, h.Config().AppSetting.UserAgents)
	if err != nil {
		return nil, err
	}
	content = decompressUnlabeled(content)
	if err := h.validateProfileContent(content, profile.AgeSecretKey); err != nil {
		return nil, err
	}
	if profile.Label == "" {
		profile.Label = fileNameForDisposition(disposition)
		if profile.Label == "" {
			if uri, err := url.Parse(profile.URL); err == nil && uri.Host != "" {
				profile.Label = uri.Host
			}
		}
	}
	if userinfo != "" {
		profile.SubscriptionInfo = parseSubscriptionInfo(userinfo)
	}
	profile.LastUpdateDate = time.Now().UnixMilli()
	if err := os.WriteFile(h.profilePath(profile.ID), content, 0o644); err != nil {
		return nil, err
	}
	return &ProfileState{Profile: *profile}, nil
}

// UpdateProfile re-downloads a URL profile.
func (h *Hub) UpdateProfile(id int64) (*ProfileState, error) {
	profile := h.Profile(id)
	if profile == nil {
		return nil, fmt.Errorf("profile not found")
	}
	if profile.Type != "url" {
		return profile, nil
	}
	h.markUpdating(id, true)
	defer h.markUpdating(id, false)
	state, err := h.downloadProfile(&profile.Profile)
	if err != nil {
		return nil, err
	}
	// keep local selections
	state.SelectedMap = profile.SelectedMap
	state.UnfoldSet = profile.UnfoldSet
	h.PutProfile(*state)
	// a synced subscription changes the running config; the original
	// debounce-applies it right away
	if h.Config().CurrentProfileID != nil && *h.Config().CurrentProfileID == id {
		_ = h.ApplyProfile(true)
	}
	return state, nil
}

func (h *Hub) markUpdating(id int64, updating bool) {
	h.mu.Lock()
	for _, p := range h.profiles {
		if p.ID == id {
			p.Updating = updating
		}
	}
	h.mu.Unlock()
	h.emitter("profiles", h.Profiles())
}

func (h *Hub) validateProfileContent(content []byte, ageKey string) error {
	temp := filepath.Join(h.dataDir, ".validate.yaml")
	if err := os.WriteFile(temp, content, 0o644); err != nil {
		return err
	}
	defer os.Remove(temp)
	if ageKey != "" {
		params := map[string]any{"path": temp, "age-secret-key": ageKey}
		if _, err := h.client.Call("config.decrypt", params, 0); err != nil {
			// not age-encrypted: fall through to plain validation
		}
	}
	params := map[string]any{"path": temp}
	raw, err := h.client.Call("config.validate", params, 0)
	if err != nil {
		return err
	}
	if raw != nil && string(raw) != "null" && len(raw) > 0 {
		var message string
		if json.Unmarshal(raw, &message) == nil && message != "" {
			return fmt.Errorf("%s", message)
		}
	}
	return nil
}

var (
	client = &http.Client{Timeout: 60 * time.Second}
)

func fetchSubscription(target string, userAgents []string) ([]byte, string, string, error) {
	uri, headers, err := downloadTarget(target, userAgents)
	if err != nil {
		return nil, "", "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, "", "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("http %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, "", "", err
	}
	return data, resp.Header.Get("content-disposition"), resp.Header.Get("subscription-userinfo"), nil
}

// downloadTarget mirrors Dart's downloadTarget: userinfo becomes Basic auth.
func downloadTarget(target string, userAgents []string) (string, map[string]string, error) {
	uri, err := url.Parse(target)
	if err != nil {
		return "", nil, err
	}
	headers := map[string]string{"User-Agent": pickUA(userAgents)}
	if uri.User != nil {
		pass, _ := uri.User.Password()
		credentials := uri.User.Username() + ":" + pass
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
		uri.User = nil
	}
	return uri.String(), headers, nil
}

func pickUA(userAgents []string) string {
	if len(userAgents) > 0 && userAgents[0] != "" {
		return userAgents[0]
	}
	return defaultUA
}

// decompressUnlabeled mirrors Dart's decompressUnlabeled.
func decompressUnlabeled(data []byte) []byte {
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err == nil {
			if out, err := io.ReadAll(reader); err == nil {
				return out
			}
		}
		return data
	}
	if len(data) >= 2 && data[0] == 0x78 && (int(data[0])<<8|int(data[1]))%31 == 0 {
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err == nil {
			if out, err := io.ReadAll(reader); err == nil {
				return out
			}
		}
	}
	return data
}

var filenameStarRe = regexp.MustCompile(`filename\*=UTF-8''([^;]+)`)
var filenameRe = regexp.MustCompile(`filename="?([^";]+)"?`)

func fileNameForDisposition(disposition string) string {
	if disposition == "" {
		return ""
	}
	if match := filenameStarRe.FindStringSubmatch(disposition); match != nil {
		if decoded, err := url.PathUnescape(match[1]); err == nil {
			return decoded
		}
		return match[1]
	}
	if match := filenameRe.FindStringSubmatch(disposition); match != nil {
		return match[1]
	}
	return ""
}

// parseSubscriptionInfo mirrors SubscriptionInfo.formHString.
func parseSubscriptionInfo(info string) *models.SubscriptionInfo {
	out := models.SubscriptionInfo{}
	for _, field := range strings.Split(info, ";") {
		sep := strings.Index(field, "=")
		if sep < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(field[:sep]))
		value := strings.TrimSpace(field[sep+1:])
		count := parseCount(value)
		switch key {
		case "upload":
			out.Upload = count
		case "download":
			out.Download = count
		case "total":
			out.Total = count
		case "expire":
			out.Expire = count
		}
	}
	return &out
}

func parseCount(value string) int64 {
	var count int64
	if _, err := fmt.Sscanf(value, "%d", &count); err == nil {
		return count
	}
	var decimal float64
	if _, err := fmt.Sscanf(value, "%f", &decimal); err == nil {
		return int64(decimal)
	}
	return 0
}

var _ = rpc.SetupParams{}

func orDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
