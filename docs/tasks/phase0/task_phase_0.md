# Task 0 — Phase 0: Shared Skeleton & Locked Contracts

**Owner:** the whole team (Aris · Spiros · Kostis), together in one session
**Priority:** ⚫ do this first — nothing else can start cleanly until it is done

This is not a feature. It is the common foundation that Packages A, B, and C all
build on. Build it once, together (one person types, the other two watch and
agree), commit it, and only then branch off. Everything locked here is a
**contract**: once agreed, each person works behind their boundary without
touching the others' files.

---

## Why this comes first

The vertical split means all three people touch the same `main.go`, the same
router, and the same template. If everyone starts on an empty repo, you get three
different `main.go` files and constant merge conflicts. Phase 0 removes that by
fixing the skeleton and the seams up front.

---

## Steps

### 0.1 — Folder structure + `go.mod`
- [x] Create the layout:
      ```text
      .
      ├── main.go
      ├── templates/
      │   └── index.html
      ├── static/
      │   ├── style.css
      │   └── app.js
      ├── banners/
      │   ├── standard.txt
      │   ├── shadow.txt
      │   └── thinkertoy.txt
      └── internal/
          ├── banner/      (unchanged)
          ├── cli/         (kept temporarily)
          ├── output/      (unchanged)
          ├── renderer/    (NormalizeInput extended later by A)
          └── web/
              ├── app.go
              ├── routes.go
              ├── types.go
              ├── response.go
              ├── page_handlers.go
              ├── download_handler.go
              ├── form.go
              ├── generate.go
              ├── presenter.go
              └── middleware.go
      ```
- [x] Confirm `go.mod` module path and that `go build ./...` compiles the skeleton

### 0.2 — `main.go` skeleton
- [x] Construct the logger
- [x] Parse `templates/index.html` once
- [x] Load and validate all three banners
- [x] Construct the web application (inject template, banner registry, logger)
- [x] Configure `http.Server` with the timeouts (see contract below)
- [x] Start listening; handle graceful shutdown (5s deadline)
- [x] Non-zero exit if any startup resource is missing/invalid

### 0.3 — `internal/web/app.go`
- [x] Application struct holding immutable dependencies + constructor

### 0.4 — `internal/web/routes.go`
- [x] Register the three routes (see contract below)
- [x] `404` for unknown routes; `405` + `Allow` for wrong methods

### 0.5 — `templates/index.html` skeleton
- [x] One empty `{{define}}` block per agreed feature region (names below)
- [x] No feature markup yet — just the shell that calls the blocks

### 0.6 — `static/style.css` base
- [x] Define the palette CSS variables (below)
- [x] Base layout only; each package styles its own controls later

### 0.7 — Lock the contracts
- [x] Paste the structs (below) into the code
- [x] Agree the generator signature (below)
- [x] Agree the template block names (below)
- [x] Confirm the form field names (below)

**Definition of Done:** `go build ./...` compiles; the server starts, serves an
empty shell page, and shuts down gracefully; all contracts below exist in code
and are agreed by all three.

---

## Contract 1 — Shared structs

Paste these into `internal/web` (names may be refined, but the trust boundaries
and one-way flow must stay):

```go
// Raw, untrusted form values — preserved for error redisplay.
type FormState struct {
	Text      string
	Banner    string
	UseColor  bool
	Color     string
	Substring string
	Alignment string
	Width     string
}

// Validated color.
type SelectedColor struct {
	Enabled          bool
	Hex              string
	Red, Green, Blue int
}

// Validated, bounded generation values. Only this enters generation.
type GenerationInput struct {
	Text      string
	Banner    string
	Color     SelectedColor
	Substring string
	Alignment string
	Width     int
}

// Shared generator output — aligned ANSI, no HTML concerns.
type GeneratedASCII struct {
	ANSIText string
	Width    int
}

// Safe presentation data for the template.
type StyledRun struct {
	Text    string
	Colored bool
}

type StyledLine struct {
	Runs []StyledRun
}

// Template model.
type PageData struct {
	Form          FormState
	Output        []StyledLine
	HasResult     bool
	Error         string
	RenderedWidth int
	Color         SelectedColor
}
```

One-way data flow (must not be violated):

```text
HTTP request → FormState → validation → GenerationInput
             → shared generator → GeneratedASCII
             → presenter → browser PageData or plain download
```

