# Product Requirements Document - ASCII Art Web

## Document Status

This is a living PRD. It records the product and architectural decisions implemented by the project.

Current status: **implemented and maintained as the product contract evolves**.

The previous version of this document described the ASCII Art alignment project. Alignment is now implemented in the current source, so this PRD supersedes that older planning document and defines the next project: ASCII Art Web.

---

## 1. Problem Statement

The existing program is a Go command-line application that renders text as ASCII art. It already supports banners, multiline input, foreground colors, substring coloring, alignment, terminal output, and file output.

The next project must expose those capabilities through a browser. Users should not need to understand CLI flags, positional arguments, or terminal behavior. They should be able to configure and generate ASCII art through a clear HTML form, see the result in the browser, clear the page, and download generated output.

The web project must reuse the existing rendering engine. It must not become a second implementation of banner rendering, substring matching, or alignment.

---

## 2. Product Goal

Build a web-only ASCII-art application that:

- starts an HTTP server with `go run .`;
- presents a browser form for all generation choices;
- accepts both Enter-generated line breaks and literal `\n` sequences;
- renders with one of the three bundled banners;
- optionally colors all text or a selected substring;
- supports left, center, right, and justify alignment;
- displays generated output safely in the browser;
- clears both the form and result when requested;
- allows the browser to download generated output;
- uses the existing banner, renderer, and output behavior wherever possible;
- uses only the Go standard library on the server.

---

## 3. Existing Baseline

The current application already has working support for:

- `standard`, `shadow`, and `thinkertoy` banners;
- printable ASCII characters from 32 through 126;
- literal escaped-newline input using `\n`;
- structured word and space segments;
- named, hexadecimal, RGB, and HSL color parsing;
- whole-input and substring coloring;
- case-sensitive, repeated, and overlapping substring matches;
- left, center, right, and justify alignment;
- ANSI-aware visible-width measurement;
- live terminal-width detection with `COLUMNS` and 80-column fallbacks;
- terminal output and ANSI-preserving file output;
- argument and interactive CLI modes.

The current source and tests are authoritative. `project_handoff_doc.md` is a historical audit from before structured rendering and visual alignment were completed. Its statements that alignment is unimplemented do not describe the current code.

At the beginning of web design, the verified baseline is:

```text
go test ./...

ok  ascii-art
ok  ascii-art/internal/banner
ok  ascii-art/internal/cli
ok  ascii-art/internal/output
ok  ascii-art/internal/renderer
```

---

## 4. Product Scope

### 4.1 Web-only runtime

ASCII Art Web is a web-only product. Running:

```bash
go run .
```

must start the HTTP server.

The previous CLI flags, positional syntax, prompts, looping interaction, `/exit` command, and exact CLI usage messages are not part of the new public interface.

Preserving previous functionality means exposing the previous capabilities through web equivalents; it does not mean preserving the CLI interaction model.

### 4.2 CLI transition

The current `main.go` will eventually be replaced by a web-server entry point. Existing CLI-specific integration tests in `main_test.go` will no longer describe the intended executable and must eventually be replaced by HTTP integration tests.

The existing `internal/cli` package is not used by the web application. It remains temporarily while the web path is built so the first implementation changes do not combine feature replacement with broad deletion.

When `main.go` becomes the long-running server entry point, the old CLI subprocess tests in `main_test.go` no longer describe the product and must be replaced with HTTP-oriented tests. They must not be rewritten to pretend that CLI behavior is still supported.

After the complete web application passes its renderer, output, HTTP, presentation, download, and security tests, the unused `internal/cli` package and its package tests are removed in a separate cleanup change. Before removal, the repository is searched to confirm that no supported code imports it.

The following packages remain because they provide the preserved application behavior:

```text
internal/banner
internal/renderer
internal/output
```

Removing the CLI does not remove Git history. Historical handoff documents may remain when they are clearly identified as descriptions of an earlier project stage.

---

## 5. Architectural Boundaries

The HTTP layer must remain thin. It is responsible for HTTP behavior, not ASCII-art algorithms.

### 5.1 Existing package ownership

- `internal/banner` owns banner loading and banner-format validation.
- `internal/renderer` owns input normalization, color validation, substring matching, source segmentation, and ASCII glyph rendering.
- `internal/output` owns visible-width calculation, alignment, flattening, and text writing.
- HTML templates own document markup and presentation structure.
- CSS owns visual styling, including browser color display.

### 5.2 Web-layer ownership

The new web layer owns:

- route and method handling;
- bounded form parsing;
- web-form validation;
- mapping public banner identifiers to bundled banner files;
- calling the shared generation workflow;
- converting renderer-controlled color state into safe template data;
- executing templates;
- returning HTML or download responses;
- mapping failures to HTTP status codes.

The web layer must not:

- construct banner glyphs;
- implement substring matching;
- duplicate color-selection rules from the renderer;
- justify flattened ASCII rows;
- accept user-controlled banner or template paths;
- write a browser-provided path on the server filesystem.

### 5.3 Shared generation workflow

Browser display and download must use one shared generation workflow:

```text
Validated form values
    -> map banner identifier to bundled path
    -> load banner
    -> normalize logical input lines
    -> build structured, optionally colored ASCII art
    -> apply the selected alignment
    -> produce aligned output
```

After generation, presentation branches by response type:

```text
Aligned result
    |-> browser presenter -> HTML template
    |-> download response -> attachment body
```

The HTML and download handlers must not independently reimplement the generation pipeline.

### 5.4 Package and file layout

The web application uses the following initial organization:

