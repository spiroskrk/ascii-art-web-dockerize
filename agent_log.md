# Agent Log

This log summarizes the evolution of the ASCII Art CLI through color, output,
and alignment work, followed by the design of the ASCII Art Web application.

## Initial State

- The project already rendered ASCII art from provided banner files.
- Argument mode supported text and optional banner selection.
- Interactive mode existed as an extra feature.
- Color support existed around preset names and earlier ANSI-style input experiments.

## Color Input Redesign

The user-facing color system was updated while keeping ANSI escape sequences as the internal terminal output mechanism.

Supported color inputs now are:

- preset names: `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `black`, `orange`
- full hexadecimal colors: `#RRGGBB`
- RGB comma form: `rgb,<red>,<green>,<blue>`
- HSL comma form: `hsl,<hue>,<saturation>%,<lightness>%`

RGB and HSL prefixes are case-insensitive. Preset names remain case-sensitive.

Unsupported user-facing inputs include:

- raw ANSI notation like `ansi:31`
- CSS-style `rgb(...)` and `hsl(...)`
- alpha forms such as `rgba(...)` and `hsla(...)`

## Renderer Changes

Updated `internal/renderer/color.go`:

- kept the function name `getANSIColor`
- changed it to return `(string, error)`
- added validation for `#RRGGBB`
- added RGB parsing and validation
- added HSL parsing and conversion to RGB
- added exact error messages:
  - `Invalid color name`
  - `Invalid hexadecimal color code`
  - `Invalid RGB color code`
  - `Invalid HSL color code`

Updated `internal/renderer/build.go`:

- color validation errors now come from `getANSIColor`
- `BuildASCII` propagates the exact renderer error
- ASCII-art coloring still happens during rendering by wrapping marked character block rows

## CLI Behavior

The CLI keeps color value validation out of the parser.

`internal/cli/args.go` is responsible for:

- accepting only exact `--color=<color>` flag spelling
- rejecting empty color values
- rejecting multiple color flags
- preserving non-empty color values for renderer validation
- supporting banner-compatible color forms:
  - `go run . [STRING] [BANNER] --color=<color>`
  - `go run . [STRING] [BANNER] --color=<color> [SUBSTRING]`

When `--color=<color>` is first, the default banner is `standard`.

When banner and color are used together, the banner must come before the color flag:

```bash
go run . "hello world" shadow --color=cyan world
```

## Interactive Color Support

Interactive mode now collects color settings after banner selection.

The interactive flow is:

1. Prompt for text.
2. Prompt for banner choice.
3. Prompt for optional color.
4. If a color is entered, prompt for optional substring.
5. Render and print the ASCII-art output.
6. Return to the text prompt.

The color prompt shown to the user is:

```text
Enter color, or press Enter for no color (/exit to quit)
Examples: red, #ff0000, rgb,255,0,0, hsl,0,100%,50%
Color:
```

Interactive color rules:

- empty color input means no color
- non-empty color input is passed to the renderer for validation
- empty substring input after a color means color the whole string
- non-empty substring input colors every matching occurrence
- substring input preserves spaces and only removes the line ending
- `/exit` works at the text, banner, color, and substring prompts
- invalid color values print the renderer error and return to the text prompt

The exported interactive loop reuses one buffered reader for `os.Stdin`, which keeps scripted interactive runs from losing buffered input between cycles.

## Output CLI Parser Handoff

Task 01 for output mode is implemented in the CLI package, with `main.go` intentionally left unchanged for the main integration owner.

Added `cli.Options`:

```go
type Options struct {
	Text       string
	BannerPath string
	Color      string
	Substring  string
	OutputFile string
}
```

Added output-aware entry points:

- `RunArgumentModeOptions(args []string) (Options, error)`
- `PickModeOptions(args []string) (Options, bool, error)`

The old tuple-returning functions still exist for compatibility:

- `RunArgumentMode(args []string) (string, string, string, string, error)`
- `PickMode(args []string) (string, string, string, string, bool, error)`

Output parsing rules:

- `--output=<fileName.txt>` is accepted only as the first argument after the program name
- malformed output flags return the exact output usage message
- `OutputFile == ""` means normal terminal output
- `OutputFile != ""` means the rendered ASCII-art rows should be written to that file by the integration layer
- output mode composes with existing banner and color forms
- ANSI color codes are preserved because output files should remain color-readable in the terminal

Interactive output collection was also added to the options-aware interactive path:

