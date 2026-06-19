package ffmpeg

import (
	"regexp"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/style"
)

type CodecOpts struct {
	Attachment bool
	Audio      bool
	Data       bool
	Subtitle   bool
	Video      bool
}

func (o CodecOpts) Default() CodecOpts {
	o.Attachment = true
	o.Audio = true
	o.Data = true
	o.Subtitle = true
	o.Video = true
	return o
}

// ActionCodecs completes codecs
//
//	4gv (4GV (Fourth Generation Vocoder))
//	4xm (4X Movie)
func ActionCodecs(opts CodecOpts) carapace.Action {
	return actionCodecs(opts, nil)
}

// ActionEncodableCodecs completes codecs with encoding support
//
//	amv (AMV Video)
//	anull (Null audio codec)
func ActionEncodableCodecs(opts CodecOpts) carapace.Action {
	return actionCodecs(opts, func(s string) bool {
		return s[1] != 'E'
	})
}

// ActionDecodableCodecs completes codecs with decoding support
//
//	avrn (Avid AVI Codec)
//	avrp (Avid 1:1 10-bit RGB Packer)
func ActionDecodableCodecs(opts CodecOpts) carapace.Action {
	return actionCodecs(opts, func(s string) bool {
		return s[0] != 'D'
	})
}

func actionCodecs(opts CodecOpts, filter func(s string) bool) carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-codecs")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " -------")
		if !ok {
			return carapace.ActionMessage("failed to parse codecs")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^ (?P<decoding>.)(?P<encoding>.)(?P<type>.).{3} (?P<codec>[^ ]+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				if filter != nil && filter(line[1:7]) {
					continue
				}

				switch matches[3] {
				case "A":
					if opts.Audio {
						vals = append(vals, matches[4], matches[5], style.Yellow)
					}
				case "D":
					if opts.Data {
						vals = append(vals, matches[4], matches[5], style.Cyan)
					}
				case "S":
					if opts.Subtitle {
						vals = append(vals, matches[4], matches[5], style.Magenta)
					}
				case "T":
					if opts.Attachment {
						vals = append(vals, matches[4], matches[5], style.Green)
					}
				case "V":
					if opts.Video {
						vals = append(vals, matches[4], matches[5], style.Blue)
					}
				}
			}
		}

		if filter == nil || !filter("D      ") {
			vals = append(vals, "copy", "copy the codec of the input", style.Default)
		}
		return carapace.ActionStyledValuesDescribed(vals...)
	}).Tag("codecs").UidF(Uid("codec"))
}

type DecoderOpts struct {
	Audio    bool
	Subtitle bool
	Video    bool
}

func (o DecoderOpts) Default() DecoderOpts {
	o.Audio = true
	o.Subtitle = true
	o.Video = true
	return o
}

// ActionDecoders completes decoders
//
//	4xm (4X Movie)
//	8bps (QuickTime 8BPS video)
func ActionDecoders(opts DecoderOpts) carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-decoders")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " ------")
		if !ok {
			return carapace.ActionMessage("failed to parse decoders")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^ (?P<type>.).{5} (?P<name>[^ ]+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				switch matches[1] {
				case "A":
					if opts.Audio {
						vals = append(vals, matches[2], matches[3], style.Yellow)
					}
				case "S":
					if opts.Subtitle {
						vals = append(vals, matches[2], matches[3], style.Magenta)
					}
				case "V":
					if opts.Video {
						vals = append(vals, matches[2], matches[3], style.Blue)
					}
				}
			}
		}
		return carapace.ActionStyledValuesDescribed(vals...)
	}).Tag("decoders").UidF(Uid("decoder"))
}

type EncoderOpts struct {
	Audio    bool
	Subtitle bool
	Video    bool
}

func (o EncoderOpts) Default() EncoderOpts {
	o.Audio = true
	o.Subtitle = true
	o.Video = true
	return o
}