```text
.
├── main.go
├── templates/
│   ├── index.html
│   └── 404.html
├── static/
│   ├── style.css
│   └── app.js
├── banners/
└── internal/
    ├── banner/
    ├── cli/
    ├── output/
    ├── renderer/
    └── web/
        ├── app.go
        ├── types.go
        ├── routes.go
        ├── response.go
        ├── page_handlers.go
        ├── download_handler.go
        ├── form.go
        ├── generate.go
        ├── presenter.go
        ├── middleware.go
        └── *_test.go
```

The filenames inside `internal/web` are responsibility guides, not a requirement to keep every concern in a separate tiny file. Closely related files may be combined when that improves readability.

`main.go` owns process-level orchestration only:

- construct the logger;
- parse templates;
- load all bundled banners;
- construct the web application;
- configure `http.Server`;
- start listening and handle graceful shutdown.

The `internal/web` package owns:

- application dependencies;
- route registration;
- form parsing and web-specific validation;
- thin page, generation, and download handlers;
- shared generation orchestration;
- safe browser presentation data;
- request logging middleware.

An application value conceptually holds immutable dependencies such as the parsed template, preloaded banner registry, and logger. These dependencies are injected during construction rather than stored as mutable package globals.

Request-specific form values, validation errors, generated output, and page data remain local to each request.

No additional `service`, `handlers`, `forms`, or `utils` packages are introduced for the first version. A separate application-layer package may be extracted later only if a real additional consumer needs the shared generation workflow.

### 5.5 Staged data model

The web package separates untrusted form values, validated generation values, generated output, and template presentation.

Conceptual raw form state:

```go
type FormState struct {
    Text       string
    Banner     string
    UseColor  bool
    Color      string
    Substring  string
    Alignment string
    Width      string
}
```

`FormState` preserves submitted browser values for validation feedback. All fields remain untrusted at this stage.

Conceptual validated types:

```go
type SelectedColor struct {
    Enabled bool
    Hex     string
    Red     int
    Green   int
    Blue    int
}

type GenerationInput struct {
    Text       string
    Banner     string
    Color      SelectedColor
    Substring  string
    Alignment string
    Width      int
}
```

Only `GenerationInput` is passed into generation. Its banner and alignment are recognized values, width is a bounded integer, and enabled color contains canonical hexadecimal and numeric RGB representations.

Conceptual generated result:

```go
type GeneratedASCII struct {
    ANSIText string
    Width    int
}
```

The shared generator returns aligned renderer output without HTML concerns.

Conceptual presentation result:

```go
type StyledRun struct {
    Text    string
    Colored bool
}

type StyledLine struct {
    Runs []StyledRun
}

type PresentedOutput struct {
    Lines     []StyledLine
    PlainText string
}
```

The presenter recognizes only renderer-generated foreground and reset sequences. It produces normal strings for `html/template` and ANSI-free text for initial downloads. Unexpected ANSI commands are internal presentation failures and return a safe `500` response.

Conceptual template data:

```go
type PageData struct {
    Form          FormState
    Output        []StyledLine
    HasResult     bool
    Error         string
    RenderedWidth int
    Color         SelectedColor
}
```

`HasResult` distinguishes successful visually blank output from the absence of a result. Templates do not receive HTTP requests, banner maps, renderer functions, ANSI parsers, or raw internal errors.

Names and minor grouping may change during implementation, but the trust boundaries and one-way data flow must remain:

```text
HTTP request
    -> FormState
    -> validation
    -> GenerationInput
    -> shared generator
    -> GeneratedASCII
    -> presenter
    -> browser PageData or plain download
```

---

## 6. Main Page and User Journey

The application uses one main page.

A user can:

1. Enter text in a multiline textarea.
2. Select exactly one banner.
3. Enable or disable coloring.
4. Select a color with a color wheel.
5. Optionally enter a substring to color.
6. Select exactly one alignment.
7. Generate and view the result on the same page.
8. Clear the form and displayed result.
9. Download generated output.

After generation or a validation error, the page should preserve submitted form values. Clear is the explicit action that returns the page to its initial state.

---

## 7. Form Contract

The agreed form controls are:

| Field | HTML control | Required value |
| --- | --- | --- |
| `text` | Multiline `<textarea>` | Non-empty text, maximum 4,096 characters |
| `banner` | Radio-button group | `standard`, `shadow`, or `thinkertoy` |
| `use_color` | Checkbox | Present/on when coloring is enabled |
| `color` | `<input type="color">` | `#RRGGBB` when coloring is enabled |
| `substring` | Single-line text input | Optional, including spaces, maximum 4,096 characters |
| `align` | Radio-button group | `left`, `center`, `right`, or `justify` |
| `width` | Hidden input | Measured terminal width in text columns |

### 7.1 Why radio buttons are required

Banner and alignment are mutually exclusive choices. Checkboxes would allow contradictory submissions such as two banners or two alignments. Radio buttons communicate and enforce one selection within each group.

The server must still validate submitted values because a client can bypass browser controls.

### 7.2 Default form state

The initial page uses:

```text
text       = empty
banner     = standard
use_color = false
color      = #ff0000
substring  = empty
align      = left
result     = absent
error      = absent
```

The color input needs a default because native color controls always contain a value. The color is ignored until `use_color` is enabled.

### 7.3 Client-side form validation

The form uses native HTML constraint validation as the first line of feedback:

- the text textarea is marked `required`;
- the banner radio group is required and defaults to `standard`;
- the alignment radio group is required and defaults to `left`;
- the native color input supplies a valid picker value, while color is applied only when `use_color` is enabled;
- the form must not use `novalidate`;
- Generate and Download are submit controls and therefore trigger browser constraint validation;
- Clear is a GET action and does not require form validation.

