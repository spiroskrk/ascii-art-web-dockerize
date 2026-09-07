# Task C Progress Log

Branch: `task-c`

## Done So Far

### C1 — Presenter
- Implemented `Present` in `internal/web/presenter.go`.
- Converts renderer ANSI output into safe `StyledLine` / `StyledRun` data.
- Produces ANSI-free plain text for downloads.
- Supports renderer-generated ANSI colors:
  - `ESC[30m` through `ESC[37m`
  - `ESC[38;5;208m`
  - `ESC[38;2;R;G;Bm`
  - `ESC[0m` reset
- Rejects incomplete or unsupported ANSI sequences.
- Added presenter tests in `internal/web/presenter_test.go`.

### C2 — Color Conversion and Safe HTML Spans
- Added conversion from validated `#RRGGBB` values into RGB integers.
- Wired RGB conversion into validated form color state.
- Updated form tests for RGB expectations.
- Updated the output template to render colored runs with safe RGB span styling.
- Added template test coverage for safe span rendering and HTML escaping.

### C5 — Color, Substring, and Width Controls
- Added `use_color` checkbox to the main form.
- Added native `<input type="color">` with default `#ff0000`.
- Added `substring` text input with `maxlength="4096"`.
- Added hidden `width` input.
- Added home-page tests for these controls.

### C6 — Width Measurement JavaScript
- Implemented `static/app.js` to measure output width before form submission.
- Uses fallback width `80`.
- Clamps measured width to the server-supported range `20` through `300`.
- Does not listen to resize events or trigger extra requests.

### C4 — Static Asset Serving
- Added fixed routes for:
  - `/static/style.css`
  - `/static/app.js`
- Static paths are hardcoded and not derived from browser input.
- Added route tests confirming only known static files are served.

### C3 — Request Logging Middleware
- Implemented request logging middleware in `internal/web/middleware.go`.
- Middleware records safe metadata:
  - method
  - path
  - status
  - duration
- Middleware does not read request bodies.
- Added tests confirming logs do not include submitted text, substrings, cookies, or encoded form bodies.
- Wired middleware into `Routes()`.

## Verification

- `node --check static/app.js` passed after the C6 implementation.
- `go test ./internal/web` passed after C3/C4 test additions.
- `go test ./...` passed after the current Task C changes.

## Not Done / Deferred

### C7 — Dark Theme
- Optional carry-over task.
- Not started.

### C8 — CLI Cleanup
- Not started.
- Must wait until the full web application is complete.
- Should happen after A3 Generate and B3 Download are implemented and the full test suite passes.

## External Blockers

- A3 `POST /ascii-art` is still outside Task C and may still need final integration.
- B3 `POST /ascii-art/download` is still outside Task C and may still need final integration.
