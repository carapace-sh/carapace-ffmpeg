package ffmpeg

import (
	"strings"

	"github.com/carapace-sh/carapace"
)

// ActionCodecs completes codec names.
// Shells out to ffmpeg -codecs to get the list.
func ActionCodecs() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-codecs")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			if len(line) < 7 {
				continue
			}
			// Format: " D.VSI.S..... codec_name   description"
			// The codec name starts at position 8 (after flags+spaces)
			name := extractCodecName(line)
			if name != "" {
				values = append(values, name)
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionDecoders completes decoder names.
func ActionDecoders() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-decoders")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			name := extractCodecName(line)
			if name != "" {
				values = append(values, name)
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionEncoders completes encoder names.
func ActionEncoders() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-encoders")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			name := extractCodecName(line)
			if name != "" {
				values = append(values, name)
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionFormats completes container format names.
func ActionFormats() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-formats")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			if len(line) < 5 {
				continue
			}
			// Format: " D  format_name   description"
			name := extractFormatName(line)
			if name != "" {
				values = append(values, name)
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionPixelFormats completes pixel format names.
func ActionPixelFormats() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-pix_fmts")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			fields := splitFields(line)
			if len(fields) >= 2 {
				values = append(values, fields[1])
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionSampleFormats completes sample format names.
func ActionSampleFormats() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-sample_fmts")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			fields := splitFields(line)
			if len(fields) >= 1 {
				values = append(values, fields[0])
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionChannelLayouts completes channel layout names.
func ActionChannelLayouts() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-layouts")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			fields := splitFields(line)
			if len(fields) >= 1 {
				values = append(values, fields[0])
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionFilters completes filter names for filtergraph.
func ActionFilters() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-filters")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			if len(line) < 7 {
				continue
			}
			name := extractFilterName(line)
			if name != "" {
				values = append(values, name)
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionVideoSizes completes video size abbreviations.
func ActionVideoSizes() carapace.Action {
	return carapace.ActionValues(
		"ntsc", "pal", "qntsc", "qpal", "sntsc", "spal",
		"film", "ntsc-film",
		"sqcif", "qcif", "cif", "4cif", "16cif",
		"qqvga", "qvga", "vga", "svga", "xga", "uxga",
		"qxga", "sxga", "qsxga", "hsxga",
		"hd1080", "hd720", "hd480",
		"uhd2160", "uhd4320", "4k", "2k",
	)
}

// ActionFrameRates completes frame rate abbreviations.
func ActionFrameRates() carapace.Action {
	return carapace.ActionValues(
		"ntsc", "pal", "qntsc", "qpal", "sntsc", "spal",
		"film", "ntsc-film",
	)
}

// ActionLogLevels completes log level values.
func ActionLogLevels() carapace.Action {
	return carapace.ActionValuesDescribed(
		"quiet", "show nothing",
		"error", "show only errors",
		"warning", "show warnings and errors",
		"info", "show informational messages (default)",
		"verbose", "show verbose messages",
		"debug", "show debug messages",
		"trace", "show all internal messages",
	)
}

// ActionFPSModes completes fps_mode/vsync values.
func ActionFPSModes() carapace.Action {
	return carapace.ActionValuesDescribed(
		"passthrough", "each frame with its timestamp from demuxer to muxer",
		"cfr", "constant frame rate (duplicate/drop frames)",
		"vfr", "variable frame rate (prevent duplicate timestamps)",
		"auto", "automatically choose between cfr and vfr (default)",
		"drop", "same as passthrough but drop all frames (deprecated)",
	)
}

// ActionCopyTB completes copytb values.
func ActionCopyTB() carapace.Action {
	return carapace.ActionValuesDescribed(
		"-1", "choose automatically (default)",
		"0", "use decoder timebase",
		"1", "use demuxer timebase",
	)
}

// ActionAbortOn completes abort_on flag values.
func ActionAbortOn() carapace.Action {
	return carapace.ActionValuesDescribed(
		"empty_output", "abort when no packets were passed to the muxer",
		"empty_output_stream", "abort when some output streams are empty",
	)
}

// ActionDiscard completes discard values.
func ActionDiscard() carapace.Action {
	return carapace.ActionValuesDescribed(
		"none", "discard nothing",
		"default", "discard useless packets (default)",
		"noref", "discard all non-reference frames",
		"bidir", "discard all bidirectional frames",
		"nointra", "discard all non-intra frames",
		"nokey", "discard all frames except keyframes",
		"all", "discard all frames",
	)
}

// ActionBitstreamFilters completes bitstream filter names.
func ActionBitstreamFilters() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-bsfs")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			fields := splitFields(line)
			if len(fields) >= 1 && !strings.Contains(fields[0], "Bitstream") {
				values = append(values, fields[0])
			}
		}
		return carapace.ActionValues(values...)
	})
}

// ActionPrintGraphsFormats completes print_graphs_format values.
func ActionPrintGraphsFormats() carapace.Action {
	return carapace.ActionValuesDescribed(
		"default", "human-readable default format",
		"compact", "compact format",
		"csv", "CSV format",
		"flat", "flat key=value format",
		"ini", "INI format",
		"json", "JSON format",
		"xml", "XML format",
		"mermaid", "Mermaid flowchart format",
		"mermaidhtml", "Mermaid flowchart as HTML",
	)
}

// ActionTargets completes target file type values.
func ActionTargets() carapace.Action {
	return carapace.ActionValuesDescribed(
		"vcd", "Video CD (PAL or NTSC)",
		"svcd", "Super Video CD",
		"dvd", "DVD (PAL or NTSC)",
		"dv", "DV (PAL or NTSC)",
		"dv50", "DV50 (PAL or NTSC)",
		"pal-vcd", "PAL Video CD",
		"pal-svcd", "PAL Super Video CD",
		"pal-dvd", "PAL DVD",
		"ntsc-vcd", "NTSC Video CD",
		"ntsc-svcd", "NTSC Super Video CD",
		"ntsc-dvd", "NTSC DVD",
		"film-vcd", "FILM Video CD",
		"film-dvd", "FILM DVD",
	)
}

// ActionDispositions completes stream disposition names.
func ActionDispositions() carapace.Action {
	return carapace.ActionValues(
		"default", "dub", "original", "comment", "lyrics", "karaoke",
		"forced", "hearing_impaired", "visual_impaired", "clean_effects",
		"attached_pic", "timed_thumbnails", "non_diegetic", "captions",
		"descriptions", "metadata", "dependent", "still_image", "multilayer",
	)
}

// ActionBoolean completes boolean value options.
func ActionBoolean() carapace.Action {
	return carapace.ActionValues("true", "false", "1", "0")
}

// ActionBitrates completes common bitrate values.
func ActionBitrates() carapace.Action {
	return carapace.ActionValuesDescribed(
		"96k", "96 kbit/s",
		"128k", "128 kbit/s",
		"160k", "160 kbit/s",
		"192k", "192 kbit/s",
		"256k", "256 kbit/s",
		"320k", "320 kbit/s",
		"500k", "500 kbit/s",
		"1M", "1 Mbit/s",
		"2M", "2 Mbit/s",
		"5M", "5 Mbit/s",
		"8M", "8 Mbit/s",
		"10M", "10 Mbit/s",
		"15M", "15 Mbit/s",
		"20M", "20 Mbit/s",
		"25M", "25 Mbit/s",
		"50M", "50 Mbit/s",
	)
}

// ActionHWAccels completes hardware acceleration method names.
func ActionHWAccels() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hwaccels")(func(output []byte) carapace.Action {
		lines := splitLines(output)
		var values []string
		for _, line := range lines {
			fields := splitFields(line)
			if len(fields) >= 1 && !strings.Contains(fields[0], "Hardware") && fields[0] != "Type" {
				values = append(values, fields[0])
			}
		}
		return carapace.ActionValues(values...)
	})
}
