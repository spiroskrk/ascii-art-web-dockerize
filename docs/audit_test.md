# ASCII Art Web — Test Cases

## 1. Test Objectives

The test suite should verify that:

* the Go server starts and runs correctly;
* only Go standard-library packages are used;
* the required HTML interface exists;
* all three bundled banners work;
* ASCII output matches the expected renderer output;
* HTTP routes use the correct methods;
* HTTP status codes are handled correctly;
* invalid requests do not crash the server;
* user input is safely displayed;
* browser input cannot control server-side file paths;
* existing renderer and output behavior remains unchanged.

---

# 2. Preconditions

Before running tests:

```bash
go test ./...
go run .
```

Confirm that the server starts successfully.

Example expected startup behavior:

```text
Server running on http://localhost:8080
```

Open the site in a browser:

```text
http://localhost:8080/
```

---

# 3. Compilation and Package Tests

## TC-001 — Project compiles

**Command**

```bash
go build ./...
```

**Expected result**

* Build completes successfully.
* No compilation errors.
* Exit code is `0`.

---

## TC-002 — All Go tests pass

**Command**

```bash
go test ./...
```

**Expected result**

* All packages compile.
* Existing renderer tests pass.
* Existing output tests pass.
* New HTTP tests pass.
* Exit code is `0`.

---

## TC-003 — Standard-library-only requirement

**Command**

```bash
go list -deps ./...
```

Review imports in:

```text
go.mod
main.go
internal/web/page_handlers.go
internal/web/download_handler.go
internal/
```

**Expected result**

* No third-party modules are required.
* `go.mod` contains no external dependencies.
* All imports belong to the Go standard library or local project packages.

---

## TC-004 — Code formatting

**Command**

```bash
gofmt -w .
git diff --exit-code
```

**Expected result**

* Source files are correctly formatted.
* No formatting changes remain after `gofmt`.

---

## TC-005 — Static analysis

**Command**

```bash
go vet ./...
```

**Expected result**

* No suspicious constructs or vet errors are reported.

---

# 4. Required Files

## TC-006 — HTML templates exist

**Steps**

1. Inspect the project root.
2. Inspect the `templates` directory.

**Expected result**

At least one HTML template exists:

```text
templates/
    index.html
```

---

## TC-007 — Required banners exist

**Expected files**

```text
banners/standard.txt
banners/shadow.txt
banners/thinkertoy.txt
```

**Expected result**

* All three files exist.
* The server can load each one.
* The files have not been modified unexpectedly.

---

## TC-008 — README exists

**Expected file**

```text
README.md
```

**Expected sections**

* Description
* Authors
* Usage
* Implementation details

---

# 5. Home Page Tests

## TC-009 — GET `/` succeeds

**Request**

```bash
curl -i http://localhost:8080/
```

**Expected result**

```text
HTTP/1.1 200 OK
```

The response contains HTML.

---

## TC-010 — Home page contains required controls

**Steps**

Open `/` in a browser.

**Expected result**

The page contains:

* text input or textarea;
* banner selector;
* standard option;
* shadow option;
* thinkertoy option;
* submit button;
* ASCII-art output area.

---

## TC-011 — Website instructions are clear

**Expected result**

The page clearly explains:

* where to enter text;
* how to select a banner;
* how to generate ASCII art;
* where the result will appear.

---

## TC-012 — Unsupported method on `/`

**Request**

```bash
curl -i -X POST http://localhost:8080/
```

**Expected result**

Either:

```text
405 Method Not Allowed
```

or a documented project-specific response.

The server must not crash.

---

# 6. POST Endpoint Tests

## TC-013 — POST `/ascii-art` succeeds

**Request**

```bash
curl -i \
  -X POST \
  -d "text=Hello&banner=standard" \
  http://localhost:8080/ascii-art
```

**Expected result**

```text
HTTP/1.1 200 OK
```

The response contains the generated ASCII art.

---

## TC-014 — GET `/ascii-art` is rejected

**Request**

```bash
curl -i http://localhost:8080/ascii-art
```

