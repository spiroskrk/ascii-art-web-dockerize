# Card B4 — HTTP status-code contract

**Package:** B (form, validation & output) · **Owner:** Spiros
**Priority:** 🟢 follows B2

---

**Context:** own the mapping of failures to status codes across all handlers.

## Status table
| Situation | Status |
| --- | --- |
| Success | `200` |
| Empty text / unknown banner / invalid color / invalid align / bad width / unsupported char / malformed encoding | `400` |
| Body over limit | `413` |
| Unsupported media type | `415` |
| Unknown route | `404` |
| Wrong method on known route | `405` + `Allow` |
| Template / presenter / unexpected failure | `500` |

## Acceptance criteria
- [ ] Each situation returns its documented status
- [ ] `405` includes an `Allow` header
- [ ] Error responses preserve safe form state, show a message, no result
- [ ] Errors expose no filesystem paths, stack traces, or internals

## Meta
- **Files:** `internal/web/form.go`, `page_handlers.go`, and `download_handler.go`
- **PRD:** §13.1, §13.2
- **Depends on:** B2 (produces error categories)

**Definition of Done:** status codes verified by httptest for each case.

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