- after text, banner, color, and substring prompts, the CLI asks whether to write output to a file
- blank, `n`, and `no` keep `OutputFile == ""`
- `y` and `yes` prompt for an output file name
- empty output file names are rejected with `No valid output file name`
- invalid output choices are rejected with `Invalid output choice`
- `/exit` works at the output choice and output file name prompts
- the old tuple-based interactive path remains available for current `main.go` compatibility

Accepted parser examples:

```bash
go run . --output=banner.txt "hello"
go run . --output=banner.txt "hello" standard
go run . --output=banner.txt --color=red "hello"
go run . --output=banner.txt --color=red kit "a king kitten have kit"
go run . --output=banner.txt "hello world" shadow --color=cyan world
```

Invalid output formats return:

```text
Usage: go run . [OPTION] [STRING] [BANNER]

EX: go run . --output=<fileName.txt> something standard
```

Historical integration note:

- At the time of this handoff, `output.WriteASCIIFile(fileName, lines)` already
  existed in `internal/output`.
- At that time, `main.go` had not yet been updated to call
  `PickModeOptions` or the output writer.
- That integration was completed later. The current CLI uses the options-aware
  path and writes aligned terminal or file output.

## Test Coverage Added

Updated CLI tests to cover:

- hex color values
- RGB and HSL color values
- banner-compatible color forms
- empty color values
- color flags in invalid positions
- multiple color flags
- shell-split RGB/HSL values
- invalid banner with color mode
- interactive color collection, whole-string color, substring color, exit prompts, and EOF behavior
- output option parsing, malformed output flags, output plus banner, and output plus color
- interactive output choice, output file name collection, output prompt exits, invalid output choices, and empty output file retries

Updated renderer color tests to cover:

- valid preset names
- valid hex values, including uppercase hex
- valid RGB values with case-insensitive prefix
- valid HSL values with case-insensitive prefix
- invalid preset names
- invalid hex lengths and characters
- invalid RGB values and shapes
- invalid HSL values and shapes
- unsupported CSS-style color functions

Updated renderer build tests to cover:

- invalid color name errors
- invalid hex errors
- invalid RGB errors
- invalid HSL errors
- whole-string color rendering
- substring color rendering

Verification run:

```bash
go test ./...
```

## Documentation Updates

Updated project documentation to match the working program:

- `README.md`
- `docs/prd.md`
- `docs/golden_tests.md`
- `docs/audit_test.md`
- `docs/exercise.md`
- `docs/tasks/tasks01_cli.md`
- `docs/tasks/task02_banner.md`
- `docs/tasks/task03_renderer.md`
- `docs/tasks/task04_main.md`
- `docs/tasks/Task01_ai_log.md`
- `docs/workFlow.md`
- `internal/cli/test_cases.md`

Added:

- `to_study.md`
- `agent_log.md`

## Alignment and Output Integration

Alignment was implemented after the earlier output-parser handoff.

Supported values are:

```text
left
center
right
justify
```

The output package owns:

- ANSI-aware visible-width calculation;
- left, center, right, and justify behavior;
- terminal-width detection and the 80-column fallback;
- terminal output;
- aligned file output.

Justify uses structured renderer segments. It expands only spaces between
source word segments and does not expand visual spaces inside ASCII glyphs.

`main.go` now calls `cli.PickModeOptions`. It prints aligned output to the
terminal or calls `output.WriteAlignedASCIIFileWithTerminalWidth` when an output
filename is present.

## ASCII Art Web Direction

The next application is ASCII Art Web. It extends the existing renderer rather
than creating a second ASCII implementation.

The final runtime will be web-only:

```text
go run .
    -> initialize required resources
    -> start HTTP server on :8080
```

The existing CLI remains temporarily while the web application is built and
tested. After the complete web suite passes, obsolete root CLI tests will be
replaced and the unused CLI package can be removed.

The agreed package boundary is:

- `internal/banner` owns banner loading and banner-format validation;
- `internal/renderer` owns normalization, color behavior, substring matching,
  segmentation, and glyph construction;
- `internal/output` owns alignment, visible width, flattening, and text writing;
- `internal/web` will own form parsing, web validation, routing, orchestration,
  safe presentation data, templates, responses, and request logging;
- `main.go` will own only process initialization, server configuration,
  listening, and graceful shutdown.

HTTP handlers must not build glyphs, parse semantic colors, search substrings,
or calculate justify spacing.

## Agreed Browser Form

The main page will contain:

