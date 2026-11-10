# Platform verification status

| Host platform | Current status | Evidence |
| --- | --- | --- |
| WSL2 / Docker Desktop | Passed | [Final live record](wsl2.json): explicit path, configured-path-only, inherited PATH/asdf; each independently verified a fresh target marker and the expected task-failure control. Owned cleanup passed. |
| Native Linux / Docker Engine | Passed | [Final live record](native-linux.json): explicit path, configured-path-only and inherited PATH, each with independent marker and task-failure checks. [Preparation and staging cleanup](native-linux-preparation.json) passed on `docker-infra`; temporary tools/source/caches and handoff files were removed, with no owned containers left. |

Unit fixtures are regression coverage, not platform evidence. Use the
[runbook](linux-wsl.md) to produce sanitized records from the final identical
source and fixture before qualifying later changes.

Both records qualify the same final source/fixture identity
`8eff2ad665bf2a164719a780899f681f5ad30f04b5bf2d0e542044493b0be267`,
immutable image profile and normal bridge routing. Shared prepared images were
retained. The completed OpenSpec change was synced and archived on November 9,
2026. Merge is a separate follow-up; no release or version bump is requested.
