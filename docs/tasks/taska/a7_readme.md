# Card A7 — README

**Package:** A (core pipeline) · **Owner:** Aris
**Priority:** 🟡 do last (describes the final project)

---

**Context:** the required root README, written last so it describes the final,
cleaned-up project.

## Steps
- [ ] **Description** — what the app does
- [ ] **Authors** — K.S.A
- [ ] **Usage** — `go run .`, open `http://localhost:8080`, run from repo root
- [ ] **Implementation details** — the renderer pipeline / algorithm

## Acceptance criteria
- [ ] All four required sections are present
- [ ] Documentation states the server runs from the repository root
- [ ] Implementation section explains the generation algorithm

## Meta
- **Files:** `README.md`
- **PRD:** §14.4 + subject requirement
- **Depends on:** everything else complete

**Definition of Done:** README has all four sections; run-from-root note present;
algorithm explained.

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