| Field | Control | Contract |
| --- | --- | --- |
| Text | Multiline textarea | Required, maximum 4,096 characters |
| Banner | Radio group | `standard`, `shadow`, or `thinkertoy` |
| Enable color | Checkbox | Absent when disabled, `on` when enabled |
| Color | Native color wheel | Public format is `#RRGGBB` |
| Substring | Single-line text input | Optional, maximum 4,096 characters |
| Alignment | Radio group | `left`, `center`, `right`, or `justify` |
| Width | Hidden input | Measured terminal width in columns |

Defaults are:

```text
text       = empty
banner     = standard
use_color = false
color      = #ff0000
substring  = empty
align      = left
width      = 80
```

Native HTML validation provides immediate feedback, but server validation is
authoritative. Empty text is invalid; whitespace-only text remains valid for
compatibility.

## Text and Newline Contract

Browser textarea input and previous CLI-style input must converge on the same
logical lines.

Supported separators are:

```text
LF
CRLF
CR
literal \n
```

Normalization belongs in `internal/renderer`. The intended order is:

1. Convert CRLF to LF.
2. Convert remaining CR to LF.
3. Treat LF and literal `\n` as logical separators.
4. Preserve existing empty-line and trailing-separator behavior.

Normalization must not trim spaces, collapse repeated spaces, change case,
escape HTML, or render glyphs.

## Color and Browser Presentation

The renderer retains named, hex, RGB, and HSL support for previous-program
compatibility. The web form exposes only color-wheel `#RRGGBB` values.

Color behavior is:

```text
use_color absent
    -> ignore color and substring for rendering

use_color=on
    -> validate #RRGGBB
    -> convert to bounded RGB components
    -> render selected glyphs with color state
```

An empty substring colors the whole input. Non-empty substrings retain existing
case-sensitive, repeated, overlapping, and line-local matching behavior. A
missing match is successful and produces uncolored output.

Renderer output may continue using ANSI internally, but browsers do not
interpret ANSI. The presentation order is:

```text
render ANSI-aware structured output
    -> align while ignoring ANSI width
    -> convert ANSI state to ordinary styled-run data
    -> let html/template create fixed RGB span markup
```

User values are never marked as trusted HTML or arbitrary CSS. Rendered text is
escaped by `html/template`, and only validated RGB integers can enter the fixed
CSS syntax.

## Result Width and Responsive Behavior

The page will contain a terminal-like, whitespace-preserving, monospace output
container.

Immediately before Generate, a small JavaScript helper measures the container
in text columns and submits the value once. Explicit widths must be integers
from 20 through 300. Missing measurement falls back to 80.

The server generates and aligns one fixed snapshot. Resizing the browser:

- does not send another request;
- does not dynamically realign existing output;
- uses horizontal scrolling if the result no longer fits.

The user can press Generate again to render at the newly measured width.
Dynamic resize rendering is outside the first implementation.

## Page Actions and Routes

The agreed routes are:

```text
GET  /                    -> fresh default page and Clear behavior
POST /ascii-art           -> validate, generate, and return the main page
POST /ascii-art/download  -> validate, regenerate, and return an attachment
GET  /static/style.css    -> fixed stylesheet
GET  /static/app.js       -> fixed browser helper
```

Generate submits the editable form. Clear requests `GET /` and removes both the
form state and displayed output.

The initial Download format is:

```text
Content-Type: text/plain; charset=utf-8
Content-Disposition: attachment; filename="ascii-art.txt"
```

Download represents the last successfully generated result. A separate,
non-nested form stores an escaped snapshot of the successful generation values.
Editing the main form or resizing the browser does not change that snapshot
until Generate succeeds again.

The download endpoint does not trust hidden values. It revalidates and
regenerates through the shared workflow. It does not accept a browser filename,
create a server-side output file, or store sessions. ANSI codes are removed
from the initial plain-text download. A standalone colored HTML download is a
future enhancement.

## HTTP Validation and Security

Generation and download accept the
`application/x-www-form-urlencoded` media type. A valid charset parameter is
accepted after parsing the base media type.

Request limits and rules are:

```text
encoded body   = at most 64 KiB
text           = at most 4,096 characters
substring      = at most 4,096 characters
explicit width = 20 through 300
```

The body limit is applied before form parsing. Values are read from `PostForm`,
not merged `Form`, so query-string values cannot override or satisfy POST-body
values.

Accepted form keys are exactly:

```text
text
banner
use_color
color
substring
align
width
```

Every key is single-valued. Duplicate and unexpected keys are rejected.

The status contract is:

