# Card A5 — Frontend: text + banner + Generate

**Package:** A (core pipeline) · **Owner:** Aris
**Priority:** 🟢 start immediately

---

**Context:** the primary input controls in the template.

## Steps
- [ ] `<textarea name="text">` with `required` and `maxlength="4096"`
- [ ] Banner **radio group** (`name="banner"`): standard/shadow/thinkertoy,
      default standard
- [ ] Generate **submit** button
- [ ] Do not use `novalidate`

## Acceptance criteria
- [ ] Main page contains text and banner controls
- [ ] Banner uses a radio-button group (not a dropdown)
- [ ] Native validation blocks an ordinary empty-text submission
- [ ] Text control exposes the 4,096-char client limit

## Meta
- **Files:** `templates/index.html` (this package's block)
- **PRD:** §7, §7.1
- **Depends on:** Phase 0 skeleton + block names

**Definition of Done:** controls render with correct `name` attributes and
constraints; radio group defaults to standard.

## Styling notes
- Textarea and radios use `--field-bg` fill, `--border` outline, `--text` text.
- Labels in `--text-muted`. Generate button filled with `--accent`, white label.

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
