package argstream

import (
	"testing"
)

func TestFFplayProfileHasNoOutputSection(t *testing.T) {
	if DefaultFFplayProfile.HasOutputSection {
		t.Error("ffplay profile should not have output section")
	}
}

func TestFFprobeProfileHasNoOutputSection(t *testing.T) {
	if DefaultFFprobeProfile.HasOutputSection {
		t.Error("ffprobe profile should not have output section")
	}
}

func TestFFmpegProfileHasOutputSection(t *testing.T) {
	if !DefaultFFmpegProfile.HasOutputSection {
		t.Error("ffmpeg profile should have output section")
	}
}

func TestFFplayProfileOptionsExist(t *testing.T) {
	opt := DefaultFFplayProfile.LookupOption("showmode")
	if opt == nil {
		t.Fatal("expected showmode option in ffplay profile")
	}
	if opt.ValueType != ValueShowMode {
		t.Errorf("expected showmode ValueType to be ValueShowMode, got %v", opt.ValueType)
	}

	// ffplay-specific option
	opt = DefaultFFplayProfile.LookupOption("fs")
	if opt == nil {
		t.Fatal("expected fs option in ffplay profile")
	}
	if opt.Type != TypeBoolean {
		t.Errorf("expected fs to be boolean, got %v", opt.Type)
	}

	// -y in ffplay means "display height", not "overwrite"
	opt = DefaultFFplayProfile.LookupOption("y")
	if opt == nil {
		t.Fatal("expected y option in ffplay profile")
	}
	if opt.Type != TypeValue {
		t.Errorf("expected y to be a value option (display height), got %v", opt.Type)
	}
}

func TestFFprobeProfileOptionsExist(t *testing.T) {
	opt := DefaultFFprobeProfile.LookupOption("show_streams")
	if opt == nil {
		t.Fatal("expected show_streams option in ffprobe profile")
	}
	if opt.Type != TypeBoolean {
		t.Errorf("expected show_streams to be boolean, got %v", opt.Type)
	}

	opt = DefaultFFprobeProfile.LookupOption("output_format")
	if opt == nil {
		t.Fatal("expected output_format option in ffprobe profile")
	}
	if opt.ValueType != ValueProbeOutputFmt {
		t.Errorf("expected output_format ValueType to be ValueProbeOutputFmt, got %v", opt.ValueType)
	}

	// -o option in ffprobe
	opt = DefaultFFprobeProfile.LookupOption("o")
	if opt == nil {
		t.Fatal("expected o option in ffprobe profile")
	}
	if opt.ValueType != ValueFileURL {
		t.Errorf("expected o to be ValueFileURL, got %v", opt.ValueType)
	}
}

func TestFFplayProfileNoOutputOnlyOptions(t *testing.T) {
	// ffmpeg output-only options should not be in the ffplay profile
	opt := DefaultFFplayProfile.LookupOption("map")
	if opt != nil {
		t.Error("ffplay should not have -map option (output-only)")
	}
	opt = DefaultFFplayProfile.LookupOption("shortest")
	if opt != nil {
		t.Error("ffplay should not have -shortest option (output-only)")
	}
}

func TestFFprobeProfileNoOutputOnlyOptions(t *testing.T) {
	opt := DefaultFFprobeProfile.LookupOption("map")
	if opt != nil {
		t.Error("ffprobe should not have -map option (output-only)")
	}
}

func TestFFplayCompletionNoOutputContext(t *testing.T) {
	// After -i input.mp4, ffplay should expect input options, not output
	ctx := ParseForCompletionWithProfile([]string{"-i", "input.mp4"}, true, DefaultFFplayProfile)
	if ctx.Scope != ScopeInputFile {
		t.Errorf("expected ScopeInputFile after -i, got %v", ctx.Scope)
	}
	AssertHasExpected(t, ctx, ExpectedInputOption)
	AssertNotHasExpected(t, ctx, ExpectedOutputOption)
	AssertNotHasExpected(t, ctx, ExpectedOutputURL)
}

func TestFFprobeCompletionNoOutputContext(t *testing.T) {
	// After -f mp4 -i input.mp4, ffprobe should stay in input context
	ctx := ParseForCompletionWithProfile([]string{"-f", "mp4", "-i", "input.mp4"}, true, DefaultFFprobeProfile)
	if ctx.Scope != ScopeInputFile {
		t.Errorf("expected ScopeInputFile, got %v", ctx.Scope)
	}
	AssertHasExpected(t, ctx, ExpectedInputOption)
	AssertNotHasExpected(t, ctx, ExpectedOutputOption)
	AssertNotHasExpected(t, ctx, ExpectedOutputURL)
}

func TestFFplayPositionalInputWithoutI(t *testing.T) {
	// ffplay input.mp4 — positional arg should be input URL
	ctx := ParseForCompletionWithProfile([]string{}, true, DefaultFFplayProfile)
	AssertHasExpected(t, ctx, ExpectedInputURL)
}

func TestFFplayShowmodeOption(t *testing.T) {
	ctx := ParseForCompletionWithProfile([]string{"-showmode"}, true, DefaultFFplayProfile)
	AssertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.ValueType != ValueShowMode {
		t.Errorf("expected ValueShowMode, got %v", ctx.CurrentOption.ValueType)
	}
}

func TestFFprobeOutputFormatOption(t *testing.T) {
	ctx := ParseForCompletionWithProfile([]string{"-of"}, true, DefaultFFprobeProfile)
	AssertHasExpected(t, ctx, ExpectedOptionValue)
	if ctx.CurrentOption == nil {
		t.Fatal("expected CurrentOption")
	}
	if ctx.CurrentOption.ValueType != ValueProbeOutputFmt {
		t.Errorf("expected ValueProbeOutputFmt, got %v", ctx.CurrentOption.ValueType)
	}
}

func AssertHasExpected(t *testing.T, ctx *CompletionContext, expected ExpectedToken) {
	t.Helper()
	for _, tok := range ctx.ExpectedTokens {
		if tok == expected {
			return
		}
	}
	t.Errorf("expected token %v not found in %v", expected, ctx.ExpectedTokens)
}

func AssertNotHasExpected(t *testing.T, ctx *CompletionContext, expected ExpectedToken) {
	t.Helper()
	for _, tok := range ctx.ExpectedTokens {
		if tok == expected {
			t.Errorf("did not expect token %v in %v", expected, ctx.ExpectedTokens)
		}
	}
}