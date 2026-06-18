package filtergraph

import (
	"testing"
)

func TestParseSimpleFilter(t *testing.T) {
	fg, err := Parse("yadif")
	if err != nil {
		t.Fatal(err)
	}
	if len(fg.Chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(fg.Chains))
	}
	if len(fg.Chains[0].Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(fg.Chains[0].Filters))
	}
	if fg.Chains[0].Filters[0].Name != "yadif" {
		t.Errorf("expected filter name 'yadif', got %q", fg.Chains[0].Filters[0].Name)
	}
}

func TestParseFilterWithOptions(t *testing.T) {
	fg, err := Parse("scale=1280:720")
	if err != nil {
		t.Fatal(err)
	}
	f := fg.Chains[0].Filters[0]
	if f.Name != "scale" {
		t.Errorf("expected 'scale', got %q", f.Name)
	}
	if len(f.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(f.Options))
	}
	if f.Options[0].Value != "1280" {
		t.Errorf("expected option value '1280', got %q", f.Options[0].Value)
	}
	if f.Options[1].Value != "720" {
		t.Errorf("expected option value '720', got %q", f.Options[1].Value)
	}
}

func TestParseFilterWithKeyValues(t *testing.T) {
	fg, err := Parse("scale=w=1280:h=720")
	if err != nil {
		t.Fatal(err)
	}
	f := fg.Chains[0].Filters[0]
	if len(f.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(f.Options))
	}
	if !f.Options[0].IsKeyed() || f.Options[0].Key != "w" || f.Options[0].Value != "1280" {
		t.Errorf("expected key='w' value='1280', got key=%q value=%q", f.Options[0].Key, f.Options[0].Value)
	}
	if !f.Options[1].IsKeyed() || f.Options[1].Key != "h" || f.Options[1].Value != "720" {
		t.Errorf("expected key='h' value='720', got key=%q value=%q", f.Options[1].Key, f.Options[1].Value)
	}
}

func TestParseChainedFilters(t *testing.T) {
	fg, err := Parse("yadif,scale=1280:720,format=yuv420p")
	if err != nil {
		t.Fatal(err)
	}
	if len(fg.Chains[0].Filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(fg.Chains[0].Filters))
	}
}

func TestParseMultipleChains(t *testing.T) {
	fg, err := Parse("yadif;scale=1280:720")
	if err != nil {
		t.Fatal(err)
	}
	if len(fg.Chains) != 2 {
		t.Fatalf("expected 2 chains, got %d", len(fg.Chains))
	}
}

func TestParseWithLabels(t *testing.T) {
	fg, err := Parse("[0:v]scale=1280:720[out]")
	if err != nil {
		t.Fatal(err)
	}
	c := fg.Chains[0]
	if c.InputLabel() != "0:v" {
		t.Errorf("expected input label '0:v', got %q", c.InputLabel())
	}
	if c.OutputLabel() != "out" {
		t.Errorf("expected output label 'out', got %q", c.OutputLabel())
	}
}

func TestParseComplexFiltergraph(t *testing.T) {
	fg, err := Parse("[0:v][1:v]overlay[out]")
	if err != nil {
		t.Fatal(err)
	}
	c := fg.Chains[0]
	if c.InputLabel() != "0:v" {
		t.Errorf("expected input label '0:v', got %q", c.InputLabel())
	}
	// overlay is the filter, [out] is the output label
	if len(c.Filters) != 1 || c.Filters[0].Name != "overlay" {
		t.Errorf("expected filter 'overlay'")
	}
	if c.OutputLabel() != "out" {
		t.Errorf("expected output label 'out', got %q", c.OutputLabel())
	}
}

func TestFormatRoundtrip(t *testing.T) {
	tests := []string{
		"yadif",
		"scale=1280:720",
		"scale=w=1280:h=720",
		"yadif,scale=1280:720",
		"[0:v]scale=1280:720[out]",
		"yadif;scale=1280:720",
	}
	for _, tt := range tests {
		fg, err := Parse(tt)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt, err)
		}
		got := Format(fg)
		if got != tt {
			t.Errorf("Format(Parse(%q)) = %q, want %q", tt, got, tt)
		}
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty filtergraph")
	}
}

func TestParseMixedOptions(t *testing.T) {
	fg, err := Parse("scale=1280:720:force_original_aspect_ratio=decrease")
	if err != nil {
		t.Fatal(err)
	}
	f := fg.Chains[0].Filters[0]
	if len(f.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(f.Options))
	}
	if f.Options[0].IsKeyed() {
		t.Errorf("expected positional, got key=%q", f.Options[0].Key)
	}
	if !f.Options[2].IsKeyed() || f.Options[2].Key != "force_original_aspect_ratio" {
		t.Errorf("expected keyed option, got key=%q", f.Options[2].Key)
	}
}
