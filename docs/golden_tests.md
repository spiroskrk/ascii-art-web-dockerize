# Test Design and Golden Contracts

This document is the bridge between the product requirements and executable
tests. It records both exact golden-output contracts and the wider test coverage
needed by the application.

A **golden test** compares complete output byte for byte with a known-good
result. Not every test should be golden. Validation, HTTP status, startup, and
browser-interaction behavior are usually clearer as focused assertions.

## Sources of Expected Behavior

Use the documents in this order:

1. `docs/prd.md` defines the agreed behavior of the new web application.
2. Existing package tests and the legacy cases below protect functionality
   inherited from the previous application.
3. `docs/audit_test.md` is a coverage checklist that helps us find missing
   scenarios.

When these sources disagree, do not write a test that guesses the answer.
Compare the behavior with the PRD, record whether the old behavior is being
preserved or intentionally changed, and then write one unambiguous expectation.

For example, the previous normalizer splits literal `\n` but not real newline
characters. The web PRD intentionally extends that contract so LF, CRLF, CR,
and literal `\n` all act as logical separators. Existing literal-`\n` behavior
remains a regression case, while new tests describe the expanded behavior.

## Test Layers

Choose the smallest layer that can prove the requirement:

| Layer | What it proves | Typical tool or location |
| --- | --- | --- |
| Package unit test | One package's rules and edge cases | `internal/*/*_test.go` |
| Golden-output test | Exact ASCII rows, spaces, newlines, or control sequences | Inline expected strings or `testdata/*.golden` |
| HTTP handler test | Routes, methods, forms, status codes, headers, and rendered state | `net/http/httptest` |
| Startup or lifecycle test | Dependency loading, configuration, and controlled failures | Focused startup helpers without a fixed network port |
| Manual browser check | Native validation, color picker, scrolling, and download prompts | Browser checklist |
| Tool verification | Compilation, formatting, static checks, and data races | `go test`, `go test -race`, `go vet`, and `gofmt` |

One requirement can need checks at more than one layer, but each test should
have one primary reason to fail. For example, the empty-text requirement leads
to separate checks that:

- HTML marks the textarea as required;
- server validation rejects an empty value;
- the HTTP response is `400 Bad Request`, preserves safe form state, and has no
  generated result;
- a manual browser check confirms native validation prevents an ordinary empty
  submission.

The HTTP test should not retest the renderer's glyph construction, and a
renderer test should not know about HTTP status codes.

## Test-Case Design Recipe

Before implementing a case, write down:

1. **Requirement** - the single promise being tested.
2. **Owner** - the package or browser boundary responsible for it.
3. **Given** - the starting state and inputs.
4. **When** - the one action being performed.
5. **Then** - the observable result, including exact status or error when
   applicable.
6. **Test layer** - unit, golden, HTTP, startup, manual, or tool verification.

Prefer deterministic inputs, explicit widths, and exact expectations. Do not
trim or normalize generated ASCII output inside a comparison unless that
normalization is itself part of the stated contract.

## Web Contract Test Inventory

The web cases are added feature by feature. They describe the behavior we have
agreed to build; they do not imply that the implementation already exists.

### Text and Newline Contract

#### Equivalence Classes

Equivalence partitioning groups inputs that should receive the same kind of
result. We test representatives from each group instead of trying every
possible string.

| Class | Representative | Expected behavior |
| --- | --- | --- |
| Valid ordinary text | `Hello` | Accepted and rendered without changing the text |
| Valid whitespace text | Three spaces | Accepted because spaces are renderable and must not be trimmed |
| Valid multiline text | LF, CRLF, CR, or literal `\n` between `Hello` and `World` | Each representation produces the same two logical lines |
| Valid empty logical line | Two consecutive separators | The empty line between surrounding lines is preserved |
| Invalid absent text | No `text` form field | Rejected as a bad request |
| Invalid empty text | Present `text` field with an empty value | Rejected as a bad request |
| Invalid unsupported text | `é` or a tab character | Rejected as unsupported banner input |
| Invalid excessive text | 4,097 printable ASCII characters | Rejected because the limit is 4,096 characters |

#### Renderer Normalization Cases

These are package unit tests for `renderer.NormalizeInput`. The notation in the
input column uses Go string literals so a real LF (`"\n"`) can be distinguished
from typed backslash-plus-`n` (`"\\n"`).

| ID | Given input | Expected logical lines | Reason |
| --- | --- | --- | --- |
| TXT-N01 | `""` | `[]string{}` | Preserve the normalizer's empty-input contract even though the web validator rejects empty text |
| TXT-N02 | `"Hello"` | `[]string{"Hello"}` | Ordinary text remains unchanged |
| TXT-N03 | `"Hello\\nWorld"` | `[]string{"Hello", "World"}` | Preserve literal-`\n` compatibility |
| TXT-N04 | `"Hello\nWorld"` | `[]string{"Hello", "World"}` | Accept the LF produced by a browser textarea |
| TXT-N05 | `"Hello\r\nWorld"` | `[]string{"Hello", "World"}` | Canonicalize a Windows CRLF as one separator |
| TXT-N06 | `"Hello\rWorld"` | `[]string{"Hello", "World"}` | Accept a remaining CR as a separator |
| TXT-N07 | `"A\r\nB\rC\nD\\nE"` | `[]string{"A", "B", "C", "D", "E"}` | Prove mixed separator forms are normalized in the correct order |
| TXT-N08 | `"Hello\n\nWorld"` | `[]string{"Hello", "", "World"}` | Preserve an empty logical line |
| TXT-N09 | `"\nHello"` | `[]string{"", "Hello"}` | Preserve a leading empty logical line |
| TXT-N10 | `"Hello\n"` | `[]string{"Hello"}` | Preserve the existing single trailing-separator behavior |
| TXT-N11 | `"Hello\n\n"` | `[]string{"Hello", ""}` | Remove only the final split element and preserve the preceding empty line |
| TXT-N12 | `"  Hello  "` | `[]string{"  Hello  "}` | Do not trim or collapse spaces |

TXT-N03 through TXT-N06 are an equivalence set: different input
representations must produce the same logical lines. TXT-N07 is useful because
it can catch an incorrect implementation that processes `\r` before `\r\n` and
accidentally turns one CRLF into two separators.

#### Text Validation Boundaries

Boundary-value analysis checks values directly around a limit. For a range of
1 through 4,096 characters, the important sizes are 0, 1, 4,096, and 4,097.

| ID | Given `text` value | Expected result | Primary owner and layer |
| --- | --- | --- | --- |
| TXT-V01 | Field is missing | Validation error | Web form-validation unit test |
| TXT-V02 | `""` (0 characters) | Validation error | Web form-validation unit test |
| TXT-V03 | `"A"` (1 character) | Accepted unchanged | Web form-validation unit test |
| TXT-V04 | `"   "` | Accepted unchanged | Web form-validation unit test |
| TXT-V05 | 4,096 `A` characters | Accepted unchanged | Web form-validation unit test |
| TXT-V06 | 4,097 `A` characters | Validation error | Web form-validation unit test |
| TXT-V07 | `"café"` | Unsupported-character error | Renderer or shared-generation test |
| TXT-V08 | `"Hello\tWorld"` | Unsupported-character error | Renderer or shared-generation test |

The length limit and the request-body limit prove different things. TXT-V06
tests a validly encoded form field that violates the product's character limit.
A body larger than 64 KiB belongs to the request-parsing tests and must return
`413 Content Too Large`.

#### Integration and Browser Representatives

Unit tests cover the complete normalization and boundary tables. Higher layers
need only representative cases that prove the pieces are connected correctly.

| ID | Layer | Scenario | Expected result |
| --- | --- | --- | --- |
| TXT-G01 | Shared generation | Generate the four separator representations at the same banner, alignment, and width | All four produce equivalent generated output |
| TXT-H01 | HTTP | Submit a valid real-LF multiline form | `200 OK`, preserved form state, and a generated result |
| TXT-H02 | HTTP | Submit an empty `text` value | `400 Bad Request`, preserved safe form state, and no result |
| TXT-H03 | HTTP | Submit whitespace-only text | `200 OK`; the server must not trim it into an empty value |
| TXT-H04 | HTTP | Submit unsupported input | `400 Bad Request` with a safe error and no result |
| TXT-H05 | HTTP | Submit 4,097 printable characters in a body below 64 KiB | `400 Bad Request`, proving the field limit independently of the body limit |
| TXT-B01 | Template test | Render the initial form | Textarea has `required` and `maxlength="4096"` |
| TXT-B02 | Manual browser | Press Generate with an empty textarea | Native browser validation prevents submission |
| TXT-B03 | Manual browser | Press Enter between two words and generate | Both logical lines appear in the result |

We intentionally do not compare the entire HTML page byte for byte. Handler
tests assert the status, form state, result state, and relevant escaped content;
existing renderer golden tests remain responsible for exact glyph rows and
spaces.

### Request Parsing and Transport Contract

Request tests follow the order in which untrusted data crosses the HTTP
boundary:

```text
known route and method
    -> supported media type
    -> bounded request body
    -> URL-form decoding
    -> allowed keys and single-value cardinality
    -> field-value validation
    -> generation
```

This order separates transport failures from semantic validation:

| Failure stage | Example | Expected status |
| --- | --- | --- |
| Media type | JSON sent to a form endpoint | `415 Unsupported Media Type` |
| Raw body size | Actual body exceeds 64 KiB | `413 Content Too Large` |
| Form decoding | Invalid percent encoding | `400 Bad Request` |
| Form structure | Duplicate or unexpected fields | `400 Bad Request` |
| Field meaning | Unknown banner or invalid width | `400 Bad Request` |
| Unexpected application failure | Template or presenter failure | `500 Internal Server Error` |

Each test should otherwise use a valid request. A test that combines an
unsupported media type, oversized body, duplicate text, and invalid banner
cannot tell us which check failed or which status should take precedence.

#### Media-Type Cases

