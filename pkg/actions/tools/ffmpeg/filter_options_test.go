package ffmpeg

import (
	"testing"
)

const scaleFilterHelp = `Filter scale
  Scale the input video size and/or convert the image format.
    Inputs:
       #0: default (video)
        dynamic (depending on the options)
    Outputs:
       #0: default (video)
scale AVOptions:
   w                 <string>     ..FV.....T. Output video width
   width             <string>     ..FV.....T. Output video width
   h                 <string>     ..FV.....T. Output video height
   height            <string>     ..FV.....T. Output video height
   flags             <string>     ..FV....... Flags to pass to libswscale (default "")
   interl            <boolean>    ..FV....... set interlacing (default false)
   in_color_matrix   <int>        ..FV....... set input YCbCr type (from -1 to 17) (default auto)
     auto            -1           ..FV.......
     bt601           5            ..FV.......
     bt709           1            ..FV.......
   in_range          <int>        ..FV....... set input color range (from 0 to 2) (default auto)
     auto            0            ..FV.......
     full            2            ..FV.......
     limited         1            ..FV.......

This filter has support for timeline through the 'enable' option.
`

const overlayFilterHelp = `Filter overlay
  Overlay a video source on top of the input.
    Inputs:
       #0: main (video)
       #1: overlay (video)
    Outputs:
       #0: default (video)
overlay AVOptions:
   x                 <string>     ..FV.....T. set the x expression (default "0")
   eof_action        <int>        ..FV....... Action to take when encountering EOF from secondary input  (from 0 to 2) (default repeat)
     repeat          0            ..FV....... Repeat the previous frame.
     endall          1            ..FV....... End both streams.
     pass            2            ..FV....... Pass through the main input.
   format            <int>        ..FV....... set output format (from 0 to 8) (default yuv420)
     yuv420          0            ..FV....... 
     rgb             6            ..FV....... 
flags AVOptions:
   eof_action        <int>        ..FV....... Action to take when encountering EOF from secondary input  (from 0 to 2) (default repeat)
     repeat          0            ..FV....... Repeat the previous frame.
`

const amixFilterHelp = `Filter amix
  Audio mixing.
    Inputs:
        dynamic (depending on the options)
    Outputs:
       #0: default (audio)
amix AVOptions:
   inputs            <int>        ..F.A...... Number of inputs. (from 1 to 32767) (default 2)
   duration          <int>        ..F.A...... How to determine the end-of-stream. (from 0 to 2) (default longest)
     longest         0            ..F.A...... Duration of longest input.
     shortest        1            ..F.A...... Duration of shortest input.
     first           2            ..F.A...... Duration of first input.
   dropout_transition <float>      ..F.A...... Transition time, in seconds, for volume renormalization when an input stream ends. (from 0 to INT_MAX) (default 2)
   normalize         <boolean>    ..F.A....T. Scale inputs (default true)

Exiting with exit code 0
`

const colorFilterHelp = `Filter color
  Provide an uniformly colored input.
    Inputs:
        none (source filter)
    Outputs:
       #0: default (video)
color AVOptions:
   color             <color>      ..FV.....T. set color (default "black")
   c                 <color>      ..FV.....T. set color (default "black")
   size              <image_size> ..FV....... set video size (default "320x240")

Exiting with exit code 0
`

func TestParseFilterHelpBasic(t *testing.T) {
	options := parseFilterHelp(scaleFilterHelp)
	if len(options) == 0 {
		t.Fatal("expected options to be parsed")
	}

	// Check that "w" option was parsed
	foundW := false
	for _, opt := range options {
		if opt.Name == "w" {
			foundW = true
			if opt.Type != "string" {
				t.Errorf("expected type 'string' for 'w', got %q", opt.Type)
			}
			if opt.Description != "Output video width" {
				t.Errorf("expected description 'Output video width' for 'w', got %q", opt.Description)
			}
		}
	}
	if !foundW {
		t.Error("expected option 'w' to be found")
	}
}

func TestParseFilterHelpBooleanOption(t *testing.T) {
	options := parseFilterHelp(scaleFilterHelp)

	foundInterl := false
	for _, opt := range options {
		if opt.Name == "interl" {
			foundInterl = true
			if opt.Type != "boolean" {
				t.Errorf("expected type 'boolean' for 'interl', got %q", opt.Type)
			}
		}
	}
	if !foundInterl {
		t.Error("expected option 'interl' to be found")
	}
}

func TestParseFilterHelpEnumValues(t *testing.T) {
	options := parseFilterHelp(scaleFilterHelp)

	foundInRange := false
	for _, opt := range options {
		if opt.Name == "in_range" {
			foundInRange = true
			if len(opt.EnumValues) < 3 {
				t.Errorf("expected at least 3 enum values for 'in_range', got %d", len(opt.EnumValues))
			}
			foundAuto := false
			for _, ev := range opt.EnumValues {
				if ev.Name == "auto" {
					foundAuto = true
					if ev.Value != "0" {
						t.Errorf("expected value '0' for enum 'auto', got %q", ev.Value)
					}
				}
			}
			if !foundAuto {
				t.Error("expected enum value 'auto' for 'in_range'")
			}
		}
	}
	if !foundInRange {
		t.Error("expected option 'in_range' to be found")
	}
}