The browser must prevent an ordinary empty-text submission and show its native validation feedback. JavaScript may call standard validity APIs when needed, but it must not duplicate a separate validation system for constraints already expressed in HTML.

Client-side validation is a usability feature, not a security boundary. A client can modify HTML or send a request without using the form, so the server must repeat all authoritative validation.

Compatibility rule: HTML `required` considers a value containing only spaces to be non-empty. This matches the existing application, where whitespace-only text is valid. Text must not be trimmed merely to make client and server validation stricter.

### 7.4 Request parsing limits

Generation and download accept only:

```text
Content-Type: application/x-www-form-urlencoded
```

Unsupported request media types return `415 Unsupported Media Type`.

Before parsing a POST form, the handler wraps the request body with a 64 KiB limit. The limit is applied before `ParseForm` so the parser cannot consume an unbounded body. Exceeding the limit returns `413 Content Too Large`.

Field limits:

```text
text       = at most 4,096 characters
substring  = at most 4,096 characters
width      = integer from 20 through 300
color      = exactly #RRGGBB when enabled
```

The textarea and substring input use matching `maxlength="4096"` attributes for immediate browser feedback. Server checks remain authoritative.

Generation values are read from `PostForm`, not from the merged `Form`, so query-string values cannot override or duplicate POST-body values.

The accepted form keys are exactly:

```text
text
banner
use_color
color
substring
align
width
```

Each is a single-value field. Duplicate values and unexpected form keys return `400 Bad Request` rather than being silently ignored or resolved by first-value selection. Submit buttons do not need names and therefore do not add form keys.

---

## 8. Page Actions and Initial Route Contract

### 8.1 Generate

The Generate button submits the current form:

```text
POST /ascii-art
```

Immediately before submission, a small JavaScript helper measures how many monospace text columns fit inside the terminal output area and stores that number in the hidden `width` field. The server validates the submitted width, generates and aligns the requested output once, then returns the main page with the submitted values and rendered result.

JavaScript is responsible only for measuring the terminal area. It must not render glyphs or implement left, center, right, or justify algorithms. If JavaScript is unavailable or measurement fails, the form submits the default width of 80 columns.

### 8.2 Clear

The Clear control requests a fresh main page:

```text
GET /
```

The response contains the default form state, no result, and no error.

Clear must not rely only on `<button type="reset">`. An HTML reset restores server-rendered initial field values and does not remove server-rendered output. Clear may be implemented as a link styled as a button or as a separate GET form, but HTML forms must not be nested.

### 8.3 Download

The Download button submits the current generation values to:

```text
POST /ascii-art/download
```

The server uses the same validation and generation workflow as browser display, then prompts browser download behavior with an attachment response. Immediately before submission, JavaScript measures the usable browser viewport in monospace columns and replaces only the download form's `width` value. The already-rendered browser result is not changed.

The initial download format is portable plain text:

```text
Content-Type: text/plain; charset=utf-8
Content-Disposition: attachment; filename="ascii-art.txt"
Content-Length: exact attachment body size in bytes
```

The downloaded file preserves ASCII characters, spaces, logical rows, and server-generated alignment. ANSI color control sequences are removed, so the file remains readable in ordinary text editors. Browser color remains a presentation feature of the HTML result.

The page remembers the width used to generate the displayed result as a no-JavaScript and measurement-failure fallback. On Download, the client recalculates width from the current browser viewport and active monospace character width. The server validates the resulting 20–300 column value and regenerates the snapshot with the selected alignment. Resizing after generation therefore leaves the browser result fixed but can intentionally change the downloaded alignment.

Download represents the successfully displayed result, not ungenerated edits in the main form. After successful generation, the page contains a separate, non-nested download form with an escaped snapshot of:

```text
text
banner
use_color, when enabled
color
substring
align
download width, initialized from the rendered width as a fallback
```

The editable generation form and download snapshot are independent. Editing visible fields does not change the downloadable content until Generate succeeds again. Browser resizing may change only the alignment width calculated when Download is pressed.

Multiline snapshot text may be held in a hidden textarea so real line endings are preserved safely. Other snapshot fields may use hidden inputs. All values remain normal `html/template` data and are escaped.

Hidden fields are not trusted. The download endpoint applies the same body, encoding, duplicate-field, field-value, and generation validation as the main POST endpoint.

The complete generated ASCII output is not stored in the browser form, and the server does not introduce sessions for this feature. The download endpoint regenerates the validated snapshot through the shared workflow, keeping the renderer as the source of truth.

A standalone colored HTML download is a planned future enhancement. It is not part of the first implementation and must be designed separately with its own safe template and tests.

---

## 9. Text and Newline Contract

### 9.1 Accepted text

The application accepts:

- printable ASCII input supported by the banner files;
- spaces, including whitespace-only input;
- real line endings produced by pressing Enter in the textarea;
- literal backslash-plus-`n` sequences.

An entirely empty text value is invalid. Whitespace-only text is valid and must not be rejected by trimming.

### 9.2 Supported line-separator representations

The following representations are logical line separators:

```text
literal \n
LF
CRLF
CR
```

They must produce equivalent logical lines.

Examples:

```text
Textarea with:  Hello<Enter>World
Literal input:  Hello\nWorld
Result:         ["Hello", "World"]
```

### 9.3 Normalization ownership

Input normalization belongs to `internal/renderer`, not individual HTTP handlers. `renderer.NormalizeInput` should eventually be extended to canonicalize real browser line endings while retaining all existing literal-`\n` behavior.

The conceptual normalization order is:

1. Convert CRLF to LF.
2. Convert remaining CR to LF.
3. Treat LF and literal `\n` as equivalent logical separators.
4. Preserve the existing behavior for empty logical lines and trailing escaped newlines unless a later decision explicitly changes it.