**Expected result**

Recommended:

```text
405 Method Not Allowed
```

The handler must not process the request as a normal generation request.

---

## TC-015 — Client-server communication works

**Steps**

1. Open the home page.
2. Enter text.
3. Select a banner.
4. Click the generate button.

**Expected result**

* The browser sends a POST request to `/ascii-art`.
* The server receives the form values.
* The renderer is called.
* The generated result appears on the page.
* No console or server errors occur.

---

# 7. Standard Banner Functional Tests

## TC-016 — Standard banner with multiline special characters

**Input**

```text
{123}
<Hello> (World)!
```

Use the `standard` banner.

**Expected result**

The output matches:

```text
   __                     __
  / /  _   ____    _____  \ \
 | |  / | |___ \  |___ /   | |
/ /   | |   __) |   |_ \    \ \
\ \   | |  / __/   ___) |   / /
 | |  |_| |_____| |____/   | |
  \_\                     /_/

   __  _    _          _   _          __            __ __          __                 _       _  __    _
  / / | |  | |        | | | |         \ \          / / \ \        / /                | |     | | \ \  | |
 / /  | |__| |   ___  | | | |   ___    \ \        | |   \ \  /\  / /    ___    _ __  | |   __| |  | | | |
< <   |  __  |  / _ \ | | | |  / _ \    > >       | |    \ \/  \/ /    / _ \  | '__| | |  / _` |  | | | |
 \ \  | |  | | |  __/ | | | | | (_) |  / /        | |     \  /\  /    | (_) | | |    | | | (_| |  | | |_|
  \_\ |_|  |_|  \___| |_| |_|  \___/  /_/         | |      \/  \/      \___/  |_|    |_|  \__,_|  | | (_)
                                                   \_\                                           /_/
```

**Additional checks**

* The browser displays literal `<Hello>`.
* `<Hello>` is not interpreted as an HTML element.
* Braces and punctuation appear correctly.
* Line breaks are preserved.

---

## TC-017 — Standard banner with question marks

**Input**

```text
123??
```

Use the `standard` banner.

**Expected result**

```text
                     ___    ___
 _   ____    _____  |__ \  |__ \
/ | |___ \  |___ /     ) |    ) |
| |   __) |   |_ \    / /    / /
| |  / __/   ___) |  |_|    |_|
|_| |_____| |____/   (_)    (_)
```

---

# 8. Shadow Banner Functional Test

## TC-018 — Shadow banner with punctuation

**Input**

```text
$% "=
```

Use the `shadow` banner.

**Expected result**

```text
                        _|  _|
  _|   _|_|    _|       _|  _|
_|_|_| _|_|  _|                _|_|_|_|_|
_|_|       _|
  _|_|   _|  _|_|              _|_|_|_|_|
_|_|_| _|    _|_|
  _|
```

**Additional checks**

* The quote character is handled correctly.
* The form submission is not broken by punctuation.
* Output whitespace is preserved.

---

# 9. Thinkertoy Banner Functional Test

## TC-019 — Thinkertoy banner with mixed characters

**Input**

```text
123 T/fs#R
```

Use the `thinkertoy` banner.

**Expected result**

```text
  0    --  o-o        o-O-o     o  o-o      | |  o--o
 /|   o  o    |         |      /   |       -O-O- |   |
o |     /   oo          |     o   -O-  o-o  | |  O-Oo
  |    /      |         |    /     |    \  -O-O- |  \
o-o-o o--o o-o          o   o      o   o-o  | |  o   o
```

---

# 10. Banner Selection Tests

## TC-020 — Standard banner accepted

**Request**

```bash
curl -i \
  -X POST \
  -d "text=Test&banner=standard" \
  http://localhost:8080/ascii-art
```

**Expected result**

* Status `200`.
* Standard banner output is returned.

---

## TC-021 — Shadow banner accepted

Use:

```text
banner=shadow
```

**Expected result**

* Status `200`.
* Shadow banner output is returned.

---

## TC-022 — Thinkertoy banner accepted

Use:

```text
banner=thinkertoy
```

**Expected result**

* Status `200`.
* Thinkertoy banner output is returned.

---

## TC-023 — Unknown banner rejected

**Request**

```bash
curl -i \
  -X POST \
  -d "text=Hello&banner=unknown" \
  http://localhost:8080/ascii-art
