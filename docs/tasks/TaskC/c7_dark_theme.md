# Card C7 — Dark-theme toggle (optional carry-over)

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🟡 optional — drop if time is short

---

**Context:** the light/dark toggle from the mockup, over the palette variables.

## Steps
- [ ] Add a theme toggle control
- [ ] Swap palette variables for a dark variant via CSS
- [ ] Small JS in `app.js` only if needed

## Acceptance criteria
- [ ] Toggle switches theme without breaking layout or contrast
- [ ] Purely presentational; does not affect generation

## Meta
- **Files:** `templates/index.html` · `static/style.css` · (`app.js` if needed)
- **PRD:** not required — design carry-over
- **Depends on:** palette variables (Phase 0)

**Definition of Done:** toggle works over the palette; safe to drop if time is short.

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
