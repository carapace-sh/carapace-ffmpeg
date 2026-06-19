package argstream

import (
	"github.com/carapace-sh/carapace/pkg/style"
)

// OptionScope defines where an option applies.
type OptionScope int

const (
	ScopeGlobalOpt    OptionScope = iota
	ScopePerFileOpt               // applies to next input or output
	ScopeInputOnlyOpt             // applies to next input only
	ScopeOutputOnlyOpt            // applies to next output only
	ScopePerStreamOpt             // applies to specific streams with specifier
)

func (s OptionScope) String() string {
	switch s {
	case ScopeGlobalOpt:
		return "Global"
	case ScopePerFileOpt:
		return "PerFile"
	case ScopeInputOnlyOpt:
		return "InputOnly"
	case ScopeOutputOnlyOpt:
		return "OutputOnly"
	case ScopePerStreamOpt:
		return "PerStream"
	}
	return "Unknown"
}

// OptionType defines whether an option takes a value.
type OptionType int

const (
	TypeBoolean  OptionType = iota
	TypeValue              // takes a value argument
)

// ValueType defines the type of value an option expects.
type ValueType string

const (
	ValueString        ValueType = "string"
	ValueInt           ValueType = "int"
	ValueInt64          ValueType = "int64"
	ValueFloat         ValueType = "float"
	ValueDuration      ValueType = "duration"
	ValueTimestamp     ValueType = "timestamp"
	ValueRatio         ValueType = "ratio"
	ValueVideoSize     ValueType = "video_size"
	ValueVideoRate     ValueType = "video_rate"
	ValuePixelFormat   ValueType = "pixel_format"
	ValueSampleFmt     ValueType = "sample_format"
	ValueChannelLayout ValueType = "channel_layout"
	ValueCodec         ValueType = "codec"
	ValueFormat        ValueType = "format"
	ValueBoolean       ValueType = "boolean"
	ValueMapSpec       ValueType = "map_spec"
	ValueFilter        ValueType = "filter"
	ValueMetadata      ValueType = "metadata"
	ValueDisposition   ValueType = "disposition"
	ValueBitrate       ValueType = "bitrate"
	ValueFileURL       ValueType = "file_url"
	ValueHWAccel       ValueType = "hwaccel"
	ValueLogLevel      ValueType = "loglevel"
	ValueFPSMode       ValueType = "fps_mode"
	ValueCopyTB        ValueType = "copytb"
	ValueAbortOn       ValueType = "abort_on"
	ValueDiscard       ValueType = "discard"
	ValueBSF           ValueType = "bsf"
	ValuePrintGraphFmt ValueType = "print_graphs_format"
	ValueTarget        ValueType = "target"
	ValueSwsFlags       ValueType = "sws_flags"
	ValueDevice        ValueType = "device"
)

// OptionDef defines a single ffmpeg option.
type OptionDef struct {
	CanonicalName string      // primary name (e.g. "codec")
	ShortName     string      // short name (e.g. "c")
	Aliases       []string    // other aliases
	Description   string      // short help text
	Scope         OptionScope
	Type          OptionType  // boolean or value-taking
	ValueType     ValueType   // type of value
	AcceptsSpec   bool        // whether stream specifier suffix is valid
	ImplicitSpec  string      // implied stream specifier for aliases (e.g. "v" for vcodec)
}

// OptionIndex maps option names (including aliases) to their definitions.
// Key is the option name WITHOUT the leading dash, e.g. "c" for -c.
var OptionIndex map[string]*OptionDef

func init() {
	OptionIndex = buildOptionIndex()
}

