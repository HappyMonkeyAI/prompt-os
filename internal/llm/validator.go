package llm

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
)

var ErrInvalidBlueprint = errors.New("llm: invalid blueprint JSON")
var ErrUnsafePath = errors.New("llm: unsafe config path detected")

// ValidateBlueprint checks that the JSON decodes into a valid Blueprint
// and performs basic sanity checks including path safety.
func ValidateBlueprint(data []byte) (*Blueprint, error) {
	var bp Blueprint
	if err := json.Unmarshal(data, &bp); err != nil {
		return nil, ErrInvalidBlueprint
	}

	if bp.BaseDistro == "" {
		return nil, ErrInvalidBlueprint
	}
	if bp.StabilityPreference == "" {
		bp.StabilityPreference = "stable"
	}

	validDistros := map[string]bool{"arch": true, "ubuntu": true, "debian": true}
	if !validDistros[bp.BaseDistro] {
		return nil, ErrInvalidBlueprint
	}

	// Path traversal protection for configs map
	for path := range bp.Configs {
		if strings.Contains(path, "..") || !filepath.IsAbs(path) {
			return nil, ErrUnsafePath
		}
		// Restrict to safe directories only
		if !strings.HasPrefix(path, "/etc/") && !strings.HasPrefix(path, "/opt/") {
			return nil, ErrUnsafePath
		}
	}

	return &bp, nil
}