CRLF must be handled before individual CR and LF values so one Windows line ending does not become two separators.

Normalization must not:

- trim text;
- remove leading or trailing spaces;
- collapse repeated spaces;
- change case;
- perform HTML escaping;
- perform glyph rendering.

---

## 10. Banner Contract

The form exposes exactly three bundled banners:

```text
standard
shadow
thinkertoy
```

The selected radio value is a public identifier, not a path. The web layer maps it to a fixed internal path:

```text
standard   -> banners/standard.txt
shadow     -> banners/shadow.txt
thinkertoy -> banners/thinkertoy.txt
```

Any other value is invalid. Browser input must never be concatenated into a filesystem path.

---

## 11. Color Contract

### 11.1 Enabling color

Color is controlled by the independent `use_color` checkbox.

```text
use_color off
    -> ignore the picker value for rendering
    -> ignore substring for rendering
    -> generate uncolored output

use_color on
    -> validate and apply the selected color
```

### 11.2 Public web color format

The browser color wheel submits `#RRGGBB`. This is the public color format for the web form.

The server must not trust the browser control alone. When color is enabled, it must validate:

- exactly seven characters;
- a leading `#`;
- exactly six hexadecimal digits.

The validated hexadecimal components are converted into three integers:

```text
Red   = 0..255
Green = 0..255
Blue  = 0..255
```

The existing renderer may retain named, RGB, and HSL parsing for compatibility with existing internal behavior, but the web form does not expose those textual syntaxes. The color wheel provides their RGB-equivalent colors without requiring users to know a notation.

### 11.3 Substring coloring

Existing substring semantics are preserved:

- color enabled with an empty substring colors the complete input;
- color enabled with a non-empty substring colors every matching occurrence;
- a substring with no match is successful and produces no colored glyphs;
- matching is case-sensitive;
- repeated matches are selected;
- overlapping matches are selected;
- matching occurs independently within each normalized line;
- matching does not cross line boundaries;
- substring spaces are significant and must not be trimmed.

When color is disabled, substring input does not affect rendering.

### 11.4 Safe browser presentation

The renderer currently represents colored glyph rows with ANSI foreground and reset sequences. Browsers do not interpret ANSI sequences as colors.

Browser output will use inline `<span>` elements styled with CSS. Go code must not concatenate untrusted values into raw HTML or arbitrary CSS.

The browser presenter should produce ordinary template data conceptually equivalent to:

```go
type RGB struct {
    Red   int
    Green int
    Blue  int
}

type StyledRun struct {
    Text    string
    Colored bool
}
```

The presenter identifies renderer-generated colored and uncolored runs. The HTML template creates spans for colored runs and builds fixed CSS syntax from validated numeric RGB components:

```html
<span style="color: rgb(R, G, B)">colored run</span>
```

Security requirements:

- rendered text remains a normal string escaped by `html/template`;
- RGB components are validated integers from 0 through 255;
- the user cannot provide arbitrary CSS declarations;
- no user-controlled value is marked as `template.HTML`;
- no user-controlled value is marked as arbitrary `template.CSS`;
- unexpected ANSI commands are not interpreted as HTML or general terminal behavior.

Adjacent runs with the same color state may be merged without changing visible output.

### 11.5 Presentation order

Color presentation occurs after server-side alignment:

```text
render with ANSI color state
    -> align while ignoring ANSI width
    -> convert aligned output to safe styled runs
    -> execute HTML template
```

Converting to HTML before alignment would introduce markup that does not represent visible columns and would complicate width calculation.

---

## 12. Browser Result

The generated result is displayed on the main page in a whitespace-preserving element, normally `<pre>`.

The result presentation must:

- preserve every generated space and line break;
- use a monospace font;
- allow horizontal scrolling when content exceeds the available viewport;
- wrap only colored runs in inline spans;
- avoid adding template indentation or whitespace inside the ASCII output;
- contain no visible ANSI escape sequences;
- safely escape all rendered text.

CSS is responsible for visual color and browser layout. The existing Go output logic remains responsible for ASCII-art alignment semantics.

### 12.1 Width measurement and resize behavior

The terminal-like output container is present before generation so its usable content width can be measured. On Generate, JavaScript measures the container and the active monospace character width, calculates a text-column count, and submits that count once.

Browser display and download have deliberately separate width measurements. On Download, JavaScript measures `document.documentElement.clientWidth`, divides it by the same monospace character width, clamps the result to 20–300 columns, and updates only the download snapshot's hidden `width` field. If viewport measurement is unavailable, that field retains the server-rendered browser width as its fallback.

The server treats `width` as untrusted input, validates it, and applies alignment through the existing width-explicit output logic. Missing generation width falls back to 80 columns; failed download measurement retains the generated browser width already present in the download form. Explicit widths must be integers from 20 through 300 columns; malformed or out-of-range explicit values are invalid requests.

Generated output is a fixed snapshot at the submitted width. Browser resizing does not send another HTTP request and does not dynamically realign existing output. If the available area later becomes narrower than the result, CSS displays horizontal overflow:

```css
.terminal-body {
    max-width: 100%;
    overflow-x: auto;
    white-space: pre;
    font-family: monospace;
}
```

The user can press Generate again to render the browser result for the newly measured output-panel size. Dynamic browser re-alignment during window resizing is outside the implementation; only a later Download measures the current viewport again for the text attachment.

---

## 13. Form State and Errors

After either successful generation or a validation error, the returned page preserves:

- text;
- banner selection;
- color-enabled state;
- selected color;
- substring;
- alignment.

The result is shown only after successful generation. A validation failure shows a useful error without exposing internal filesystem paths, stack traces, or implementation details.