func TestParseFilterHelpMultipleSections(t *testing.T) {
	options := parseFilterHelp(overlayFilterHelp)

	// Should have options from both "overlay AVOptions:" and "flags AVOptions:" sections
	if len(options) == 0 {
		t.Fatal("expected options to be parsed")
	}

	// Check that "x" from overlay section exists
	foundX := false
	foundFlagsEofAction := false
	for _, opt := range options {
		if opt.Name == "x" {
			foundX = true
		}
		if opt.Name == "eof_action" {
			foundFlagsEofAction = true
		}
	}
	if !foundX {
		t.Error("expected option 'x' from overlay AVOptions")
	}
	if !foundFlagsEofAction {
		t.Error("expected option 'eof_action' from flags AVOptions (or overlay)")
	}
}

func TestParseFilterHelpEnumWithFloatType(t *testing.T) {
	options := parseFilterHelp(amixFilterHelp)

	// "dropout_transition" is a <float> without enum values
	foundDropout := false
	for _, opt := range options {
		if opt.Name == "dropout_transition" {
			foundDropout = true
			if opt.Type != "float" {
				t.Errorf("expected type 'float' for 'dropout_transition', got %q", opt.Type)
			}
			if len(opt.EnumValues) != 0 {
				t.Errorf("expected no enum values for 'dropout_transition', got %d", len(opt.EnumValues))
			}
		}
	}
	if !foundDropout {
		t.Error("expected option 'dropout_transition' to be found")
	}
}

func TestParseFilterHelpEnumWithIntType(t *testing.T) {
	options := parseFilterHelp(amixFilterHelp)

	// "duration" is an <int> with enum values
	foundDuration := false
	for _, opt := range options {
		if opt.Name == "duration" {
			foundDuration = true
			if len(opt.EnumValues) < 3 {
				t.Errorf("expected at least 3 enum values for 'duration', got %d", len(opt.EnumValues))
			}
		}
	}
	if !foundDuration {
		t.Error("expected option 'duration' to be found")
	}
}

func TestParseFilterHelpSpecialTypes(t *testing.T) {
	options := parseFilterHelp(colorFilterHelp)

	foundColor := false
	foundSize := false
	for _, opt := range options {
		if opt.Name == "color" {
			foundColor = true
			if opt.Type != "color" {
				t.Errorf("expected type 'color', got %q", opt.Type)
			}
		}
		if opt.Name == "size" {
			foundSize = true
			if opt.Type != "image_size" {
				t.Errorf("expected type 'image_size', got %q", opt.Type)
			}
		}
	}
	if !foundColor {
		t.Error("expected option 'color' to be found")
	}
	if !foundSize {
		t.Error("expected option 'size' to be found")
	}
}

func TestParseFilterHelpEmpty(t *testing.T) {
	options := parseFilterHelp("")
	if len(options) != 0 {
		t.Errorf("expected 0 options for empty input, got %d", len(options))
	}
}

func TestParseFilterHelpUnknownFilter(t *testing.T) {
	options := parseFilterHelp("Unknown filter 'nonexistent'.\n\nExiting with exit code 0\n")
	if len(options) != 0 {
		t.Errorf("expected 0 options for unknown filter, got %d", len(options))
	}
}

func TestParseFilterHelpEnumValueDescription(t *testing.T) {
	options := parseFilterHelp(overlayFilterHelp)

	for _, opt := range options {
		if opt.Name == "eof_action" && len(opt.EnumValues) > 0 {
			foundRepeat := false
			for _, ev := range opt.EnumValues {
				if ev.Name == "repeat" {
					foundRepeat = true
					if ev.Description != "Repeat the previous frame." {
						t.Errorf("expected description for 'repeat', got %q", ev.Description)
					}
				}
			}
			if !foundRepeat {
				t.Error("expected enum value 'repeat' for 'eof_action'")
			}
			return
		}
	}
	t.Error("expected option 'eof_action' with enum values")
}

const swscalerFilterHelp = `Filter scale
  Scale the input video size and/or convert the image format.
scale AVOptions:
   w                 <string>     ..FV.....T. Output video width
SWScaler AVOptions:
  -sws_flags         <flags>      E..V....... swscale flags (default bicubic)
     fast_bilinear                E..V....... fast bilinear
     bicubic                      E..V....... bicubic
  -param0            <double>     E..V....... scaler param 0 (from INT_MIN to INT_MAX) (default 123456)

Exiting with exit code 0
`

func TestParseFilterHelpSecondaryOptions(t *testing.T) {
	options := parseFilterHelp(swscalerFilterHelp)

	// Should have primary option "w" and secondary options "-sws_flags", "-param0"
	foundW := false
	foundSwsFlags := false
	foundParam0 := false
	for _, opt := range options {
		switch opt.Name {
		case "w":
			foundW = true
		case "-sws_flags":
			foundSwsFlags = true
			if opt.Type != "flags" {
				t.Errorf("expected type 'flags' for '-sws_flags', got %q", opt.Type)
			}
			if len(opt.EnumValues) < 2 {
				t.Errorf("expected at least 2 enum values for '-sws_flags', got %d", len(opt.EnumValues))
			}
		case "-param0":
			foundParam0 = true
			if opt.Type != "double" {
				t.Errorf("expected type 'double' for '-param0', got %q", opt.Type)
			}
		}
	}
	if !foundW {
		t.Error("expected primary option 'w'")
	}
	if !foundSwsFlags {
		t.Error("expected secondary option '-sws_flags'")
	}
	if !foundParam0 {
		t.Error("expected secondary option '-param0'")
	}
}
