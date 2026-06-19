package filtergraph

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
	ctx := ParseForCompletion("")
	assertHasExpected(t, ctx, ExpectedFilterName)
}

func TestCompletionPartialFilterName(t *testing.T) {
	ctx := ParseForCompletion("sca")
	assertHasExpected(t, ctx, ExpectedFilterName)
	if ctx.PartialIdent != "sca" {
		t.Errorf("expected PartialIdent 'sca', got %q", ctx.PartialIdent)
	}
}

func TestCompletionAfterFilterName(t *testing.T) {
	ctx := ParseForCompletion("scale")
	assertHasExpected(t, ctx, ExpectedFilterOption)
	assertHasExpected(t, ctx, ExpectedFilterName)
}

func TestCompletionFilterOptionValue(t *testing.T) {
	ctx := ParseForCompletion("scale=w=1280")
	assertHasExpected(t, ctx, ExpectedFilterOptionValue)
	if ctx.Filter == nil {
		t.Fatal("expected filter context")
	}
	if ctx.Filter.Name != "scale" {
		t.Errorf("expected filter name 'scale', got %q", ctx.Filter.Name)
	}
}

func TestCompletionNewLabel(t *testing.T) {
	ctx := ParseForCompletion("[0:v]")
	if ctx.PartialIdent != "0:v" {
		t.Errorf("expected PartialIdent '0:v', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInLabel(t *testing.T) {
	ctx := ParseForCompletion("[0:v")
	if !ctx.InLabel {
		t.Error("expected InLabel=true")
	}
}

func TestCompletionOptionKeyTracking(t *testing.T) {
	ctx := ParseForCompletion("scale=w=1280:h=720")
	assertHasExpected(t, ctx, ExpectedFilterOptionValue)
	if ctx.Filter == nil {
		t.Fatal("expected filter context")
	}
	if ctx.Filter.Name != "scale" {
		t.Errorf("expected filter name 'scale', got %q", ctx.Filter.Name)
	}
	if ctx.Filter.OptionKey != "h" {
		t.Errorf("expected optionKey 'h', got %q", ctx.Filter.OptionKey)
	}
	if !ctx.Filter.InValue {
		t.Error("expected InValue=true")
	}
	expectedKeys := []string{"w", "h"}
	if len(ctx.Filter.OptionKeys) != len(expectedKeys) {
		t.Errorf("expected %d optionKeys, got %d", len(expectedKeys), len(ctx.Filter.OptionKeys))
	}
	for i, k := range expectedKeys {
		if ctx.Filter.OptionKeys[i] != k {
			t.Errorf("expected optionKeys[%d] = %q, got %q", i, k, ctx.Filter.OptionKeys[i])
		}
	}
}

func TestCompletionOptionKeyAfterEquals(t *testing.T) {
	ctx := ParseForCompletion("overlay=format=")
	assertHasExpected(t, ctx, ExpectedFilterOptionValue)
	if ctx.Filter == nil {
		t.Fatal("expected filter context")
	}
	if ctx.Filter.OptionKey != "format" {
		t.Errorf("expected optionKey 'format', got %q", ctx.Filter.OptionKey)
	}
	if !ctx.Filter.InValue {
		t.Error("expected InValue=true")
	}
}

func TestCompletionAmbiguousKeyValue(t *testing.T) {
	ctx := ParseForCompletion("scale=w")
	assertHasExpected(t, ctx, ExpectedFilterOptionKey)
	assertHasExpected(t, ctx, ExpectedFilterOptionValue)
	if ctx.Filter == nil {
		t.Fatal("expected filter context")
	}
	if ctx.Filter.Name != "scale" {
		t.Errorf("expected filter name 'scale', got %q", ctx.Filter.Name)
	}
}

func TestCompletionPositionalValue(t *testing.T) {
	ctx := ParseForCompletion("scale=1280")
	assertHasExpected(t, ctx, ExpectedFilterOptionKey)
	assertHasExpected(t, ctx, ExpectedFilterOptionValue)
	if ctx.Filter == nil {
		t.Fatal("expected filter context")
	}
	if ctx.Filter.ArgIndex != 0 {
		t.Errorf("expected argIndex 0, got %d", ctx.Filter.ArgIndex)
	}
}