Clear is the only normal action that intentionally returns all fields and output to their default state.

### 13.1 HTTP status contract

| Situation | Status |
| --- | --- |
| Successful page, generation, or download | `200 OK` |
| Empty text | `400 Bad Request` |
| Unknown banner | `400 Bad Request` |
| Invalid enabled color | `400 Bad Request` |
| Invalid alignment | `400 Bad Request` |
| Malformed or out-of-range explicit width | `400 Bad Request` |
| Unsupported input character | `400 Bad Request` |
| Malformed form encoding | `400 Bad Request` |
| Request body exceeds the configured limit | `413 Content Too Large` |
| Unsupported request media type | `415 Unsupported Media Type` |
| Unknown route | `404 Not Found` |
| Unsupported method on a known route | `405 Method Not Allowed` |
| Template, presenter, or unexpected generation failure | `500 Internal Server Error` |

Known routes that reject a method include the appropriate `Allow` header.

Unknown routes render the dedicated minimal `not_found` template with `404 Not Found`, an HTML media type, and a link back to `/`. The response does not echo the requested path. Known routes with unsupported methods remain `405 Method Not Allowed`; the not-found fallback must not hide that distinction.

Missing templates and bundled banners are initialization failures, not user-facing missing resources. Users select stable public identifiers; they do not request internal files directly. The server does not begin listening when a required startup resource is unavailable or invalid.

### 13.2 Error response behavior

Invalid generation input returns the main HTML page with:

- the actual error status rather than `200`;
- a safe, useful validation message;
- submitted form values preserved;
- no generated result.

Internal errors are logged with enough detail for developers, while client responses omit filesystem paths, stack traces, raw internal errors, and template details.

Templates should be executed into a temporary buffer before committing the response status and body. This allows a template-execution error to become a clean `500` instead of leaving a partial response with an already-written status. Successful buffered HTML responses set `Content-Length` from the final byte count.

A failed download must not include `Content-Disposition: attachment`. Validation failures return the normal HTML form with an error; unexpected failures return a safe `500` response.

---

## 14. Server Lifecycle and Logging

### 14.1 Address and timeouts

Running `go run .` starts the server on:

```text
:8080
```

Users access it locally at `http://localhost:8080`. The first version uses a fixed port; environment or flag configuration is outside the current scope.

The application uses an explicit `http.Server` with:

```text
ReadHeaderTimeout = 5 seconds
ReadTimeout       = 15 seconds
WriteTimeout      = 30 seconds
IdleTimeout       = 60 seconds
```

Required templates and handlers are initialized before the server begins listening. Initialization or listen failures are logged and cause a non-zero process exit rather than leaving a partially working server.

### 14.2 Graceful shutdown

Interrupt and termination signals start graceful shutdown. The server stops accepting new connections and gives active requests up to five seconds to finish.

`http.ErrServerClosed` after an intentional shutdown is treated as a normal result. Other listen or shutdown errors are reported as failures.

### 14.3 Logging destination

The application does not create or manage `log.txt` or another log file. A standard-library structured logger writes to standard error. The execution environment may display, redirect, rotate, or collect those logs.

The logger is constructed during startup and passed to components that need it. It is not kept as mutable global state.

Logs may include:

- server startup and listening address;
- startup and shutdown failures;
- graceful shutdown lifecycle;
- request method and safe route path;
- response status and request duration;
- safe identifiers such as the selected bundled banner;
- internal errors needed for diagnosis.

Logs must not include:

- submitted text;
- substring values;
- complete form bodies;
- generated ASCII output;
- cookies or authorization headers;
- arbitrary request headers.

Request metadata may be recorded through focused HTTP middleware. Logging middleware must not read or copy request bodies and must not contain application or rendering logic.

Client responses receive safe error messages. Detailed diagnostic errors remain in server logs.

### 14.4 Templates and static assets

The first version reads web assets from disk rather than embedding them in the executable.

Required layout:

```text
templates/
├── index.html
└── 404.html

static/
├── style.css
└── app.js
```

`index.html` and `404.html` are parsed together once during startup with `html/template`. The application validates both the main-page definitions and the `not_found` definition before listening. A missing file or template syntax error prevents the server from starting. The parsed template set is passed to handlers and is not reparsed for each request.

CSS and JavaScript are served from their fixed files under public URLs:

```text
/static/style.css -> static/style.css
/static/app.js     -> static/app.js
```

The first version should expose only known static assets rather than an unrestricted project-directory file server. Directory listing is not required.

All filesystem paths are application constants. Route parameters, form values, and other browser input must never be joined to template or static-asset paths.

The application must be started with the repository root as its working directory so the relative `templates/`, `static/`, and existing `banners/` paths resolve correctly. This working-directory requirement must be documented in the README.

Embedding templates and assets in the executable is a possible future packaging improvement, not a first-version requirement.

### 14.5 Banner initialization

The application loads and validates all three bundled banner files from fixed disk paths during startup:

```text
standard   -> banners/standard.txt
shadow     -> banners/shadow.txt
thinkertoy -> banners/thinkertoy.txt
```

The server begins listening only after all three banners load successfully. A missing or malformed required banner is logged as an initialization failure and causes a non-zero exit.

Successfully loaded glyph maps are treated as immutable and shared between request handlers. Handlers and renderers may read them concurrently but must not modify them.

Browser requests contain only the exact public identifiers `standard`, `shadow`, or `thinkertoy`. Lookup occurs in the preloaded registry. Browser values are never concatenated into banner paths, and no banner is read from disk in response to a client-supplied path.

Changing a banner file while the server is running does not change the in-memory banner. Restarting the server reloads the fixed files. Runtime banner editing is outside the product scope.

