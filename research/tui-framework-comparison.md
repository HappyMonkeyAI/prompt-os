# TUI Framework Comparison for PromptOS (Phase 0.1)

**Date:** 2026-06-25

**Frameworks Evaluated**
- Go + Bubble Tea (+ Lip Gloss, Bubbles)
- Python + Textual (+ Rich)

**Key Criteria for PromptOS**
- Binary / package size on a minimal live ISO (< 300 MB target)
- RAM usage during wizard
- Ease of structured/JSON output + LLM integration
- Packaging simplicity for Alpine live or archiso
- Static binary vs interpreter/runtime requirements

**Findings (from web research)**

**Go + Bubble Tea**
- Produces a single static binary (typically 8–20 MB for a full TUI app).
- No runtime dependencies beyond the kernel — ideal for tiny live ISOs.
- Excellent for Alpine Linux live environments.
- Strong ecosystem for terminal styling and components.
- LLM integration requires calling external HTTP clients (standard library or lightweight libs).
- Packaging: Just copy the binary into the ISO initramfs or squashfs.

**Python + Textual**
- Requires Python interpreter + dependencies (even with PyInstaller or similar, results in 50–150+ MB).
- Higher RAM footprint.
- Faster development iteration and richer async/LLM SDK support (native OpenAI, Anthropic clients).
- Packaging for live ISO is heavier (need to embed Python + wheels).
- Better if rapid prototyping of complex LLM flows is prioritized over size.

**Recommendation**
**Go + Bubble Tea** is the better fit for PromptOS.

**Rationale**
- Matches the "minimal bootstrap DIY" goal.
- Static binary aligns perfectly with shipping a tiny Alpine-based live environment.
- Size and dependency profile give the highest chance of staying under reasonable ISO limits.
- LLM calls can be handled with standard `net/http` + JSON (or a small wrapper).
- If rapid LLM experimentation becomes painful, we can still prototype critical flows in Python first and port.

**Next Steps**
- Proceed with Go module setup in Phase 1.
- Document any size measurements once the first binary is built.

**Sources**
- Multiple comparisons on binary size and live ISO suitability for Charmbracelet tools (2025–2026 discussions).
- Textual packaging challenges noted in several minimal Linux tool projects.