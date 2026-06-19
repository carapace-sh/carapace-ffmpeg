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

func TestCompletionAfterIFlag(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i"}, true)
	assertHasExpected(t, ctx, ExpectedInputURL)
	assertNotHasExpected(t, ctx, ExpectedInputOption)
	assertNotHasExpected(t, ctx, ExpectedGlobalOption)
}

func TestCompletionAfterPreOptionAndIFlag(t *testing.T) {
	ctx := ParseForCompletion([]string{"-f", "mp4", "-i"}, true)
	assertHasExpected(t, ctx, ExpectedInputURL)
	assertNotHasExpected(t, ctx, ExpectedInputOption)
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

func TestCompletionShellSplitColonOnly(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c", ":"}, true)
	assertHasExpected(t, ctx, ExpectedStreamSpecifier)
	assertNotHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.Name != "c" {
		t.Errorf("expected option name 'c', got %q", ctx.CurrentOption.Name)
	}
}

func TestCompletionShellSplitColonMidToken(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c", ":"}, false)
	assertHasExpected(t, ctx, ExpectedStreamSpecifier)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
}

func TestCompletionShellSplitColonSpecMidToken(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c", ":v"}, false)
	assertHasExpected(t, ctx, ExpectedStreamSpecifier)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.PartialSpec != "v" {
		t.Errorf("expected partial spec 'v', got %q", ctx.PartialSpec)
	}
}

func TestCompletionShellSplitColonSpecTrailingSpace(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c", ":v"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.StreamSpecifier != "v" {
		t.Errorf("expected stream specifier 'v', got %q", ctx.CurrentOption.StreamSpecifier)
	}
}

func TestCompletionMidTokenOptionWithColon(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c:"}, false)
	assertHasExpected(t, ctx, ExpectedStreamSpecifier)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.Name != "c" {
		t.Errorf("expected option name 'c', got %q", ctx.CurrentOption.Name)
	}
}

func TestCompletionMidTokenOptionWithSpec(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c:v"}, false)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.StreamSpecifier != "v" {
		t.Errorf("expected stream specifier 'v', got %q", ctx.CurrentOption.StreamSpecifier)
	}
}

func TestCompletionShellSplitNonSpecOption(t *testing.T) {
	// -loglevel :  → the ":" is consumed as the loglevel value (not a stream specifier)
	ctx := ParseForCompletion([]string{"-loglevel", ":"}, true)
	assertNotHasExpected(t, ctx, ExpectedOptionValue)
	assertNotHasExpected(t, ctx, ExpectedStreamSpecifier)
}

func TestCompletionShellSplitEmptySpecifierNextArg(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c", ":", "v"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.StreamSpecifier != "v" {
		t.Errorf("expected stream specifier 'v', got %q", ctx.CurrentOption.StreamSpecifier)
	}
}

func TestCompletionShellSplitEmptySpecifierNextArgMidToken(t *testing.T) {
	ctx := ParseForCompletion([]string{"-c", ":"}, true)
	assertHasExpected(t, ctx, ExpectedStreamSpecifier)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
}

func TestCompletionImplicitSpecNoStreamSpecifier(t *testing.T) {
	ctx := ParseForCompletion([]string{"-acodec"}, true)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.AcceptsSpec {
		t.Error("acodec should not accept additional stream specifier")
	}
}

func TestCompletionImplicitSpecMidTokenWithColon(t *testing.T) {
	ctx := ParseForCompletion([]string{"-acodec:"}, false)
	assertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
}

func TestCompletionPartialDashAfterValue(t *testing.T) {
	// "-i test -c:v 012v -" with no trailing space: the dash is a partial option
	ctx := ParseForCompletion([]string{"-i", "test", "-c:v", "012v", "-"}, false)
	assertHasExpected(t, ctx, ExpectedInputOption)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertNotHasExpected(t, ctx, ExpectedOutputURL)
	if ctx.Scope != ScopeInputFile {
		t.Errorf("expected ScopeInputFile, got %v", ctx.Scope)
	}
}

func TestCompletionPartialDashAfterValueTrailingSpace(t *testing.T) {
	// "-i test -c:v libx264 -" with trailing space: the dash is a complete token (stdin)
	ctx := ParseForCompletion([]string{"-i", "test", "-c:v", "libx264", "-"}, true)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertHasExpected(t, ctx, ExpectedOutputURL)
	if ctx.Scope != ScopeOutputFile {
		t.Errorf("expected ScopeOutputFile, got %v", ctx.Scope)
	}
}

