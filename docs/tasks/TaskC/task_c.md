# Package C — Color, Presentation & Infrastructure

**Owner:** Kostis
**Theme:** the most *technical* slice — turning renderer ANSI into safe browser
output, the color controls, the width-measuring JavaScript, logging, and static
assets.

This package spans the widest set of technologies: Go, JavaScript, CSS, and the
security model. It is the most challenging conceptually (the ANSI presenter and
the no-injection rules) but its core piece, the presenter, lands relatively
early — which is why the final CLI cleanup is attached here.

Each card is tagged with a priority colour:

- 🟢 **Start immediately** — no blockers, begin right after Phase 0
- 🔵 **Critical path** — unblocks A3 + B3, or waits on A6
- 🟡 **Last / optional** — do at the end, or drop if time is short

| Priority | Cards |
| --- | --- |
| 🟢 Start immediately | C3, C4, C5 |
| 🔵 Critical path (C1 unblocks A3 & B3) | C1, C2, C6 |
| 🟡 Last / optional | C7, C8 |

---

## Scope at a glance

**Backend**
- `internal/web/presenter.go` — ANSI → safe styled runs + plain text
- Color validation: `#RRGGBB` → RGB integers
- `internal/web/middleware.go` — request logging
- Safe static-asset serving

**Frontend**
- `use_color` checkbox + native `<input type="color">`
- `substring` text input
- Hidden `width` input
- `static/app.js` — width measurement
- The colored `<span>` markup in the template
- Dark-theme toggle

**Tests**
- ANSI presenter tests
- Security tests
- Static-asset / logging tests

**Extra deliverable**
- CLI cleanup (delete `internal/cli`)

---

## Task C1 — `presenter.go`: ANSI → safe styled runs

**What**
The renderer represents colored glyphs with ANSI foreground and reset sequences.
Browsers don't interpret ANSI. Convert the aligned ANSI output into safe template
data:
- `StyledRun { Text string; Colored bool }` and `StyledLine { Runs []StyledRun }`
  for HTML
- an ANSI-free `PlainText` string for downloads

The presenter must recognise **every** renderer-generated sequence:

```
ESC[30m through ESC[37m
ESC[38;5;208m
ESC[38;2;R;G;Bm
ESC[0m   (reset)
```

**Why**
This is the bridge between terminal output and the browser. It must strip ANSI
completely (none may appear in HTML text or download text) while preserving every
visible character, space, row, and the final newline convention.

**Must NOT**
Emit raw HTML. An **unexpected or incomplete** ANSI command is an internal
presentation failure → safe `500`, never interpreted as HTML or terminal
behavior. Adjacent runs with the same color state may be merged, but visible text
must not change.

**Files**
- `internal/web/presenter.go` + `presenter_test.go`

**PRD** §5.5, §11.4, §16.4

**Depends on** the `StyledRun` / `StyledLine` / `PresentedOutput` structs locked
in Phase 0 · A2 produces the ANSI text it consumes.

---

## Task C2 — Color validation and safe span rendering

**What**
Two connected pieces:
1. Convert the validated `#RRGGBB` (format already checked by B) into three
   bounded integers, `Red/Green/Blue = 0..255`
2. In the template, wrap colored runs in fixed, safe CSS built only from those
   validated integers:

```html
<span style="color: rgb(R, G, B)">colored run</span>
```

**Why**
This is the security core of coloring. The user must never be able to inject
arbitrary CSS or HTML. The only user-derived values that reach the style are
three integers, each proven to be 0–255.

**Security requirements (all mandatory)**
- rendered text stays a normal `html/template`-escaped string
- RGB components are validated integers 0–255
- no user value is marked `template.HTML` or arbitrary `template.CSS`
- arbitrary CSS/HTML injection attempts are rejected or safely escaped

**Substring behavior** is preserved from the renderer (case-sensitive, repeated,
overlapping, per-line, spaces significant). C exposes the control; the renderer
does the matching.

**Files**
- `internal/web/presenter.go` (RGB conversion) · `templates/index.html` (span
  block)

**PRD** §11.2, §11.3, §11.4, §16.6

**Depends on** B's format check · C1 (run classification).

---

## Task C3 — `middleware.go`: request logging

**What**
Focused HTTP middleware that logs request metadata with a standard-library
structured logger writing to **standard error**:
- method, safe route path, response status, request duration
- safe identifiers such as the selected banner