Generation and download accept the
`application/x-www-form-urlencoded` media type. A syntactically valid parameter
such as `charset=UTF-8` does not change the base media type and is accepted
after parsing the header; it does not cause a different body parser to be used.

| ID | `Content-Type` | Expected result |
| --- | --- | --- |
| REQ-M01 | `application/x-www-form-urlencoded` | Parsing continues |
| REQ-M02 | `application/x-www-form-urlencoded; charset=UTF-8` | Parsing continues as URL-encoded form data |
| REQ-M03 | Header is missing | `415 Unsupported Media Type` |
| REQ-M04 | `text/plain` | `415 Unsupported Media Type` |
| REQ-M05 | `application/json` | `415 Unsupported Media Type` |
| REQ-M06 | `multipart/form-data; boundary=example` | `415 Unsupported Media Type` |
| REQ-M07 | Malformed media-type syntax | `415 Unsupported Media Type` |

These tests check the parsed media type, not a fragile case-sensitive comparison
of the complete header string.

#### Body-Limit and Decoding Cases

The 64 KiB limit applies to the encoded request body before form parsing:

```text
64 KiB = 65,536 bytes
```

| ID | Given request body | Expected result |
| --- | --- | --- |
| REQ-B01 | Small, valid URL-encoded form | Parsing continues |
| REQ-B02 | Exactly 65,536 bytes, using an overlong but syntactically valid `text` field | `400 Bad Request` from field validation, specifically not `413` |
| REQ-B03 | 65,537 bytes | `413 Content Too Large` |
| REQ-B04 | More than 65,536 bytes with no trustworthy `Content-Length` | `413 Content Too Large` based on bytes actually read |
| REQ-B05 | `text=%` inside an otherwise valid form | `400 Bad Request` |
| REQ-B06 | `text=%GG` inside an otherwise valid form | `400 Bad Request` |
| REQ-B07 | URL-encoded `Hello+World` | Decodes to `Hello World` |
| REQ-B08 | URL-encoded `Hello%0AWorld` | Decodes to a real LF for renderer normalization |

REQ-B02 distinguishes an inclusive body limit from the 4,096-character field
limit. The body may fit under the transport limit and still fail a later rule.
REQ-B04 prevents an implementation from trusting only the client-controlled
`Content-Length` header.

#### Field-Shape Cases

The accepted POST-form keys are exactly:

```text
text
banner
use_color
color
substring
align
width
```

Every accepted key is single-valued. Required and conditional field meanings
are covered in their feature sections; these cases first establish an
unambiguous form shape.

| ID | Given form structure | Expected result |
| --- | --- | --- |
| REQ-F01 | Each supplied key occurs once and no unknown key exists | Structural parsing succeeds |
| REQ-F02 | Any accepted key occurs twice | `400 Bad Request`; test all seven keys with a table |
| REQ-F03 | Unknown key such as `admin=true` | `400 Bad Request` |
| REQ-F04 | Named submit field such as `action=generate` | `400 Bad Request`; the HTML buttons therefore remain unnamed |
| REQ-F05 | Fields appear in a different order | Same parsed form state |
| REQ-F06 | Valid body plus conflicting query values | Body values are used; query values neither override nor create duplicates |
| REQ-F07 | Required value exists only in the query string | `400 Bad Request`; query values do not satisfy POST-form requirements |
| REQ-F08 | Unknown value exists only in the query string | It is not treated as a POST form field or passed into generation |

REQ-F06 through REQ-F08 prove that generation reads `PostForm` rather than the
merged `Form`. They should assert the resulting validated values, not merely
the response status, because a `200` response alone would not reveal that a
query value silently replaced a body value.

#### Endpoint-Parity Cases

The detailed parser and validator tables belong in focused
`internal/web/form_test.go` tests. A smaller `httptest` table runs representative
boundary failures against both:

```text
POST /ascii-art
POST /ascii-art/download
```

| ID | Representative failure | Expected result on both endpoints |
| --- | --- | --- |
| REQ-H01 | Unsupported media type | `415 Unsupported Media Type` |
| REQ-H02 | Oversized body | `413 Content Too Large` |
| REQ-H03 | Malformed URL encoding | `400 Bad Request` |
| REQ-H04 | Duplicate field | `400 Bad Request` |
| REQ-H05 | Unexpected field | `400 Bad Request` |

This parity table catches a handler that accidentally bypasses the shared
request workflow. Every failed download response must additionally omit the
`Content-Disposition: attachment` header.

### Banner Selection and Filesystem Contract

Banner testing distinguishes three values that must never be confused:

```text
public identifier -> trusted startup path -> preloaded glyph map
standard          -> banners/standard.txt -> map[rune][]string
shadow            -> banners/shadow.txt   -> map[rune][]string
thinkertoy         -> banners/thinkertoy.txt -> map[rune][]string
```

Only the public identifier crosses the HTTP boundary. Trusted paths exist in
application startup configuration, and handlers receive the completed registry
of glyph maps. A request value is never concatenated with `banners/`, passed to
`filepath.Join`, or sent to `banner.LoadBanner`.

#### Existing Banner-Loader Boundary

`internal/banner` accepts a trusted path and owns file-format validation. These
tests are independent from HTTP:

| ID | Loader scenario | Expected result | Coverage |
| --- | --- | --- | --- |
| BAN-L01 | Load each official banner | Each map contains 95 printable runes and 8 rows for `' '` and `'~'` | Existing |
| BAN-L02 | File does not exist | Error and nil banner map | Existing |
| BAN-L03 | Banner has too few character blocks | Invalid-format error and nil map | Existing |
| BAN-L04 | Banner has too many character blocks | Invalid-format error and nil map | Planned |
| BAN-L05 | Banner uses CRLF line endings | Successful load with no hidden `\r` in glyph rows | Existing |
| BAN-L06 | File content has an invalid overall shape | Invalid-format error and nil map | Existing |

Exact ASCII golden outputs provide regression protection for the official
banner contents. We do not need to hard-code file hashes unless the product
later requires byte-identical banner files; valid intentional banner updates
would otherwise fail a checksum test even when their format is correct.

#### Startup Registry Cases

Startup owns trusted disk access. It must finish loading every required banner
before the server begins accepting requests.

| ID | Startup condition | Expected result |
| --- | --- | --- |
| BAN-S01 | All three fixed files are valid | Registry contains exactly `standard`, `shadow`, and `thinkertoy` |
| BAN-S02 | `standard` is missing | Initialization fails; no usable partial registry is returned |
| BAN-S03 | `shadow` is missing | Initialization fails; no usable partial registry is returned |
| BAN-S04 | `thinkertoy` is missing | Initialization fails; no usable partial registry is returned |
| BAN-S05 | Any required banner is malformed | Initialization fails before listening |
| BAN-S06 | Files change after successful initialization | Existing registry remains unchanged until application restart |
| BAN-S07 | Concurrent requests read the registry | Requests do not mutate maps or leak state; race-enabled tests remain clean |

Failure cases should use `t.TempDir`, temporary fixture copies, or an injected
loader in a focused initialization helper. Tests must not rename or modify the
repository's real banner files. This keeps tests isolated and allows parallel
or interrupted test runs without leaving the project broken.

BAN-S06 can preload a temporary banner, alter or remove only that temporary
source file, and then prove generation still uses the in-memory map. It should
not depend on sleep calls or filesystem modification timestamps.

#### Public Identifier Validation

Banner names are exact, case-sensitive identifiers. They are not filenames and
are not normalized by trimming.

| ID | Submitted `banner` | Expected result |
| --- | --- | --- |
| BAN-V01 | `standard` | Accepted and selects the standard preloaded map |
| BAN-V02 | `shadow` | Accepted and selects the shadow preloaded map |
| BAN-V03 | `thinkertoy` | Accepted and selects the thinkertoy preloaded map |
| BAN-V04 | Field is missing | `400 Bad Request` |
| BAN-V05 | Empty value | `400 Bad Request` |
| BAN-V06 | `STANDARD` | `400 Bad Request` |
| BAN-V07 | ` standard` or `standard ` | `400 Bad Request`; do not trim into a valid identifier |
| BAN-V08 | `standard.txt` | `400 Bad Request` |
| BAN-V09 | `banners/standard.txt` | `400 Bad Request`, even though that trusted file exists |
| BAN-V10 | `../../etc/passwd` | `400 Bad Request` |
| BAN-V11 | `/etc/passwd` | `400 Bad Request` |
| BAN-V12 | `..\..\windows\system.ini` | `400 Bad Request` |
| BAN-V13 | Percent-encoded traversal that decodes to a path | `400 Bad Request` after normal URL-form decoding |

Testing an invalid value that happens to name a real banner file, such as
`banners/standard.txt`, is stronger than testing only a nonexistent path. It
proves that rejection comes from exact registry lookup rather than from a
failed attempt to open the submitted path.

#### Template, HTTP, and Exposure Cases

| ID | Layer | Scenario | Expected result |
| --- | --- | --- | --- |
| BAN-T01 | Template | Render the initial form | Exactly three `banner` radio values exist and `standard` is checked |
| BAN-T02 | Template | Render a successful request for each banner | Submitted valid banner remains selected |
| BAN-G01 | Shared generation | Generate the same small input with each registry entry | Each selection produces the corresponding banner output |
| BAN-H01 | HTTP generation | Submit each valid identifier | `200 OK` with a generated result |
| BAN-H02 | HTTP generation | Submit each invalid identifier class | `400 Bad Request`, safe preserved state, and no result |
| BAN-H03 | HTTP download | Submit an invalid or traversal-like identifier | `400 Bad Request` and no attachment header |
| BAN-H04 | Public route | `GET /banners/standard.txt` | `404 Not Found`; banner source files are not public assets |
| BAN-H05 | Public route | `GET /templates/index.html` | `404 Not Found`; templates are not public assets |

The audit case that temporarily renames a live banner and expects an HTTP `404`
does not match the agreed architecture. Missing or malformed bundled banners
are startup failures. Once the application is serving, every valid public
identifier already has a loaded map, so request-time banner-file `404` errors
do not exist.

### Color and Substring Contract