func TestCompletionPartialDoubleDashAfterValue(t *testing.T) {
	// "--" with no trailing space: partial long option being typed
	ctx := ParseForCompletion([]string{"-i", "test", "-c:v", "libx264", "--"}, false)
	assertHasExpected(t, ctx, ExpectedInputOption)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	if ctx.Scope != ScopeInputFile {
		t.Errorf("expected ScopeInputFile, got %v", ctx.Scope)
	}
}

func TestCompletionPartialUnknownOptionAtGlobalScope(t *testing.T) {
	// "-l" at global scope should include both GlobalOption and InputOption
	// so that global options like -lavfi and -loglevel are shown alongside
	// per-stream options like -level
	ctx := ParseForCompletion([]string{"-l"}, false)
	assertHasExpected(t, ctx, ExpectedGlobalOption)
	assertHasExpected(t, ctx, ExpectedInputOption)
	if ctx.Scope != ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", ctx.Scope)
	}
}

func TestCompletionPartialKnownGlobalOptionAtGlobalScope(t *testing.T) {
	// "-loglev" is a partial match for -loglevel, LookupOption returns nil
	ctx := ParseForCompletion([]string{"-loglev"}, false)
	assertHasExpected(t, ctx, ExpectedGlobalOption)
	assertHasExpected(t, ctx, ExpectedInputOption)
	if ctx.Scope != ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", ctx.Scope)
	}
}

func TestCompletionPartialKnownPerStreamOptionAtGlobalScope(t *testing.T) {
	// "-leve" matches no exact option; per-stream option -level should
	// be reachable via InputOption, and global options via GlobalOption
	ctx := ParseForCompletion([]string{"-leve"}, false)
	assertHasExpected(t, ctx, ExpectedGlobalOption)
	assertHasExpected(t, ctx, ExpectedInputOption)
	if ctx.Scope != ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", ctx.Scope)
	}
}

func TestCompletionPartialUnknownOptionAtInputScope(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "test", "-l"}, false)
	assertHasExpected(t, ctx, ExpectedInputOption)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertNotHasExpected(t, ctx, ExpectedGlobalOption)
	if ctx.Scope != ScopeInputFile {
		t.Errorf("expected ScopeInputFile, got %v", ctx.Scope)
	}
}

func TestCompletionPartialUnknownOptionAtOutputScope(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "test", "out.mp4", "-l"}, false)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	assertNotHasExpected(t, ctx, ExpectedGlobalOption)
	assertNotHasExpected(t, ctx, ExpectedInputOption)
	if ctx.Scope != ScopeOutputFile {
		t.Errorf("expected ScopeOutputFile, got %v", ctx.Scope)
	}
}

func TestCompletionKnownPerStreamOptionMidTokenAtGlobalScope(t *testing.T) {
	// Typing "-level" mid-token: LookupOption("level") succeeds,
	// but option is per-stream so the default branch adds scope-appropriate tokens
	ctx := ParseForCompletion([]string{"-level"}, false)
	assertHasExpected(t, ctx, ExpectedGlobalOption)
	assertHasExpected(t, ctx, ExpectedInputOption)
	if ctx.Scope != ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", ctx.Scope)
	}
}

func TestCompletionKnownPerStreamOptionMidTokenAtInputScope(t *testing.T) {
	ctx := ParseForCompletion([]string{"-i", "test", "-level"}, false)
	assertHasExpected(t, ctx, ExpectedInputOption)
	assertHasExpected(t, ctx, ExpectedOutputOption)
	if ctx.Scope != ScopeInputFile {
		t.Errorf("expected ScopeInputFile, got %v", ctx.Scope)
	}
}

func TestCompletionPartialDashAtGlobalScope(t *testing.T) {
	// bare "-" at global scope: the non-option partial path
	ctx := ParseForCompletion([]string{"-"}, false)
	assertHasExpected(t, ctx, ExpectedGlobalOption)
	assertHasExpected(t, ctx, ExpectedInputOption)
	if ctx.Scope != ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", ctx.Scope)
	}
}