---

## 15. Initial Acceptance Criteria

The following criteria cover decisions made so far. More criteria will be added as remaining architecture decisions are resolved.

### Runtime and form

- [ ] `go run .` starts the web server rather than a CLI prompt.
- [ ] The main page contains text, banner, color, substring, and alignment controls.
- [ ] Banner and alignment use radio-button groups.
- [ ] Color uses an enable checkbox and a native color wheel.
- [ ] Generate displays a result on the main page.
- [ ] Clear returns an empty/default form and removes the result.
- [ ] Download submits the same generation values through the shared workflow.
- [ ] Generate measures the terminal area once and submits its width in text columns.
- [ ] Download measures the current browser viewport and submits a separate width without altering the browser result.
- [ ] Missing generation measurement falls back to 80 columns, while failed download measurement retains the rendered browser width.
- [ ] Resizing after generation does not send a request or alter the generated spacing.
- [ ] Narrow terminal containers provide horizontal scrolling.
- [ ] Native browser validation blocks ordinary empty-text submissions.
- [ ] The form does not disable native constraint validation.
- [ ] Client validation does not replace server validation.
- [ ] Text and substring controls expose a 4,096-character client limit.
- [ ] Generation and download accept URL-encoded forms only.
- [ ] Form values come from the POST body rather than merged query values.
- [ ] Duplicate and unexpected form fields are rejected.

### Text

- [ ] Empty text is rejected.
- [ ] Whitespace-only text is accepted.
- [ ] LF, CRLF, CR, and literal `\n` separators produce equivalent logical lines.
- [ ] Existing literal-`\n` compatibility behavior remains tested.
- [ ] Leading, trailing, and repeated spaces are preserved.

### Color

- [ ] Color is not applied when `use_color` is disabled.
- [ ] Enabling color requires a valid `#RRGGBB` value.
- [ ] The server converts the validated value to RGB integers.
- [ ] An empty substring colors all text.
- [ ] Repeated, overlapping, and case-sensitive substring behavior is preserved.
- [ ] A missing substring match is not an error.
- [ ] Browser output uses styled spans and contains no visible ANSI codes.
- [ ] Rendered text is escaped by `html/template`.
- [ ] Arbitrary CSS and HTML injection attempts are rejected or safely escaped.

### Download

- [ ] Download returns `ascii-art.txt` as a plain-text attachment.
- [ ] Download declares its exact byte size with `Content-Length`.
- [ ] Download uses viewport-derived alignment width and falls back to the displayed result's recorded width when measurement is unavailable.
- [ ] Download preserves ASCII spaces, rows, and alignment.
- [ ] Download contains no ANSI color control sequences.
- [ ] Browser input cannot select or influence a server-side output path.
- [ ] Download appears only after successful generation.
- [ ] Download uses a separate form containing the successful generation snapshot.
- [ ] Editing the main form does not change the downloadable content.
- [ ] Snapshot values are escaped, revalidated, and regenerated server-side.
- [ ] Generated ASCII output is not trusted or round-tripped through hidden fields.
- [ ] Download requires no server-side session state.

### Regression

- [ ] Existing banner, renderer, and output behavior remains covered by tests.
- [ ] The HTTP and download paths do not duplicate rendering algorithms.
- [ ] `main.go` contains process lifecycle wiring rather than HTTP or rendering algorithms.
- [ ] HTTP behavior is independently testable through `internal/web` without opening a network port.
- [ ] Templates, banner maps, and the logger are injected dependencies rather than mutable globals.
- [ ] Request-specific state is never shared between concurrent requests.
- [ ] Raw browser strings are not passed directly into generation.
- [ ] Validated widths and RGB components use bounded numeric values.
- [ ] Browser and plain-text output use the same generation and presentation pipeline while allowing their validated alignment widths to differ.
- [ ] Templates receive presentation data rather than renderer or HTTP internals.
- [ ] CLI code remains during initial web construction rather than being deleted in the first feature change.
- [ ] Obsolete root CLI tests are replaced when `main.go` becomes a server.
- [ ] No supported source imports `internal/cli` before final CLI removal.
- [ ] The unused CLI package and its tests are removed only after the web application passes its complete test suite.
- [ ] Banner, renderer, and output packages and their regression tests remain.

### HTTP errors

- [ ] Invalid form values return `400` with preserved state and no result.
- [ ] Oversized request bodies return `413`.
- [ ] Unsupported request media types return `415`.
- [ ] Unknown routes return the minimal HTML 404 page without echoing the requested path.
- [ ] Wrong methods return `405` with an `Allow` header.
- [ ] Missing required startup resources prevent the server from listening.
- [ ] Unexpected internal request failures return `500`.
- [ ] Error responses do not expose internal implementation details.
- [ ] Failed downloads do not include attachment headers.

### Server lifecycle and logging

- [ ] The server listens on `:8080` with the documented timeouts.
- [ ] Initialization completes before the server starts accepting requests.
- [ ] Startup failures produce a non-zero exit.
- [ ] Interrupt and termination signals trigger graceful shutdown with a five-second deadline.
- [ ] Intentional `http.ErrServerClosed` is not reported as a failure.
- [ ] Structured logs are written to standard error rather than an application-managed file.
- [ ] Logs contain safe lifecycle and request metadata but not submitted or generated content.
- [ ] `templates/index.html` and `templates/404.html` are read from disk and parsed together once during startup.
- [ ] Missing or invalid templates prevent startup.
- [ ] All three bundled banners are loaded and validated before listening.
- [ ] Preloaded banner maps are treated as immutable shared data.
- [ ] Requests select banners only through exact registry identifiers.
- [ ] Only fixed, known CSS and JavaScript paths are exposed publicly.
- [ ] No browser value can influence a template or static-asset filesystem path.
- [ ] Documentation states that the server runs from the repository root.