// ActionEncoders completes encoders
//
//	ac3 (ATSC A/52A (AC-3))
//	ac3_fixed (ATSC A/52A (AC-3) (codec ac3))
func ActionEncoders(opts EncoderOpts) carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-encoders")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " ------")
		if !ok {
			return carapace.ActionMessage("failed to parse encoders")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^ (?P<type>.).{5} (?P<name>[^ ]+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				switch matches[1] {
				case "A":
					if opts.Audio {
						vals = append(vals, matches[2], matches[3], style.Yellow)
					}
				case "S":
					if opts.Subtitle {
						vals = append(vals, matches[2], matches[3], style.Magenta)
					}
				case "V":
					if opts.Video {
						vals = append(vals, matches[2], matches[3], style.Blue)
					}
				}
			}
		}
		return carapace.ActionStyledValuesDescribed(vals...)
	}).Tag("encoders").UidF(Uid("encoder"))
}

// ActionFormats completes formats
//
//	aax (CRI AAX)
//	ac3 (raw AC-3)
func ActionFormats() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-formats")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " ---")
		if !ok {
			return carapace.ActionMessage("failed to parse formats")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^ (?P<decoding>D?)(?P<muxing>E?)(?P<device>d?) +(?P<names>[^ ]+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				s := style.Default
				switch {
				case matches[1] == "D" && matches[2] == "E":
					s = style.Magenta
				case matches[1] == "D":
					s = style.Blue
				case matches[2] == "E":
					s = style.Yellow
				}
				for _, name := range strings.Split(matches[4], ",") {
					vals = append(vals, name, matches[5], s)
				}
			}
		}
		return carapace.ActionStyledValuesDescribed(vals...)
	}).Tag("formats").UidF(Uid("format"))
}

// ActionPixelFormats completes pixel formats
//
//	0rgb (0rgb)
//	0bgr (0bgr)
func ActionPixelFormats() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-pix_fmts")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), "-----")
		if !ok {
			return carapace.ActionMessage("failed to parse pixel formats")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^.{4}(?P<flags>[^ ]+) +(?P<name>[^ ]+) +(?P<nb_components>\d+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				vals = append(vals, matches[2], strings.TrimSpace(matches[4]))
			}
		}
		return carapace.ActionValuesDescribed(vals...)
	}).Tag("pixel formats").UidF(Uid("pixel-format"))
}

// ActionSampleFormats completes sample formats
//
//	dbl (64-bit double-precision floating-point)
//	dblp (64-bit double-precision floating-point planar)
func ActionSampleFormats() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-sample_fmts")(func(output []byte) carapace.Action {
		lines := strings.Split(string(output), "\n")

		r := regexp.MustCompile(`^(?P<name>[^ ]+) +(?P<depth>\d+)$`)

		vals := make([]string, 0)
		for _, line := range lines[1:] {
			if matches := r.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
				vals = append(vals, matches[1], matches[1]+" ("+matches[2]+"-bit)")
			}
		}
		return carapace.ActionValuesDescribed(vals...)
	}).Tag("sample formats").UidF(Uid("sample-format"))
}

// ActionChannelLayouts completes channel layouts
//
//	mono (FC)
//	stereo (FL+FR)
func ActionChannelLayouts() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-layouts")(func(output []byte) carapace.Action {
		content := string(output)
		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^(?P<name>[^ ]+) +(?P<channels>.+)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "Individual") || strings.HasPrefix(trimmed, "Standard") || strings.HasPrefix(trimmed, "NAME") {
				continue
			}
			if matches := r.FindStringSubmatch(trimmed); matches != nil {
				vals = append(vals, matches[1], matches[2])
			}
		}
		return carapace.ActionValuesDescribed(vals...)
	}).Tag("channel layouts").UidF(Uid("channel-layout"))
}

// ActionFilters completes filters
//
//	acrusher (Reduce audio bit resolution.)
//	acue (Delay filtering to match a cue.)
func ActionFilters() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-filters")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " ------")
		if !ok {
			return carapace.ActionMessage("failed to parse filters")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^ .{2,3} (?P<name>[^ ]+) +[^ ]+ *(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				vals = append(vals, matches[1], matches[2])
			}
		}
		return carapace.ActionValuesDescribed(vals...)
	}).Tag("filters").UidF(Uid("filter"))
}

