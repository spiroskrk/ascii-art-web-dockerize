# Card C3 — `middleware.go`: request logging

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🟢 start immediately

---

**Context:** structured logging middleware writing to standard error.

## Steps
- [ ] Log method, safe path, status, duration, safe identifiers (e.g. banner)
- [ ] Write to standard error via a standard-library structured logger
- [ ] Do not read/copy request bodies; no rendering logic inside

## Must NOT log
Submitted text, substrings, full form bodies, generated output, cookies, or
arbitrary headers.

## Acceptance criteria
- [ ] Logs contain safe lifecycle/request metadata only
- [ ] No submitted or generated content is logged
- [ ] Logs go to standard error, not an app-managed file

## Meta
- **Files:** `internal/web/middleware.go` + test (in-memory destination)
- **PRD:** §14.3, §16.8
- **Depends on:** logger injected in Phase 0

**Definition of Done:** middleware test confirms safe fields logged, secrets absent.

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
