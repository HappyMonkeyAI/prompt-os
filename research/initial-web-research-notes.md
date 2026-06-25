# Initial Web Research Notes - PromptOS

**Date:** 2026-06-25

**Topics Covered:**
1. AI/LLM-driven Linux installers or orchestration (web:0-8 from first batch)
2. Key storage best practices (web:9-13)
3. Rescue/emergency targets (web:14-18)
4. Name/competitor scan and ISO examples (partial)

**Key Verified Findings (with citations where applicable):**
- **Novelty:** No direct live LLM-orchestrated conversational installer exists. Existing work is post-install (LLM server agents, AI package mgmt like Renovate) or heavy enterprise (RHEL AI). archinstall is the closest flexible base. Supports "high-impact minimal project" thesis.
- **Key Baking:** Environment variables via /etc/environment.d or systemd EnvironmentFile= are standard/recommended (OpenAI docs explicitly suggest OPENAI_API_KEY env var on Linux). LoadCredential for service isolation also viable. Matches transcript proposal exactly. Caveats on exposure risks addressed by permissions and non-interactive use in installer context.
- **Triage/Repair:** systemd rescue.target and emergency.target provide clean hooks for custom single-user shells with diagnostics. Easy to extend with journalctl parsing + LLM calls using baked keys for "auto-repair or revert" suggestions.
- **Name:** PromptOS remains clean (no active conflicting distro/OS).

**Gaps for Deeper Research:**
- Specific TUI framework comparisons and minimal-ISO integration examples.
- Concrete Alpine live ISO customization scripts or archiso + TUI packaging.
- Full blueprint JSON schema examples or similar orchestration patterns.
- Real-world security audits of baked env vars in desktop installs.

**Next Research Steps Recommendation:**
- Targeted searches or deep-research invocation on TUI choice, ISO build, and security implementation.
- Review archinstall source for integration points.
- X search or Perplexica for recent discussions on AI Linux tooling.

All external claims in this project will be re-verified before reuse. Research notes are reference-only; architecture decisions live in ADRs/CONTEXT.