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

func assertNotHasExpected(t *testing.T, ctx *CompletionContext, expected ExpectedToken) {
	t.Helper()
	if slices.Contains(ctx.ExpectedTokens, expected) {
		t.Errorf("did not expect token %v in %v", expected, ctx.ExpectedTokens)
	}
}

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion([]string{}, true)
	assertHasExpected(t, ctx, ExpectedGlobalOption)
	assertHasExpected(t, ctx, ExpectedInputOption)
	assertHasExpected(t, ctx, ExpectedInputURL)
}

func TestCompletionAfterGlobalOption(t *testing.T) {
	ctx := ParseForCompletion([]string{"-y"}, true)
	if ctx.Scope != ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", ctx.Scope)
	}
}

func TestCompletionAfterInput(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4"}, true)
	if ctx.InputCount != 1 {
		t.Errorf("expected InputCount=1, got %d", ctx.InputCount)
	}
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertHasExpected(t, ctx, ExpectedOutputURL)
}

func TestCompletionPartialOption(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c:v"}, false)
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

func TestCompletionOptionValueMidToken(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-c:v", "libx26"}, false)
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

func TestCompletionOptionValueTrailingSpace(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-c:v", "libx264"}, true)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertHasExpected(t, ctx, ExpectedOutputURL)
	assertNotHasExpected(t, ctx, ExpectedOptionValue)
}

func TestCompletionAfterOptionWithValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-c:v", "libx264"}, true)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertHasExpected(t, ctx, ExpectedOutputURL)
}

func TestCompletionVcodecExpectsValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-vcodec"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.Name != "vcodec" {
		t.Errorf("expected option name 'vcodec', got %q", ctx.CurrentOption.Name)
	}
	if ctx.CurrentOption.StreamSpecifier != "v" {
		t.Errorf("expected implicit stream specifier 'v', got %q", ctx.CurrentOption.StreamSpecifier)
	}
	if ctx.CurrentOption.AcceptsSpec {
		t.Error("vcodec should not accept additional stream specifier")
	}
}

func TestCompletionVcodecMidToken(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-vcodec"}, false)
	assertHasExpected(t, ctx, ExpectedOptionValue)
}

func TestCompletionAcodecExpectsValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-acodec"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.StreamSpecifier != "a" {
		t.Errorf("expected implicit stream specifier 'a', got %q", ctx.CurrentOption.StreamSpecifier)
	}
}

func TestCompletionVcodecValueTrailingSpace(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-vcodec", "libx264"}, true)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertHasExpected(t, ctx, ExpectedOutputURL)
	assertNotHasExpected(t, ctx, ExpectedOptionValue)
}

func TestCompletionVcodecValueMidToken(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-vcodec", "libx26"}, false)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.StreamSpecifier != "v" {
		t.Errorf("expected implicit stream specifier 'v', got %q", ctx.CurrentOption.StreamSpecifier)
	}
	if ctx.PartialValue != "libx26" {
		t.Errorf("expected partial value 'libx26', got %q", ctx.PartialValue)
	}
}

func TestCompletionVfExpectsValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-vf"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.Name != "vf" {
		t.Errorf("expected option name 'vf', got %q", ctx.CurrentOption.Name)
	}
	if ctx.CurrentOption.StreamSpecifier != "v" {
		t.Errorf("expected implicit stream specifier 'v', got %q", ctx.CurrentOption.StreamSpecifier)
	}
	if ctx.CurrentOption.AcceptsSpec {
		t.Error("vf should not accept additional stream specifier")
	}
}

func TestCompletionAfExpectsValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-af"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.StreamSpecifier != "a" {
		t.Errorf("expected implicit stream specifier 'a', got %q", ctx.CurrentOption.StreamSpecifier)
	}
}

func TestCompletionAbExpectsValue(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "input.mp4", "-ab"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.StreamSpecifier != "a" {
		t.Errorf("expected implicit stream specifier 'a', got %q", ctx.CurrentOption.StreamSpecifier)
	}
}
