// Package renderer converts normalized input text into structured ASCII-art rows.
package renderer

import "strings"

// NormalizeInput splits input text into logical lines.
//
// A "logical line" is produced by any of the following separators, all of
// which are treated as equivalent:
//   - a literal backslash-n sequence (\n typed as two characters), which is
//     how the original CLI accepted embedded newlines;
//   - a real CRLF line ending ("\r\n"), typical of Windows browsers;
//   - a real lone CR line ending ("\r"), typical of old Mac line endings;
//   - a real LF line ending ("\n"), typical of Unix/browser textareas.
//
// All forms are normalized to a single LF separator before splitting, so the
// rest of the renderer only ever has to reason about one kind of line break.
//
// Trailing whitespace, leading whitespace, and repeated spaces inside a line
// are preserved exactly as received. This function only decides where one
// logical line ends and the next begins — it never trims, cases, escapes, or
// renders anything.
func NormalizeInput(input string) []string {
	if input == "" {
		return []string{}
	}

	// Step 1 — collapse real newline variants to a single LF.
	//
	// CRLF must be replaced before a lone CR. If we replaced CR first, the
	// "\r" half of every CRLF pair would already be gone, and the trailing
	// "\n" would look like an independent LF — turning one Windows line
	// ending into what looks like two separators. Handling CRLF first
	// guarantees each real line ending contributes exactly one logical
	// separator.
	normalized := strings.ReplaceAll(input, "\r\n", "\n")   // CRLF -> LF
	normalized = strings.ReplaceAll(normalized, "\r", "\n") // lone CR -> LF

	// Step 2 — make the literal "\n" escape sequence (backslash + n, as
	// typed by a CLI user) equivalent to a real LF. This preserves the
	// original CLI behavior: existing callers that send literal "\n" keep
	// producing the same logical lines as before.
	normalized = strings.ReplaceAll(normalized, `\n`, "\n")

	// Step 3 — every separator is now a single LF character, so a plain
	// split is enough to produce the logical lines.
	lines := strings.Split(normalized, "\n")

	// Step 4 — preserve existing behavior: input that ends with a newline
	// (real or literal) does not produce a trailing empty logical line.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}
