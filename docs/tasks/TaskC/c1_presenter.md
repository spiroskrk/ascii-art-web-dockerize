# Card C1 — `presenter.go`: ANSI → safe styled runs

**Package:** C (color, presentation & infra) · **Owner:** Kostis
**Priority:** 🔵 on critical path (A3 + B3 need it)

---

**Context:** convert aligned ANSI output into safe template data and ANSI-free
plain text.

## Sequences to recognise
```
ESC[30m through ESC[37m
ESC[38;5;208m
ESC[38;2;R;G;Bm
ESC[0m   (reset)
```

## Steps
- [ ] Produce `StyledRun { Text, Colored }` and `StyledLine { Runs }` for HTML
- [ ] Produce an ANSI-free `PlainText` string for downloads
- [ ] Merge adjacent same-state runs without changing visible text
- [ ] Unexpected/incomplete ANSI → internal failure → safe `500`

## Acceptance criteria
- [ ] No ANSI appears in styled-run text or plain download text
- [ ] Visible characters, spaces, rows, final newline preserved
- [ ] Colored and uncolored runs classified correctly
- [ ] Incomplete/unsupported ANSI causes a presentation error (not HTML)

## Meta
- **Files:** `internal/web/presenter.go` + `presenter_test.go`
- **PRD:** §5.5, §11.4, §16.4
- **Depends on:** Phase 0 structs · A2 (produces the ANSI text)
- **Blocks:** A3 (colored output), B3 (plain-text stripping)

**Definition of Done:** presenter tests cover every sequence form; no ANSI leaks.

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
