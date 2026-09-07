# Package A — The Core Pipeline

**Owner:** Aris
**Theme:** the central data path — *text in → ASCII out → shown on screen*

This package owns the heart of the application: the shared generation workflow
that every other feature calls, plus the two most central HTTP handlers and the
primary input/output controls in the browser.

Because the shared generator and the extended normalizer are dependencies for
Packages B and C, Package A is a **blocker**. The signatures it exposes must be
agreed during Phase 0 so B and C can code against them in parallel.

Each card is tagged with a priority colour:

- 🟢 **Start immediately** — no blockers, begin right after Phase 0
- 🔵 **Critical path** — unblocks the other packages; do these early
- 🟡 **Last** — depends on everything else being done

| Priority | Cards |
| --- | --- |
| 🟢 Start immediately | A1, A4, A5, A6 |
| 🔵 Critical path (unblocks B & C) | A1, A2, A3 |
| 🟡 Last | A7 |

---

## Scope at a glance

**Backend**
- Extend `renderer.NormalizeInput` for LF / CRLF / CR
- `internal/web/generate.go` — the shared generation workflow
- Handler for `POST /ascii-art` (Generate)
- Handler for `GET /` (main page + Clear target)

**Frontend**
- `<textarea name="text">` (required, maxlength 4096)
- Banner radio group
- Generate button
- The `<pre>` output box + terminal CSS

**Tests**
- New normalization cases
- Shared generation workflow tests
- `POST /ascii-art` happy path
- `GET /` default-page test

**Extra deliverable**
- `README.md`

---

## Task A1 — Extend `NormalizeInput` for real newlines

**What**
The existing normalizer splits literal `\n` sequences into logical lines. Extend
it so real browser line endings — **LF, CRLF, and CR** — also act as logical
separators, while keeping every existing literal-`\n` behavior unchanged.

**Why**
A browser textarea produces real newline characters when the user presses Enter,
but the form can also carry literal backslash-n. The PRD requires both to yield
equivalent logical lines. Normalization is the renderer's job, not the HTTP
handler's, so this belongs in `internal/renderer`.

**Normalization order (must be exact)**
1. Convert CRLF → LF
2. Convert remaining CR → LF
3. Treat LF and literal `\n` as equivalent separators
4. Preserve existing behavior for empty logical lines and trailing escaped
   newlines

CRLF must be handled **before** individual CR/LF so one Windows line ending does
not become two separators.

**Must NOT**
Trim text, strip leading/trailing spaces, collapse repeated spaces, change case,
HTML-escape, or render glyphs.

**Files**
- `internal/renderer/` (the normalizer + its tests)

**PRD** §9.3, §16.1

**Depends on** nothing — can start immediately.

---

## Task A2 — The shared generation workflow (`generate.go`)

**What**
One function that turns a validated `GenerationInput` into a `GeneratedASCII`
result, following the fixed pipeline:

```
validated input
  → map banner identifier to bundled path
  → load banner (from preloaded registry)
  → normalize logical input lines
  → build structured, optionally colored ASCII
  → apply selected alignment
  → produce aligned output (ANSI text + recorded width)
```

**Why**
Both the browser display and the download endpoint must produce output the
**same way**. If each handler built its own pipeline, they could drift apart.
This function is the single source of truth; B (download) and C (color
presentation) both call it.

**Must NOT**
Construct banner glyphs, implement substring matching, justify rows, or
re-implement any renderer logic. It **calls** `banner`, `renderer`, and `output`
— it does not duplicate them.

**Returns** `GeneratedASCII { ANSIText string; Width int }` — aligned ANSI text,
no HTML concerns.

**Files**
- `internal/web/generate.go` + `generate_test.go`

**PRD** §5.3, §5.5

**Depends on** A1 (normalization) · the shared structs locked in Phase 0.
**Blocks** Package B (download) and Package C (presenter).

---

## Task A3 — Handler: `POST /ascii-art` (Generate)