// ActionVideoSizes completes video size abbreviations
//
//	ntsc
//	pal
func ActionVideoSizes() carapace.Action {
	return carapace.ActionValues(
		"ntsc", "pal", "qntsc", "qpal", "sntsc", "spal",
		"film", "ntsc-film",
		"sqcif", "qcif", "cif", "4cif", "16cif",
		"qqvga", "qvga", "vga", "svga", "xga", "uxga",
		"qxga", "sxga", "qsxga", "hsxga",
		"hd1080", "hd720", "hd480",
		"uhd2160", "uhd4320", "4k", "2k",
	).Tag("video sizes").Uid("ffmpeg", "video-size")
}

// ActionFrameRates completes frame rate abbreviations
//
//	ntsc
//	pal
func ActionFrameRates() carapace.Action {
	return carapace.ActionValues(
		"ntsc", "pal", "qntsc", "qpal", "sntsc", "spal",
		"film", "ntsc-film",
	).Tag("frame rates").Uid("ffmpeg", "frame-rate")
}

// ActionLogLevels completes log levels
//
//	verbose (Same as "info", except more verbose)
//	warning (Show all warnings and errors)
func ActionLogLevels() carapace.Action {
	return carapace.ActionValuesDescribed(
		"quiet", "Show nothing at all; be silent",
		"panic", "Only show fatal errors which could lead the process to crash",
		"fatal", "Only show fatal errors",
		"error", "Show all errors, including ones which can be recovered from",
		"warning", "Show all warnings and errors",
		"info", "Show informative messages during processing",
		"verbose", "Same as \"info\", except more verbose",
		"debug", "Show everything, including debugging information",
		"trace", "",
	).StyleF(style.ForLogLevel).Tag("log levels").Uid("ffmpeg", "log-level")
}

// ActionFPSModes completes fps_mode/vsync values
//
//	cfr (constant frame rate (duplicate/drop frames))
//	vfr (variable frame rate (prevent duplicate timestamps))
func ActionFPSModes() carapace.Action {
	return carapace.ActionValuesDescribed(
		"passthrough", "each frame with its timestamp from demuxer to muxer",
		"cfr", "constant frame rate (duplicate/drop frames)",
		"vfr", "variable frame rate (prevent duplicate timestamps)",
		"auto", "automatically choose between cfr and vfr (default)",
		"drop", "same as passthrough but drop all frames (deprecated)",
	).Tag("fps modes").Uid("ffmpeg", "fps-mode")
}

// ActionCopyTB completes copytb values
//
//	-1 (choose automatically (default))
//	0 (use decoder timebase)
func ActionCopyTB() carapace.Action {
	return carapace.ActionValuesDescribed(
		"-1", "choose automatically (default)",
		"0", "use decoder timebase",
		"1", "use demuxer timebase",
	).Tag("copy timebase").Uid("ffmpeg", "copytb")
}

// ActionAbortOn completes abort_on flag values
//
//	empty_output (abort when no packets were passed to the muxer)
//	empty_output_stream (abort when some output streams are empty)
func ActionAbortOn() carapace.Action {
	return carapace.ActionValuesDescribed(
		"empty_output", "abort when no packets were passed to the muxer",
		"empty_output_stream", "abort when some output streams are empty",
	).Tag("abort on flags").Uid("ffmpeg", "abort-on")
}

// ActionDiscard completes discard values
//
//	none (discard nothing)
//	default (discard useless packets (default))
func ActionDiscard() carapace.Action {
	return carapace.ActionValuesDescribed(
		"none", "discard nothing",
		"default", "discard useless packets (default)",
		"noref", "discard all non-reference frames",
		"bidir", "discard all bidirectional frames",
		"nointra", "discard all non-intra frames",
		"nokey", "discard all frames except keyframes",
		"all", "discard all frames",
	).Tag("discard values").Uid("ffmpeg", "discard")
}

