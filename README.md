# ascii-art-web

Render text as ASCII art in the browser. A Go HTTP server wraps an existing
ASCII-art engine in a web interface: choose a banner, an alignment, and an
optional colour, and the rendered art appears on the same page — ready to read
or download.

Built with the Go standard library only. No third-party modules, no build step.

---

## Description

The command-line version of this project rendered ASCII art in a terminal.
ascii-art-web exposes the same capabilities through a browser: the flags become
form controls, the terminal becomes an output panel, and the output file becomes
a download.

What you can do:

| | |
| --- | --- |
| **Banners** | `standard`, `shadow`, `thinkertoy` |
| **Alignments** | `left`, `center`, `right`, `justify` |
| **Colour** | any `#RRGGBB`, applied to the whole text or to every occurrence of a substring |
| **Multiline input** | real newlines from the textarea, plus the literal `\n` escape |
| **Download** | the displayed result as `ascii-art.txt` |
| **Theme** | light or dark, remembered for the browser tab |

The web layer does not reimplement any of the rendering. It reuses the original
`banner`, `renderer`, and `output` packages unchanged, and adds only what HTTP
requires: routing, request validation, and safe presentation.

---

## Authors

| Name | Handle | Package |
| --- | --- | --- |
| Aris Kasapidis | `akasapid` | A — core pipeline |
| Spiros | `skourou` | B — form, validation, download |
| Kostis | `ksfakiana` | C — colour, presentation, infrastructure |


The work was split **vertically**: instead of one person per layer, each of us
owned a slice of the product end-to-end, from the Go handler down to the HTML
and CSS. `docs/workFlow.md` records the split and `docs/tasks/` the individual
task cards.

---

## Usage: how to run

### Requirements

- **Go 1.22.2** or newer
- The bundled `banners/`, `templates/`, and `static/` directories

### Start the server

Run from the **repository root**:

```bash
go run .
```

```text
http://localhost:8080
```

The server resolves `templates/index.html`, `templates/404.html`,
`banners/*.txt`, and `static/*`
relative to the current working directory. Starting it from anywhere else fails
immediately at startup with a clear message rather than serving a broken page.

Stop it with `Ctrl+C`. Shutdown is graceful: in-flight requests get up to five
seconds to finish before the process exits.

### Run with Docker

Build the production image from the repository root:

```bash
docker image build -f Dockerfile -t ascii-art-web-docker .
```

Start the container and publish the web server on port 8080:

```bash
docker container run \
  --detach \
  --name dockerize \
  --publish 8080:8080 \
  ascii-art-web-docker
```

Open `http://localhost:8080`, then inspect the container and its logs with:

```bash
docker ps -a
docker logs dockerize
docker exec -it dockerize /bin/bash
```

Inside the container, `/app` contains only the compiled `server` and the
`banners/`, `templates/`, and `static/` runtime directories. The server runs as
the unprivileged `app` user. Stop and remove the container with:

```bash
docker stop dockerize
docker rm dockerize
```

The convenience script builds the image, replaces an existing container named
`dockerize`, and starts the new container:

```bash
./build.sh
```

The script accepts optional `IMAGE_NAME`, `CONTAINER_NAME`, and `HOST_PORT`
environment variables. If the local Docker daemon requires elevated access,
run Docker commands—or the script—with the privileges configured for that
machine.

The Dockerfile uses a multi-stage build. Tests and compilation run in the Go
builder, while the final Debian image contains no Go toolchain or source code.
OCI labels provide the image title, description, and authors. The final image
documents port 8080, starts the server directly so it receives termination
signals, and retains `/bin/bash` for audit-time filesystem inspection.

### Using the page

1. Type the text to render. Press Enter for a new line; a literal `\n` typed as
   two characters works too.
2. Pick a banner and an alignment.
3. Optionally tick **Use color**, choose a colour, and enter a substring to
   colour only the matching parts. An empty substring colours everything.
4. Press **Generate**. The art appears below and your form values are kept.
5. Press **Download** to save the result as `ascii-art.txt`. At download time,
   alignment is recalculated against the current browser viewport width.
6. Press **Clear** to start over.

### Routes

```text
GET   /                       main page, and the target of Clear
POST  /ascii-art              generate and display
POST  /ascii-art/download     download the displayed result
GET   /static/style.css       stylesheet
GET   /static/app.js          column measurement and theme toggle
*     all other paths         custom 404 page
```

### Form fields

Exactly seven keys are accepted. Anything else — an unknown key, or the same key
twice — is rejected.

| Field | Type | Rule |
| --- | --- | --- |
| `text` | textarea | required, up to 4096 characters (whitespace-only is valid) |
| `banner` | radio | `standard` \| `shadow` \| `thinkertoy` |
| `use_color` | checkbox | `on` when present |
| `color` | colour input | `#RRGGBB`, validated only when colouring is on |
| `substring` | text | optional, up to 4096 characters |
| `align` | radio | `left` \| `center` \| `right` \| `justify` |
| `width` | hidden | integer 20–300; generation defaults to 80 and download retains the rendered width if viewport measurement fails |

