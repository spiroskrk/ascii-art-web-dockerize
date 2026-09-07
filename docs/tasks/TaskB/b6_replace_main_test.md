# Card B6 — Replace `main_test.go`

**Package:** B (form, validation & output) · **Owner:** Spiros
**Priority:** 🟡 expand near the end

---

**Context:** Phase 0 removed the obsolete CLI subprocess tests when `main.go`
became a server and added startup/server-contract coverage. Expand that baseline
with complete HTTP integration tests once the feature handlers exist.

## Steps
- [x] Remove the CLI subprocess integration tests (completed in Phase 0)
- [ ] Expand HTTP integration tests via `httptest`
- [ ] Do NOT rewrite old tests to pretend CLI is still supported

## Acceptance criteria
- [x] Obsolete root CLI tests are replaced when `main.go` becomes a server
- [ ] HTTP behavior is testable without opening a network port

## Meta
- **Files:** `main_test.go`
- **PRD:** §4.2
- **Depends on:** `main.go` = server (Phase 0) + handlers exist

**Definition of Done:** the Phase 0 server tests are expanded with end-to-end
generation, validation, and download coverage; no CLI pretence remains.

---

## Design tokens (locked in Phase 0)

Style every control with these shared CSS variables in `static/style.css`.
**Do not introduce new colours.**

| Token | Hex | Use |
| --- | --- | --- |
| `--page-bg` | `#FAF8F3` | page background — soft warm off-white |
| `--card` | `#FFFFFF` | card surface |
| `--accent` | `#8F3B3B` | wine/μπορντό — logo accent, primary action, active states |
| `--text` | `#3B322A` | espresso — primary text |
| `--text-muted` | `#7A7062` | labels, secondary text, placeholders |
| `--border` | `#E8E4DA` | field and card borders |
| `--field-bg` | `#EEEDE7` | input / control fill |

**Output box (terminal):** dark surface `#2E2820`, light text `#F6EFE2`, monospace.
Colored runs use `rgb(R, G, B)` from validated integers only.
