# Card B2 — `form.go`: `FormState` and field validation

**Package:** B (form, validation & output) · **Owner:** Spiros
**Priority:** 🟢 start immediately

---

**Context:** capture raw values into `FormState` (untrusted), validate into
`GenerationInput` (clean, bounded).

## Field rules
| Field | Rule |
| --- | --- |
| `text` | non-empty, ≤ 4096 (whitespace-only is **valid**) |
| `banner` | `standard` / `shadow` / `thinkertoy` |
| `use_color` | present/on toggles coloring |
| `color` | exactly `#RRGGBB` **when enabled** (C converts to RGB) |
| `substring` | optional, ≤ 4096 |
| `align` | `left` / `center` / `right` / `justify` |
| `width` | integer 20–300; missing → 80 |

## Steps
- [ ] Build `FormState` from raw submitted values
- [ ] Validate each field per the table
- [ ] Accepted keys are **exactly** the seven above
- [ ] Reject duplicate single-value fields and unexpected keys → `400`
- [ ] Distinguish invalid color (color on) from ignored color (color off)
- [ ] Do NOT trim text (whitespace-only stays valid)

## Acceptance criteria
- [ ] Empty text rejected; whitespace-only accepted
- [ ] Unknown banner / alignment → `400`
- [ ] Text or substring over 4096 → `400`
- [ ] Duplicate and unexpected fields rejected
- [ ] Raw browser strings are not passed directly into generation

## Meta
- **Files:** `internal/web/form.go`
- **PRD:** §7, §7.4, §9.1, §13.1
- **Depends on:** `FormState` / `GenerationInput` structs (Phase 0)

**Definition of Done:** table-driven validation suite passes for valid and
invalid combinations.

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
