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

func TestParseMetadataKeyCollisionWithViewKeyword(t *testing.T) {
	// The metadata specifier is greedy in ffmpeg: "m:view:all" is metadata
	// key "view" with value "all", not a view specifier
	mv, err := Parse("0:m:view:all")
	if err != nil {
		t.Fatal(err)
	}
	if mv.Specifier != "m:view:all" {
		t.Errorf("expected specifier 'm:view:all', got %q", mv.Specifier)
	}
	if mv.HasView {
		t.Error("expected HasView=false")
	}
}

func TestParseViewSpecifierAfterIndex(t *testing.T) {
	mv, err := Parse("0:v:vidx:0")
	if err != nil {
		t.Fatal(err)
	}
	if mv.Specifier != "v" {
		t.Errorf("expected specifier 'v', got %q", mv.Specifier)
	}
	if !mv.HasView || mv.ViewSpec != "vidx:0" {
		t.Errorf("expected view spec 'vidx:0', got %q", mv.ViewSpec)
	}
}

func TestParseViewSpecifierWithOptional(t *testing.T) {
	mv, err := Parse("0:v:0:view:all?")
	if err != nil {
		t.Fatal(err)
	}
	if !mv.HasView || mv.ViewSpec != "view:all" || !mv.Optional {
		t.Errorf("expected view:all optional, got view %q optional %v", mv.ViewSpec, mv.Optional)
	}
}

func TestParseOptionalAfterEmptySpecifier(t *testing.T) {
	mv, err := Parse("0:a:?")
	if err != nil {
		t.Fatal(err)
	}
	if mv.Specifier != "a" || !mv.Optional {
		t.Errorf("expected specifier 'a' optional, got %q optional %v", mv.Specifier, mv.Optional)
	}
}

func TestParseViewPosition(t *testing.T) {
	mv, err := Parse("0:v:vpos:left")
	if err != nil {
		t.Fatal(err)
	}
	if !mv.HasView || mv.ViewSpec != "vpos:left" {
		t.Errorf("expected view spec 'vpos:left', got %q", mv.ViewSpec)
	}
}

func TestParseInvalidViewPosition(t *testing.T) {
	if _, err := Parse("0:v:vpos:bogus"); err == nil {
		t.Error("expected error for invalid view position")
	}
}

func TestParseTrailingGarbageAfterSpecifier(t *testing.T) {
	if _, err := Parse("0:v:x"); err == nil {
		t.Error("expected error for trailing garbage")
	}
}

func TestParseHexFileIndex(t *testing.T) {
	mv, err := Parse("0x1:v")
	if err != nil {
		t.Fatal(err)
	}
	if mv.FileIndex != 1 {
		t.Errorf("expected file index 1, got %d", mv.FileIndex)
	}
}

func TestParseMetadataEscapedColon(t *testing.T) {
	mv, err := Parse(`0:m:lang\:x:eng`)
	if err != nil {
		t.Fatal(err)
	}
	if mv.Specifier != `m:lang\:x:eng` {
		t.Errorf("expected specifier 'm:lang\\:x:eng', got %q", mv.Specifier)
	}
}

func TestParseLinkLabelOptionalError(t *testing.T) {
	// '?' is only valid on the stream form, not on link labels
	if _, err := Parse("[out]?"); err == nil {
		t.Error("expected error for optional link label")
	}
}

func TestFormatRoundtripViewSpecifier(t *testing.T) {
	tests := []string{
		"0:m:view:all",
		"0:v:vidx:0",
		"0:a?",
		"0:",
		"0:v:vpos:right",
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

func TestFormatNormalizesOptionalSeparator(t *testing.T) {
	// "0:a:?" and "0:a?" are equivalent; Format emits the canonical form
	mv, err := Parse("0:a:?")
	if err != nil {
		t.Fatal(err)
	}
	if got := Format(mv); got != "0:a?" {
		t.Errorf("Format(Parse(\"0:a:?\")) = %q, want \"0:a?\"", got)
	}
}