```

**Expected result**

Recommended:

```text
400 Bad Request
```

The request must not cause the server to try to open:

```text
banners/unknown.txt
```

---

## TC-024 — Filesystem-path banner rejected

**Request**

```bash
curl -i \
  -X POST \
  --data-urlencode "text=Hello" \
  --data-urlencode "banner=../../etc/passwd" \
  http://localhost:8080/ascii-art
```

**Expected result**

```text
400 Bad Request
```

The server must not use the submitted value as a filesystem path.

---

## TC-025 — Absolute path banner rejected

**Request**

```bash
curl -i \
  -X POST \
  --data-urlencode "text=Hello" \
  --data-urlencode "banner=/etc/passwd" \
  http://localhost:8080/ascii-art
```

**Expected result**

```text
400 Bad Request
```

---

# 11. Input Validation Tests

## TC-026 — Missing text field

**Request**

```bash
curl -i \
  -X POST \
  -d "banner=standard" \
  http://localhost:8080/ascii-art
```

**Expected result**

```text
400 Bad Request
```

---

## TC-027 — Missing banner field

**Request**

```bash
curl -i \
  -X POST \
  -d "text=Hello" \
  http://localhost:8080/ascii-art
```

**Expected result**

```text
400 Bad Request
```

---

## TC-028 — Empty text

**Request**

```bash
curl -i \
  -X POST \
  -d "text=&banner=standard" \
  http://localhost:8080/ascii-art
```

**Expected result**

Use the behavior required by the project specification.

Recommended:

```text
400 Bad Request
```

The server must not crash.

---

## TC-029 — Whitespace-only text

**Input**

```text
   
```

**Expected result**

Confirm the intended project behavior.

If the existing renderer accepts spaces, the server should return `200` and render the blank glyph spacing consistently.

---

## TC-030 — Unsupported character

Use a character outside the supported banner range, such as:

```text
Hello é
```

**Expected result**

Recommended:

```text
400 Bad Request
```

An understandable error should be returned.

The server must remain running.

---

## TC-031 — Very long text

Submit a long printable ASCII string.

**Expected result**

* The request completes.
* The server does not panic.
* The output area handles overflow.
* The page remains usable.
* No uncontrolled memory growth is observed.

---

## TC-032 — Multiline input

Submit:

```text
Hello
World
```

or the exact newline format supported by the application.

**Expected result**

* Both lines are rendered.
* Their separation matches existing renderer behavior.
* No text is lost.

---

# 12. HTML Escaping and Injection Tests

## TC-033 — HTML tags are escaped

**Input**

```html
<script>alert("xss")</script>
```

**Expected result**

* The text is displayed as text or rejected as unsupported input.
* No JavaScript executes.
* No alert appears.
* The DOM does not gain a new `<script>` element from user input.

---

## TC-034 — Image error-handler injection

**Input**

```html
<img src=x onerror=alert(1)>
```

**Expected result**

* No image is created from user input.
* No JavaScript executes.
* User-controlled text is escaped.

---

## TC-035 — Template expression input

**Input**

```text
{{.ASCIIArt}}
```

**Expected result**

* It is treated as ordinary input text.
* It is not evaluated as a second Go template.
* The server does not expose template data.

---

## TC-036 — HTML-looking valid banner input

**Input**

```text
<Hello>
```

**Expected result**

* Angle brackets appear in the rendered ASCII result.
* They are not interpreted as HTML tags.

---

# 13. HTTP 404 Tests

## TC-037 — Unknown route

**Request**

```bash
curl -i http://localhost:8080/not-found
```

**Expected result**

```text
404 Not Found
```

The home page must not be returned for every unknown path.

---

## TC-038 — Nested unknown route

**Request**

```bash
curl -i http://localhost:8080/ascii-art/unknown
```

**Expected result**

```text
404 Not Found
```

---

## TC-039 — Missing banner file

Temporarily rename one bundled banner during a controlled test:

```bash
mv banners/standard.txt banners/standard.txt.bak
```

Submit a request using `standard`.

**Expected result**

```text
404 Not Found
```

Restore the file afterward:

```bash
mv banners/standard.txt.bak banners/standard.txt
```

The server must not crash.

---

# 14. HTTP 400 Tests

## TC-040 — Malformed form body

**Request**

```bash
curl -i \
  -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data "%" \
  http://localhost:8080/ascii-art