Color crosses several boundaries, so one large end-to-end color test would be
difficult to diagnose. The responsibilities are:

```text
raw web fields
    -> web-form rules: enabled state and public #RRGGBB syntax
    -> renderer-owned color conversion and substring selection
    -> ANSI-colored structured rendering
    -> ANSI-aware alignment
    -> safe presenter data
    -> template-created CSS spans
```

The thin HTTP handler orchestrates these steps. It does not parse colors, find
substring matches, interpret arbitrary CSS, or build HTML strings. A focused
validator enforces the web-only syntax, while renderer-owned color logic remains
the semantic source of truth for conversion and rendering.

#### Color Enablement and Web Validation

An unchecked checkbox is absent from an HTML form. A checked checkbox submits
the browser value `on`.

| ID | `use_color` | Submitted `color` | Expected validated behavior |
| --- | --- | --- | --- |
| COL-V01 | Absent | Missing or `#ff0000` | Color disabled; picker value is ignored for rendering |
| COL-V02 | Absent | Invalid or malicious string | Color disabled; value cannot create CSS and does not fail generation |
| COL-V03 | `on` | `#000000` | Enabled with RGB `(0, 0, 0)` |
| COL-V04 | `on` | `#ffffff` | Enabled with RGB `(255, 255, 255)` |
| COL-V05 | `on` | `#00fF88` | Enabled with RGB `(0, 255, 136)`; hex digits are case-insensitive |
| COL-V06 | `on` | Missing or empty | `400 Bad Request` |
| COL-V07 | `on` | `#f00`, `#ff000`, or `#ff00000` | `400 Bad Request` |
| COL-V08 | `on` | `#gg0000` or another non-hex value | `400 Bad Request` |
| COL-V09 | `on` | `red`, `rgb,255,0,0`, or `hsl,0,100%,50%` | `400 Bad Request`; these remain internal legacy formats, not public web formats |
| COL-V10 | `on` | CSS-like input such as `red; background:url(x)` | `400 Bad Request` |
| COL-V11 | `on` | HTML-like input such as `\"><script>alert(1)</script>` | `400 Bad Request` and safely escaped form state |
| COL-V12 | Any other checkbox value | Valid-looking color | `400 Bad Request`; only absent or `on` matches the form contract |

The validator must produce bounded numeric RGB components rather than passing a
raw browser string to the template. Tests assert the resulting integers, not
only that the input matched a regular expression.

The initial template uses `#ff0000` as the picker value but leaves color
disabled. Successful and invalid submissions preserve safe user-visible form
state without turning an invalid raw value into a style declaration.

#### Substring Selection and Limits

Substring matching remains renderer behavior. Spaces are significant, matching
is case-sensitive, and matches never cross normalized logical lines.

| ID | Input text | Substring | Expected selected characters |
| --- | --- | --- | --- |
| SUB-R01 | `hello` | Empty | All characters |
| SUB-R02 | `hello` | `ell` | Exactly the first `ell` |
| SUB-R03 | `a king kitten have kit` | `kit` | Every occurrence |
| SUB-R04 | `banana` | `an` | Both overlapping matches, covering positions 1 through 4 |
| SUB-R05 | `Kit kit` | `kit` | Lowercase occurrence only |
| SUB-R06 | `hello` | `xyz` | No selected characters and no error |
| SUB-R07 | `a  b` | Two spaces | The exact two-space match |
| SUB-R08 | Logical lines `hello` and `hello` | `lohe` | No cross-line match |
| SUB-R09 | `RGB()` | `B` | Only the `B` |

Field validation adds these boundaries:

| ID | Submitted `substring` | Expected result |
| --- | --- | --- |
| SUB-V01 | Field missing | Treated as empty |
| SUB-V02 | Empty value | Accepted |
| SUB-V03 | Spaces only | Accepted unchanged; do not trim |
| SUB-V04 | 4,096 characters | Accepted unchanged |
| SUB-V05 | 4,097 characters | `400 Bad Request` |

The substring length limit applies even when color is disabled because it is a
request-size contract. Rendering behavior is different: when color is
disabled, every valid substring value is ignored and output must be
byte-for-byte equivalent to ordinary uncolored generation.

#### Renderer and Shared-Generation Color Cases

Existing renderer tests continue to prove named, hex, RGB, and HSL parsing for
legacy compatibility. New shared-generation cases prove the web connection:

| ID | Scenario | Expected result |
| --- | --- | --- |
| COL-G01 | Color disabled with empty substring | Uncolored generated output |
| COL-G02 | Color disabled with matching substring | Same uncolored output as COL-G01 |
| COL-G03 | Color enabled with empty substring | Every rendered character block carries foreground state |
| COL-G04 | Color enabled with one matching substring | Only matching glyph blocks carry foreground state |
| COL-G05 | Color enabled with repeated or overlapping matches | Every selected glyph block carries foreground state |
| COL-G06 | Color enabled with missing substring | Successful uncolored output |
| COL-G07 | Colored multiline input | Selection is recalculated independently for each logical line |
| COL-G08 | Colored and uncolored generation at the same width and alignment | Visible rows and spacing are identical after ANSI is removed |

These tests compare complete ANSI output where exact control-sequence placement
is the contract. They do not search loosely for a color code while ignoring
incorrect reset placement.

#### ANSI Presenter Cases

The presenter accepts only foreground and reset sequences generated by the
renderer:

```text
ESC[30m through ESC[37m
ESC[38;5;208m
ESC[38;2;R;G;Bm
ESC[0m
```

| ID | Presenter input | Expected result |
| --- | --- | --- |
| PRE-C01 | Plain text only | One or more uncolored runs with identical visible text |
| PRE-C02 | Each named foreground form followed by reset | Colored run followed by uncolored state |
| PRE-C03 | Orange 256-color form followed by reset | Correct colored-state boundaries |
| PRE-C04 | Truecolor form with components 0 and 255 | Correct colored-state boundaries |
| PRE-C05 | Adjacent fragments with the same state | May merge without changing text |
| PRE-C06 | Spaces and empty logical rows | Preserved exactly |
| PRE-C07 | Incomplete escape sequence | Internal presentation error |
| PRE-C08 | Unsupported ANSI command | Internal presentation error; never interpreted as HTML or terminal behavior |
| PRE-C09 | Truecolor component outside 0 through 255 | Internal presentation error |
| PRE-C10 | Foreground state without required reset at end | Internal presentation error |

For every successful presenter case:

- styled-run text contains no ANSI bytes;
- plain download text contains no ANSI bytes;
- concatenating all run text reproduces the visible aligned text exactly;
- row count, spaces, blank rows, and final newline convention are preserved.

Malformed renderer output is an internal failure and maps to a safe `500`; it
is not treated as invalid browser color input.

#### Color Presentation and Security Cases

| ID | Layer | Scenario | Expected result |
| --- | --- | --- | --- |
| COL-T01 | Template | Render a colored run with validated RGB components | Template creates fixed `color: rgb(R, G, B)` syntax |
| COL-T02 | Template | Render uncolored runs | No unnecessary user-controlled style syntax |
| COL-T03 | HTTP | Generate colored output | `200 OK`, styled spans present, and no ANSI bytes in HTML |
| COL-T04 | HTTP | Generate with color disabled | No colored output spans caused by picker or substring values |
| COL-T05 | Security | Submit CSS, quotes, tags, and template syntax in color-related fields | Rejected or escaped; no new element, attribute, or CSS declaration is created |
| COL-T06 | Security | Preserve raw text such as `<script>alert("x")</script>` in the form | It appears only as escaped textarea text and never executes |

Tests should inspect parsed or focused HTML properties rather than accepting a
raw `<script>` substring anywhere in the response: escaped textarea content can
legitimately contain those characters as text. No user value is converted to
`template.HTML` or arbitrary `template.CSS`.

### Alignment and Width Contract

Alignment has two boundaries: the web layer validates the requested mode and
explicit width, while `internal/output` owns all spacing algorithms. The HTTP
handler passes validated values to output code and never calculates padding.

#### Alignment Validation

| ID | Submitted `align` | Expected result |
| --- | --- | --- |
| ALN-V01 | `left` | Accepted |
| ALN-V02 | `center` | Accepted |
| ALN-V03 | `right` | Accepted |
| ALN-V04 | `justify` | Accepted |
| ALN-V05 | Field missing or empty | `400 Bad Request` |
| ALN-V06 | `LEFT` or another case variant | `400 Bad Request` |
| ALN-V07 | Leading or trailing spaces | `400 Bad Request`; do not trim |
| ALN-V08 | Unsupported value such as `middle` | `400 Bad Request` |

The template contains exactly these four radio values, uses one shared name,
and selects `left` by default.

#### Width Validation Boundaries

The browser submits a measured width in text columns. Explicit widths must be
integers from 20 through 300 inclusive.

| ID | Submitted `width` | Expected validated width |
| --- | --- | --- |
| WID-V01 | Field missing | `80` fallback |
| WID-V02 | `20` | `20` |
| WID-V03 | `80` | `80` |
| WID-V04 | `300` | `300` |
| WID-V05 | Empty value | `400 Bad Request` |
| WID-V06 | `19` | `400 Bad Request` |
| WID-V07 | `301` | `400 Bad Request` |
| WID-V08 | `0` or a negative value | `400 Bad Request` |
| WID-V09 | `abc` or `20.5` | `400 Bad Request` |
| WID-V10 | Extremely large decimal value | `400 Bad Request` without overflow or large allocation |

The missing-field fallback is intentionally different from an explicitly empty
or malformed field. It supports no-JavaScript submission without allowing a
client to submit ambiguous width data.

#### Output Alignment Cases

Existing output tests already cover most spacing behavior. The completed
contract is:

