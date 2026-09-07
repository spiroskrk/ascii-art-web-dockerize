# Card A6 — Frontend: output box + terminal CSS

**Package:** A (core pipeline) · **Owner:** Aris
**Priority:** 🟢 start immediately (blocks C6)

---

**Context:** the `<pre>` result area that preserves the ASCII drawing.

## Steps
- [ ] `<pre>` result element in the template
- [ ] CSS: `white-space: pre`, `overflow-x: auto`, `max-width: 100%`, monospace
- [ ] No template indentation inside the ASCII block
- [ ] Leave a clean seam for C's colored `<span>` markup

## Acceptance criteria
- [ ] Result preserves every generated space and line break
- [ ] Monospace font is used
- [ ] Narrow containers provide horizontal scrolling
- [ ] No visible ANSI escape sequences in the output

## Meta
- **Files:** `templates/index.html` (output block) · `static/style.css`
- **PRD:** §12, §12.1
- **Depends on:** Phase 0 palette variables
- **Blocks:** C6 (JS width measurement needs the output area to measure)

**Definition of Done:** output area renders monospace with horizontal overflow;
seam for colored spans is in place.

## Styling notes
- Terminal surface `#2E2820`, text `#F6EFE2`, monospace.
- The `.terminal-body` rule:
  ```css
  .terminal-body {
    max-width: 100%;
    overflow-x: auto;
    white-space: pre;
    font-family: monospace;
  }
  ```

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
