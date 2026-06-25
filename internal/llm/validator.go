package llm

import (
	"encoding/json"
	"errors"
)

var ErrInvalidBlueprint = errors.New("llm: invalid blueprint JSON")

// ValidateBlueprint checks that the JSON decodes into a valid Blueprint
// and performs basic sanity checks.
func ValidateBlueprint(data []byte) (*Blueprint, error) {
	var bp Blueprint
	if err := json.Unmarshal(data, &bp); err != nil {
		return nil, ErrInvalidBlueprint
	}

	// Basic validation rules (expand in later phases)
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

	return &bp, nil
}