| ID | Output scenario | Expected result |
| --- | --- | --- |
| ALN-O01 | Left alignment | Byte-for-byte flattened rows; width adds no padding |
| ALN-O02 | Center with row narrower than width | `(width - visible row width) / 2` spaces added on the left |
| ALN-O03 | Right with row narrower than width | `width - visible row width` spaces added on the left |
| ALN-O04 | Any row equal to or wider than width | Row remains unchanged |
| ALN-O05 | ANSI-colored rows | Escape bytes do not count toward visible width |
| ALN-O06 | Justify with multiple word gaps | Extra columns distributed between word segments |
| ALN-O07 | Justify with an uneven remainder | Earlier gaps receive remainder columns from left to right |
| ALN-O08 | Justify with one word | Rows remain left-aligned |
| ALN-O09 | Justify with leading or trailing spaces | Outer spaces are preserved and are not expandable gaps |
| ALN-O10 | Spaces inside ASCII glyph rows | Never treated as word gaps |
| ALN-O11 | Multiple source spaces between words | Original space segment is preserved before extra spacing |
| ALN-O12 | Empty logical line | Remains one blank output row |
| ALN-O13 | Multiline input | Every logical line is aligned independently |
| ALN-O14 | Invalid alignment passed directly to output package | Explicit error; no partial output |
| ALN-O15 | Underlying writer fails | Error is returned instead of being swallowed |

Deterministic unit tests use explicit widths rather than live terminal
detection. Existing CLI terminal-width tests remain legacy regression coverage,
but web generation does not consult terminal streams or `COLUMNS`.

#### Browser Width and Snapshot Cases

| ID | Layer | Scenario | Expected result |
| --- | --- | --- | --- |
| WID-T01 | Template | Render initial page | Hidden `width` field exists with fallback value `80` |
| WID-J01 | JavaScript/manual | Submit Generate with a measurable terminal | Width is measured once in monospace columns immediately before submission |
| WID-J02 | JavaScript/manual | Measurement is unavailable | Existing hidden value `80` is submitted |
| WID-J03 | Browser/manual | Resize after generation | No request is sent and existing output spacing does not change |
| WID-J04 | Browser/manual | Container becomes narrower than output | Horizontal scrolling appears; text does not wrap |
| WID-J05 | Browser/manual | Generate again after resizing | New request uses the newly measured width |
| WID-H01 | HTTP | Successful generation | Response records the validated width used for output |
| WID-H02 | Template | Successful generation | Download snapshot contains the recorded generated width |

JavaScript measures only layout. It does not render glyphs, align rows, resize
font to hide overflow, or send requests from a resize event.

### Shared Generation Contract

The shared generator is tested without HTTP or templates. It consumes only a
validated `GenerationInput` and immutable preloaded banners:

```text
validated values
    -> banner registry lookup
    -> newline normalization
    -> structured rendering and substring color
    -> explicit-width alignment
    -> GeneratedASCII
```

| ID | Scenario | Expected result |
| --- | --- | --- |
| GEN-01 | Each official banner at left alignment | Corresponding exact renderer output |
| GEN-02 | LF, CRLF, CR, and literal `\n` variants | Equivalent output |
| GEN-03 | Each alignment at widths 20, 80, and 300 | Deterministic aligned output |
| GEN-04 | Color disabled and enabled | Same visible characters and spacing |
| GEN-05 | Empty, matching, repeated, overlapping, and missing substrings | Existing renderer selection semantics |
| GEN-06 | Unsupported input rune | Typed user-input failure that the handler maps to `400` |
| GEN-07 | Unexpected output-writer or generation dependency failure | Internal failure that maps to `500` |
| GEN-08 | Successful visually blank whitespace input | Successful result remains distinguishable from no result |
| GEN-09 | Same validated input used for browser and download | Both presentations derive from the same aligned generated text |

Generation tests may use small deterministic banner maps when testing
orchestration, while a smaller set with official banners protects real
integration. They should not load a banner from a browser-derived path or
reimplement glyph, substring, or alignment algorithms in expected-value helper
code.

### Routes, Page State, and Error Contract

Route tests use `net/http/httptest` against the constructed handler. They do not
bind `:8080` or start the production server.

#### Route and Method Matrix

| ID | Method and path | Expected result |
| --- | --- | --- |
| RTE-01 | `GET /` | `200 OK` HTML containing the default page |
| RTE-02 | `POST /` | `405 Method Not Allowed` with `Allow: GET` |
| RTE-03 | `POST /ascii-art` with valid form | `200 OK` HTML containing generated state |
| RTE-04 | `GET /ascii-art` | `405 Method Not Allowed` with `Allow: POST` |
| RTE-05 | `PUT`, `DELETE`, or `OPTIONS /ascii-art` | `405 Method Not Allowed` with `Allow: POST` |
| RTE-06 | `POST /ascii-art/download` with valid form | `200 OK` plain-text attachment |
| RTE-07 | `GET /ascii-art/download` | `405 Method Not Allowed` with `Allow: POST` |
| RTE-08 | `PUT`, `DELETE`, or `OPTIONS /ascii-art/download` | `405 Method Not Allowed` with `Allow: POST` |
| RTE-09 | `GET /not-found` | `404 Not Found` |
| RTE-10 | `GET /ascii-art/unknown` | `404 Not Found` |
| RTE-11 | `GET /ascii-art/` | `404 Not Found`; do not silently treat nested paths as the exact route |

The wrong-method cases use otherwise ordinary requests. Media-type validation
does not replace method validation, and unknown paths must not fall through to
the home page.

#### Initial Page and Form Contract

Focused template assertions verify:

| ID | Initial-page requirement | Expected result |
| --- | --- | --- |
| PAGE-01 | Response media type | HTML with UTF-8 character encoding |
| PAGE-02 | Text control | Multiline textarea with `name="text"`, `required`, and `maxlength="4096"` |
| PAGE-03 | Banner controls | Three labeled radios; `standard` selected |
| PAGE-04 | Color controls | Unchecked `use_color` checkbox and color input default `#ff0000` |
| PAGE-05 | Substring control | Empty single-line input with `maxlength="4096"` |
| PAGE-06 | Alignment controls | Four labeled radios; `left` selected |
| PAGE-07 | Width control | Hidden value `80` |
| PAGE-08 | Generate action | Submit control targets `POST /ascii-art` |
| PAGE-09 | Clear action | Non-nested GET action targets `/` |
| PAGE-10 | Result state | Measurable terminal container exists, but no generated result or download snapshot is present |
| PAGE-11 | Error state | No error message |
| PAGE-12 | Native validation | Form does not use `novalidate` |
| PAGE-13 | Public assets | Page references only `/static/style.css` and `/static/app.js` |

Tests should inspect relevant attributes and values rather than comparing the
complete template byte for byte. Whitespace, accessibility improvements, and
unrelated styling can then change without rewriting every handler test.

#### Successful, Invalid, and Clear State

| ID | Scenario | Expected page state |
| --- | --- | --- |
| STATE-01 | Valid generation | `200`, preserved submitted controls, `HasResult=true`, output present, and download snapshot present |
| STATE-02 | Valid whitespace-only generation | `200` and `HasResult=true` even if output is visually blank |
| STATE-03 | Invalid form values | Documented `400`, safe submitted state preserved, useful error, no result, and no download snapshot |
| STATE-04 | Unexpected generation or presentation failure | Safe `500`; no partial generated page and no internal details |
| STATE-05 | `GET /` after any prior response | Fresh defaults, no result, and no error |
| STATE-06 | Browser Clear control | Requests `GET /` and clears both editable input and displayed output |

Clear is not implemented only as `type="reset"`, because reset would restore
server-rendered values and leave server-rendered output in place.

#### Error Mapping and Recovery

| ID | Error class | Expected response |
| --- | --- | --- |
| ERR-01 | Invalid field meaning | `400 Bad Request` HTML with safe feedback |
| ERR-02 | Malformed URL form | `400 Bad Request` |
| ERR-03 | Body over 64 KiB | `413 Content Too Large` |
| ERR-04 | Unsupported media type | `415 Unsupported Media Type` |
| ERR-05 | Unknown route | `404 Not Found` |
| ERR-06 | Wrong method | `405 Method Not Allowed` and correct `Allow` |
| ERR-07 | Controlled template execution failure | Safe `500 Internal Server Error` |
| ERR-08 | Controlled presenter failure | Safe `500 Internal Server Error` |
| ERR-09 | Controlled unexpected generator failure | Safe `500 Internal Server Error` |
| ERR-10 | Normal request after any controlled error | Succeeds; the application remains usable |

Template execution is tested through a deliberately failing injected template
or writer, not by corrupting the production template. The response is first
rendered into a temporary buffer so execution failure cannot leave a partial
`200 OK` body.

For all client-facing errors, responses omit stack traces, repository paths,
template source, raw wrapped errors, and generated output. Internal diagnostic
detail belongs in logs.

### HTML and Browser Presentation Security

Security assertions test observable output and trust boundaries rather than
only searching source code for forbidden function names.

| ID | Submitted or generated content | Expected result |
| --- | --- | --- |
| SEC-H01 | `<script>alert("x")</script>` in text | Preserved form text is escaped; no executable script element is created |
| SEC-H02 | `<img src=x onerror=alert(1)>` in text | No image or event handler is created |
| SEC-H03 | `{{.Output}}` in text | Treated as ordinary data and never evaluated as a second template |
| SEC-H04 | `& < > " '` in form state | Correct HTML escaping without data loss |
| SEC-H05 | ASCII-art rows containing `<`, `>`, or `&` glyph text | Safely escaped while spacing remains intact |
| SEC-H06 | ANSI-colored renderer output | Browser HTML contains styled runs but no raw escape bytes |
| SEC-H07 | User-supplied CSS-like color string | Rejected when enabled and unable to influence styles when disabled |
| SEC-H08 | Internal error containing a filesystem path | Path appears only in captured logs, not the response |
| SEC-H09 | Traversal-like banner, template, or static path | Cannot influence a filesystem operation |

Additional presentation assertions verify:

- output uses a whitespace-preserving monospace container;
- template indentation adds no characters inside generated output;
- result text remains selectable and copyable;
- overflow scrolls horizontally rather than wrapping;
- JavaScript does not insert generated output with `innerHTML`;
- request values are never marked as trusted HTML or arbitrary CSS.

