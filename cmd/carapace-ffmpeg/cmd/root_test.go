package cmd

import (
	"slices"
	"testing"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/argstream"
	"github.com/carapace-sh/carapace-ffmpeg/pkg/completer"
)

func TestContextToArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		value          string
		expectedArgs   []string
		expectedTrail  bool
	}{
		{
			name:          "empty with trailing space",
			args:          []string{""},
			value:         "",
			expectedArgs:  []string{},
			expectedTrail: true,
		},
		{
			name:          "single arg with trailing space",
			args:          []string{"-codec", ""},
			value:         "",
			expectedArgs:  []string{"-codec"},
			expectedTrail: true,
		},
		{
			name:          "option with value and new option mid-token",
			args:          []string{"-codec", "4gv"},
			value:         "-co",
			expectedArgs:  []string{"-codec", "4gv", "-co"},
			expectedTrail: false,
		},
		{
			name:          "option with value and dash mid-token",
			args:          []string{"-codec", "4gv"},
			value:         "-",
			expectedArgs:  []string{"-codec", "4gv", "-"},
			expectedTrail: false,
		},
		{
			name:          "option with value and trailing space",
			args:          []string{"-codec", "4gv", ""},
			value:         "",
			expectedArgs:  []string{"-codec", "4gv"},
			expectedTrail: true,
		},
		{
			name:          "multiple args with trailing space",
			args:          []string{"-i", "input.mp4", "-c:v", "libx264", ""},
			value:         "",
			expectedArgs:  []string{"-i", "input.mp4", "-c:v", "libx264"},
			expectedTrail: true,
		},
		{
			name:          "option with value and new partial option trailing space",
			args:          []string{"-codec", "4gv", "-co"},
			value:         "",
			expectedArgs:  []string{"-codec", "4gv", "-co"},
			expectedTrail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := carapace.Context{Args: tt.args, Value: tt.value}
			args, trailingSpace := completer.ContextToArgs(c)
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected args %v, got %v", tt.expectedArgs, args)
			}
			for i, v := range args {
				if v != tt.expectedArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, v, tt.expectedArgs[i])
				}
			}
			if trailingSpace != tt.expectedTrail {
				t.Errorf("trailingSpace = %v, want %v", trailingSpace, tt.expectedTrail)
			}
		})
	}
}

func TestContextToArgsArgstreamIntegration(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		value           string
		expectedTokens  []argstream.ExpectedToken
		expectedScope   argstream.Scope
		expectedPartial string
	}{
		{
			name:            "option with value then partial option",
			args:            []string{"-codec", "4gv"},
			value:           "-co",
			expectedTokens:  []argstream.ExpectedToken{argstream.ExpectedInputOption},
			expectedScope:   argstream.ScopeGlobal,
			expectedPartial: "co",
		},
		{
			name:            "option with value then dash",
			args:            []string{"-codec", "4gv"},
			value:           "-",
			expectedTokens:  []argstream.ExpectedToken{argstream.ExpectedGlobalOption, argstream.ExpectedInputOption},
			expectedScope:   argstream.ScopeGlobal,
			expectedPartial: "",
		},
		{
			name:            "option with value then trailing space",
			args:            []string{"-codec", "4gv", ""},
			value:           "",
			expectedTokens:  []argstream.ExpectedToken{argstream.ExpectedGlobalOption, argstream.ExpectedInputOption, argstream.ExpectedInputURL},
			expectedScope:   argstream.ScopeGlobal,
			expectedPartial: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := carapace.Context{Args: tt.args, Value: tt.value}
			args, trailingSpace := completer.ContextToArgs(c)
			ctx := argstream.ParseForCompletion(args, trailingSpace)

			if ctx.Scope != tt.expectedScope {
				t.Errorf("scope = %v, want %v", ctx.Scope, tt.expectedScope)
			}
			if ctx.PartialOption != tt.expectedPartial {
				t.Errorf("partialOption = %q, want %q", ctx.PartialOption, tt.expectedPartial)
			}
			for _, token := range tt.expectedTokens {
				if !slices.Contains(ctx.ExpectedTokens, token) {
					t.Errorf("expected token %v in %v", token, ctx.ExpectedTokens)
				}
			}
		})
	}
}