**What**
The thin handler that receives the submitted form, runs it through validation
(from Package B's `form.go`) and the shared generator (A2), then returns the main
page with the submitted values and the rendered result at `200 OK`.

**Why**
This is the primary action of the whole app. "Thin" means it orchestrates —
parse, validate, generate, present, respond — but contains no rendering or
alignment logic itself.

**Behavior**
- Success → `200`, preserved form values, rendered output, recorded width, and
  the download snapshot
- The template is executed into a buffer first, so a template error becomes a
  clean `500` instead of a half-written response

**Files**
- `internal/web/page_handlers.go` (this handler)

**PRD** §8.1, §13.2

**Depends on** A2 · B's validation (`form.go`) · C's presenter for colored output.
Agree the seams in Phase 0; a plain (uncolored) path can be built first.

---

## Task A4 — Handler: `GET /` (main page + Clear)

**What**
Serves the main page with default form state, no result, and no error. This same
handler is the target of the Clear action.

**Why**
`GET /` is both the first thing a visitor sees and the "reset" destination.
Clear is defined as a fresh GET of `/`, not an HTML form reset (a reset would not
remove server-rendered output).

**Default state**
```
text = empty, banner = standard, use_color = false,
color = #ff0000, substring = empty, align = left,
result = absent, error = absent
```

**Files**
- `internal/web/page_handlers.go` (this handler)

**PRD** §8.2, §7.2

**Depends on** the template skeleton (Phase 0). Independent of A2 — can be built
early.

---

## Task A5 — Frontend: primary input controls

**What**
Inside the page template, the controls this package owns:
- `<textarea name="text">` with `required` and `maxlength="4096"`
- Banner **radio group** (`name="banner"`): `standard`, `shadow`, `thinkertoy`,
  default `standard`
- The **Generate** submit button

**Why radio, not dropdown**
Banner is a mutually-exclusive choice. The PRD mandates radio groups so a client
cannot submit two banners. (The mockup used a dropdown — this must change.)

**Files**
- `templates/index.html` (this package's block only)

**PRD** §7, §7.1

**Depends on** the template skeleton + block names locked in Phase 0.

---

## Task A6 — Frontend: the output box + terminal CSS

**What**
The result area where generated ASCII appears:
- A `<pre>` element that preserves every space and line break
- Monospace font, `white-space: pre`, `overflow-x: auto` for horizontal scroll
- No template indentation inside the ASCII, no visible ANSI sequences

**Why**
ASCII art depends on every character occupying equal width; only monospace +
`pre` preserves the drawing. Horizontal scroll handles output wider than the
viewport without reflowing it.

```css
.terminal-body {
  max-width: 100%;
  overflow-x: auto;
  white-space: pre;
  font-family: monospace;
}
```

**Files**
- `templates/index.html` (output block) · `static/style.css` (terminal styles)

**PRD** §12, §12.1

**Depends on** the CSS-variable palette locked in Phase 0. The colored `<span>`
markup inside is Package C's concern — leave a clean seam.

---

## Task A7 — README

**What**
The required root `README.md` with four sections:
- **Description** — what the app does
- **Authors** — the team (K.S.A)
- **Usage** — how to run (`go run .`, open `http://localhost:8080`), and the note
  that the server must run from the repository root
- **Implementation details** — the algorithm (the renderer pipeline)

**Why it's yours**
The Implementation section must explain the renderer, which you wrote and know
best. Written **last**, so it describes the final, cleaned-up project.

**Files**
- `README.md`

**PRD** §14.4 (working-directory note) · subject requirement

**Depends on** everything else being done.

---

## Dependency summary for Package A

```
A1 (normalize) ──┐
                 ├─> A2 (generate.go) ──> A3 (POST handler)
Phase 0 structs ─┘
Phase 0 skeleton ──> A4 (GET handler)
Phase 0 blocks   ──> A5, A6 (frontend)
everything       ──> A7 (README)
```

**Start-first, no-blockers:** A1, A4, A5, A6
**Critical path (unblocks B and C):** A1 → A2
**Last:** A7

**Coordination note:** because A2 blocks the other packages, prioritise landing
its signature (even as a stub) in Phase 0, then implement it early.