| Failure | Status |
| --- | --- |
| Invalid encoding, structure, or field value | `400 Bad Request` |
| Unknown route | `404 Not Found` |
| Wrong method | `405 Method Not Allowed` with `Allow` |
| Body over 64 KiB | `413 Content Too Large` |
| Unsupported media type | `415 Unsupported Media Type` |
| Unexpected template, presenter, or generation failure | `500 Internal Server Error` |

Browser banner input is an exact public identifier, never a path. The three
fixed banner files are loaded before listening and stored in an immutable
registry. Missing or malformed banners are startup failures, not request-time
`404` responses.

Templates, banners, logs, static assets, and download destinations cannot be
selected through browser input. Errors returned to clients omit internal paths,
stack traces, template details, and raw wrapped errors.

## Server Lifecycle, Assets, and Logging

The first version reads resources from disk and must be started from the
repository root.

Required web layout:

```text
templates/index.html
static/style.css
static/app.js
```

The template is parsed once during startup. All three banners are loaded and
validated before listening. Missing or invalid required resources cause
startup failure.

The explicit HTTP server configuration is:

```text
address           = :8080
ReadHeaderTimeout = 5 seconds
ReadTimeout       = 15 seconds
WriteTimeout      = 30 seconds
IdleTimeout       = 60 seconds
shutdown deadline = 5 seconds
```

Interrupt and termination signals trigger graceful shutdown.
`http.ErrServerClosed` after intentional shutdown is normal.

The application does not create `log.txt`. A standard-library structured logger
writes to standard error. Logs may contain safe lifecycle and request metadata,
but must not contain submitted text, substring values, complete form bodies,
generated output, cookies, authorization headers, arbitrary headers, or query
strings.

## Test Design Work

`docs/golden_tests.md` was expanded from a mostly CLI-oriented list into the
complete test-design contract for the web project.

It now distinguishes:

- package unit tests;
- exact golden-output tests;
- shared-generation tests;
- HTTP handler tests using `httptest`;
- startup and lifecycle tests;
- manual browser checks;
- tool and race verification.

The design uses:

- equivalence partitioning for newline, identifier, media-type, and input
  classes;
- boundary-value analysis for 0, 1, 4,096, 4,097, 65,536, 65,537, 20, and 300;
- focused assertions instead of brittle full-page HTML snapshots;
- reviewed golden fixtures rather than expected output generated by the code
  under test;
- temporary fixtures instead of renaming production banners or templates;
- representative HTTP integration cases while exhaustive rules remain in
  package tests.

Coverage includes text, newline normalization, request parsing, banners,
filesystem safety, color, substring matching, ANSI presentation, alignment,
width, routes, page state, errors, HTML escaping, download snapshots, static
assets, startup, logging, concurrency, races, resource lifecycle, and browser
behavior.

Important audit corrections recorded in the test design:

- missing banners prevent startup instead of returning HTTP `404`;
- resize behavior uses a fixed generated snapshot and horizontal scrolling;
- early POST examples are extended to include the agreed form contract;
- vague “recommended” statuses are replaced with exact expectations;
- download, `413`, `415`, duplicate fields, static assets, safe logging, and
  startup failures receive explicit coverage.

The planned test implementation order is:

1. Extend renderer newline tests.
2. Complete remaining banner and output regression tests.
3. Add web parsing and validation tests.
4. Add shared generation and ANSI presenter tests.
5. Add template and route tests.
6. Add download, static, logging, and startup tests.
7. Run test, race, vet, build, formatting, and dependency verification.
8. Complete the manual browser checklist.

The inherited baseline currently passes:

```bash
go test ./...
```

## Documentation and Repository Decisions

`docs/prd.md` is the authoritative product contract. `docs/audit_test.md` is a
coverage source and external-review checklist. `docs/golden_tests.md` bridges
the requirements to future executable tests.

`docs/exercise.md` and `project_handoff_doc.md` were working documents used to
construct the PRD. They are kept locally for private and historical use, listed
in `.gitignore`, and removed from current repository tracking. Their previous
contents remain in existing Git history unless history is explicitly rewritten.

## Current Working State

The current executable is still the completed CLI baseline. It supports banner
selection, color, substring color, output files, and all four alignment modes.
Interactive mode supports the corresponding options.

`main.go` currently uses `PickModeOptions`, renders through the existing domain
packages, and writes aligned terminal or file output.

The web application has not been implemented yet. In particular, the planned
`internal/web`, `templates`, and `static` components do not exist yet.

The requirements and test design are now complete enough to begin test-driven
implementation without reopening the main product decisions. The project
continues to use only Go standard-library and local packages.
