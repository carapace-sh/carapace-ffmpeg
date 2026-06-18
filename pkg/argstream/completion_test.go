package argstream

import (
	"slices"
	"testing"
)

func assertHasExpected(t *testing.T, ctx *CompletionContext, expected ExpectedToken) {
	t.Helper()
	if !slices.Contains(ctx.ExpectedTokens, expected) {
		t.Errorf("expected token %v not found in %v", expected, ctx.ExpectedTokens)
	}
}

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion([]string{})
	assertHasExpected(t, ctx, ExpectedGlobalOption)
	assertHasExpected(t, ctx, ExpectedInputOption)
	assertHasExpected(t, ctx, ExpectedInputURL)
}

func TestCompletionAfterGlobalOption(t *testing.T) {
	ctx := ParseForCompletion([]string{"-y"})
	if ctx.Scope != ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", ctx.Scope)
	}
}

func TestCompletionAfterInput(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4"})
	if ctx.InputCount != 1 {
		t.Errorf("expected InputCount=1, got %d", ctx.InputCount)
	}
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertHasExpected(t, ctx, ExpectedOutputURL)
}

func TestCompletionPartialOption(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c:v"})
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.Name != "c" {
		t.Errorf("expected option name 'c', got %q", ctx.CurrentOption.Name)
	}
	if ctx.CurrentOption.StreamSpecifier != "v" {
		t.Errorf("expected specifier 'v', got %q", ctx.CurrentOption.StreamSpecifier)
	}
}

func TestCompletionOptionValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-c:v", "libx26"})
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.CanonicalName != "codec" {
		t.Errorf("expected canonical name 'codec', got %q", ctx.CurrentOption.CanonicalName)
	}
	if ctx.PartialValue != "libx26" {
		t.Errorf("expected partial value 'libx26', got %q", ctx.PartialValue)
	}
}

func TestCompletionOptionValueEmpty(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-ab", ""})
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.CanonicalName != "b" {
		t.Errorf("expected canonical name 'b', got %q", ctx.CurrentOption.CanonicalName)
	}
}

func TestCompletionInOutputContext(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-c:v", "libx264"})
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.CanonicalName != "codec" {
		t.Errorf("expected canonical name 'codec', got %q", ctx.CurrentOption.CanonicalName)
	}
}
