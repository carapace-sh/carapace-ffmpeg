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

func TestCompletionAfterOptionValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c:v", "libx264", "-i", "input.mp4"})
	if ctx.InputCount != 1 {
		t.Errorf("expected InputCount=1, got %d", ctx.InputCount)
	}
}

func TestCompletionInOutputContext(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-c:v", "libx264"})
	if ctx.Scope != ScopeInputFile {
		t.Errorf("expected ScopeInputFile, got %v", ctx.Scope)
	}
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertHasExpected(t, ctx, ExpectedOutputURL)
}