// ActionBitstreamFilters completes bitstream filters
//
//	dca_core
//	dts2pts
func ActionBitstreamFilters() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-bsfs")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), ":")
		if !ok {
			return carapace.ActionMessage("failed to parse bitstream filters")
		}

		lines := strings.Split(content, "\n")
		vals := make([]string, 0)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				vals = append(vals, line)
			}
		}
		return carapace.ActionValues(vals...)
	}).Tag("bitstream filters").UidF(Uid("bitstream-filter"))
}

// ActionPrintGraphsFormats completes print_graphs_format values
//
//	default (human-readable default format)
//	compact (compact format)
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
	).Tag("print graphs formats").Uid("ffmpeg", "print-graphs-format")
}

// ActionTargets completes target file type values
//
//	vcd (Video CD (PAL or NTSC))
//	svcd (Super Video CD)
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
	).Tag("targets").Uid("ffmpeg", "target")
}

// ActionDispositions completes stream disposition names
//
//	default
//	dub
func ActionDispositions() carapace.Action {
	return carapace.ActionValues(
		"default", "dub", "original", "comment", "lyrics", "karaoke",
		"forced", "hearing_impaired", "visual_impaired", "clean_effects",
		"attached_pic", "timed_thumbnails", "non_diegetic", "captions",
		"descriptions", "metadata", "dependent", "still_image", "multilayer",
	).Tag("dispositions").Uid("ffmpeg", "disposition")
}

// ActionBoolean completes boolean value options
//
//	true
//	false
func ActionBoolean() carapace.Action {
	return carapace.ActionValues("true", "false", "1", "0").Tag("booleans").Uid("ffmpeg", "boolean")
}

// ActionBitrates completes common bitrate values
//
//	96k (96 kbit/s)
//	128k (128 kbit/s)
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
	).Tag("bitrates").Uid("ffmpeg", "bitrate")
}

// ActionHWAccels completes hardware acceleration method names
//
//	cuda
//	drm
func ActionHWAccels() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-hwaccels")(func(output []byte) carapace.Action {
		lines := strings.Split(string(output), "\n")

		vals := make([]string, 0)
		for _, line := range lines[1:] {
			if line != "" {
				vals = append(vals, line)
			}
		}
		return carapace.ActionValues(vals...)
	}).Tag("hardware accelerators").UidF(Uid("hwaccel"))
}

// ActionSwsFlags completes sws_flags/scaler algorithm values
//
//	fast_bilinear (Select fast bilinear scaling algorithm)
//	bilinear (Select bilinear scaling algorithm)
func ActionSwsFlags() carapace.Action {
	return carapace.ActionValuesDescribed(
		"fast_bilinear", "Select fast bilinear scaling algorithm",
		"bilinear", "Select bilinear scaling algorithm",
		"bicubic", "Select bicubic scaling algorithm (default)",
		"experimental", "Select experimental scaling algorithm",
		"neighbor", "Select nearest neighbor rescaling algorithm",
		"area", "Select averaging area rescaling algorithm",
		"bicublin", "Select bicubic for luma, bilinear for chroma",
		"gauss", "Select Gaussian rescaling algorithm",
		"sinc", "Select sinc rescaling algorithm",
		"lanczos", "Select Lanczos rescaling algorithm",
		"spline", "Select natural bicubic spline rescaling algorithm",
		"print_info", "Enable printing/debug logging",
		"accurate_rnd", "Enable accurate rounding",
		"full_chroma_int", "Enable full chroma interpolation",
		"full_chroma_inp", "Select full chroma input",
		"bitexact", "Enable bitexact output",
	).Tag("sws flags").Uid("ffmpeg", "sws-flags")
}

type DeviceOpts struct {
	Demuxing bool
	Muxing   bool
}

func (o DeviceOpts) Default() DeviceOpts {
	o.Demuxing = true
	o.Muxing = true
	return o
}

