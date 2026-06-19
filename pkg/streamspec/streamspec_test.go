package streamspec

import (
	"testing"
)

func TestParseStreamIndex(t *testing.T) {
	spec, err := Parse("0")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindStreamIndex {
		t.Errorf("expected KindStreamIndex, got %v", spec.Kind)
	}
	if spec.StreamIndex() != 0 {
		t.Errorf("expected index 0, got %d", spec.StreamIndex())
	}
	if Format(spec) != "0" {
		t.Errorf("expected format '0', got %q", Format(spec))
	}
}

func TestParseStreamIndexNonZero(t *testing.T) {
	spec, err := Parse("3")
	if err != nil {
		t.Fatal(err)
	}
	if spec.StreamIndex() != 3 {
		t.Errorf("expected index 3, got %d", spec.StreamIndex())
	}
}

func TestParseStreamType(t *testing.T) {
	spec, err := Parse("a")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindStreamType {
		t.Errorf("expected KindStreamType, got %v", spec.Kind)
	}
	st := spec.StreamType()
	if st.Type != TypeAudio {
		t.Errorf("expected TypeAudio, got %v", st.Type)
	}
}

func TestParseStreamTypeWithIndex(t *testing.T) {
	spec, err := Parse("a:1")
	if err != nil {
		t.Fatal(err)
	}
	st := spec.StreamType()
	if st.Type != TypeAudio {
		t.Errorf("expected TypeAudio, got %v", st.Type)
	}
	if spec.Additional == nil {
		t.Fatal("expected additional specifier")
	}
	if spec.Additional.Kind != KindStreamIndex {
		t.Errorf("expected additional KindStreamIndex, got %v", spec.Additional.Kind)
	}
	if spec.Additional.StreamIndex() != 1 {
		t.Errorf("expected additional index 1, got %d", spec.Additional.StreamIndex())
	}
	if Format(spec) != "a:1" {
		t.Errorf("expected format 'a:1', got %q", Format(spec))
	}
}

func TestParseVideoType(t *testing.T) {
	for _, s := range []string{"v", "V"} {
		spec, err := Parse(s)
		if err != nil {
			t.Fatal(err)
		}
		if spec.Kind != KindStreamType {
			t.Errorf("expected KindStreamType for %q, got %v", s, spec.Kind)
		}
	}
}

func TestParseAllStreamTypes(t *testing.T) {
	type test struct {
		input string
		want  StreamType
	}
	tests := []test{
		{"v", TypeVideo},
		{"V", TypeVideoNoAttached},
		{"a", TypeAudio},
		{"s", TypeSubtitle},
		{"d", TypeData},
		{"t", TypeAttachment},
	}
	for _, tt := range tests {
		spec, err := Parse(tt.input)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.input, err)
		}
		if spec.StreamType().Type != tt.want {
			t.Errorf("Parse(%q): expected %v, got %v", tt.input, tt.want, spec.StreamType().Type)
		}
	}
}

func TestParseGroupByIndex(t *testing.T) {
	spec, err := Parse("g:0")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindGroup {
		t.Errorf("expected KindGroup, got %v", spec.Kind)
	}
	g := spec.Group()
	if g.Kind != GroupByIndex || g.Index != 0 {
		t.Errorf("expected group by index 0, got %v %d", g.Kind, g.Index)
	}
}

func TestParseProgram(t *testing.T) {
	spec, err := Parse("p:0x1")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindProgram {
		t.Errorf("expected KindProgram, got %v", spec.Kind)
	}
	if spec.Program().ID != "0x1" {
		t.Errorf("expected program ID '0x1', got %q", spec.Program().ID)
	}
}

func TestParseStreamID(t *testing.T) {
	spec, err := Parse("#0x1F3")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindStreamID {
		t.Errorf("expected KindStreamID, got %v", spec.Kind)
	}
	sid := spec.StreamID()
	if sid.Alt {
		t.Errorf("expected Alt=false")
	}
	if sid.ID != "0x1F3" {
		t.Errorf("expected ID '0x1F3', got %q", sid.ID)
	}
}

func TestParseStreamIDAlternate(t *testing.T) {
	spec, err := Parse("i:0x1F3")
	if err != nil {
		t.Fatal(err)
	}
	sid := spec.StreamID()
	if !sid.Alt {
		t.Errorf("expected Alt=true")
	}
	if sid.ID != "0x1F3" {
		t.Errorf("expected ID '0x1F3', got %q", sid.ID)
	}
}

func TestParseMetadataKeyOnly(t *testing.T) {
	spec, err := Parse("m:language")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindMetadata {
		t.Errorf("expected KindMetadata, got %v", spec.Kind)
	}
	m := spec.Metadata()
	if m.Key != "language" {
		t.Errorf("expected key 'language', got %q", m.Key)
	}
	if m.Value != "" {
		t.Errorf("expected empty value, got %q", m.Value)
	}
}

func TestParseMetadataKeyAndValue(t *testing.T) {
	spec, err := Parse("m:language:eng")
	if err != nil {
		t.Fatal(err)
	}
	m := spec.Metadata()
	if m.Key != "language" {
		t.Errorf("expected key 'language', got %q", m.Key)
	}
	if m.Value != "eng" {
		t.Errorf("expected value 'eng', got %q", m.Value)
	}
}

func TestParseDisposition(t *testing.T) {
	spec, err := Parse("disp:default")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindDisposition {
		t.Errorf("expected KindDisposition, got %v", spec.Kind)
	}
	d := spec.Disposition()
	if len(d.Dispositions) != 1 || d.Dispositions[0] != "default" {
		t.Errorf("expected ['default'], got %v", d.Dispositions)
	}
}

func TestParseDispositionMultiple(t *testing.T) {
	spec, err := Parse("disp:default+forced")
	if err != nil {
		t.Fatal(err)
	}
	d := spec.Disposition()
	if len(d.Dispositions) != 2 {
		t.Fatalf("expected 2 dispositions, got %d", len(d.Dispositions))
	}
	if d.Dispositions[0] != "default" || d.Dispositions[1] != "forced" {
		t.Errorf("expected ['default','forced'], got %v", d.Dispositions)
	}
}

func TestParseUsable(t *testing.T) {
	spec, err := Parse("u")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Kind != KindUsable {
		t.Errorf("expected KindUsable, got %v", spec.Kind)
	}
}

func TestParseStreamTypeWithAdditionalStreamType(t *testing.T) {
	spec, err := Parse("a:g:0:1")
	if err != nil {
		t.Fatal(err)
	}
	st := spec.StreamType()
	if st.Type != TypeAudio {
		t.Errorf("expected TypeAudio")
	}
	// a:g:0:1 -> spec.Additional is g:0:1
	if spec.Additional == nil {
		t.Fatal("expected additional specifier")
	}
	if spec.Additional.Kind != KindGroup {
		t.Errorf("expected KindGroup, got %v", spec.Additional.Kind)
	}
}

func TestFormatRoundtrip(t *testing.T) {
	tests := []string{
		"0",
		"3",
		"v",
		"V",
		"a",
		"a:1",
		"g:0",
		"p:0x1",
		"#0x1F3",
		"m:language",
		"m:language:eng",
		"disp:default",
		"disp:default+forced",
		"u",
	}
	for _, tt := range tests {
		spec, err := Parse(tt)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt, err)
		}
		got := Format(spec)
		if got != tt {
			t.Errorf("Format(Parse(%q)) = %q, want %q", tt, got, tt)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("x")
	if err == nil {
		t.Error("expected error for invalid specifier 'x'")
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty specifier")
	}
}