```

**Expected result**

```text
400 Bad Request
```

---

## TC-041 — Invalid banner value

Use:

```text
banner=STANDARD
```

if banner names are case-sensitive.

**Expected result**

```text
400 Bad Request
```

---

## TC-042 — Invalid width value

Applicable if the browser sends terminal width.

**Request**

```bash
curl -i \
  -X POST \
  -d "text=Hello&banner=standard&columns=abc" \
  http://localhost:8080/ascii-art
```

**Expected result**

```text
400 Bad Request
```

---

## TC-043 — Zero or negative width

**Inputs**

```text
columns=0
columns=-10
```

**Expected result**

```text
400 Bad Request
```

or a documented safe fallback.

---

## TC-044 — Excessively large width

**Input**

```text
columns=999999999
```

**Expected result**

* The value is rejected or safely limited.
* The server does not allocate excessive memory.
* The server does not hang.

---

# 15. HTTP 500 Tests

## TC-045 — Template execution failure

Create a controlled test using an injected or intentionally invalid template.

**Expected result**

```text
500 Internal Server Error
```

The server must:

* log the internal error;
* return a safe client response;
* avoid exposing stack traces or filesystem details;
* continue serving later requests.

---

## TC-046 — Unexpected internal renderer failure

Use a controlled test double or helper that returns an unexpected error.

**Expected result**

```text
500 Internal Server Error
```

This test should distinguish internal errors from invalid user input.

---

## TC-047 — Server survives after internal error

**Steps**

1. Trigger a controlled `500`.
2. Send a normal `GET /`.
3. Submit a valid ASCII-art request.

**Expected result**

* The later requests succeed.
* The process remains alive.

---

# 16. HTTP Method Tests

## TC-048 — Correct GET method

`GET /` must return the main page.

---

## TC-049 — Correct POST method

`POST /ascii-art` must generate ASCII art.

---

## TC-050 — PUT rejected

**Request**

```bash
curl -i -X PUT http://localhost:8080/ascii-art
```

**Expected result**

```text
405 Method Not Allowed
```

---

## TC-051 — DELETE rejected

**Request**

```bash
curl -i -X DELETE http://localhost:8080/ascii-art
```

**Expected result**

```text
405 Method Not Allowed
```

---

## TC-052 — OPTIONS handled safely

**Request**

```bash
curl -i -X OPTIONS http://localhost:8080/ascii-art
```

**Expected result**

* The server returns a controlled response.
* The server does not crash.

A `405` response is acceptable unless OPTIONS support is intentionally implemented.

---

# 17. Handler and Architecture Tests

## TC-053 — Web handler does not build glyphs

**Code review**

Inspect handlers.

**Expected result**

The web handler:

* reads request values;
* validates them;
* maps banner names;
* calls existing renderer functions;
* executes templates.

It must not contain character-to-glyph assembly loops.

---

## TC-054 — Web handler does not parse colors

**Code review**

**Expected result**

No color parsing implementation exists inside the HTTP handler.

Existing renderer or color packages remain responsible for color behavior.

---

## TC-055 — Web handler does not search substrings

**Code review**

**Expected result**

The HTTP handler does not implement substring matching or selected-character tracking.

---

## TC-056 — Web handler does not implement alignment

**Code review**

**Expected result**

The handler passes alignment and width values to the appropriate output or renderer package.

It does not calculate left, center, right, or justify spacing itself.

---

## TC-057 — Only bundled banner mapping is used

**Expected implementation pattern**

```go
var bannerPaths = map[string]string{
	"standard":   "banners/standard.txt",
	"shadow":     "banners/shadow.txt",
	"thinkertoy": "banners/thinkertoy.txt",
}
```

**Expected result**

The request value is used only as a lookup key.

It is never joined directly into a file path.

---

# 18. Browser-Aware Terminal Tests

Applicable if the terminal-like output area is resizable.

## TC-058 — Terminal resize works on frontend

**Steps**

1. Generate ASCII art.
2. Resize the terminal-like output area.
3. Observe the displayed output.

**Expected result**

* Font size or scale updates.
* The result remains readable.
* The page does not reload unnecessarily.

---

## TC-059 — Resize does not flood the server

**Steps**

1. Open browser developer tools.
2. Select the Network tab.
3. Resize the terminal repeatedly.

**Expected result**

For frontend-only scaling:

* no HTTP requests are sent during resizing.

For debounced server alignment:

* requests are delayed;
* only a small number are sent after resizing stops;
* no request is sent for every resize event.

---

## TC-060 — Output remains inside terminal

**Expected result**

* Content either scales or scrolls.
* It does not break the entire page layout.
* Horizontal overflow is handled deliberately.

---

## TC-061 — Text remains selectable

**Expected result**

* ASCII text can still be selected and copied.
* Resizing does not convert it to an image.

---

# 19. User Interface Tests

## TC-062 — Output is understandable

**Expected result**

* ASCII result is displayed inside a `<pre>` or equivalent.
* Spaces and line breaks are preserved.
* Font is monospace.
* Contrast is readable.
* The terminal-like container clearly separates output from controls.

---

## TC-063 — Banner selection is understandable

**Expected result**

The user can clearly distinguish:

* Standard
* Shadow
* Thinkertoy

---

## TC-064 — Form preserves useful values

After generating output:

**Expected result**

Recommended behavior:

* entered text remains visible;
* selected banner remains selected;
* result appears without confusing navigation.

---

## TC-065 — Errors are understandable

Submit an invalid request through the interface.

**Expected result**

The page provides a readable error message rather than:

* a blank page;
* raw Go error text;
* a stack trace;
* an unexplained status page.

---

# 20. Navigation Tests

## TC-066 — All available links work

**Steps**

Click every link or navigation control.

**Expected result**

* Each intended page opens.
* No broken links exist.
* No valid navigation returns `404`.

---

## TC-067 — Browser back and refresh behavior

**Steps**

1. Generate ASCII art.
2. Refresh the page.
3. Use browser back and forward controls.

**Expected result**

* The application remains stable.
* No crash occurs.
* Repeated POST behavior is understandable.

---

# 21. Stability Tests

## TC-068 — Repeated valid requests

Send multiple valid requests:

```bash
for i in $(seq 1 100); do
  curl -s \
    -X POST \
    -d "text=Hello&banner=standard" \
    http://localhost:8080/ascii-art > /dev/null
