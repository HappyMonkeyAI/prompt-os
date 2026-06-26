# Article Research MCP Integration — Research Outcome

**Date:** 2026-06-26

## Result: Deferred

The project backlog includes "Integration with Article Research MCP for ongoing research."

Investigating this via the available article research surface showed it is primarily oriented to general DEV.to / HN article discovery. Targeted searches for Go MCP server integration patterns did not return actionable implementation material from this corpus.

## Decision
Do not integrate the Article Research MCP into PromptOS yet.

## Rationale
- No concrete code references or integration examples surfaced from the available article corpus.
- The project has not reached a stage where ongoing research automation justifies additional infra.
- Adding the MCP now would likely become dead weight without a verified transport/config path.

## Revisit when
- We need live article feeds in the installer workflow.
- We have a verified Go client or HTTP wrapper for article_research.
- The roadmap explicitly includes research-powered onboarding or post-install guidance.

## Artifacts
- Task remains in `TASKS.md` backlog.
