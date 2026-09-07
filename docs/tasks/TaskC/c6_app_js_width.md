# Card C6 — `static/app.js`: width measurement

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🔵 waits on A6 (needs the output area)

---

**Context:** measure how many monospace columns fit the output area and write it
to the hidden `width` field right before Generate submits.

## Steps
- [ ] Measure the output container + monospace char width
- [ ] Compute column count; write to hidden `width`
- [ ] Fallback to 80 if JS/measurement unavailable
- [ ] Keep it small — no JS test framework in v1

## Must NOT
Render glyphs or implement any alignment algorithm. Only measure.

## Acceptance criteria
- [ ] Generate measures the terminal area once and submits width in columns
- [ ] Missing measurement falls back to 80
- [ ] Resizing after generation sends no request and does not realign

## Meta
- **Files:** `static/app.js`
- **PRD:** §8.1, §12.1
- **Depends on:** A6 (output area exists)

**Definition of Done:** width is measured once and submitted; 80 fallback works.

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