### Status codes

| Code | When |
| --- | --- |
| `200` | success |
| `400` | empty text, unknown banner or alignment, invalid colour, bad width, unsupported character, malformed encoding, duplicate or unexpected field |
| `404` | unknown route, rendered with the minimal not-found page |
| `405` | wrong method on a known route (with an `Allow` header) |
| `413` | request body over 64 KiB |
| `415` | content type other than `application/x-www-form-urlencoded` |
| `500` | template, presenter, or unexpected internal failure |

Validation error pages keep the values you submitted, so nothing has to be
retyped. Unknown routes use a separate minimal 404 page with a link home.

### Tests

```bash
go test ./...        # full suite
go test -race ./...  # concurrency
go vet ./...         # static checks
gofmt -l .           # formatting (silence means clean)
```

---

## Implementation details: algorithm

### One pipeline, two endings

The page and the download run the **same generation pipeline**. The browser
aligns against the output panel; immediately before download, JavaScript
replaces only the download form's width with the current viewport width in
monospace columns. The server then regenerates the validated snapshot as plain
text, so content stays consistent while the download can use a wider alignment.

```text
HTTP request
  |
  |- parse & validate ----------- internal/web/form.go
  |     untrusted strings -> bounded, checked values
  |
  |- choose width --------------- output panel | download viewport
  |
  |- generate ------------------- internal/web/generate.go
  |     banner -> normalize -> build -> align at chosen width
  |
  |- present -------------------- internal/web/presenter.go
  |     ANSI -> styled runs (HTML) + plain text (download)
  |
  `- respond -------------------- HTML page  |  text attachment
