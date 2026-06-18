package argstream

import (
	"testing"
)

func TestParseGlobalOption(t *testing.T) {
	prog, err := Parse([]string{"-y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(prog.Tokens))
	}
	if prog.Tokens[0].Kind != KindGlobalOption {
		t.Errorf("expected KindGlobalOption, got %v", prog.Tokens[0].Kind)
	}
	if prog.Tokens[0].OptionName != "y" {
		t.Errorf("expected option 'y', got %q", prog.Tokens[0].OptionName)
	}
}

func TestParseGlobalOptionWithValue(t *testing.T) {
	prog, err := Parse([]string{"-v", "error"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(prog.Tokens))
	}
	if prog.Tokens[0].Value != "error" {
		t.Errorf("expected value 'error', got %q", prog.Tokens[0].Value)
	}
}

func TestParseInputFile(t *testing.T) {
	prog, err := Parse([]string{"-i", "input.mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.InputFiles) != 1 {
		t.Fatalf("expected 1 input, got %d", len(prog.InputFiles))
	}
	if prog.InputFiles[0].URL != "input.mp4" {
		t.Errorf("expected URL 'input.mp4', got %q", prog.InputFiles[0].URL)
	}
}

func TestParseInputWithOption(t *testing.T) {
	prog, err := Parse([]string{"-f", "mp4", "-i", "input.mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(prog.Tokens))
	}
	if prog.Tokens[0].Kind != KindInputOption {
		t.Errorf("expected KindInputOption, got %v", prog.Tokens[0].Kind)
	}
}

func TestParseOutputStreamOption(t *testing.T) {
	prog, err := Parse([]string{"-i", "input.mp4", "-c:v", "libx264", "output.mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.OutputFiles) != 1 {
		t.Fatalf("expected 1 output, got %d", len(prog.OutputFiles))
	}
	if prog.OutputFiles[0].URL != "output.mp4" {
		t.Errorf("expected URL 'output.mp4', got %q", prog.OutputFiles[0].URL)
	}
}

func TestParseStreamSpecifier(t *testing.T) {
	prog, err := Parse([]string{"-c:v", "libx264"})
	if err != nil {
		t.Fatal(err)
	}
	if prog.Tokens[0].OptionName != "c" {
		t.Errorf("expected option 'c', got %q", prog.Tokens[0].OptionName)
	}
	if prog.Tokens[0].StreamSpecifier != "v" {
		t.Errorf("expected specifier 'v', got %q", prog.Tokens[0].StreamSpecifier)
	}
}

func TestParseBooleanFlags(t *testing.T) {
	prog, err := Parse([]string{"-y", "-hide_banner"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(prog.Tokens))
	}
	if prog.Tokens[0].Value != "" || prog.Tokens[1].Value != "" {
		t.Error("boolean flags should have empty values")
	}
}

func TestParseFullCommand(t *testing.T) {
	prog, err := Parse([]string{"-y", "-i", "input.mp4", "-c:v", "libx264", "-c:a", "aac", "output.mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.InputFiles) != 1 {
		t.Errorf("expected 1 input, got %d", len(prog.InputFiles))
	}
	if len(prog.OutputFiles) != 1 {
		t.Errorf("expected 1 output, got %d", len(prog.OutputFiles))
	}
}

func TestParseMultipleInputs(t *testing.T) {
	prog, err := Parse([]string{"-i", "input1.mp4", "-i", "input2.aac", "output.mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.InputFiles) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(prog.InputFiles))
	}
	if prog.InputFiles[0].URL != "input1.mp4" {
		t.Errorf("expected 'input1.mp4', got %q", prog.InputFiles[0].URL)
	}
	if prog.InputFiles[1].URL != "input2.aac" {
		t.Errorf("expected 'input2.aac', got %q", prog.InputFiles[1].URL)
	}
}

func TestParseFilterComplex(t *testing.T) {
	prog, err := Parse([]string{"-filter_complex", "[0:v]scale=1280:720[out]", "-i", "input.mp4", "output.mp4"})
	if err != nil {
		t.Fatal(err)
	}
	filterToken := prog.Tokens[0]
	if filterToken.Kind != KindGlobalOption {
		t.Errorf("expected KindGlobalOption for filter_complex, got %v", filterToken.Kind)
	}
	if filterToken.OptionName != "filter_complex" {
		t.Errorf("expected option 'filter_complex', got %q", filterToken.OptionName)
	}
}