done
```

**Expected result**

* All requests complete.
* Server remains responsive.
* No crashes occur.

---

## TC-069 — Mixed valid and invalid requests

Alternate:

* valid banners;
* invalid banners;
* missing fields;
* unknown routes.

**Expected result**

* Each request receives the correct status.
* One bad request does not affect later requests.

---

## TC-070 — Concurrent requests

Example:

```bash
seq 1 50 | xargs -n1 -P10 -I{} \
  curl -s \
  -X POST \
  -d "text=Hello{}&banner=standard" \
  http://localhost:8080/ascii-art > /dev/null
```

**Expected result**

* The server handles concurrent requests.
* Results do not leak between users.
* No data races or crashes occur.

Run additionally:

```bash
go test -race ./...
```

---

## TC-071 — Resource-leak review

**Expected result**

* Request bodies are managed by `net/http`.
* Opened files are closed.
* No goroutine is started per request without a termination path.
* No unbounded global result history is stored.

---

# 22. Automated HTTP Test Cases

Use `net/http/httptest`.

Recommended test table:

```go
func TestRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		contentType string
		wantStatus int
	}{
		{
			name:       "home page",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid standard render",
			method:     http.MethodPost,
			path:       "/ascii-art",
			body:       "text=Hello&banner=standard",
			contentType: "application/x-www-form-urlencoded",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			path:       "/missing",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid method",
			method:     http.MethodGet,
			path:       "/ascii-art",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "missing text",
			method:     http.MethodPost,
			path:       "/ascii-art",
			body:       "banner=standard",
			contentType: "application/x-www-form-urlencoded",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown banner",
			method:     http.MethodPost,
			path:       "/ascii-art",
			body:       "text=Hello&banner=unknown",
			contentType: "application/x-www-form-urlencoded",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "path traversal banner",
			method:     http.MethodPost,
			path:       "/ascii-art",
			body:       "text=Hello&banner=..%2F..%2Fetc%2Fpasswd",
			contentType: "application/x-www-form-urlencoded",
			wantStatus: http.StatusBadRequest,
		},
	}
}
```

---

# 23. Golden Output Tests

Golden tests should compare exact renderer output.

Recommended cases:

```text
standard:
{123}
<Hello> (World)!

