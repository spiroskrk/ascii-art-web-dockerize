# Card C4 — Safe static-asset serving

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🟢 start immediately

---

**Context:** serve exactly the known assets under fixed public URLs.

## Steps
- [ ] `/static/style.css` → `static/style.css`
- [ ] `/static/app.js` → `static/app.js`
- [ ] No directory listing; paths are constants, never request-derived

## Acceptance criteria
- [ ] CSS and JS return expected content types
- [ ] `/static/` exposes no directory listing
- [ ] Templates, banners, traversal-like paths are not publicly served
- [ ] No browser value influences a filesystem path

## Meta
- **Files:** `internal/web/` (asset wiring)
- **PRD:** §14.4, §16.8
- **Depends on:** Phase 0 routing

**Definition of Done:** asset tests confirm content types and no traversal/listing.

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
