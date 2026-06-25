# PromptOS Blueprint JSON Schema (Draft)

**Version:** 0.1 (2026-06-25)

This is the contract the LLM must return after the wizard completes.

```json
{
  "base_distro": "arch" | "ubuntu" | "debian",
  "stability_preference": "bleeding" | "stable",
  "display": {
    "server": "wayland" | "x11",
    "manager": "gdm" | "sddm" | "lightdm"
  },
  "packages": ["string"],
  "drivers": {
    "gpu": "nvidia" | "amd" | "intel" | "none",
    "extra": ["string"]
  },
  "configs": {
    "/etc/some.conf": "file contents...",
    "/etc/environment.d/ai-keys.conf": "..."
  },
  "services": {
    "enable": ["string"],
    "disable": ["string"]
  },
  "remote_access": {
    "enabled": boolean,
    "method": "opencloud" | "other"
  }
}
```

**Validation Rules**
- `packages` and `drivers.extra` must be valid for the chosen base_distro.
- All config paths must be absolute.
- Keys in `configs` for AI providers must be written to protected locations only.

This schema will be refined after more research and first LLM tests.