Browser DOM behavior is confirmed manually where Go template tests cannot prove
execution behavior. Server-side tests still inspect escaped output so manual
testing is not the only security check.

### Download Contract

Download is a second presentation of the same successful generation, not a
server-side file-writing feature.

#### Successful Attachment Cases

| ID | Download scenario | Expected result |
| --- | --- | --- |
| DLD-01 | Valid snapshot | `200 OK` |
| DLD-02 | Response media type | `text/plain; charset=utf-8` |
| DLD-03 | Response disposition | `attachment; filename="ascii-art.txt"` |
| DLD-04 | Each official banner | Corresponding plain ASCII output |
| DLD-05 | Multiline text | All logical lines and blank rows preserved |
| DLD-06 | Each alignment | Same aligned visible rows as browser generation |
| DLD-07 | Color enabled | No ANSI bytes; ASCII characters and spacing preserved |
| DLD-08 | Whitespace-only successful result | Download remains available and preserves generated blank spacing |
| DLD-09 | Final newline | Same documented plain-output newline convention as shared generation |

The body is compared exactly for a small representative set. It must not be
trimmed before comparison because leading spaces, trailing spaces, blank rows,
and the final newline are output data.

#### Snapshot Semantics

After successful generation, the page contains a separate non-nested download
form with an escaped snapshot of:

```text
text
banner
use_color, when enabled
color
substring
align
rendered width
```

| ID | Snapshot scenario | Expected result |
| --- | --- | --- |
| DLD-S01 | Successful generation | Download form appears with the values that produced the displayed result |
| DLD-S02 | Initial page or failed generation | Download form is absent |
| DLD-S03 | Edit main text after successful generation without regenerating | Download still uses snapshot text |
| DLD-S04 | Edit banner, color, substring, or alignment without regenerating | Download still uses snapshot choices |
| DLD-S05 | Resize browser after successful generation | Download uses recorded generated width, not a new measurement |
| DLD-S06 | Generate successfully again | Snapshot is replaced by the newly successful values |
| DLD-S07 | Multiline snapshot text | Real line endings survive the hidden textarea or equivalent escaped control |
| DLD-S08 | Complete generated ASCII submitted as a hidden field | Rejected as an unexpected field; output is regenerated server-side |

The handler does not trust hidden controls. Download parses, validates, and
regenerates the snapshot through the same workflow as Generate. No server
session or browser-supplied generated-output blob becomes a source of truth.

#### Download Failure and Filesystem Safety

| ID | Failure scenario | Expected result |
| --- | --- | --- |
| DLD-F01 | Invalid snapshot field | Documented `400` HTML error and no attachment header |
| DLD-F02 | Oversized body | `413` and no attachment header |
| DLD-F03 | Unsupported media type | `415` and no attachment header |
| DLD-F04 | Unexpected generator or presenter failure | Safe `500` and no attachment header |
| DLD-F05 | Submitted `filename`, `path`, or `output` field | `400` as an unexpected field |
| DLD-F06 | Traversal-like value in any identifier field | Rejected without opening or creating that path |
| DLD-F07 | Successful download | Response body is streamed to the client; no server-side output file is created |

Tests do not grant the application a temporary output path because the web
download design has no file-writing dependency. The constant client-facing
filename exists only in `Content-Disposition`.

### Static Assets and Client Script Contract

Only two fixed public asset paths are exposed:

```text
/static/style.css
/static/app.js
```

| ID | Asset request | Expected result |
| --- | --- | --- |
| AST-01 | `GET /static/style.css` | `200 OK` with a CSS media type |
| AST-02 | `GET /static/app.js` | `200 OK` with a JavaScript media type |
| AST-03 | `GET /static/` | `404 Not Found`; no directory listing |
| AST-04 | `GET /static/unknown` | `404 Not Found` |
| AST-05 | `POST` on a known asset | `405 Method Not Allowed` with `Allow: GET` |
| AST-06 | Traversal-like static URL | Never returns template, banner, repository, or arbitrary filesystem content |
| AST-07 | Direct template or banner URL | `404 Not Found` |

Focused script review and manual checks verify that `app.js`:

- measures terminal content width only immediately before Generate;
- updates only the hidden width field;
- leaves the fallback value `80` when measurement fails;
- does not construct ASCII glyphs or implement alignment;
- does not send a request from every resize event;
- does not inject generated output with `innerHTML`;
- does not invent a second client-side validation system.

CSS and browser checks verify monospace output, preserved whitespace, readable
contrast, responsive controls, and deliberate horizontal overflow.

### Startup, Template, and Server Lifecycle Contract

Process-level wiring is tested through focused helpers. Tests avoid starting the
production address or waiting for operating-system signals when the underlying
behavior can be invoked directly.

#### Dependency Initialization

| ID | Initialization scenario | Expected result |
| --- | --- | --- |
| STA-01 | Valid template and all banners | Complete dependencies are constructed before route registration |
| STA-02 | Template file missing | Initialization error and no listener |
| STA-03 | Template syntax invalid | Initialization error and no listener |
| STA-04 | Any banner missing or malformed | Initialization error and no listener |
| STA-05 | Valid template | Parsed once and reused concurrently |
| STA-06 | Construct app with missing required dependency | Constructor error rather than later panic |
| STA-07 | Run from unsupported working directory | Clear startup failure consistent with documented repository-root requirement |

Fixtures use temporary directories or injected readers. They do not modify
production templates or banners.

#### Server Configuration and Shutdown

| ID | Lifecycle property | Expected result |
| --- | --- | --- |
| SRV-00 | Application entry point | Starts the web lifecycle without entering a CLI prompt |
| SRV-01 | Listen address | `:8080` |
| SRV-02 | `ReadHeaderTimeout` | 5 seconds |
| SRV-03 | `ReadTimeout` | 15 seconds |
| SRV-04 | `WriteTimeout` | 30 seconds |
| SRV-05 | `IdleTimeout` | 60 seconds |
| SRV-06 | Interrupt or termination | Graceful shutdown begins |
| SRV-07 | Shutdown deadline | At most 5 seconds for active requests |
| SRV-08 | Intentional `http.ErrServerClosed` | Normal completion, not a reported failure |
| SRV-09 | Other listen error | Logged failure and non-zero process result |
| SRV-10 | Shutdown error | Logged and returned as a failure |

Lifecycle tests inject a controllable server or call the serve/shutdown helper
with cancellable contexts. They avoid fixed ports, long sleeps, and timing
assertions narrower than the documented deadline.

### Logging Contract

Tests inject an in-memory log destination. Production constructs a
standard-library structured logger writing to standard error and never creates
`log.txt`.

| ID | Logging scenario | Expected result |
| --- | --- | --- |
| LOG-01 | Server startup | Safe address and lifecycle event are recorded |
| LOG-02 | Completed request | Method, safe route path, status, and duration are recorded |
| LOG-03 | Safe banner identifier | May be recorded |
| LOG-04 | Validation error | Status is recorded without form body |
| LOG-05 | Internal failure | Diagnostic error is recorded while client receives a safe response |
| LOG-06 | Graceful shutdown | Start and completion or failure are recorded |

A request containing a unique secret marker in each of these locations is used
to prove the marker does not appear in logs:

```text
text
substring
complete form body
generated ASCII
cookie
authorization header
arbitrary header
query string
```

Middleware logs `URL.Path` or another safe route identity rather than the raw
request URI, and it does not read or copy the request body. Duration is checked
for presence and a valid non-negative form, not an exact timing value.

### Concurrency, Stability, and Tool Verification

#### Concurrent Request Cases

| ID | Scenario | Expected result |
| --- | --- | --- |
| CON-01 | Many independent valid requests | Every response matches its own submitted values |
| CON-02 | Mixed valid and invalid requests | Each receives its own documented status |
| CON-03 | Concurrent requests using different banners, colors, and widths | No output or form state leaks between requests |
| CON-04 | Request following a controlled `500` | Succeeds normally |
| CON-05 | Race-enabled suite | No mutation race in templates, banner maps, logger, or request state |
| CON-06 | Resource-lifecycle review | Bodies and files close, goroutines terminate, and no unbounded result history is retained |

Tests use bounded goroutine counts and `httptest`, not shell loops against a
fixed production port. Independent cases may use `t.Parallel()` only when they
do not mutate environment variables, working directories, package globals, or
shared fixtures.

#### Verification Commands

Run in this order so basic failures are reported before slower checks:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./...
rg --files -g '*.go' | xargs gofmt -l
go list -m all
```

Expected results:

- every command exits successfully;
- the `gofmt -l` command prints no Go source files;
- `go list -m all` lists no third-party module dependency;
- existing banner, renderer, and output regression tests still pass;
- HTTP tests use the standard library, especially `httptest`;
- no test changes production expectations merely to make a failure disappear.

#### Manual Browser Checklist

Manual verification covers behavior the Go test process does not execute:

- [ ] The server starts at `http://localhost:8080/`.
- [ ] Labels, radio groups, and controls are keyboard-accessible.
- [ ] Page instructions explain text entry, banner selection, generation, output, and download.
- [ ] Native required-field feedback blocks empty text.
- [ ] Whitespace-only input can still be generated.
- [ ] Enter and typed `\n` produce equivalent logical lines.
- [ ] All three banner selections visibly differ.
- [ ] The color picker enables correct whole-input and substring color.
- [ ] Disabled color ignores picker and substring values.
- [ ] All four alignments render at the measured terminal width.
- [ ] Width is measured once on Generate.
- [ ] Resizing sends no request and existing output does not realign.
- [ ] Narrow output scrolls horizontally without wrapping.
- [ ] Result text remains selectable and copyable.
- [ ] Clear removes both form state and output.
- [ ] Download appears only after success and prompts for `ascii-art.txt`.
- [ ] Editing the main form does not alter the existing download snapshot.
- [ ] Downloaded text preserves spacing and contains no ANSI codes.
- [ ] Invalid submissions show useful errors without raw Go details.
- [ ] Browser back, forward, refresh, and repeated use do not destabilize the app.
- [ ] Developer tools show no unexpected console errors or resize request flood.

