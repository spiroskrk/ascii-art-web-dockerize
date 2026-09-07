// Package renderer tests substring position detection for coloring.
package renderer

import (
	"reflect"
	"testing"
)

func TestColoredPositionsFoundOnce(t *testing.T) {
	result := coloredPositions("ABC", "B")
	expected := map[int]bool{1: true}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}

func TestColoredPositionsFoundMultipleTimes(t *testing.T) {
	result := coloredPositions("a king kitten have kit", "kit")
	expected := map[int]bool{
		7:  true,
		8:  true,
		9:  true,
		19: true,
		20: true,
		21: true,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}

func TestColoredPositionsNotFound(t *testing.T) {
	result := coloredPositions("hello", "z")
	expected := map[int]bool{}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}

func TestColoredPositionsEmptySubstringColorsWholeString(t *testing.T) {
	result := coloredPositions("abc", "")
	expected := map[int]bool{0: true, 1: true, 2: true}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}

func TestColoredPositionsCaseSensitive(t *testing.T) {
	result := coloredPositions("Kit kit", "kit")
	expected := map[int]bool{4: true, 5: true, 6: true}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}

func TestColoredPositionsOverlappingMatches(t *testing.T) {
	result := coloredPositions("banana", "an")
	expected := map[int]bool{1: true, 2: true, 3: true, 4: true}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}

func TestColoredPositionsEmptyInputString(t *testing.T) {
	result := coloredPositions("", "a")
	expected := map[int]bool{}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}

func TestColoredPositionsDoesNotCrossNewline(t *testing.T) {
	result := coloredPositions("ab\ncd", "b\nc")
	expected := map[int]bool{}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("coloredPositions() = %v, expected %v", result, expected)
	}
}