**Why**
Observability without leaking user data. The app must never manage its own log
file; the runtime environment owns persistence.

**Must NOT log**
Submitted text, substrings, full form bodies, generated output, cookies, or
arbitrary headers. The middleware must not read or copy request bodies, and must
contain no application or rendering logic.

**Files**
- `internal/web/middleware.go` + test (in-memory log destination)

**PRD** §14.3, §16.8

**Depends on** the logger injected in Phase 0.

---

## Task C4 — Safe static-asset serving

**What**
Serve exactly the known assets under fixed public URLs:

```
/static/style.css -> static/style.css
/static/app.js    -> static/app.js
```

**Why**
Assets must be served without exposing the project directory. All filesystem
paths are application constants; browser input is never joined to a path.

**Must NOT**
Expose a directory listing. Serve templates, banners, or traversal-like paths.
Derive any path from a request value.

**Files**
- `internal/web/` (asset serving wiring)

**PRD** §14.4, §16.8

**Depends on** Phase 0 routing.

---

## Task C5 — Frontend: color, substring, width controls

**What**
The input controls this package owns:
- `use_color` **checkbox** (`name="use_color"`)
- native **`<input type="color" name="color">`** (default `#ff0000`)
- `substring` single-line text input (`name="substring"`, `maxlength="4096"`)
- hidden `width` input (`name="width"`)

**Why**
Color uses the native browser color wheel — no custom spectrum bar (the mockup's
bar must be replaced). The `use_color` checkbox makes coloring independent: the
color value is ignored until it is on.

**Files**
- `templates/index.html` (this package's blocks)

**PRD** §7, §11.1, §17

**Depends on** the template skeleton + block names (Phase 0).

---

## Task C6 — `static/app.js`: width measurement

**What**
A small JavaScript helper that, immediately before Generate submits, measures how
many monospace text columns fit inside the output area and writes that number
into the hidden `width` field.

**Why**
Alignment (especially center/right/justify) needs a column count. Measuring in
the browser lets the server align to the real visible width. If JS is
unavailable or measurement fails, the form submits the fallback **80**.

**Must NOT**
Render glyphs or implement any alignment algorithm. It only measures. Explicit
widths must be 20–300; the server validates.

**Files**
- `static/app.js`

**PRD** §8.1, §12.1

**Depends on** the output area existing (A6) so it has something to measure.
Keep it intentionally small — no JS test framework in the first version.

---

## Task C7 — Dark-theme toggle (optional carry-over)

**What**
If the team keeps it from the mockup, the light/dark toggle: a CSS-driven theme
switch over the palette variables.

**Why**
A nice-to-have from the design. Purely presentational; not required by the PRD.

**Files**
- `templates/index.html` · `static/style.css` · (small JS in `app.js` if needed)

**PRD** — not required; design carry-over.

**Depends on** the CSS-variable palette (Phase 0). Lowest priority; drop if time
is short.

---

## Task C8 — CLI cleanup

**What**
After the whole web app passes its test suite, remove the now-unused CLI:
- search the repo to confirm no supported source imports `internal/cli`
- delete `internal/cli` and its package tests
- run `go test ./...` to confirm nothing broke

**Why it's yours**
It is mechanical, and your package's core (the presenter) finishes relatively
early, so you have room at the end. It must be a **separate** cleanup change, not
mixed into a feature.

**Files**
- delete `internal/cli/`

**PRD** §4.2

**Depends on** the full app passing its tests. Strictly last.

---

## Dependency summary for Package C

```
Phase 0 structs ──> C1 (presenter) ──> C2 (color/spans)
Phase 0 logger  ──> C3 (middleware)
Phase 0 routing ──> C4 (static assets)
Phase 0 blocks  ──> C5 (color/substring/width controls)
A6 (output area) ─> C6 (app.js width)
palette         ──> C7 (dark theme, optional)
full test pass  ──> C8 (CLI cleanup, last)
```

**Start-first, no-blockers:** C1, C3, C4, C5
**Waits on A6:** C6
**Optional / droppable:** C7
**Strictly last:** C8

**Coordination note:** C1 (the presenter) is on A's and B's critical path — the
Generate handler needs it for colored output and Download needs its plain-text
stripping. Land the presenter early so A3 and B3 aren't blocked.