```

### 1. Validation

Two boundaries are crossed before a single value is trusted.

**Request level.** The content type must be `application/x-www-form-urlencoded`.
The body is wrapped in a 64 KiB limit **before** `ParseForm` reads it, so a
malicious client can never make the parser allocate without bound. Values are
read from `PostForm` rather than the merged `Form`, which means a query-string
parameter cannot smuggle in a value or conflict with a submitted one.

**Field level.** Raw values are copied into a `FormState` that stays untrusted —
its only job is to be echoed back onto an error page unchanged. Validation then
produces a `GenerationInput` containing nothing but bounded, checked values.
Only that struct reaches generation.

The types enforce the boundary rather than relying on discipline:

```go
FormState        // raw, untrusted — for redisplay
GenerationInput  // validated, bounded — for generation
GeneratedASCII   // aligned ANSI text + the width it was aligned at
StyledLine       // safe runs — for the template
```

Text is deliberately **not** trimmed. HTML's `required` treats spaces as
content, and the original engine accepts whitespace-only input; silently
trimming would change established behaviour to make validation tidier.

### 2. Banner lookup

All three banner files are read and validated **once at startup** into a
read-only registry. Each file is 95 blocks of 9 lines — one block per printable
ASCII character, 32 to 126, of which 8 lines are art. A malformed file stops the
server rather than producing broken output at request time.

A request selects a banner by **identifier**, never by path. No browser value is
ever joined into a filename, which removes path traversal as a category of bug
rather than as a case to be caught.

### 3. Normalization

`renderer.NormalizeInput` splits the text into logical lines, treating four
separators as equivalent: `LF`, `CRLF`, a lone `CR`, and the literal `\n`
escape.

Order matters:

```text
"\r\n" -> "\n"     first  — a Windows line ending is one separator
"\r"   -> "\n"     second — a lone CR is also one
"\\n"  -> "\n"     the CLI-era escape joins the same alphabet
split on "\n"
```

Collapsing `CRLF` first is what keeps one Windows line ending from becoming two
blank-separated lines. After this step, exactly one kind of line break exists,
and every later stage can stop thinking about the question.

Spaces are never trimmed, collapsed, or reordered — leading, trailing, and
repeated spaces all survive into the rendered art.

### 4. Rendering

Each character maps to eight banner rows. Instead of returning flat rows, the
renderer returns **word and space segments**:

```text
"hi there" -> [word "hi"] [space " "] [word "there"]
```

This is what makes justify possible. Once rows are flattened into strings, a
space between words is indistinguishable from a space inside a glyph — and
expanding the wrong one shreds the letters. Keeping the structure lets alignment
expand only the real gaps.

When colouring is on, the substring matcher marks which character positions to
colour. Matching is case-sensitive, evaluated per line, and **includes
overlapping matches**: searching for `aa` in `aaa` colours all three characters,
because both matches are honoured rather than the first one consuming the input.
Those glyph rows are then wrapped in ANSI colour codes.

### 5. Alignment

| Mode | Behaviour |
| --- | --- |
| `left` | rows unchanged |
| `center` | pad each row by half the remaining width |
| `right` | pad each row by the full remaining width |
| `justify` | distribute the remaining columns across the gaps between words |

Justify gives any remainder to the **leftmost** gaps, matching how text
justification reads left to right. Only a space segment with a word on both
sides counts as a gap; leading and trailing spaces are never stretched.

All padding is computed from **visible** width. The width function walks the
string and skips ANSI escape sequences entirely, so a coloured row and an
uncoloured row of the same art pad identically. Without this, turning on colour
would silently shift the layout.

### 6. Width

Alignment needs a column count, and the two browser actions measure different
targets. Immediately before Generate, a hidden probe rendered in the terminal
font converts the output panel's pixel width into columns. Immediately before
Download, the same character measurement converts the current browser viewport
into columns without changing the browser result. Both values are clamped to
20–300. Generation falls back to 80; failed download measurement retains the
server-rendered generation width.

Each measurement happens **once**, when its form is submitted. Resizing sends no
request and does not re-align an existing browser result: the output is a
snapshot of the width it was generated at, and it scrolls sideways rather than
reflowing. A later Download intentionally observes the then-current viewport.

### 7. Presentation

The renderer speaks ANSI; browsers do not. The presenter scans the aligned
output byte by byte and produces both outputs in a single pass — styled runs for
HTML and an ANSI-free string for the download — which is what guarantees the two
can never disagree.

ANSI is treated as renderer-owned control data. Recognised foreground sequences
toggle a colour flag and are consumed:

```text
ESC[30m ... ESC[37m    basic foreground
ESC[38;5;208m          256-colour orange
ESC[38;2;R;G;Bm        truecolor
ESC[0m                 reset
```

Anything incomplete or unrecognised is an **internal failure**, not something to
pass through or guess at. No escape byte ever reaches the HTML or the downloaded
file.

Colour reaches the page as three integers. A `#RRGGBB` value is validated,
converted to `0–255` components, and rendered inside a fixed inline style:

```html
<span style="color: rgb(191, 215, 234)">...</span>
```

The rendered text stays an ordinary escaped template string. No user value is
ever marked as trusted HTML or CSS, so the only thing a request can influence is
three bounded numbers — injection has nowhere to land.

Output is written into a `<pre>` with `white-space: pre` inside a horizontally
scrolling container, preserving every space and line break exactly as generated.

### 8. Server lifecycle

```text
address              :8080
ReadHeaderTimeout    5s
ReadTimeout          15s
WriteTimeout         30s
IdleTimeout          60s
shutdown deadline    5s
```

Templates and banners load once at startup, so no request touches the filesystem
for them. Every page renders into a buffer first: a template failure becomes a
clean `500` instead of a half-written page with a `200` already sent. Buffered
HTML and successful text downloads declare their exact byte size with
`Content-Length`.

Structured logs go to **standard error** — method, route, status, duration, and
the selected banner. Submitted text, substrings, form bodies, generated output,
and headers are never logged. The application manages no log file of its own;
persistence belongs to whatever runs it.

---

## Project layout

```text
ascii-art-web/
|-- Dockerfile               multi-stage production image
|-- .dockerignore            files excluded from the build context
|-- build.sh                 build-and-run convenience script
|-- main.go                  startup, timeouts, graceful shutdown
|-- banners/                 standard.txt, shadow.txt, thinkertoy.txt
|-- templates/
|   |-- index.html           page shell and its blocks
|   `-- 404.html             minimal not-found page
|-- static/
|   |-- style.css
|   `-- app.js               panel/viewport measurement, theme toggle
|-- internal/
|   |-- banner/              banner files -> glyph maps
|   |-- renderer/            normalization, colour, substrings, glyph rows
|   |-- output/              visible width, alignment, writing
|   `-- web/                 routes, form, generation, presentation, handlers
`-- docs/                    PRD, workflow, task cards
```

`banner`, `renderer`, and `output` are carried over from the command-line
version **unchanged**. The web layer wraps them; it never duplicates them.

`templates/index.html` is split into named blocks — one owner per block — so
three people could edit the same page without stepping on each other.

---

## Limitations

- Only printable ASCII characters present in the banner files can be rendered;
  anything else is a `400` with a clear message.
- One colour and one substring per request. A substring containing a newline
  matches nothing.
- The browser result is a fixed snapshot: resizing the window does not re-align
  it until the next Generate.
- Download alignment targets the browser viewport at the moment Download is
  pressed. A text editor with a different font or window width may show a
  different visual center because plain text has no dynamic layout metadata.
- Text and substring are capped at 4,096 characters, the request body at 64 KiB.
- The theme is remembered per browser tab, not across tabs or sessions.