### Audit Reconciliation

The audit is fully represented, with vague or outdated expectations resolved by
the PRD:

| Audit area | Covered here | Resolution |
| --- | --- | --- |
| Build, tests, standard library, formatting, vet | Verification commands and architecture review | Automated, with `gofmt -l` used as a non-mutating check |
| Required templates, banners, and README | Startup, banner, static, and documentation sections | Required internal files are startup dependencies, not public URLs |
| Home page and form controls | Route and initial-page sections | Includes color, substring, alignment, Clear, Download, limits, and defaults omitted by early audit cases |
| Generation routes and methods | Route, request, and shared-generation sections | Statuses and `Allow` headers are exact rather than “recommended” |
| Standard, shadow, and thinkertoy outputs | Exact golden fixture set | Compared at renderer/generation boundaries instead of as brittle full HTML |
| Banner traversal | Banner and filesystem section | Exact registry lookup; browser values never become paths |
| Text and multiline validation | Text/newline and request sections | Empty is `400`, whitespace is valid, unsupported text is `400`, and limits are explicit |
| HTML injection | Color and presentation-security sections | Escaped form state, safe runs, no raw ANSI, HTML, or arbitrary CSS trust |
| `404`, `400`, `413`, `415`, and `500` | Request and error sections | Missing banners are corrected to startup failures rather than request-time `404` |
| Width and resize behavior | Alignment and browser-width sections | Fixed generated snapshot plus horizontal scrolling; no dynamic scaling or resize requests |
| Architecture review | Documentation and architecture section | Thin handlers reuse banner, renderer, and output packages |
| UI, navigation, and browser stability | Initial-page and manual-browser sections | Clear, refresh, links, accessibility, copyability, and instructions included |
| Repetition, concurrency, races, and leaks | Concurrency and verification sections | Bounded `httptest` concurrency replaces shell load loops for automation |
| Download | Download section | Adds snapshot validation, fixed client filename, ANSI-free content, and no server file |

### Exact Golden Fixture Set

Golden fixtures are reviewed expected results, not output generated during the
test by the same implementation being tested. Otherwise both `got` and `want`
would contain the same bug.

The initial exact fixture set is:

| ID | Banner and input | Contract protected |
| --- | --- | --- |
| GLD-01 | Standard: `{123}` followed by `<Hello> (World)!` on the next logical line | Multiline standard output, punctuation, angle brackets, rows, and spaces |
| GLD-02 | Standard: `123??` | Repeated punctuation and numeric glyph composition |
| GLD-03 | Shadow: `$% "=` | Shadow punctuation and quote handling |
| GLD-04 | Thinkertoy: `123 T/fs#R` | Thinkertoy mixed letters, numbers, spaces, and symbols |
| GLD-05 | Small structured `A B C` fixture at explicit widths | Left, center, right, justify, and remainder distribution |
| GLD-06 | Small whole-input truecolor fixture | Exact ANSI foreground and reset placement |
| GLD-07 | Small repeated and overlapping substring fixture | Exact colored versus uncolored glyph segments |
| GLD-08 | One aligned colored fixture passed through presenter | Visible text equals ANSI-free plain output exactly |
| GLD-09 | One generated snapshot downloaded as text | Exact browser-visible spacing, rows, and final newline without ANSI |

Large ASCII expectations may live under package-local `testdata/*.golden`
files. Small structural expectations stay inline when that makes the behavior
easier to read.

Golden comparison rules:

- compare exact bytes;
- preserve leading, internal, and trailing spaces;
- preserve empty rows and the documented final newline;
- store repository fixtures with LF endings;
- normalize CRLF only in a test specifically exercising cross-platform fixture
  loading, not in every output assertion;
- produce a useful line or quoted-byte diff on failure;
- never call `strings.TrimSpace` on generated output;
- update a fixture only after reviewing the intentional behavioral change.

### Documentation and Architecture Review

Some acceptance requirements are clearer as focused review items than runtime
assertions:

- [ ] `README.md` explains the web application, authors, repository-root startup,
      usage, and implementation structure.
- [ ] `main.go` contains lifecycle wiring rather than form, rendering, or
      alignment algorithms.
- [ ] HTTP handlers parse, validate, orchestrate, and respond without building
      glyphs, parsing semantic colors, finding substrings, or justifying rows.
- [ ] `internal/banner`, `internal/renderer`, and `internal/output` retain their
      documented responsibilities.
- [ ] Templates contain presentation structure but no application algorithms.
- [ ] Browser values cannot choose banner, template, static-asset, log, or
      output filesystem paths.
- [ ] Parsed templates, banner maps, and the logger are injected dependencies,
      not mutable package globals.
- [ ] Request-specific values remain local to each request.
- [ ] Only Go standard-library and local project packages are imported.
- [ ] Public behavior is tested through package APIs and `httptest` rather than
      by depending on unexported implementation layout without reason.

### Test Implementation Order

Implement from narrowest responsibility to widest integration:

1. Extend renderer newline unit tests.
2. Add missing banner-loader and output regression cases.
3. Add web form parsing and semantic validation tests.
4. Add shared generation and presenter tests.
5. Add template and route tests with `httptest`.
6. Add download, static-asset, logging, and startup tests.
7. Run race, vet, build, formatting, and dependency verification.
8. Complete the manual browser checklist.

This order gives failures a useful location. Test completeness is judged by
public behavior, branch and boundary coverage, security controls, and preserved
domain contracts—not by an arbitrary coverage percentage.

## Legacy Regression Inventory

These edge cases describe behavior inherited from the previous application at
the existing package boundaries. Keep the tests small and focused so each
package proves only its own responsibility.

Transition rule:

- banner, renderer, and output regression tests remain part of the web project;
- CLI tests remain temporarily while the HTTP implementation is being built;
- after the complete web suite passes and `main.go` becomes the server entry
  point, obsolete root CLI tests are replaced and the unused CLI package is
  removed;
- the legacy CLI command cases below then become historical documentation, not
  executable web acceptance tests.

