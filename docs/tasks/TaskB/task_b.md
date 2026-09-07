# Package B — Form, Validation & Output Management

**Owner:** Spiros
**Theme:** everything that comes *in* and everything that goes *out* — request
parsing, authoritative validation, error handling, Clear, and Download.

This package is the most **detail-heavy** in the project. It is not conceptually
hard, but it is exacting: content-type checks, byte limits, duplicate-key
rejection, five distinct status codes, and dozens of table-driven test cases.

Its big advantage: it is the most **independent** package. Validation does not
wait on anyone and can begin right after Phase 0.

Each card is tagged with a priority colour:

- 🟢 **Start immediately** — no blockers, begin right after Phase 0
- 🔵 **Critical path** — waits on A2 + C presenter
- 🟡 **Last** — depends on everything else being done

| Priority | Cards |
| --- | --- |
| 🟢 Start immediately | B1, B2, B4, B5 |
| 🔵 Waits on A2 + C | B3 |
| 🟡 Last | B6 |

---

## Scope at a glance

**Backend**
- `internal/web/form.go` — parsing + authoritative validation
- Handler for `POST /ascii-art/download`
- The HTTP status-code contract for validation errors

**Frontend**
- Alignment radio group (all four, including `justify`)
- Clear as a GET action
- The separate, non-nested download form with a hidden snapshot
- Error display + form-value preservation
- CSS for the form controls

**Tests**
- Form parsing & validation (the big table-driven suite)
- Download tests

**Extra deliverable**
- Replace `main_test.go` CLI tests with HTTP integration tests

---

## Task B1 — `form.go`: request parsing with limits

**What**
Before any validation, guard the request itself:
- Accept only `Content-Type: application/x-www-form-urlencoded`; anything else →
  `415 Unsupported Media Type`
- Wrap the body with a **64 KiB limit applied *before* `ParseForm`**, so the
  parser can never consume an unbounded body; exceeding it → `413 Content Too
  Large`
- Read values from `PostForm`, **not** the merged `Form`, so query-string values
  cannot override POST-body values

**Why**
These are security and robustness boundaries. Limiting the body before parsing
prevents a malicious client from exhausting memory. Reading `PostForm` prevents
query-string smuggling.

**Files**
- `internal/web/form.go`

**PRD** §7.4, §13.1

**Depends on** nothing beyond Phase 0 — can start immediately.

---

## Task B2 — `form.go`: `FormState` and field validation

**What**
Capture the raw submitted values into a `FormState` (all fields untrusted), then
validate them into a `GenerationInput`:

| Field | Rule |
| --- | --- |
| `text` | non-empty, ≤ 4096 chars (whitespace-only is **valid**) |
| `banner` | exactly `standard`, `shadow`, or `thinkertoy` |
| `use_color` | present/on toggles coloring |
| `color` | exactly `#RRGGBB` **when color enabled** |
| `substring` | optional, ≤ 4096 chars |
| `align` | exactly `left`, `center`, `right`, or `justify` |
| `width` | integer 20–300; missing → fallback 80 |

The accepted keys are **exactly** those seven. **Duplicate single-value fields**
and **unexpected keys** → `400 Bad Request` (not silently ignored).

**Why**
Client-side validation is only usability; a client can bypass the form entirely.
The server must re-validate every request. `FormState` preserves the raw values
so an error page can show them back to the user; `GenerationInput` holds only
clean, bounded values that are safe to generate from.

**Note on color**
Distinguish *invalid color while `use_color` is on* (→ `400`) from *ignored color
while coloring is off* (no error). The `#RRGGBB` → RGB integer conversion itself
is Package C's concern — B validates presence/format, C converts and presents.

**Compatibility rule**
Do **not** trim text to make validation stricter. HTML `required` treats
spaces-only as non-empty, and the existing app accepts whitespace-only text.

**Files**
- `internal/web/form.go`

**PRD** §7, §7.4, §9.1, §13.1

**Depends on** the `FormState` / `GenerationInput` structs locked in Phase 0.

---

## Task B3 — Handler: `POST /ascii-art/download`

**What**
Handle the download request. It uses the **same** validation (B1/B2) and the
**same** shared generator (A2) as the display path, then returns the result as a
file attachment:

