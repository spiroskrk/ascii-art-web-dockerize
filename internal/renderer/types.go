// Package renderer converts normalized input text into structured ASCII-art rows.
package renderer

// SegmentKind identifies whether a rendered segment is a word run or a space run.
type SegmentKind int

const (
	// SegmentWord is a contiguous run of non-space characters.
	SegmentWord SegmentKind = iota
	// SegmentSpace is a contiguous run of spaces.
	SegmentSpace
)

// RenderedSegment stores the original text run and its 8 rendered ASCII-art rows.
type RenderedSegment struct {
	Kind SegmentKind
	Text string
	Rows []string
}

// RenderedLine stores all word and space segments for one normalized input line.
type RenderedLine struct {
	Segments []RenderedSegment
}

// RenderedASCII is the structured renderer output consumed by the output package.
type RenderedASCII []RenderedLine