standard:
123??

shadow:
$% "=

thinkertoy:
123 T/fs#R
```

Store expected output either:

* directly in test strings; or
* in `testdata/*.golden` files.

Comparison should be exact:

```go
if got != want {
	t.Fatalf("unexpected output\nwant:\n%s\ngot:\n%s", want, got)
}
```

Do not trim spaces from ASCII output before comparison. Leading and trailing spaces are part of the result.

Normalize line endings only when necessary:

```go
strings.ReplaceAll(value, "\r\n", "\n")
```

---

# 24. Security Acceptance Tests

The project passes the security review only when:

* HTML user input is escaped.
* ASCII output is inserted safely.
* JavaScript uses `textContent`, not `innerHTML`, for generated text.
* Browser banner values are mapped to fixed server paths.
* No request field becomes a file path.
* Unknown banners are rejected.
* Template names are not selected from request input.
* Error messages do not expose internal filesystem paths.
* The server does not execute user input.
* The server does not create arbitrary files from browser input.

---

# 25. Final Audit Checklist

## Functional

* [ ] Only standard Go packages are used.
* [ ] HTML files exist.
* [ ] Standard banner golden tests pass.
* [ ] Shadow banner golden test passes.
* [ ] Thinkertoy banner golden test passes.
* [ ] ASCII output is visually understandable.
* [ ] All available pages work.
* [ ] Unknown routes return `404`.
* [ ] Invalid requests return `400`.
* [ ] Controlled internal failures return `500`.
* [ ] Browser-server communication works.
* [ ] Correct HTTP methods are used.
* [ ] Server does not crash.
* [ ] Server is written in Go.

## General

* [ ] Required handlers and patterns exist.
* [ ] Server starts quickly.
* [ ] No unnecessary requests are made.
* [ ] Code follows Go best practices.
* [ ] Automated tests exist.
* [ ] Tests cover successful and failure cases.
* [ ] Website instructions are clear.
* [ ] The web interface uses HTTP endpoints as its API.

## Architecture

* [ ] Handler does not build glyphs.
* [ ] Handler does not parse colors.
* [ ] Handler does not search substrings.
* [ ] Handler does not calculate justify spacing.
* [ ] Only bundled banners are selectable.
* [ ] Renderer tests still pass.
* [ ] Output tests still pass.
* [ ] HTML output escapes user content.
* [ ] Browser input cannot create filesystem paths.

## Failure Conditions

The project should fail the audit for:

* empty or effectively empty work;
* incomplete required functionality;
* invalid compilation;
* prohibited third-party packages;
* copied or cheating implementation;
* crashes during normal use;
* uncontrolled resource leaks;
* missing required endpoints;
* absent HTML interface;
* unsafe filesystem path handling;
* unescaped user-controlled HTML;
* replacement or duplication of the existing renderer without justification.