// ActionDemuxers completes demuxers
//
//	aax (CRI AAX)
//	ac3 (raw AC-3)
func ActionDemuxers() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-demuxers")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " ---")
		if !ok {
			return carapace.ActionMessage("failed to parse demuxers")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^.{5}(?P<name>[^ ]+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				vals = append(vals, matches[1], matches[2])
			}
		}
		return carapace.ActionValuesDescribed(vals...)
	}).Tag("demuxers").UidF(Uid("demuxer"))
}

// ActionMuxers completes muxers
//
//	aax (CRI AAX)
//	ac3 (raw AC-3)
func ActionMuxers() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-muxers")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " ---")
		if !ok {
			return carapace.ActionMessage("failed to parse muxers")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^.{5}(?P<name>[^ ]+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				vals = append(vals, matches[1], matches[2])
			}
		}
		return carapace.ActionValuesDescribed(vals...)
	}).Tag("muxers").UidF(Uid("muxer"))
}

// ActionProtocols completes input/output protocols
//
//	concatf
//	crypto
func ActionProtocols() carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-protocols")(func(output []byte) carapace.Action {
		lines := strings.Split(string(output), "\n")

		found := false
		vals := make([]string, 0)
		for _, line := range lines[2 : len(lines)-1] {
			if !found && line == "Output:" {
				found = true
				continue
			}

			switch found {
			case true:
				vals = append(vals, strings.TrimSpace(line), style.Yellow)
			default:
				vals = append(vals, strings.TrimSpace(line), style.Blue)
			}
		}
		return carapace.ActionStyledValues(vals...)
	}).Tag("protocols").UidF(Uid("protocol"))
}

// ActionDevices completes device names for -sinks/-sources
//
//	alsa (ALSA audio output)
//	pulse (Pulse audio output)
func ActionDevices(opts DeviceOpts) carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-devices")(func(output []byte) carapace.Action {
		_, content, ok := strings.Cut(string(output), " ---")
		if !ok {
			return carapace.ActionMessage("failed to parse devices")
		}

		lines := strings.Split(content, "\n")
		r := regexp.MustCompile(`^ (?P<demuxing>.)(?P<muxing>.) (?P<name>[^ ]+) +(?P<description>.*)$`)

		vals := make([]string, 0)
		for _, line := range lines {
			if matches := r.FindStringSubmatch(line); matches != nil {
				demuxing := matches[1] == "D" && opts.Demuxing
				muxing := matches[2] == "E" && opts.Muxing
				var s string
				switch {
				case demuxing && muxing:
					s = style.Magenta
				case demuxing:
					s = style.Blue
				case muxing:
					s = style.Yellow
				default:
					continue
				}
				for _, name := range strings.Split(matches[3], ",") {
					vals = append(vals, name, matches[4], s)
				}
			}
		}
		return carapace.ActionStyledValuesDescribed(vals...)
	}).Tag("devices").UidF(Uid("device"))
}

// ActionHelpTopics completes help topic values for -h
//
//	long (print more options)
//	full (print all options)
//	decoder= (print detailed information about the decoder)
func ActionHelpTopics() carapace.Action {
	return carapace.ActionMultiPartsN("=", 2, func(c carapace.Context) carapace.Action {
		switch len(c.Parts) {
		case 0:
			return carapace.ActionValuesDescribed(
				"long", "Print advanced tool options in addition to the basic tool options.",
				"full", "Print complete list of options, including all format and codec specific options",
				"decoder=", "Print detailed information about the decoder",
				"encoder=", "Print detailed information about the encoder",
				"demuxer=", "Print detailed information about the demuxer",
				"muxer=", "Print detailed information about the muxer",
				"filter=", "Print detailed information about the filter",
				"bsf=", "Print detailed information about the bitstream filter",
				"protocol=", "Print detailed information about the protocol",
			).Tag("help topics").Uid("ffmpeg", "help-topic")
		default:
			switch c.Parts[0] {
			case "decoder":
				return ActionDecoders(DecoderOpts{}.Default())
			case "encoder":
				return ActionEncoders(EncoderOpts{}.Default())
			case "demuxer":
				return ActionDemuxers()
			case "muxer":
				return ActionMuxers()
			case "filter":
				return ActionFilters()
			case "bsf":
				return ActionBitstreamFilters()
			case "protocol":
				return ActionProtocols()
			default:
				return carapace.ActionValues()
			}
		}
	})
}