```
Content-Type        = text/plain; charset=utf-8
Content-Disposition = attachment; filename="ascii-art.txt"
```

**Why**
Download must represent the *successfully displayed* result, not later edits to
the form. The client never supplies the filename or a path — it is the fixed
constant `ascii-art.txt`. Hidden snapshot fields are **not** trusted: the
endpoint re-validates and regenerates server-side through the shared workflow.

**Must NOT**
Write any file on the server. Trust hidden fields. Include a
`Content-Disposition` header on a **failed** download (a failure returns the
normal HTML page with an error, or a safe `500`).

**Output detail**
The downloaded file is plain text: ASCII, spaces, rows, and alignment preserved;
**ANSI color sequences removed** (that stripping is done by C's presenter, which
B calls).

**Files**
- `internal/web/download_handler.go` (this handler)

**PRD** §8.3, §16.7

**Depends on** A2 (shared generator) · C's plain-text presenter · B1/B2
(validation). Build against the Phase 0 seams.

---

## Task B4 — The HTTP status-code contract

**What**
Own the mapping of failure situations to status codes, and make sure each
handler returns the right one:

| Situation | Status |
| --- | --- |
| Success | `200` |
| Empty text / unknown banner / invalid color / invalid align / bad width / unsupported char / malformed encoding | `400` |
| Body over limit | `413` |
| Unsupported media type | `415` |
| Unknown route | `404` |
| Wrong method on known route | `405` + `Allow` header |
| Template / presenter / unexpected failure | `500` |

Error responses must **preserve safe form state**, show a useful message, and
never expose filesystem paths, stack traces, or internal details.

**Why**
The subject and PRD both require appropriate status codes; this is a core grading
point. Centralising the contract here keeps the three handlers consistent.

**Files**
- `internal/web/form.go`, `page_handlers.go`, and `download_handler.go`

**PRD** §13.1, §13.2

**Depends on** B2 (validation produces the error categories).

---

## Task B5 — Frontend: alignment, Clear, download form, errors

**What**
The controls this package owns in the template:
- Alignment **radio group** (`name="align"`): `left`, `center`, `right`,
  **`justify`** — all four, default `left`
- **Clear** as a GET action — a link styled as a button, or a separate GET form.
  **HTML forms must not be nested.**
- The **separate download form** shown only after a successful generation,
  holding an escaped hidden snapshot (text in a hidden textarea to preserve real
  line endings; other fields as hidden inputs)
- **Error message** display + preserved form values after a `400`

**Why**
Radio enforces one alignment. Clear can't be a plain HTML reset (it wouldn't
clear server-rendered output). The download form is separate and non-nested so
editing the main form doesn't change the downloadable result until Generate
succeeds again.

**Files**
- `templates/index.html` (this package's blocks) · `static/style.css` (control
  styles)

**PRD** §8.2, §8.3, §13, §7.1

**Depends on** the template skeleton + block names (Phase 0). The `width` and
color controls live in Package C — leave clean seams.

---

## Task B6 — Replace `main_test.go`

**What**
The old CLI subprocess integration tests no longer describe the product once
`main.go` is a server. Replace them with HTTP integration tests.

**Why it's yours**
You already own the status-code contract and write the most `httptest` in the
project, so the HTTP integration tests are a natural fit.

**Must NOT**
Rewrite the old tests to pretend CLI behavior is still supported.

**Files**
- `main_test.go`

**PRD** §4.2

**Depends on** `main.go` becoming the server (Phase 0) and the handlers existing.
Runs near the end.

---

## Dependency summary for Package B

```
Phase 0 ──> B1 (parsing) ──> B2 (validation) ──> B4 (status contract)
                                    │
A2 (generate) + C (presenter) ──────┼──> B3 (download handler)
Phase 0 blocks ─────────────────────┴──> B5 (frontend)
main.go = server + handlers ─────────────> B6 (main_test.go)
```

**Start-first, no-blockers:** B1, B2, B5
**Waits on A2 + C:** B3
**Last:** B6

**Coordination note:** B is the most self-contained package — validation and the
form frontend can progress fully in parallel while A builds the generator. Only
the download handler (B3) needs A and C in place.
