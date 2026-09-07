# Card A4 — Handler `GET /` (main page + Clear)

**Package:** A (core pipeline) · **Owner:** Aris
**Priority:** 🟢 start immediately (independent of A2)

---

**Context:** serves the default page; also the Clear target.

## Steps
- [ ] Return the main page with default form state, no result, no error
- [ ] Confirm default values: `text=empty, banner=standard, use_color=false,
      color=#ff0000, substring=empty, align=left`
- [ ] Verify Clear points here (GET `/`), not an HTML reset

## Acceptance criteria
- [ ] `GET /` returns `200` HTML with the complete default form
- [ ] Banner and alignment defaults are selected
- [ ] Initial response contains no result and no download snapshot
- [ ] Clear returns the default page with no result or error

## Meta
- **Files:** `internal/web/page_handlers.go`
- **PRD:** §8.2, §7.2
- **Depends on:** Phase 0 template skeleton (independent of A2)

**Definition of Done:** `GET /` httptest confirms default form, defaults
selected, no result.

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