// ActionShowModes completes show_mode values for ffplay -showmode
//
//	0 (Display video (default))
//	1 (Show audio waveform)
func ActionShowModes() carapace.Action {
	return carapace.ActionValuesDescribed(
		"0", "Display video (default)",
		"1", "Show audio waveform",
		"2", "Show audio frequency data (RDFT)",
		"video", "Display video",
		"waves", "Show audio waveform",
		"rdft", "Show audio frequency data (RDFT)",
	).Tag("show modes").Uid("ffmpeg", "show-mode")
}

// ActionSyncTypes completes sync_type values for ffplay -sync
//
//	audio (Audio clock is master (default))
//	video (Video clock is master)
func ActionSyncTypes() carapace.Action {
	return carapace.ActionValuesDescribed(
		"audio", "Audio clock is master (default)",
		"video", "Video clock is master",
		"ext", "External clock is master",
	).Tag("sync types").Uid("ffmpeg", "sync-type")
}

// ActionProbeOutputFormats completes output format values for ffprobe -of/-print_format
//
//	default (Human-readable key=value format (default))
//	compact (Compact one-line format)
func ActionProbeOutputFormats() carapace.Action {
	return carapace.ActionValuesDescribed(
		"default", "Human-readable key=value format (default)",
		"compact", "Compact one-line format",
		"csv", "CSV format",
		"flat", "Flat key=value with dot-separated paths",
		"ini", "INI-style sections",
		"json", "JSON format",
		"xml", "XML format",
	).Tag("probe output formats").Uid("ffmpeg", "probe-output-format")
}

// ActionDataDumpFormats completes data_dump_format values for ffprobe
//
//	xxd (Hex+ASCII dump)
//	base64 (Base64-encoded)
func ActionDataDumpFormats() carapace.Action {
	return carapace.ActionValuesDescribed(
		"xxd", "Hex+ASCII dump (default)",
		"base64", "Base64-encoded",
	).Tag("data dump formats").Uid("ffmpeg", "data-dump-format")
}

// ActionShowOptionalFields completes show_optional_fields values for ffprobe
//
//	always (Always print, even if invalid)
//	never (Never print invalid fields)
func ActionShowOptionalFields() carapace.Action {
	return carapace.ActionValuesDescribed(
		"always", "Always print, even if invalid",
		"1", "Always print, even if invalid",
		"never", "Never print invalid fields",
		"0", "Never print invalid fields",
		"auto", "Print only if valid (default)",
		"-1", "Print only if valid (default)",
	).Tag("show optional fields").Uid("ffmpeg", "show-optional-fields")
}

// ActionMetadataKeys completes common stream metadata keys
//
//	language (Language of the stream)
//	title (Title of the stream)
func ActionMetadataKeys() carapace.Action {
	return carapace.ActionValuesDescribed(
		"album", "Album name",
		"album_artist", "Album artist",
		"artist", "Artist name",
		"comment", "Comment",
		"composer", "Composer",
		"copyright", "Copyright notice",
		"date", "Date",
		"description", "Description",
		"encoder", "Encoder used",
		"genre", "Genre",
		"language", "Language of the stream",
		"lyrics", "Lyrics",
		"network_name", "Network name (DVB)",
		"provider_name", "Provider name (DVB)",
		"service_name", "Service name (DVB)",
		"sort_album_artist", "Sort album artist",
		"sort_artist", "Sort artist",
		"sort_title", "Sort title",
		"synopsis", "Synopsis",
		"title", "Title of the stream",
		"track", "Track number",
	).Tag("metadata keys").Uid("ffmpeg", "metadata-key").Suffix(":").NoSpace(':')
}
