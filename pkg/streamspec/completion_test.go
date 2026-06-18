package streamspec

import (
	"testing"
)

func assertHasExpected(t *testing.T, ctx *CompletionContext, expected ExpectedToken) {
	t.Helper()
	for _, e := range ctx.ExpectedTokens {
		if e == expected {
			return
		}
	}
	t.Errorf("expected token %v not found in %v", expected, ctx.ExpectedTokens)
}

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion("")
	assertHasExpected(t, ctx, ExpectedSpecifierType)
	assertHasExpected(t, ctx, ExpectedStreamIndex)
	assertHasExpected(t, ctx, ExpectedStreamTypeLetter)
}

func TestCompletionStreamType(t *testing.T) {
	ctx := ParseForCompletion("a")
	if ctx.CurrentKind != KindStreamType {
		t.Errorf("expected KindStreamType, got %v", ctx.CurrentKind)
	}
}

func TestCompletionStreamTypeWithIndex(t *testing.T) {
	ctx := ParseForCompletion("a:1")
	if ctx.CurrentKind != KindStreamIndex {
		t.Errorf("expected KindStreamIndex after a:1, got %v", ctx.CurrentKind)
	}
	if ctx.PartialIdent != "1" {
		t.Errorf("expected PartialIdent '1', got %q", ctx.PartialIdent)
	}
}

func TestCompletionGroupEmpty(t *testing.T) {
	ctx := ParseForCompletion("g:")
	assertHasExpected(t, ctx, ExpectedGroupIndex)
	assertHasExpected(t, ctx, ExpectedGroupID)
}

func TestCompletionMetadataKey(t *testing.T) {
	ctx := ParseForCompletion("m:lang")
	assertHasExpected(t, ctx, ExpectedMetadataKey)
	if ctx.PartialIdent != "lang" {
		t.Errorf("expected PartialIdent 'lang', got %q", ctx.PartialIdent)
	}
}

func TestCompletionMetadataValue(t *testing.T) {
	ctx := ParseForCompletion("m:language:")
	assertHasExpected(t, ctx, ExpectedMetadataValue)
	if !ctx.InMetadataValue {
		t.Error("expected InMetadataValue")
	}
}

func TestCompletionDisposition(t *testing.T) {
	ctx := ParseForCompletion("disp:default")
	assertHasExpected(t, ctx, ExpectedDispositionName)
}

func TestCompletionUsable(t *testing.T) {
	ctx := ParseForCompletion("u")
	if ctx.CurrentKind != KindUsable {
		t.Errorf("expected KindUsable, got %v", ctx.CurrentKind)
	}
}

func TestCompletionStreamID(t *testing.T) {
	ctx := ParseForCompletion("#0x1")
	if ctx.CurrentKind != KindStreamID {
		t.Errorf("expected KindStreamID, got %v", ctx.CurrentKind)
	}
}

func TestCompletionValidForms(t *testing.T) {
	ctx := ParseForCompletion("")
	if len(ctx.ValidForms) == 0 {
		t.Error("expected some valid forms for empty input")
	}
}