---

## 16. Test Strategy

Testing is layered so failures identify the responsibility that broke. Automated HTTP tests use the standard-library `httptest` package and do not bind the production port.

### 16.1 Existing regression tests

Existing tests remain for:

- official, missing, malformed, and CRLF banner loading;
- structured rendering and flattening;
- supported and unsupported input characters;
- color parsing and conversion;
- repeated, overlapping, missing, and case-sensitive substring matching;
- left, center, right, and justify alignment;
- ANSI-aware visible-width calculation;
- text and file writers.

New normalization tests cover LF, CRLF, CR, literal `\n`, mixed separators, empty logical lines, and existing trailing-newline compatibility without removing existing cases.

### 16.2 Form parsing and validation tests

Table-driven tests cover valid combinations of:

- each banner;
- each alignment;
- enabled and disabled color;
- empty and non-empty substrings;
- whitespace-only text;
- real and literal newlines;
- widths at 20, 80, and 300;
- missing width with the 80-column fallback.

Invalid cases include:

- missing or empty text;
- text or substring over 4,096 characters;
- missing or unknown banner;
- missing or unknown alignment;
- malformed enabled color;
- non-numeric, negative, below-minimum, and above-maximum widths;
- duplicate single-value fields;
- unexpected form keys;
- query values attempting to override POST values;
- malformed URL encoding;
- request bodies over 64 KiB;
- unsupported media types.

Tests distinguish invalid color with `use_color` enabled from ignored color values when coloring is disabled.

### 16.3 Shared generation tests

The generation workflow is tested independently from HTTP for:

- all three preloaded banners;
- normalized multiline input;
- all four alignments at explicit deterministic widths;
- enabled and disabled color;
- whole-input and substring coloring;
- repeated and overlapping substring matches;
- unsupported-character errors;
- derivation of browser and download content through the same workflow at independently supplied alignment widths.

### 16.4 ANSI presenter tests

Presenter tests cover every renderer-generated foreground form and reset sequence:

```text
ESC[30m through ESC[37m
ESC[38;5;208m
ESC[38;2;R;G;Bm
ESC[0m
```

They verify:

- ANSI sequences do not appear in styled-run text;
- ANSI sequences do not appear in plain download text;
- visible characters, spaces, logical rows, and the final newline convention are preserved;
- colored and uncolored runs are classified correctly;
- safe merging of adjacent runs does not change visible text;
- incomplete or unsupported ANSI commands cause internal presentation errors.

### 16.5 HTTP route and template tests

`httptest` cases verify:

- `GET /` returns `200` HTML with the complete default form;
- required and `maxlength` attributes are present;
- banner and alignment defaults are selected;
- initial responses contain no result or download snapshot;
- valid `POST /ascii-art` returns `200`, preserved form state, rendered output, rendered width, and a download snapshot;
- invalid generation returns `400`, preserved safe values, an error, and no result;
- Clear returns the default page with no result or error;
- unknown routes return the minimal HTML 404 page with a home link and do not expose the requested path;
- internal template paths such as `/templates/404.html` remain inaccessible;
- wrong methods return `405` with the correct `Allow` header.

Tests assert focused status, header, state, and security behavior rather than comparing one brittle full-page HTML string.

### 16.6 HTML and request security tests

Security cases include printable input containing HTML syntax such as:

```text
<script>alert("x")</script>
<img src=x onerror=alert(1)>
& < > " '
```

Tests verify:

- preserved form values are escaped;
- user input cannot create executable elements or event attributes;
- ANSI bytes are absent from browser HTML;
- arbitrary color/CSS syntax is rejected when coloring is enabled;
- RGB presentation components remain bounded integers;
- traversal-like banner and asset requests cannot influence filesystem paths;
- oversized, ambiguous, duplicate, and unsupported requests receive their documented statuses.

### 16.7 Download tests

Successful download tests verify:

```text
status              = 200
Content-Type        = text/plain; charset=utf-8
Content-Disposition = attachment; filename="ascii-art.txt"
Content-Length      = exact response body size in bytes
```

They also verify:

- aligned ASCII spaces and rows are preserved;
- ANSI escape bytes are absent;
- all banners and multiline input work;
- the submitted download width is used and remains independently validated;
- client-side download measurement targets the current browser viewport and retains the recorded displayed width as fallback;
- a submitted snapshot renders the displayed values rather than later editable-form values;
- no server file is created;
- clients cannot supply filenames or paths;
- failed downloads have no attachment header.

### 16.8 Static assets, logging, and startup tests

Static-asset tests verify:

- `/static/style.css` and `/static/app.js` return their expected content types;
- `/static/` does not expose a directory listing;
- templates, banners, and traversal-like paths are not publicly served.

Logging middleware is tested with an in-memory log destination. Logs include method, safe path, status, and duration while excluding a unique marker submitted as text, substrings, form bodies, generated output, cookies, and arbitrary headers.

Startup helpers are tested for:

- valid templates and all banners;
- missing and syntactically invalid templates;
- missing and malformed banners;
- the documented address and server timeouts;
- complete dependency initialization before route construction.

Graceful shutdown tests must avoid fixed ports and timing-sensitive behavior. Lifecycle helpers should be tested directly where practical.

### 16.9 Concurrency and tool verification

After normal tests pass, verification includes:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Independent handler tests may use `t.Parallel()` when they do not mutate environment variables or other process-wide state. Shared templates, banner maps, and logger dependencies are read concurrently, while form and result values remain request-local.

### 16.10 Manual browser checklist

