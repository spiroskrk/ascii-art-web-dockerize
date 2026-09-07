# Card B5 — Frontend: alignment, Clear, download form, errors

**Package:** B (form, validation & output) · **Owner:** Spiros
**Priority:** 🟢 start immediately

---

**Context:** the controls this package owns in the template.

## Steps
- [ ] Alignment **radio group** (`name="align"`): left/center/right/**justify**,
      default left
- [ ] **Clear** as a GET action (link-as-button or separate GET form; no nested forms)
- [ ] Separate **download form** shown only after success, with an escaped hidden
      snapshot (text in a hidden textarea; other fields hidden inputs)
- [ ] Error message display + preserved form values after `400`

## Acceptance criteria
- [ ] Alignment uses a radio group with all four options
- [ ] Clear returns empty/default form and removes the result
- [ ] Download appears only after successful generation
- [ ] Editing the main form does not change the downloadable result
- [ ] HTML forms are not nested

## Meta
- **Files:** `templates/index.html` (this package's blocks) · `static/style.css`
- **PRD:** §8.2, §8.3, §13, §7.1
- **Depends on:** Phase 0 skeleton + block names

## Styling notes
- Radios/inputs use `--field-bg`, `--border`, `--text`. Labels `--text-muted`.
- Clear is a quiet outline control; Generate (owned by A) is the filled `--accent` one.

**Definition of Done:** controls render with correct names; Clear is a GET;
download form is separate and non-nested.

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