func buildOptionIndex() map[string]*OptionDef {
	options := []*OptionDef{
		// Global options
		{CanonicalName: "y", ShortName: "y", Description: "overwrite output files without asking", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "n", ShortName: "n", Description: "never overwrite output files", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "loglevel", ShortName: "v", Description: "set logging level", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueLogLevel},
		{CanonicalName: "report", ShortName: "report", Description: "generate a report", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "hide_banner", ShortName: "hide_banner", Description: "suppress startup banner", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueBoolean},
		{CanonicalName: "benchmark", ShortName: "benchmark", Description: "show benchmark timing info", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "benchmark_all", ShortName: "benchmark_all", Description: "add timings for each task", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "stats", ShortName: "stats", Description: "print progress report during encoding", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "cpuflags", ShortName: "cpuflags", Description: "force specific CPU flags", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "cpucount", ShortName: "cpucount", Description: "force specific CPU count", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "max_alloc", ShortName: "max_alloc", Description: "set maximum memory allocation", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "filter_complex", ShortName: "filter_complex", Description: "create a complex filtergraph", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFilter},
		{CanonicalName: "filter_complex_threads", ShortName: "filter_complex_threads", Description: "set threads for complex filtergraph", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "filter_complex_script", ShortName: "filter_complex_script", Description: "create a complex filtergraph from file (deprecated)", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFileURL},
		{CanonicalName: "lavfi", ShortName: "lavfi", Description: "create a complex filtergraph (libavfilter)", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFilter},
		{CanonicalName: "filter_threads", ShortName: "filter_threads", Description: "number of non-complex filter threads", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "filter_buffered_frames", ShortName: "filter_buffered_frames", Description: "maximum number of buffered frames in a filter graph", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "auto_conversion_filters", ShortName: "auto_conversion_filters", Description: "enable automatic conversion filters globally", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "print_graphs", ShortName: "print_graphs", Description: "print execution graph data to stderr", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "print_graphs_file", ShortName: "print_graphs_file", Description: "write execution graph data to file", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFileURL},
		{CanonicalName: "print_graphs_format", ShortName: "print_graphs_format", Description: "set the output printing format", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValuePrintGraphFmt},
		{CanonicalName: "progress", ShortName: "progress", Description: "write program-readable progress information", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "stdin", ShortName: "stdin", Description: "enable or disable interaction on standard input", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueBoolean},
		{CanonicalName: "timelimit", ShortName: "timelimit", Description: "set max runtime in seconds (CPU user time)", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "dump", ShortName: "dump", Description: "dump each input packet", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "hex", ShortName: "hex", Description: "when dumping packets, also dump the payload", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "frame_drop_threshold", ShortName: "frame_drop_threshold", Description: "frame drop threshold", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "copyts", ShortName: "copyts", Description: "copy timestamps", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "start_at_zero", ShortName: "start_at_zero", Description: "shift input timestamps to start at 0 when using copyts", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "copytb", ShortName: "copytb", Description: "copy input stream time base when stream copying", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueCopyTB},
		{CanonicalName: "dts_delta_threshold", ShortName: "dts_delta_threshold", Description: "timestamp discontinuity delta threshold", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "dts_error_threshold", ShortName: "dts_error_threshold", Description: "timestamp error delta threshold", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "xerror", ShortName: "xerror", Description: "exit on error", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueBoolean},
		{CanonicalName: "abort_on", ShortName: "abort_on", Description: "abort on the specified condition flags", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueAbortOn},
		{CanonicalName: "stats_period", ShortName: "stats_period", Description: "set the period at which ffmpeg updates stats and -progress output", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "debug_ts", ShortName: "debug_ts", Description: "print timestamp debugging info", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "max_error_rate", ShortName: "max_error_rate", Description: "maximum error rate above which ffmpeg returns an error", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "vstats", ShortName: "vstats", Description: "dump video coding statistics to file", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "vstats_file", ShortName: "vstats_file", Description: "dump video coding statistics to file", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFileURL},
		{CanonicalName: "vstats_version", ShortName: "vstats_version", Description: "version of the vstats format to use", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "sdp_file", ShortName: "sdp_file", Description: "specify a file to print sdp information", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFileURL},
		{CanonicalName: "ignore_unknown", ShortName: "ignore_unknown", Description: "ignore unknown stream types", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "copy_unknown", ShortName: "copy_unknown", Description: "copy unknown stream types", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "recast_media", ShortName: "recast_media", Description: "allow recasting stream type to force a decoder of different media type", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "vsync", ShortName: "vsync", Description: "set video sync method globally (deprecated, use -fps_mode)", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFPSMode},
		{CanonicalName: "init_hw_device", ShortName: "init_hw_device", Description: "initialise hardware device", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "filter_hw_device", ShortName: "filter_hw_device", Description: "set hardware device used when filtering", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "vaapi_device", ShortName: "vaapi_device", Description: "set VAAPI hardware device", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "qsv_device", ShortName: "qsv_device", Description: "set QSV hardware device", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},

		// Informational options (print info and exit)
		{CanonicalName: "h", ShortName: "h", Aliases: []string{"help", "?"}, Description: "show help", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "version", ShortName: "version", Description: "show version", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "L", ShortName: "L", Aliases: []string{"license"}, Description: "show license", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "buildconf", ShortName: "buildconf", Description: "show build configuration", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "formats", ShortName: "formats", Description: "show available formats", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "codecs", ShortName: "codecs", Description: "show available codecs", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "decoders", ShortName: "decoders", Description: "show available decoders", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "encoders", ShortName: "encoders", Description: "show available encoders", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "bsfs", ShortName: "bsfs", Description: "show available bitstream filters", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "protocols", ShortName: "protocols", Description: "show available protocols", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "filters", ShortName: "filters", Description: "show available filters", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "pix_fmts", ShortName: "pix_fmts", Description: "show available pixel formats", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "layouts", ShortName: "layouts", Description: "show standard channel layouts", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "sample_fmts", ShortName: "sample_fmts", Description: "show available audio sample formats", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "colors", ShortName: "colors", Description: "show available color names", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "dispositions", ShortName: "dispositions", Description: "show available stream dispositions", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "hwaccels", ShortName: "hwaccels", Description: "show available HW acceleration methods", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "devices", ShortName: "devices", Description: "show available devices", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "muxers", ShortName: "muxers", Description: "show available muxers", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "demuxers", ShortName: "demuxers", Description: "show available demuxers", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "sources", ShortName: "sources", Description: "list sources of the input device", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueDevice},
		{CanonicalName: "sinks", ShortName: "sinks", Description: "list sinks of the output device", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueDevice},

		// Per-file options (input + output)
		{CanonicalName: "f", ShortName: "f", Description: "force container format", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueFormat},
		{CanonicalName: "t", ShortName: "t", Description: "stop transcoding after specified duration", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "to", ShortName: "to", Description: "stop transcoding after specified time is reached", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "ss", ShortName: "ss", Description: "start transcoding at specified time", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "bitexact", ShortName: "bitexact", Description: "bitexact mode", Scope: ScopePerFileOpt, Type: TypeBoolean},
		{CanonicalName: "thread_queue_size", ShortName: "thread_queue_size", Description: "set the maximum number of queued packets from the demuxer", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "probesize", ShortName: "probesize", Description: "set probing size in bytes", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueInt64},
		{CanonicalName: "analyzeduration", ShortName: "analyzeduration", Description: "set analysis duration", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "sws_flags", ShortName: "sws_flags", Description: "set default flags for the libswscale library", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueSwsFlags},

		// Input-only options
		{CanonicalName: "i", ShortName: "i", Description: "input file", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueFileURL},
		{CanonicalName: "sseof", ShortName: "sseof", Description: "set the start time offset relative to EOF", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "seek_timestamp", ShortName: "seek_timestamp", Description: "enable/disable seeking by timestamp with -ss", Scope: ScopeInputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "accurate_seek", ShortName: "accurate_seek", Description: "enable/disable accurate seeking with -ss", Scope: ScopeInputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "isync", ShortName: "isync", Description: "indicate the input index for sync reference", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "itsoffset", ShortName: "itsoffset", Description: "set the input ts offset", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "re", ShortName: "re", Description: "read input at native frame rate (equivalent to -readrate 1)", Scope: ScopeInputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "readrate", ShortName: "readrate", Description: "read input at specified rate", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "readrate_initial_burst", ShortName: "readrate_initial_burst", Description: "initial amount of input to burst read before imposing readrate", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "readrate_catchup", ShortName: "readrate_catchup", Description: "temporary readrate used to catch up if an input lags behind", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "dump_attachment", ShortName: "dump_attachment", Description: "extract an attachment into a file", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueFileURL, AcceptsSpec: true},
		{CanonicalName: "stream_loop", ShortName: "stream_loop", Description: "set number of times input stream shall be looped", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "find_stream_info", ShortName: "find_stream_info", Description: "read and decode the streams to fill missing information", Scope: ScopeInputOnlyOpt, Type: TypeBoolean},

		// Output-only options
		{CanonicalName: "map", ShortName: "map", Description: "set input stream mapping", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueMapSpec},
		{CanonicalName: "map_metadata", ShortName: "map_metadata", Description: "set metadata information of outfile from infile", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "map_chapters", ShortName: "map_chapters", Description: "set chapters mapping", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "fs", ShortName: "fs", Description: "set the limit file size in bytes", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueInt64},
		{CanonicalName: "timestamp", ShortName: "timestamp", Description: "set the recording timestamp", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueTimestamp},
		{CanonicalName: "metadata", ShortName: "metadata", Description: "add metadata", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueMetadata, AcceptsSpec: true},
		{CanonicalName: "program", ShortName: "program", Description: "add program with specified streams", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "stream_group", ShortName: "stream_group", Description: "add stream group with specified streams", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "dframes", ShortName: "dframes", Description: "set the number of data frames to output", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueInt64},
		{CanonicalName: "target", ShortName: "target", Description: "specify target file type", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueTarget},
		{CanonicalName: "shortest", ShortName: "shortest", Description: "finish encoding within shortest input", Scope: ScopeOutputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "shortest_buf_duration", ShortName: "shortest_buf_duration", Description: "maximum buffering duration for -shortest option", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "dec", ShortName: "dec", Description: "force decoder", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueCodec},
		{CanonicalName: "attach", ShortName: "attach", Description: "add an attachment to the output file", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueFileURL},
		{CanonicalName: "muxdelay", ShortName: "muxdelay", Description: "set the maximum demux-decode delay", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "muxpreload", ShortName: "muxpreload", Description: "set the initial demux-decode delay", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "fpre", ShortName: "fpre", Description: "set options from indicated preset file", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueFileURL},

		// Per-stream options
		{CanonicalName: "codec", ShortName: "c", Aliases: []string{"vcodec", "acodec", "scodec", "dcodec"}, Description: "codec name", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueCodec, AcceptsSpec: true},
		{CanonicalName: "b", ShortName: "b", Aliases: []string{"ab"}, Description: "bitrate (bits/s)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "filter", ShortName: "filter", Description: "stream filter graph", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFilter, AcceptsSpec: true, Aliases: []string{"vf", "af"}},
		{CanonicalName: "frames", ShortName: "frames", Description: "set the number of frames to output", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "bsf", ShortName: "bsf", Description: "bitstream filter", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBSF, AcceptsSpec: true},
		{CanonicalName: "disposition", ShortName: "disposition", Description: "stream disposition", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueDisposition, AcceptsSpec: true},
		{CanonicalName: "tag", ShortName: "tag", Description: "force codec tag/fourcc", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "qscale", ShortName: "q", Description: "use fixed quality scale (VBR)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "pre", ShortName: "pre", Description: "preset name", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "itsscale", ShortName: "itsscale", Description: "set the input ts scale", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "copyinkf", ShortName: "copyinkf", Description: "copy initial non-keyframes", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "copypriorss", ShortName: "copypriorss", Description: "copy or discard frames before start time", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "filter_script", ShortName: "filter_script", Description: "apply filter from file (deprecated, use -/filter)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFileURL, AcceptsSpec: true},
		{CanonicalName: "reinit_filter", ShortName: "reinit_filter", Description: "reinit filtergraph on input parameter changes", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBoolean, AcceptsSpec: true},
		{CanonicalName: "drop_changed", ShortName: "drop_changed", Description: "drop frame instead of reiniting filtergraph on input parameter changes", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBoolean, AcceptsSpec: true},
		{CanonicalName: "discard", ShortName: "discard", Description: "discard", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueDiscard, AcceptsSpec: true},
		{CanonicalName: "bits_per_raw_sample", ShortName: "bits_per_raw_sample", Description: "set the number of bits per raw sample", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "stats_enc_pre", ShortName: "stats_enc_pre", Description: "write encoding stats before encoding", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "stats_enc_post", ShortName: "stats_enc_post", Description: "write encoding stats after encoding", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "stats_mux_pre", ShortName: "stats_mux_pre", Description: "write packets stats before muxing", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "stats_enc_pre_fmt", ShortName: "stats_enc_pre_fmt", Description: "format of the stats written with -stats_enc_pre", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "stats_enc_post_fmt", ShortName: "stats_enc_post_fmt", Description: "format of the stats written with -stats_enc_post", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "stats_mux_pre_fmt", ShortName: "stats_mux_pre_fmt", Description: "format of the stats written with -stats_mux_pre", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "time_base", ShortName: "time_base", Description: "set the desired time base hint for output stream", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueRatio, AcceptsSpec: true},
		{CanonicalName: "enc_time_base", ShortName: "enc_time_base", Description: "set the desired time base for the encoder", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueRatio, AcceptsSpec: true},
		{CanonicalName: "max_muxing_queue_size", ShortName: "max_muxing_queue_size", Description: "maximum number of packets that can be buffered while waiting for all streams to initialize", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "muxing_queue_data_threshold", ShortName: "muxing_queue_data_threshold", Description: "set the threshold after which max_muxing_queue_size is taken into account", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "profile", ShortName: "profile", Description: "set codec profile", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "level", ShortName: "level", Description: "set codec level", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "threads", ShortName: "threads", Description: "number of encoding/decoding threads", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},

		// Video per-stream options
		{CanonicalName: "r", ShortName: "r", Description: "override input framerate/convert to given output framerate (Hz)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoRate, AcceptsSpec: true},
		{CanonicalName: "aspect", ShortName: "aspect", Description: "set aspect ratio (4:3, 16:9 or 1.3333, 1.7777)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueRatio, AcceptsSpec: true},
		{CanonicalName: "s", ShortName: "s", Description: "frame size (WxH or abbreviation)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoSize, AcceptsSpec: true},
		{CanonicalName: "pix_fmt", ShortName: "pix_fmt", Description: "set pixel format", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValuePixelFormat, AcceptsSpec: true},
		{CanonicalName: "fpsmax", ShortName: "fpsmax", Description: "set max frame rate (Hz)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoRate, AcceptsSpec: true},
		{CanonicalName: "g", ShortName: "g", Description: "set the group of picture (GOP) size", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "maxrate", ShortName: "maxrate", Description: "maximum bitrate (bits/s)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "minrate", ShortName: "minrate", Description: "minimum bitrate (bits/s)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "bufsize", ShortName: "bufsize", Description: "set ratecontrol buffer size", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "pass", ShortName: "pass", Description: "select the pass number (1 to 3)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "passlogfile", ShortName: "passlogfile", Description: "select two pass log file name prefix", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFileURL, AcceptsSpec: true},
		{CanonicalName: "vframes", ShortName: "vframes", Description: "set the number of video frames to output", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, ImplicitSpec: "v"},
		{CanonicalName: "display_rotation", ShortName: "display_rotation", Description: "set pure counter-clockwise rotation in degrees for stream(s)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "display_hflip", ShortName: "display_hflip", Description: "set display horizontal flip for stream(s)", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "display_vflip", ShortName: "display_vflip", Description: "set display vertical flip for stream(s)", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "rc_override", ShortName: "rc_override", Description: "rate control override for specific intervals", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "timecode", ShortName: "timecode", Description: "set initial TimeCode value", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "v"},
		{CanonicalName: "intra_matrix", ShortName: "intra_matrix", Description: "specify intra matrix coeffs", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "inter_matrix", ShortName: "inter_matrix", Description: "specify inter matrix coeffs", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "chroma_intra_matrix", ShortName: "chroma_intra_matrix", Description: "specify intra matrix coeffs", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "vtag", ShortName: "vtag", Description: "force video tag/fourcc", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "v"},
		{CanonicalName: "fps_mode", ShortName: "fps_mode", Description: "set framerate mode for matching video streams; overrides vsync", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFPSMode, AcceptsSpec: true},
		{CanonicalName: "force_fps", ShortName: "force_fps", Description: "force the selected framerate, disable the best supported framerate selection", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "streamid", ShortName: "streamid", Description: "set the value of an outfile streamid", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "v"},
		{CanonicalName: "force_key_frames", ShortName: "force_key_frames", Description: "force key frames at specified timestamps", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "hwaccel", ShortName: "hwaccel", Description: "use HW accelerated decoding", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueHWAccel, AcceptsSpec: true},
		{CanonicalName: "hwaccel_device", ShortName: "hwaccel_device", Description: "select a device for HW acceleration", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "hwaccel_output_format", ShortName: "hwaccel_output_format", Description: "select output format used with HW accelerated decoding", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "autorotate", ShortName: "autorotate", Description: "automatically insert correct rotate filters", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "autoscale", ShortName: "autoscale", Description: "automatically insert a scale filter at the end of the filter graph", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "apply_cropping", ShortName: "apply_cropping", Description: "select the cropping to apply", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBoolean, AcceptsSpec: true},
		{CanonicalName: "fix_sub_duration_heartbeat", ShortName: "fix_sub_duration_heartbeat", Description: "set this video output stream to be a heartbeat stream for fix_sub_duration", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "vpre", ShortName: "vpre", Description: "set the video options to the indicated preset", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "v"},
		{CanonicalName: "deinterlace", ShortName: "deinterlace", Description: "deinterlace pictures", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "top", ShortName: "top", Description: "deprecated, use the setfield video filter", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBoolean, AcceptsSpec: true},

		// Audio per-stream options
		{CanonicalName: "ar", ShortName: "ar", Description: "set audio sampling rate (in Hz)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "ac", ShortName: "ac", Description: "set number of audio channels", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "aq", ShortName: "aq", Description: "set audio quality (codec-specific)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, ImplicitSpec: "a"},
		{CanonicalName: "aframes", ShortName: "aframes", Description: "set the number of audio frames to output", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, ImplicitSpec: "a"},
		{CanonicalName: "apad", ShortName: "apad", Description: "audio pad", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "atag", ShortName: "atag", Description: "force audio tag/fourcc", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "a"},
		{CanonicalName: "sample_fmt", ShortName: "sample_fmt", Description: "set sample format", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueSampleFmt, AcceptsSpec: true},
		{CanonicalName: "channel_layout", ShortName: "channel_layout", Description: "set channel layout", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueChannelLayout, AcceptsSpec: true},
		{CanonicalName: "ch_layout", ShortName: "ch_layout", Description: "set channel layout", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueChannelLayout, AcceptsSpec: true},
		{CanonicalName: "guess_layout_max", ShortName: "guess_layout_max", Description: "set the maximum number of channels to try to guess the channel layout", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "apre", ShortName: "apre", Description: "set the audio options to the indicated preset", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "a"},

		// Subtitle per-stream options
		{CanonicalName: "stag", ShortName: "stag", Description: "force subtitle tag/fourcc", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "s"},
		{CanonicalName: "fix_sub_duration", ShortName: "fix_sub_duration", Description: "fix subtitles duration", Scope: ScopePerStreamOpt, Type: TypeBoolean, AcceptsSpec: true},
		{CanonicalName: "canvas_size", ShortName: "canvas_size", Description: "set canvas size (WxH or abbreviation)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoSize, AcceptsSpec: true},
		{CanonicalName: "spre", ShortName: "spre", Description: "set the subtitle options to the indicated preset", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, ImplicitSpec: "s"},

		// Boolean stream flags (per-stream, no value)
		{CanonicalName: "vn", ShortName: "vn", Description: "disable video", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "an", ShortName: "an", Description: "disable audio", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "sn", ShortName: "sn", Description: "disable subtitle", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "dn", ShortName: "dn", Description: "disable data", Scope: ScopePerStreamOpt, Type: TypeBoolean},
	}

	index := make(map[string]*OptionDef)
	for _, opt := range options {
		// Register by canonical name
		index[opt.CanonicalName] = opt
		// Register by short name (if different)
		if opt.ShortName != opt.CanonicalName {
			index[opt.ShortName] = opt
		}
		// Register aliases with implicit spec where applicable
		for _, alias := range opt.Aliases {
			aliasOpt := *opt
			switch alias {
			case "vcodec":
				aliasOpt.ImplicitSpec = "v"
			case "acodec":
				aliasOpt.ImplicitSpec = "a"
			case "scodec":
				aliasOpt.ImplicitSpec = "s"
			case "dcodec":
				aliasOpt.ImplicitSpec = "d"
			case "vf":
				aliasOpt.ImplicitSpec = "v"
			case "af":
				aliasOpt.ImplicitSpec = "a"
			case "ab":
				aliasOpt.ImplicitSpec = "a"
			}
			index[alias] = &aliasOpt
		}
	}

	return index
}

// Style returns the carapace style string for this option, matching the
// flag styling conventions used by carapace:
//   - FlagArg (blue): option takes a required value argument
//   - FlagNoArg (default): boolean option, no argument
func (o *OptionDef) Style() string {
	switch o.Type {
	case TypeValue:
		return style.Carapace.FlagArg
	default:
		return style.Carapace.FlagNoArg
	}
}

// LookupOption looks up an option by name (without leading dash).
// For options with stream specifiers (e.g. "c:v:1"), the name is the base
// before the first colon.
func LookupOption(name string) *OptionDef {
	if opt, ok := OptionIndex[name]; ok {
		return opt
	}
	return nil
}

// ParseOptionName splits a raw option token (e.g. "c:v:1") into the
// base option name and the stream specifier.
// hasColon reports whether a colon was present in the token,
// distinguishing "-c" (no colon, value expected) from "-c:" (colon, spec expected).
func ParseOptionName(raw string) (name string, specifier string, hasColon bool) {
	for i, ch := range raw {
		if ch == ':' {
			return raw[:i], raw[i+1:], true
		}
	}
	return raw, "", false
}