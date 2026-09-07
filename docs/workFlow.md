# Workflow — ASCII Art Web

## Objective

Expose the existing ASCII-art engine through a browser. Running:

```bash
go run .
```

must start an HTTP server on `:8080`. Users configure everything through an HTML
form, generate on the same page, clear the page, and download the output:

```text
GET  /                    → main page (also the Clear target)
POST /ascii-art           → generate and display the result
POST /ascii-art/download  → download the displayed result as ascii-art.txt
```

The web layer must **reuse** the existing banner, renderer, and output packages.
It must not become a second implementation of banner rendering, substring
matching, or alignment. Only the Go standard library is used on the server.

## Shared Contract

The web package separates untrusted form values, validated generation values,
generated output, and template presentation. These types are locked in Phase 0
so all three packages compile against them:

```go
type FormState struct {          // raw, untrusted
	Text, Banner, Color, Substring, Alignment, Width string
	UseColor bool
}

type SelectedColor struct {       // validated
	Enabled          bool
	Hex              string
	Red, Green, Blue int
}

type GenerationInput struct {     // validated, bounded
	Text, Banner, Substring, Alignment string
	Color                              SelectedColor
	Width                              int
}

type GeneratedASCII struct {      // shared generator output
	ANSIText string
	Width    int
}

type StyledRun struct  { Text string; Colored bool }
type StyledLine struct { Runs []StyledRun }

type PageData struct {            // template model
	Form          FormState
	Output        []StyledLine
	HasResult     bool
	Error         string
	RenderedWidth int
	Color         SelectedColor
}
```

One-way data flow:

```text
HTTP request → FormState → validation → GenerationInput
             → shared generator → GeneratedASCII
             → presenter → browser PageData or plain download
```

Accepted form keys are exactly: `text`, `banner`, `use_color`, `color`,
`substring`, `align`, `width`.

## Architecture Decision

The HTTP layer stays thin. It handles HTTP concerns only; the algorithms stay in
the preserved packages.

```text
internal/banner    → banner loading and validation      (unchanged)
internal/renderer  → normalize, color, substring, glyphs (NormalizeInput extended)
internal/output    → visible width, alignment, writing   (unchanged)
internal/web       → routes, forms, handlers, generation, presentation, middleware (new)
main.go            → server lifecycle and shutdown        (was CLI)
```

Browser display and download share **one** generation workflow:

```text
validated input
    → map banner identifier to bundled path
    → load banner
    → normalize logical input lines
    → build structured, optionally colored ASCII
    → apply the selected alignment
    → produce aligned output
```

The HTML and download handlers must not each reimplement this pipeline.

## Why This Matters

The team of three splits the work **vertically**: instead of one person per layer
(back / front / glue), each person owns a feature slice end-to-end. This lets
everyone touch a bit of Go, HTML, templates, and the request/response flow.

The risk of a vertical split is that all three touch the same `main.go`, the same
router, and the same template. The mitigation is a shared **Phase 0** that locks
the skeleton and the contracts above, after which each person works behind clean
handler files, named template blocks, and one CSS section each.

## Phase 0 — Shared skeleton (all three together, before branches)

Built once, in one session, then rarely touched:

- folder structure, `go.mod`
- `main.go` skeleton (logger, parse template, load banners, `http.Server` with
  timeouts, listen, graceful shutdown)
- `internal/web/app.go` (dependencies + constructor)
- `internal/web/routes.go` (the three routes, `404`, `405` + `Allow`)
- `internal/web/types.go` and `response.go` (shared contracts + buffered pages)
- separate page and download handler files
- `templates/index.html` skeleton with an empty named block per feature region
- `static/style.css` with the palette variables and base layout

Contracts to lock here: the structs above · the shared generator signature · the
template block names · the form field names · the CSS
variables.

## Task Split

### Package A — Core pipeline

Owner: **Aris**
Folder: `package-a/` (explanation + cards A1–A7)

Responsibilities:

- Extend `renderer.NormalizeInput` for LF / CRLF / CR.
- Build the shared generation workflow (`generate.go`).
- Handlers for `POST /ascii-art` and `GET /`.
- Frontend: text + banner controls, Generate, the `<pre>` output box.
- Write the README.

### Package B — Form, validation & output

Owner: **Spiros**
Folder: `package-b/` (explanation + cards B1–B6)

Responsibilities:

- `form.go`: bounded parsing (415/413), FormState → GenerationInput validation.
- Handler for `POST /ascii-art/download`.
- The HTTP status-code contract.
- Frontend: alignment radios, Clear (GET), the separate download form, errors.
- Replace the old CLI tests in `main_test.go` with HTTP integration tests.

### Package C — Color, presentation & infrastructure

Owner: **Kostis**
Folder: `package-c/` (explanation + cards C1–C8)

Responsibilities:

- `presenter.go`: ANSI → safe styled runs + ANSI-free plain text.
- Color validation (`#RRGGBB` → RGB integers) and safe span rendering.
- `middleware.go`: request logging.
- Safe static-asset serving.
- Frontend: `use_color`, native color input, substring, hidden width.
- `static/app.js` width measurement; optional dark-theme toggle.
- Final CLI cleanup (delete `internal/cli`).

## Implementation Order

1. All three build the **Phase 0** skeleton and lock the contracts.
2. Aris lands **A1 → A2** (extended normalizer, shared generator) early — this
   unblocks B and C.
3. Kostis lands **C1** (presenter) early — the Generate and Download handlers
   need it.
4. In parallel: A4/A5/A6, B1/B2/B4/B5, C3/C4/C5.
5. Wire the dependent handlers: A3, then B3 (download).
6. C6 (width JS) once A6 (output area) exists.
7. Run package tests continuously; then `go test ./...`, `go test -race ./...`,
   `go vet ./...`.
8. Spiros replaces `main_test.go`; Kostis removes `internal/cli` (last, separate
   commit); Aris writes the README.

## Manual Checks

Browser behavior that Go tests do not cover directly:

- native required-field feedback on empty text;
- color-wheel interaction and `use_color` toggle;
- radio groups and keyboard navigation;
- monospace spacing and terminal-like appearance;
- one-time width measurement; horizontal scroll after narrowing the window;
- no resize-triggered HTTP requests; regenerate at the new width on Generate;
- Clear removes both form state and output;
- Download prompts for `ascii-art.txt`; downloaded text is readable.

## Status Codes

```text
200  success (page, generation, download)
400  empty text, unknown banner/align, invalid color, bad width, unsupported char, bad encoding
413  body over 64 KiB
415  unsupported media type
404  unknown route
405  wrong method on a known route (+ Allow header)
500  template / presenter / unexpected failure
```
