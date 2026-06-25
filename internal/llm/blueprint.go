package llm

// Blueprint represents the JSON contract returned by the LLM.
// It mirrors docs/blueprint-schema.md (Phase 0.3).
type Blueprint struct {
	BaseDistro         string                 `json:"base_distro"`
	StabilityPreference string                `json:"stability_preference"`
	Display            DisplayConfig          `json:"display"`
	Packages           []string               `json:"packages"`
	Drivers            DriverConfig           `json:"drivers"`
	Configs            map[string]string      `json:"configs"`
	Services           ServiceConfig          `json:"services"`
	RemoteAccess       RemoteAccessConfig     `json:"remote_access"`
}

type DisplayConfig struct {
	Server  string `json:"server"`
	Manager string `json:"manager"`
}

type DriverConfig struct {
	GPU   string   `json:"gpu"`
	Extra []string `json:"extra"`
}

type ServiceConfig struct {
	Enable  []string `json:"enable"`
	Disable []string `json:"disable"`
}

type RemoteAccessConfig struct {
	Enabled bool   `json:"enabled"`
	Method  string `json:"method"`
}