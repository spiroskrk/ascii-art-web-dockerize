# Card A2 — Shared generation workflow (`generate.go`)

**Package:** A (core pipeline) · **Owner:** Aris
**Priority:** 🔵 on critical path (blocks B3, C1, C2)

---

**Context:** one function that both display and download call, turning a validated
`GenerationInput` into `GeneratedASCII`.

## Steps
- [ ] Map banner identifier → bundled path via the preloaded registry
- [ ] Load banner from the registry (never from a client path)
- [ ] Normalize logical input lines (uses A1)
- [ ] Build structured, optionally colored ASCII (call renderer)
- [ ] Apply selected alignment (call output)
- [ ] Return `GeneratedASCII { ANSIText, Width }`
- [ ] Tests: 3 banners, multiline input, all 4 alignments at fixed widths,
      color on/off, whole-input + substring coloring, unsupported-char error

## Acceptance criteria
- [ ] Display and download derive from this same aligned result
- [ ] No banner glyph construction, substring matching, or alignment is
      re-implemented here — it only calls `banner`/`renderer`/`output`
- [ ] Returns aligned ANSI text with no HTML concerns

## Meta
- **Files:** `internal/web/generate.go` + `generate_test.go`
- **PRD:** §5.3, §5.5
- **Depends on:** A1 · Phase 0 structs
- **Blocks:** B3 (download), C1/C2 (presenter/color)

**Definition of Done:** generation tests pass independently of HTTP; signature
matches the Phase 0 contract so B and C compile against it.

---

## Design tokens (locked in Phase 0)

Style every control with these shared CSS variables in `static/style.css`.
**Do not introduce new colours.**

| Token | Hex | Use |
| --- | --- | --- |
| `--page-bg` | `#FAF8F3` | page background — soft warm off-white |
| `--card` | `#FFFFFF` | card surface |
| `--accent` | `#8F3B3B` | wine/μπορντό — logo accent, primary action, active states |
| `--text` | `#3B322A` | espresso — primary text |
| `--text-muted` | `#7A7062` | labels, secondary text, placeholders |
| `--border` | `#E8E4DA` | field and card borders |
| `--field-bg` | `#EEEDE7` | input / control fill |

**Output box (terminal):** dark surface `#2E2820`, light text `#F6EFE2`, monospace.
Colored runs use `rgb(R, G, B)` from validated integers only.
