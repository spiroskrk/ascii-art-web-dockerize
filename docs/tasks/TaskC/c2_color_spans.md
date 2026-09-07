# Card C2 — Color validation and safe span rendering

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🔵 follows C1

---

**Context:** convert validated `#RRGGBB` (format checked by B) into bounded RGB
integers, and wrap colored runs in fixed, safe CSS.

## Steps
- [ ] Convert `#RRGGBB` → `Red/Green/Blue = 0..255`
- [ ] Template wraps colored runs: `<span style="color: rgb(R, G, B)">…</span>`
- [ ] Preserve substring behavior (renderer does the matching)

## Security requirements (all mandatory)
- [ ] Rendered text stays `html/template`-escaped
- [ ] RGB components are validated integers 0–255
- [ ] No user value marked `template.HTML` or arbitrary `template.CSS`
- [ ] CSS/HTML injection attempts rejected or safely escaped

## Acceptance criteria
- [ ] Enabling color requires a valid `#RRGGBB`; server converts to RGB ints
- [ ] Empty substring colors all text; matches preserved (repeated/overlapping/case)
- [ ] Browser output uses styled spans with no visible ANSI
- [ ] Injection attempts are rejected or escaped

## Meta
- **Files:** `internal/web/presenter.go` (RGB) · `templates/index.html` (span block)
- **PRD:** §11.2, §11.3, §11.4, §16.6
- **Depends on:** B format check · C1

**Definition of Done:** RGB conversion + span rendering pass security tests.

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
