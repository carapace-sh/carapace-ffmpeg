package mapvalue

import (
	"testing"
)

func TestParseSimple(t *testing.T) {
	mv, err := Parse("0")
	if err != nil {
		t.Fatal(err)
	}
	if mv.FileIndex != 0 {
		t.Errorf("expected file index 0, got %d", mv.FileIndex)
	}
	if mv.HasSpecifier {
		t.Error("expected no specifier")
	}
}

func TestParseWithSpecifier(t *testing.T) {
	mv, err := Parse("0:v")
	if err != nil {
		t.Fatal(err)
	}
	if mv.FileIndex != 0 {
		t.Errorf("expected file index 0, got %d", mv.FileIndex)
	}
	if !mv.HasSpecifier || mv.Specifier != "v" {
		t.Errorf("expected specifier 'v', got %q", mv.Specifier)
	}
}

func TestParseWithStreamIndex(t *testing.T) {
	mv, err := Parse("0:a:1")
	if err != nil {
		t.Fatal(err)
	}
	if mv.Specifier != "a:1" {
		t.Errorf("expected specifier 'a:1', got %q", mv.Specifier)
	}
}

func TestParseNegative(t *testing.T) {
	mv, err := Parse("-0:a:1")
	if err != nil {
		t.Fatal(err)
	}
	if !mv.Negate {
		t.Error("expected negate=true")
	}
	if mv.FileIndex != 0 {
		t.Errorf("expected file index 0, got %d", mv.FileIndex)
	}
}

func TestParseOptional(t *testing.T) {
	mv, err := Parse("0:a?")
	if err != nil {
		t.Fatal(err)
	}
	if !mv.Optional {
		t.Error("expected optional=true")
	}
	if mv.Specifier != "a" {
		t.Errorf("expected specifier 'a', got %q", mv.Specifier)
	}
}

func TestParseLinkLabel(t *testing.T) {
	mv, err := Parse("[out]")
	if err != nil {
		t.Fatal(err)
	}
	if !mv.IsLinkLabel {
		t.Error("expected IsLinkLabel=true")
	}
	if mv.LinkLabel != "out" {
		t.Errorf("expected link label 'out', got %q", mv.LinkLabel)
	}
}

func TestParseMapViewSpecifier(t *testing.T) {
	mv, err := Parse("0:v:0:view:all")
	if err != nil {
		t.Fatal(err)
	}
	if !mv.HasView {
		t.Error("expected HasView=true")
	}
	if mv.ViewSpec != "view:all" {
		t.Errorf("expected view spec 'view:all', got %q", mv.ViewSpec)
	}
}

func TestFormatRoundtrip(t *testing.T) {
	tests := []string{
		"0",
		"0:v",
		"0:a:1",
		"-0:a:1",
		"0:a?",
		"[out]",
		"0:v:0:view:all",
	}
	for _, tt := range tests {
		mv, err := Parse(tt)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt, err)
		}
		got := Format(mv)
		if got != tt {
			t.Errorf("Format(Parse(%q)) = %q, want %q", tt, got, tt)
		}
	}
}

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion("")
	if len(ctx.ExpectedTokens) == 0 {
		t.Error("expected some expected tokens")
	}
}

func TestCompletionFileIndex(t *testing.T) {
	ctx := ParseForCompletion("0")
	if ctx.FileIndex != 0 {
		t.Errorf("expected FileIndex 0, got %d", ctx.FileIndex)
	}
}

func TestCompletionAfterFileIndex(t *testing.T) {
	ctx := ParseForCompletion("0:")
	if !ctx.HasSpecifier {
		t.Error("expected HasSpecifier=true")
	}
}

func TestCompletionLinkLabel(t *testing.T) {
	ctx := ParseForCompletion("[out")
	if !ctx.IsLinkLabel {
		t.Error("expected IsLinkLabel=true")
	}
}
