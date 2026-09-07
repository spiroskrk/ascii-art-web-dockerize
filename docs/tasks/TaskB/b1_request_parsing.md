# Card B1 — `form.go`: request parsing with limits

**Package:** B (form, validation & output) · **Owner:** Spiros
**Priority:** 🟢 start immediately

---

**Context:** guard the request itself before any validation runs.

## Steps
- [ ] Accept only `Content-Type: application/x-www-form-urlencoded`
- [ ] Wrap the body with a **64 KiB limit applied *before* `ParseForm`**
- [ ] Read values from `PostForm`, not the merged `Form`

## Acceptance criteria
- [ ] Unsupported media type → `415`
- [ ] Body over 64 KiB → `413`
- [ ] Query-string values cannot override or duplicate POST-body values
- [ ] Generation and download accept URL-encoded forms only

## Meta
- **Files:** `internal/web/form.go`
- **PRD:** §7.4, §13.1
- **Depends on:** nothing beyond Phase 0

**Definition of Done:** limit and content-type guards proven by test before parsing.

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