Manual browser verification covers behavior that Go tests do not execute directly:

- native required-field feedback;
- color-wheel interaction;
- labels, radio groups, and keyboard navigation;
- terminal-like appearance and monospace spacing;
- correct one-time terminal-width measurement;
- horizontal scrolling after narrowing the window;
- no resize-triggered HTTP requests;
- regeneration at a newly measured width;
- Clear removing both form state and output;
- Download prompting for `ascii-art.txt`;
- viewport-based text-download alignment without changing the browser result;
- readable downloaded text.

The JavaScript measurement helper remains intentionally small rather than introducing a JavaScript testing framework for the first version.

Test completeness is judged by public behavior, validation branches, security boundaries, and preserved rendering contracts rather than an arbitrary coverage percentage.

---

## 17. Decisions Recorded So Far

| Topic | Decision |
| --- | --- |
| Runtime | Web-only; `go run .` starts the server |
| CLI | Not a public interface or compatibility requirement |
| Text control | Multiline textarea |
| Newlines | Accept literal `\n`, LF, CRLF, and CR |
| Normalization owner | Renderer |
| Banner control | Radio-button group |
| Banner values | Three fixed bundled identifiers |
| Color enablement | Independent checkbox |
| Color selection | Native color wheel |
| Web color value | Strict `#RRGGBB`, converted to numeric RGB |
| Substring control | Optional single-line text input |
| Substring matching | Preserve existing renderer behavior |
| Alignment control | Radio-button group |
| Generate | POST current form and display result |
| Clear | GET a fresh default page |
| Download | POST the displayed result's snapshot through the shared generation workflow |
| Download meaning | Download the successfully displayed result, not later form edits |
| Download state | Separate escaped hidden snapshot form; no nested forms |
| Download trust | Revalidate snapshot inputs and regenerate server-side |
| Server session | Not required for generation or download state |
| Initial download format | Portable `ascii-art.txt` without ANSI color sequences |
| Download response headers | Plain-text `Content-Type`, attachment `Content-Disposition`, and exact byte `Content-Length` |
| Future download format | Standalone colored HTML document, outside the first version |
| Browser color | Inline spans created by the template using validated RGB integers |
| HTML safety | Keep text escaped; do not inject raw user HTML or CSS |
| Generation width | Measure terminal columns once immediately before Generate |
| Download width | Measure browser viewport columns immediately before Download |
| Generation width fallback | 80 columns when output-panel measurement is unavailable |
| Download width fallback | Retain the server-rendered browser width when viewport measurement is unavailable |
| Width limits | Explicit client width must be 20–300 columns |
| Window resize | Keep the generated snapshot and show horizontal overflow |
| Client validation | Native HTML constraints provide immediate feedback |
| Authoritative validation | Server validates every request independently |
| Empty text | Rejected by both browser and server |
| Whitespace-only text | Accepted for compatibility |
| Invalid form values | `400 Bad Request` with preserved form state |
| Oversized body | `413 Content Too Large` |
| Request body limit | 64 KiB, applied before form parsing |
| Text limit | 4,096 characters |
| Substring limit | 4,096 characters |
| Form encoding | `application/x-www-form-urlencoded` only |
| Form source | Read `PostForm`; do not merge query values |
| Form cardinality | Reject duplicate single-value fields and unexpected keys |
| Unsupported media type | `415 Unsupported Media Type` |
| Unknown route | Minimal HTML `404 Not Found` page with a link to `/` |
| Wrong method | `405 Method Not Allowed` with `Allow` |
| Internal failure | Safe `500 Internal Server Error`; details logged server-side |
| Server address | Fixed `:8080` for the first version |
| Server timeouts | Header 5s, read 15s, write 30s, idle 60s |
| Shutdown | Graceful on interrupt/termination with a five-second deadline |
| Logging API | Standard-library structured logger |
| Logging destination | Standard error; runtime environment owns persistence |
| Logging privacy | Do not log form values, input text, substrings, or output |
| Template loading | Read `index.html` and `404.html` from disk and parse them together once during startup |
| Static assets | Serve fixed disk files under `/static/` |
| Asset security | No request-derived filesystem paths or directory listing |
| Working directory | Run from the repository root |
| Asset embedding | Deferred as a possible packaging improvement |
| Banner loading | Load and validate all three fixed files during startup |
| Banner lifetime | Share immutable parsed maps for all requests |
| Banner selection | Exact registry lookup; never derive a path from browser input |
| Process entry point | `main.go` owns initialization, server lifecycle, and shutdown |
| HTTP package | `internal/web` owns routes, forms, handlers, presentation, and middleware |
| Dependency ownership | Inject templates, banner registry, and logger during construction |
| Request state | Keep form, result, and error data local to each request |
| Package scope | Do not add an application/service package without another real consumer |
| Data trust boundary | Separate raw `FormState` from validated `GenerationInput` |
| Generator result | Return aligned ANSI text without HTML concerns |
| Browser presentation | Convert generated ANSI state into normal styled-run data |
| Plain presentation | Remove only recognized ANSI sequences and preserve visible text exactly |
| Template model | Provide page-specific state with an explicit `HasResult` flag |
| CLI transition | Keep temporarily during web construction |
| Root CLI tests | Replace when the executable becomes a long-running server |
| CLI final state | Remove unused package and tests in a separate final cleanup |
| Preserved packages | Keep banner, renderer, and output implementations and tests |

---

## 18. Open Design Decisions

No known product or architectural decision remains open for the first implementation. Before coding begins, this PRD must receive a final consistency review against `docs/exercise.md`, `AGENTS.md`, the current source, and the current regression tests.

Any newly discovered ambiguity must be added here and resolved explicitly rather than assumed during implementation.
