# Card A1 — Extend `NormalizeInput` for real newlines

**Package:** A (core pipeline) · **Owner:** Aris
**Priority:** 🟢 start immediately · 🔵 on critical path (blocks A2)

---

**Context:** the normalizer currently splits literal `\n`; extend it so real
browser line endings — LF, CRLF, CR — also split into logical lines, while
keeping every existing literal-`\n` behavior.

## Steps
- [ ] Add CRLF → LF conversion (must run **first**)
- [ ] Add remaining CR → LF conversion
- [ ] Ensure LF and literal `\n` are treated as equivalent separators
- [ ] Preserve empty-logical-line and trailing-escaped-newline behavior
- [ ] Add table-driven tests: LF, CRLF, CR, literal `\n`, mixed separators,
      empty logical lines, existing trailing-newline compatibility

## Acceptance criteria
- [ ] LF, CRLF, CR, and literal `\n` produce equivalent logical lines
- [ ] Existing literal-`\n` compatibility behavior still passes
- [ ] Leading, trailing, and repeated spaces are preserved
- [ ] No trimming, case change, escaping, or glyph rendering happens here

## Meta
- **Files:** `internal/renderer/` (normalizer + tests)
- **PRD:** §9.3, §16.1
- **Depends on:** nothing
- **Blocks:** A2

**Definition of Done:** new + existing renderer tests pass; the four separator
forms are proven equivalent by test.

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
