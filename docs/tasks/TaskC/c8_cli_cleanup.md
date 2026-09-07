# Card C8 — CLI cleanup

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🟡 strictly last

---

**Context:** after the whole web app passes its tests, remove the unused CLI in a
**separate** cleanup change.

## Steps
- [ ] Search the repo: confirm no supported source imports `internal/cli`
- [ ] Delete `internal/cli` and its package tests
- [ ] Run `go test ./...` to confirm nothing broke

## Acceptance criteria
- [ ] No supported source imports `internal/cli` before removal
- [ ] Package and its tests removed only after the full suite passes
- [ ] Not mixed into a feature change

## Meta
- **Files:** delete `internal/cli/`
- **PRD:** §4.2
- **Depends on:** full app passing its tests

**Definition of Done:** `internal/cli` gone; `go test ./...` green; separate commit.

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
