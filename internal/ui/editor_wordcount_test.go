package ui

import (
	"testing"
)

func TestCountWordsAndCharsEmptyText(t *testing.T) {
	words, chars := countWordsAndChars("")
	if words != 0 || chars != 0 {
		t.Fatalf("unexpected count: words=%d want=0, chars=%d want=0", words, chars)
	}
}

func TestCountWordsAndCharsSingleWord(t *testing.T) {
	words, chars := countWordsAndChars("hello")
	if words != 1 {
		t.Fatalf("unexpected word count: got=%d want=1", words)
	}
	if chars != 5 {
		t.Fatalf("unexpected char count: got=%d want=5", chars)
	}
}

func TestCountWordsAndCharsMultipleWords(t *testing.T) {
	text := "hello world how are you"
	words, chars := countWordsAndChars(text)
	if words != 5 {
		t.Fatalf("unexpected word count: got=%d want=5", words)
	}
	if chars != 23 {
		t.Fatalf("unexpected char count: got=%d want=23", chars)
	}
}

func TestCountWordsAndCharsWithNewlines(t *testing.T) {
	text := "line one\nline two\nline three"
	words, chars := countWordsAndChars(text)
	if words != 6 {
		t.Fatalf("unexpected word count: got=%d want=6", words)
	}
	if chars != 28 {
		t.Fatalf("unexpected char count: got=%d want=28", chars)
	}
}

func TestCountWordsAndCharsWithExtraSpaces(t *testing.T) {
	text := "hello    world    test"
	words, chars := countWordsAndChars(text)
	if words != 3 {
		t.Fatalf("unexpected word count: got=%d want=3", words)
	}
	if chars != 22 {
		t.Fatalf("unexpected char count: got=%d want=22", chars)
	}
}

func TestCountWordsAndCharsWithTabs(t *testing.T) {
	text := "hello\tworld\ttest"
	words, chars := countWordsAndChars(text)
	if words != 3 {
		t.Fatalf("unexpected word count: got=%d want=3", words)
	}
	if chars != 16 {
		t.Fatalf("unexpected char count: got=%d want=16", chars)
	}
}

func TestCountWordsAndCharsSpecialCharacters(t *testing.T) {
	text := "hello! world? test."
	words, chars := countWordsAndChars(text)
	if words != 3 {
		t.Fatalf("unexpected word count: got=%d want=3", words)
	}
	if chars != 19 {
		t.Fatalf("unexpected char count: got=%d want=19", chars)
	}
}
