# Card C5 — Frontend: color, substring, width controls

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🟢 start immediately

---

**Context:** the input controls this package owns.

## Steps
- [ ] `use_color` **checkbox** (`name="use_color"`)
- [ ] Native `<input type="color" name="color">` (default `#ff0000`)
- [ ] `substring` text input (`name="substring"`, `maxlength="4096"`)
- [ ] Hidden `width` input (`name="width"`)

## Acceptance criteria
- [ ] Color uses an enable checkbox and a native color wheel
- [ ] Color is ignored when `use_color` is off
- [ ] Substring exposes the 4096-char client limit
- [ ] The mockup's custom spectrum bar is replaced by the native picker

## Meta
- **Files:** `templates/index.html` (this package's blocks)
- **PRD:** §7, §11.1, §17
- **Depends on:** Phase 0 skeleton + block names

## Styling notes
- Checkbox accent uses `--accent`. Inputs use `--field-bg`/`--border`/`--text`.
- The native color swatch keeps its own chrome; frame it with `--border`.

**Definition of Done:** controls render with correct names; native picker in place.

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
