# Card A3 — Handler `POST /ascii-art` (Generate)

**Package:** A (core pipeline) · **Owner:** Aris
**Priority:** 🔵 on critical path

---

**Context:** the thin handler for the primary action — parse, validate, generate,
present, respond.

## Steps
- [ ] Accept POST; call B's form parsing/validation
- [ ] On valid input, call A2 (generator)
- [ ] Convert to page data (use C's presenter for colored output)
- [ ] Execute template **into a buffer** before writing status
- [ ] Return `200` with preserved values, output, recorded width, download
      snapshot

## Acceptance criteria
- [ ] Valid `POST /ascii-art` → `200`, preserved form state, rendered output,
      rendered width, download snapshot
- [ ] Invalid generation → `400`, preserved safe values, error, no result
      (status logic owned by B, wired here)
- [ ] A template-execution error becomes a clean `500`, not a partial response

## Meta
- **Files:** `internal/web/page_handlers.go`
- **PRD:** §8.1, §13.2
- **Depends on:** A2 · B (validation) · C (presenter). Build the plain path first.

**Definition of Done:** happy-path httptest passes; template renders from a
buffer; seams to B and C are clean.

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