| Area | Edge case | Expected behavior | Test file |
| --- | --- | --- | --- |
| Banner loader | All official banners load successfully | `standard`, `shadow`, and `thinkertoy` each produce 95 mapped runes, and boundary runes `' '` and `'~'` each have 8 lines | `internal/banner/loader_test.go` |
| Banner loader | Windows CRLF banner content | Loaded banner lines must not contain hidden `\r` characters | `internal/banner/loader_test.go` |
| Banner loader | Missing character block | A banner with fewer than 95 blocks is invalid | `internal/banner/loader_test.go` |
| CLI argument mode | Whitespace-only text argument | Whitespace text is valid because spaces are renderable characters | `internal/cli/cli_test.go` |
| CLI argument mode | Escaped newline text argument | Text like `Hello\nThere` is preserved for the renderer to normalize | `internal/cli/cli_test.go` |
| CLI color mode | Whole-string color form | `--color=red "hello"` returns text `hello`, color `red`, empty substring, and `banners/standard.txt` | `internal/cli/cli_test.go` |
| CLI color mode | Whole-string hex color form | `--color=#ff0000 "hello"` returns text `hello`, color `#ff0000`, empty substring, and `banners/standard.txt` | `internal/cli/cli_test.go` |
| CLI color mode | Whole-string RGB color form | `--color=RGB,255,0,0 "hello"` returns text `hello`, color `RGB,255,0,0`, empty substring, and `banners/standard.txt` | `internal/cli/cli_test.go` |
| CLI color mode | Whole-string HSL color form | `--color=hsl,0,100%,50% "hello"` returns text `hello`, color `hsl,0,100%,50%`, empty substring, and `banners/standard.txt` | `internal/cli/cli_test.go` |
| CLI color mode | Substring color form | `--color=orange GuYs "HeY GuYs"` returns text `HeY GuYs`, color `orange`, substring `GuYs`, and `banners/standard.txt` | `internal/cli/cli_test.go` |
| CLI color mode | One-character substring | `--color=blue B "RGB()"` returns substring `B` and does not treat `RGB()` as a banner | `internal/cli/cli_test.go` |
| CLI color mode | Banner-compatible color form | `"hello world" shadow --color=cyan world` returns banner `banners/shadow.txt`, color `cyan`, substring `world` | `internal/cli/cli_test.go` |
| CLI color mode | Invalid separated color flag | `--color red "banana"` returns the exact color usage message | `internal/cli/cli_test.go` |
| CLI color mode | Empty color value | `--color= "banana"` returns the exact color usage message | `internal/cli/cli_test.go` |
| CLI color mode | Misspelled color flag | `-color=red` and `--colour=red` return the exact color usage message | `internal/cli/cli_test.go` |
| CLI color mode | Missing string after color flag | `--color=red` returns the exact color usage message | `internal/cli/cli_test.go` |
| CLI color mode | Multiple color flags | A command containing two `--color=<color>` flags is invalid | `internal/cli/cli_test.go` |
| CLI color mode | Order-flexible whole-string color | `"hello" --color=red` returns text `hello`, color `red`, and an empty substring | `internal/cli/cli_test.go` |
| CLI output mode | Order-flexible output flag | `"hello" --output=out.txt` returns text `hello` and output file `out.txt` | `internal/cli/cli_test.go` |
| CLI alignment mode | Valid alignment values | `--align=left`, `--align=center`, `--align=right`, and `--align=justify` return the selected alignment | `internal/cli/cli_test.go` |
| CLI alignment mode | Invalid alignment formats | Separated, empty, misspelled, uppercase, unsupported, and duplicate alignment flags return the exact alignment usage message | `internal/cli/cli_test.go` |
| CLI option mode | Unknown options | Unknown `--...` flags return the supported-options usage message | `internal/cli/cli_test.go` |
| CLI banner mode | Invalid banner value | Invalid banner names return `Invalid banner: <value>` followed by the general usage message | `internal/cli/cli_test.go` |
| CLI color mode | Shell-split RGB/HSL | Split color values such as `--color=rgb,255, 0, 0` are invalid argument formats | `internal/cli/cli_test.go` |
| CLI interactive mode | Windows-style user input | `\r\n` input is trimmed correctly at prompts | `internal/cli/interactive_test.go` |
| Renderer normalization | Leading escaped newline | `\nHello` becomes `["", "Hello"]` | `internal/renderer/normalize_test.go` |
| Renderer normalization | Trailing escaped newline | `Hello\n` becomes `["Hello"]` | `internal/renderer/normalize_test.go` |
| Renderer normalization (superseded expectation) | Real newline character | The current pre-web test preserves real newlines; TXT-N04 intentionally replaces this expectation so browser LF becomes a logical separator | `internal/renderer/normalize_test.go` |
| Renderer color helpers | Supported preset colors | `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, and `orange` return non-empty ANSI codes | `internal/renderer/color_test.go` |
| Renderer color helpers | Supported hex notation | `#ff0000`, `#FF0000`, and other full `#RRGGBB` values return ANSI truecolor escape codes | `internal/renderer/color_test.go` |
| Renderer color helpers | Supported RGB notation | `rgb,255,0,0` and case variants such as `RGB,255,0,0` return ANSI truecolor escape codes | `internal/renderer/color_test.go` |
| Renderer color helpers | Supported HSL notation | `hsl,0,100%,50%` and case variants such as `HSL,0,100%,50%` return ANSI truecolor escape codes | `internal/renderer/color_test.go` |
| Renderer color helpers | Invalid color name | Unknown color names return `Invalid color name` | `internal/renderer/color_test.go` |
| Renderer color helpers | Invalid hex color | Short, long, or non-hex `#...` values return `Invalid hexadecimal color code` | `internal/renderer/color_test.go` |
| Renderer color helpers | Invalid RGB color | Out-of-range, non-integer, or wrong-shape `rgb...` values return `Invalid RGB color code` | `internal/renderer/color_test.go` |
| Renderer color helpers | Invalid HSL color | Out-of-range, missing percent, non-integer, or wrong-shape `hsl...` values return `Invalid HSL color code` | `internal/renderer/color_test.go` |
| Renderer color helpers | ANSI reset wrapping | `colorize("abc", "\033[31m")` returns color code + `abc` + `\033[0m` | `internal/renderer/color_test.go` |
| Renderer substring helpers | Empty substring | Empty substring marks every character position in the line | `internal/renderer/substring_test.go` |
| Renderer substring helpers | Missing substring | Missing substring returns an empty position map | `internal/renderer/substring_test.go` |
| Renderer substring helpers | Repeated matches | Multiple occurrences of the same substring all mark their character positions | `internal/renderer/substring_test.go` |
| Renderer substring helpers | Case-sensitive matching | `GuYs` matches `GuYs` but not `guys`, `GUYS`, or `Guys` | `internal/renderer/substring_test.go` |
| Renderer substring helpers | Two-letter substring | A two-character substring marks exactly those two positions for each match | `internal/renderer/substring_test.go` |
| Renderer build | Horizontal composition | Adjacent characters are joined row-by-row, not printed as separate blocks | `internal/renderer/build_test.go` |
| Renderer build | Empty input line between rendered lines | Empty normalized lines produce one blank output line and do not break following lines | `internal/renderer/build_test.go` |
| Renderer build color | Whole-string color | `BuildASCII(..., "", "#FF0000")` wraps every rendered character block row with truecolor red and reset codes | `internal/renderer/build_test.go` |
| Renderer build color | Substring color | `BuildASCII(..., "GuYs", "orange")` wraps only the `GuYs` character blocks | `internal/renderer/build_test.go` |
| Renderer build color | One-letter color | `BuildASCII(..., "B", "blue")` colors only the `B` in `RGB()` | `internal/renderer/build_test.go` |
| Renderer build color | Missing substring | If a provided substring is not found, output is valid uncolored ASCII art | `internal/renderer/build_test.go` |
| Renderer build color | Newline boundaries | A substring never matches across normalized input lines | `internal/renderer/build_test.go` |
| Renderer build color | Invalid color | An invalid name, hex code, RGB value, or HSL value returns an error before producing output | `internal/renderer/build_test.go` |
| Renderer build color | Spaces preserved | Coloring part of a string must not remove or trim spaces from rendered output | `internal/renderer/build_test.go` |
| Output printer | Exact stdout format | Each provided output row is printed with one newline, and an empty slice prints nothing | `internal/output/print_test.go` |
| Output alignment | ANSI-aware visible width | ANSI escape sequences do not count as visible width when padding aligned output | `internal/output/alignment_test.go` |
| Output alignment | Justify uses structured segments | Justify expands only spaces between rendered word segments, not spaces inside ASCII-art letters | `internal/output/alignment_test.go` |
| Output width | Linux non-terminal file | A regular file returns width `0` and detection status `false` | `internal/output/terminal_width_linux_test.go` |
| Output width | Live terminal width | Terminal output detects the current terminal width when output is produced instead of assuming a fixed width | manual terminal check |
| Output width | Redirected standard stream | Width detection tries standard output, standard input, and standard error before using fallbacks | `internal/output/*_test.go` |
| Output width | Detection fallback | If no live terminal is available, a valid `COLUMNS` value is used; otherwise width falls back to `80` | `internal/output/alignment_test.go` |
| File output alignment | Current terminal width | `--output=<file>` calculates alignment using the current launching terminal width, and writes the resulting padding as fixed spaces | `internal/output/*_test.go` plus manual file check |

---

## Existing Renderer Color Golden Rules

Color output uses ANSI escape codes internally. User-facing color values support four notations:

- named colors, such as `red` or `orange`
- full hexadecimal colors in `#RRGGBB` form, such as `#ff0000`
- RGB colors in `rgb,<red>,<green>,<blue>` form, such as `rgb,255,0,0`
- HSL colors in `hsl,<hue>,<saturation>%,<lightness>%` form, such as `hsl,0,100%,50%`

Expected named color codes:

```text
black   -> \033[30m
red     -> \033[31m
green   -> \033[32m
yellow  -> \033[33m
blue    -> \033[34m
magenta -> \033[35m
cyan    -> \033[36m
white   -> \033[37m
orange  -> \033[38;5;208m
reset   -> \033[0m
```

Expected hexadecimal value notation:

```text
#ff0000 -> \033[38;2;255;0;0m
#FF0000 -> \033[38;2;255;0;0m
#00ff88 -> \033[38;2;0;255;136m
```

Expected RGB and HSL notation:

```text
rgb,255,0,0       -> \033[38;2;255;0;0m
RGB,0,255,0       -> \033[38;2;0;255;0m
hsl,0,100%,50%    -> \033[38;2;255;0;0m
HSL,120,100%,50%  -> \033[38;2;0;255;0m
```

Golden assertions for colored rendering:

- Color is applied to rendered ASCII-art character blocks, not to raw input before rendering.
- Every colored segment is followed by the reset code.
- Uncolored character blocks stay byte-for-byte identical to normal rendering.
- Output keeps the same visual 8-row structure per non-empty normalized line.
- Spaces and banner alignment are preserved.
- Substring matching is case-sensitive.
- Substring matching is repeated for every occurrence in a line.
- Substring matching does not cross escaped newline boundaries after normalization.
- A missing substring renders valid uncolored ASCII art.
- Multiple color flags in one command are not supported.

## Existing Output Alignment Golden Rules

Alignment output will use structured renderer data.

Expected alignment values:

```text
left
center
right
justify
```

Golden assertions for aligned rendering:

- Left alignment is byte-for-byte equivalent to the flattened structured output.
- Center alignment adds left padding based on visible row width.
- Right alignment adds left padding based on visible row width.
- Rows wider than the configured width are printed unchanged.
- ANSI color sequences do not count toward visible width.
- Justify expands only spaces between word segments.
- Justify does not expand spaces inside ASCII-art letters.
- Justify leaves one-word lines left-aligned.
- Leading and trailing spaces are preserved.
- File output and terminal output use the same aligned rows.
- Terminal and file output use the current terminal width measured when output
  is produced.
- A saved text file keeps its written padding and does not dynamically realign
  when opened in a differently sized viewer.

## Legacy CLI Audit Golden Outputs

These cases mirror the audit commands. Exact ASCII-art outputs are shown as they appear through `cat -e`, so those lines end with `$`. Color cases describe the ANSI and alignment assertions to verify.

### Invalid Argument Count

Command:

```bash
go run . "banana" standard abc
```

Expected output:

```text
Usage: go run . [STRING] [BANNER]

EX: go run . something standard
```

### Invalid Color Flag Format

Command:

```bash
go run . --color red "banana"
```

Expected output:

```text
Usage: go run . [OPTION] [STRING]

EX: go run . --color=<color> <substring to be colored> "something"
```

### Empty Color Value

Command:

```bash
go run . --color= "banana"
```

Expected output:

```text
Usage: go run . [OPTION] [STRING]

EX: go run . --color=<color> <substring to be colored> "something"
```

### Misspelled Color Flag

Commands:

```bash
go run . -color=red "banana"
go run . --colour=red "banana"
```

Expected output for each command:

```text
Usage: go run . [OPTION] [STRING]

EX: go run . --color=<color> <substring to be colored> "something"
```

### Multiple Color Flags

Command:

```bash
go run . --color=red --color=blue "hello"
```

Expected behavior:

- The command is invalid because only one color flag is supported.
- It must not attempt to render `hello`.

### Invalid Color Notation

Commands:

```bash
go run . --color=#f00 "hello"
go run . --color=rgb,256,0,0 "hello"
go run . --color=hsl,361,100%,50% "hello"
```

Expected behavior:

- Short or malformed hexadecimal values return `Invalid hexadecimal color code`.
- Invalid RGB values return `Invalid RGB color code`.
- Invalid HSL values return `Invalid HSL color code`.
- Unsupported color names return `Invalid color name`.

### Whole String Red

Command:

```bash
go run . --color=red "hello world"
```

Expected behavior:

