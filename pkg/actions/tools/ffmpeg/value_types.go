package ffmpeg

import (
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
	return carapace.ActionValues("quiet", "error", "warning", "info", "verbose", "debug", "trace")
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