---

## Contract 2 — Shared generator signature

Owned by A (`generate.go`), called by A (display) and B (download). Agree the
shape now so B and C can compile against it while A implements it:

```go
// Generate runs the one shared workflow:
//   banner lookup → normalize → build (optionally colored) → align.
// Returns aligned ANSI text and the recorded width.
func (app *App) Generate(in GenerationInput) (GeneratedASCII, error)
```

The presenter (C) then converts `GeneratedASCII` into `[]StyledLine` for HTML and
an ANSI-free plain string for download:

```go
func Present(g GeneratedASCII) (lines []StyledLine, plain string, err error)
```

---

## Contract 3 — Routes

```text
GET  /                    → main page (also the Clear target)   [A]
POST /ascii-art           → generate and display                [A]
POST /ascii-art/download  → download ascii-art.txt              [B]
GET  /static/style.css    → static asset                        [C]
GET  /static/app.js       → static asset                        [C]
```

- Unknown route → `404`
- Wrong method on a known route → `405` + `Allow` header

During Phase 0, the two POST routes return a temporary `501 Not Implemented`.
Their real handlers replace those explicit placeholders in Packages A and B.
Safe serving for the two locked static URLs is implemented by Package C.

---

## Contract 4 — Form field names

Exactly these seven keys — nothing else, or the server returns `400`:

```text
text        textarea, required, maxlength 4096          [A]
banner      radio group: standard/shadow/thinkertoy     [A]
use_color   checkbox                                    [C]
color       <input type="color">, #RRGGBB               [C]
substring   text input, maxlength 4096                  [C]
align       radio group: left/center/right/justify      [B]
width       hidden input, integer 20–300 (fallback 80)  [C]
```

---

## Contract 5 — Template block names

Blocks are split by semantic action as well as owner. This keeps keyboard order
logical, prevents nested forms, and lets CSS place related actions together
without coupling their HTTP behavior:

```text
{{define "page"}}             → shared HTML shell + form boundary
{{define "input_controls"}}   → text + banner                     [A]
{{define "generate_action"}}  → Generate submit button            [A]
{{define "output_box"}}       → the <pre> result area               [A]
{{define "align_controls"}}   → alignment radios                  [B]
{{define "form_error"}}       → safe validation feedback           [B]
{{define "clear_action"}}     → Clear link to GET /                [B]
{{define "download_action"}}  → separate download form             [B]
{{define "color_controls"}}   → use_color + color + substring     [C]
{{define "width_field"}}      → hidden width input                [C]
{{define "theme_control"}}    → optional client-side theme control [C]
```

The Generate button and Clear link live within the generation form. Download
owns a separate form outside it; HTML forms are never nested.

---

## Contract 6 — CSS variables (palette)

Defined once in `static/style.css`. **No new colours** anywhere else.

```css
:root {
  --page-bg:    #FAF8F3; /* page background — soft warm off-white */
  --card:       #FFFFFF; /* card surface */
  --accent:     #8F3B3B; /* wine/μπορντό — logo accent, primary action */
  --text:       #3B322A; /* espresso — primary text */
  --text-muted: #7A7062; /* labels, secondary text, placeholders */
  --border:     #E8E4DA; /* field and card borders */
  --field-bg:   #EEEDE7; /* input / control fill */
}
```

Output box (terminal): dark surface `#2E2820`, light text `#F6EFE2`, monospace.
Colored runs use `rgb(R, G, B)` from validated integers only.

---

## Contract 7 — Server config (for `main.go`)

```text
address           :8080
ReadHeaderTimeout 5s
ReadTimeout       15s
WriteTimeout      30s
IdleTimeout       60s
shutdown deadline 5s (on interrupt/termination)
body limit        64 KiB (applied before ParseForm — B enforces)
working directory repository root
```

---

## After Phase 0

Once this is committed and green, branch off:

- **Aris** starts A1 → A2 (extended normalizer, shared generator) — unblocks the team
- **Kostis** starts C1 (presenter) early — A3 and B3 need it
- **Spiros** starts B1/B2/B5 immediately — the most independent package

Everyone now works behind their own handler file, template block, and CSS
section. The only shared files (`main.go`, `routes.go`, the `index.html` shell)
are already locked; touch them only for a small, announced, one-line change.
