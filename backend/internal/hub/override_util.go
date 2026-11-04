package hub

import (
	"encoding/json"
	"time"
	"os"
	"strings"

	"flclashweb/backend/internal/models"

	"go.yaml.in/yaml/v3"
)

type modelsPatchClashConfig = models.PatchClashConfig
type modelsProfileOverrides = models.ProfileOverrides

func osReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

func yamlUnmarshal(data []byte) (map[string]any, error) {
	var out map[string]any
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func yamlMarshal(value map[string]any) (string, error) {
	// Node-based encode keeps every scalar as written and sorts nothing: yaml.v3
	// marshals maps in sorted key order, matching the Dart encoder closely
	// enough for the core.
	encoded, err := yaml.Marshal(value)
	return string(encoded), err
}

// overrideJSONFromRaw decodes a stored override section (models raw json).
func overrideJSONFromRaw(raw json.RawMessage) map[string]any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var out map[string]any
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	return out
}

func tunOverrideJSON(t models.ProfileTun) map[string]any {
	data, _ := json.Marshal(t)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	return out
}

// generalOverrideFragment mirrors PatchClashConfigExt.generalOverrideJson: the
// picked general keys only, test-url excluded.
func generalOverrideFragment(patch *models.PatchClashConfig) map[string]any {
	data, _ := json.Marshal(patch)
	var full map[string]any
	if json.Unmarshal(data, &full) != nil {
		return nil
	}
	picked := map[string]bool{}
	for _, key := range patch.GeneralOverrideKeys {
		if key == "test-url" {
			continue
		}
		picked[key] = true
	}
	fragment := map[string]any{}
	for _, key := range []string{"test-url", "unified-delay", "tcp-concurrent", "ipv6", "geodata-loader"} {
		if !picked[key] {
			continue
		}
		fragment[key] = full[key]
	}
	return fragment
}

// overrideFragment mirrors _overrideJson: only the selected keys survive, a
// dotted key nests under its parent, unset values drop out.
func overrideFragment(section map[string]any, keys []string) map[string]any {
	return overrideFragmentKeepEmpty(section, keys, false)
}

func overrideFragmentKeepEmpty(section map[string]any, keys []string, keepEmpty bool) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		var value any
		if dot := strings.Index(key, "."); dot >= 0 {
			parent, field := key[:dot], key[dot+1:]
			nested, ok := section[parent].(map[string]any)
			if !ok {
				continue
			}
			value = nested[field]
		} else {
			value = section[key]
		}
		if isUnset(value) && !keepEmpty {
			continue
		}
		if dot := strings.Index(key, "."); dot >= 0 {
			parent, field := key[:dot], key[dot+1:]
			nested, ok := out[parent].(map[string]any)
			if !ok {
				nested = map[string]any{}
				out[parent] = nested
			}
			nested[field] = value
		} else {
			out[key] = value
		}
	}
	return out
}

func isUnset(value any) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

func timeNow() int64 { return time.Now().UnixMilli() }

// normalizeDelayTimeout treats sub-second values as seconds (legacy storage).
func normalizeDelayTimeout(value int) int64 {
	if value > 0 && value < 1000 {
		return int64(value) * 1000
	}
	if value <= 0 {
		return 8000
	}
	return int64(value)
}
