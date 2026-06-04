---
name: MDA-sync-crack
description: Redirect-only compatibility wrapper for the canonical MDA crack-sync workflow in .agents\skills\MDA-sync-crack\SKILL.md.
license: MIT
compatibility: Redirect only; use .agents\skills\MDA-sync-crack\SKILL.md
metadata:
    version: "1.1-redirect"
    project: MDA
---

# Redirect: Canonical MDA Sync Workflow

Do not maintain crack-sync procedure text in `.claude`.

The single source of truth is:

```text
D:\NIKKE\MDA\.agents\skills\MDA-sync-crack\SKILL.md
```

For any request to sync upstream, update `crack`, rebuild the cracked Go agent, or run related automation, load and follow that `.agents` skill exactly. If this wrapper disagrees with the `.agents` file, the `.agents` file wins.