- Renders `hello world` with `banners/standard.txt`.
- Every rendered character block is wrapped with red `\033[31m` and reset `\033[0m`.
- Output remains aligned across 8 rows.

### Whole String Hex Value

Command:

```bash
go run . --color=#ff0000 "hello world"
```

Expected behavior:

- Renders `hello world` with `banners/standard.txt`.
- Every rendered character block is wrapped with `\033[38;2;255;0;0m` and reset `\033[0m`.
- CLI preserves the color value `#ff0000` exactly for renderer validation.

### Whole String Uppercase Hex Value

Command:

```bash
go run . --color=#FFAA00 "hello"
```

Expected behavior:

- Renders `hello` with `banners/standard.txt`.
- Every rendered character block is wrapped with `\033[38;2;255;170;0m` and reset `\033[0m`.
- Uppercase hex digits are accepted.

### Whole String RGB Value

Command:

```bash
go run . --color=RGB,255,0,0 "hello"
```

Expected behavior:

- Renders `hello` with `banners/standard.txt`.
- Every rendered character block is wrapped with `\033[38;2;255;0;0m` and reset `\033[0m`.
- The RGB notation prefix is case-insensitive.

### Whole String HSL Value

Command:

```bash
go run . --color=hsl,120,100%,50% "hello"
```

Expected behavior:

- Renders `hello` with `banners/standard.txt`.
- Every rendered character block is wrapped with `\033[38;2;0;255;0m` and reset `\033[0m`.
- The HSL notation prefix is case-insensitive.

### Whole String Green With Numbers

Command:

```bash
go run . --color=green "1 + 1 = 2"
```

Expected behavior:

- Renders numbers, spaces, and symbols using `banners/standard.txt`.
- Every rendered character block is wrapped with green `\033[32m` and reset `\033[0m`.
- Spaces remain visible as banner spacing; they are not trimmed.

### Whole String Yellow With Special Characters

Command:

```bash
go run . --color=yellow "(%&) ??"
```

Expected behavior:

- Renders all printable special characters using `banners/standard.txt`.
- Every rendered character block is wrapped with yellow `\033[33m` and reset `\033[0m`.
- Output remains 8 rows and aligned.

### Substring Orange

Command:

```bash
go run . --color=orange GuYs "HeY GuYs"
```

Expected behavior:

- Renders `HeY GuYs` with `banners/standard.txt`.
- Only the `G`, `u`, `Y`, and `s` character blocks in `GuYs` are wrapped with orange `\033[38;5;208m` and reset `\033[0m`.
- `HeY ` remains uncolored.
- Matching is case-sensitive.

### One Letter Blue

Command:

```bash
go run . --color=blue B 'RGB()'
```

Expected behavior:

- Renders `RGB()` with `banners/standard.txt`.
- Only the `B` character block is wrapped with blue `\033[34m` and reset `\033[0m`.
- `R`, `G`, `(`, and `)` remain uncolored.

### Two-Letter Substring

Command:

```bash
go run . --color=cyan ll "hello"
```

Expected behavior:

- Renders `hello` with `banners/standard.txt`.
- Only both `l` character blocks in the `ll` substring are wrapped with cyan `\033[36m` and reset `\033[0m`.
- `h`, `e`, and `o` remain uncolored.

### Missing Substring

Command:

```bash
go run . --color=red xyz "hello"
```

Expected behavior:

- Renders valid uncolored ASCII art for `hello`.
- Does not crash.
- Does not emit ANSI color codes because `xyz` is not present.

### Newline Boundary Color

Command:

```bash
go run . --color=magenta lohe "hello\nhello"
```

Expected behavior:

- Renders two normalized lines: `hello` and `hello`.
- Does not color across the escaped newline boundary.
- Since `lohe` does not appear inside either normalized line, output remains uncolored.

### Banner-Compatible Color Extension

Command:

```bash
go run . "hello world" shadow --color=cyan world
```

Expected behavior:

- Renders with `banners/shadow.txt`.
- Only the `world` character blocks are wrapped with cyan `\033[36m` and reset `\033[0m`.
- `hello ` remains uncolored.

### hello standard

Command:

```bash
go run . "hello" standard | cat -e
```

Expected output:

```text
 _              _   _          $
| |            | | | |         $
| |__     ___  | | | |   ___   $
|  _ \   / _ \ | | | |  / _ \  $
| | | | |  __/ | | | | | (_) | $
|_| |_|  \___| |_| |_|  \___/  $
                               $
                               $
```

### hello world shadow

Command:

```bash
go run . "hello world" shadow | cat -e
```

Expected output:

```text
                                                                                        $
_|                _| _|                                                     _|       _| $
_|_|_|     _|_|   _| _|   _|_|         _|      _|      _|   _|_|   _|  _|_| _|   _|_|_| $
_|    _| _|_|_|_| _| _| _|    _|       _|      _|      _| _|    _| _|_|     _| _|    _| $
_|    _| _|       _| _| _|    _|         _|  _|  _|  _|   _|    _| _|       _| _|    _| $
_|    _|   _|_|_| _| _|   _|_|             _|      _|       _|_|   _|       _|   _|_|_| $
                                                                                        $
                                                                                        $
```

### nice 2 meet you thinkertoy

Command:

```bash
go run . "nice 2 meet you" thinkertoy | cat -e
```

Expected output:

```text
                                                                       $
                       --                       o                      $
     o                o  o                      |                      $
o-o     o-o o-o         /        o-O-o o-o o-o -o-       o  o o-o o  o $
|  | | |    |-'        /         | | | |-' |-'  |        |  | | | |  | $
o  o |  o-o o-o       o--o       o o o o-o o-o  o        o--O o-o o--o $
                                                            |          $
                                                         o--o          $
```

### you & me standard

Command:

```bash
go run . "you & me" standard | cat -e
```

Expected output:

```text
                                                                $
                                ___                             $
 _   _    ___    _   _         ( _ )          _ __ ___     ___  $
| | | |  / _ \  | | | |        / _ \/\       | '_ ` _ \   / _ \ $
| |_| | | (_) | | |_| |       | (_>  <       | | | | | | |  __/ $
 \__, |  \___/   \__,_|        \___/\/       |_| |_| |_|  \___| $
 __/ /                                                          $
|___/                                                           $
```

### 123 shadow

Command:

```bash
go run . "123" shadow | cat -e
```

Expected output:

```text
                       $
  _|   _|_|   _|_|_|   $
_|_| _|    _|       _| $
  _|     _|     _|_|   $
  _|   _|           _| $
  _| _|_|_|_| _|_|_|   $
                       $
                       $
```

### Slash Parenthesis Quote thinkertoy

Command:

```bash
go run . "/(\")" thinkertoy | cat -e
```

Expected output:

```text
         o o    $
    o  / | | \  $
   /  o       o $
  o   |       | $
 /    o       o $
o      \     /  $
                $
                $
```

### Uppercase Alphabet shadow

Command:

```bash
go run . "ABCDEFGHIJKLMNOPQRSTUVWXYZ" shadow | cat -e
```

Expected output:

```text
                                                                                                                                                                                                                                                              $
  _|_|   _|_|_|     _|_|_| _|_|_|   _|_|_|_| _|_|_|_|   _|_|_| _|    _| _|_|_|       _| _|    _| _|       _|      _| _|      _|   _|_|   _|_|_|     _|_|     _|_|_|     _|_|_| _|_|_|_|_| _|    _| _|      _| _|          _| _|      _| _|      _| _|_|_|_|_| $
_|    _| _|    _| _|       _|    _| _|       _|       _|       _|    _|   _|         _| _|  _|   _|       _|_|  _|_| _|_|    _| _|    _| _|    _| _|    _|   _|    _| _|           _|     _|    _| _|      _| _|          _|   _|  _|     _|  _|         _|   $
_|_|_|_| _|_|_|   _|       _|    _| _|_|_|   _|_|_|   _|  _|_| _|_|_|_|   _|         _| _|_|     _|       _|  _|  _| _|  _|  _| _|    _| _|_|_|   _|  _|_|   _|_|_|     _|_|       _|     _|    _| _|      _| _|    _|    _|     _|         _|         _|     $
_|    _| _|    _| _|       _|    _| _|       _|       _|    _| _|    _|   _|   _|    _| _|  _|   _|       _|      _| _|    _|_| _|    _| _|       _|    _|   _|    _|       _|     _|     _|    _|   _|  _|     _|  _|  _|     _|  _|       _|       _|       $
_|    _| _|_|_|     _|_|_| _|_|_|   _|_|_|_| _|         _|_|_| _|    _| _|_|_|   _|_|   _|    _| _|_|_|_| _|      _| _|      _|   _|_|   _|         _|_|  _| _|    _| _|_|_|       _|       _|_|       _|         _|  _|     _|      _|     _|     _|_|_|_|_| $
                                                                                                                                                                                                                                                              $
                                                                                                                                                                                                                                                              $
```

### Special Characters thinkertoy

Command:

```bash
go run . "\"#$%&/()*+,-./" thinkertoy | cat -e
```

Expected output:

```text
o o         | |                                                  $
| |  | |   -O-O-      O          o  / \  o | o                 o $
    -O-O- o | |   o  /    o     /  o   o  \|/   |             /  $
     | |   -O-O-    /    /|    o   |   | --O-- -o-           o   $
    -O-O-   | | o  /  o o-O-  /    o   o  /|\   |    o-o    /    $
     | |   -O-O-  O       |  o      \ /  o | o     o     O o     $
            | |                                    |             $
                                                                 $
```

### It's Working thinkertoy

Command:

```bash
go run . "It's Working" thinkertoy | cat -e
```

Expected output:

```text
          o                                              $
o-O-o  o  |           o       o         o                $
  |    |              |       |         | /  o           $
  |   -o-   o-o       o   o   o o-o o-o OO     o-o  o--o $
  |    |     \         \ / \ /  | | |   | \  | |  | |  | $
o-O-o  o    o-o         o   o   o-o o   o  o | o  o o--O $
                                                       | $
                                                    o--o $
```
