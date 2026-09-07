# Card B3 — Handler `POST /ascii-art/download`

**Package:** B (form, validation & output) · **Owner:** Spiros
**Priority:** 🔵 waits on A2 + C presenter

---

**Context:** download uses the same validation (B1/B2) and the same shared
generator (A2), then returns a file attachment.

## Steps
- [ ] Re-validate the hidden snapshot (not trusted)
- [ ] Regenerate via A2 (shared workflow)
- [ ] Strip ANSI via C's plain-text presenter
- [ ] Set headers: `text/plain; charset=utf-8` and
      `Content-Disposition: attachment; filename="ascii-art.txt"`

## Acceptance criteria
- [ ] Download returns `ascii-art.txt` as a plain-text attachment
- [ ] Preserves recorded width, ASCII spaces, rows, alignment
- [ ] Contains no ANSI sequences
- [ ] No server file is created; client cannot supply filename/path
- [ ] Failed download has no attachment header
- [ ] Requires no session state

## Meta
- **Files:** `internal/web/download_handler.go`
- **PRD:** §8.3, §16.7
- **Depends on:** A2 · C plain-text presenter · B1/B2

**Definition of Done:** download tests pass; headers correct; no file written.

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
