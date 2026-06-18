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
	ValueString     ValueType = "string"
	ValueInt        ValueType = "int"
	ValueInt64      ValueType = "int64"
	ValueFloat      ValueType = "float"
	ValueDuration   ValueType = "duration"
	ValueTimestamp   ValueType = "timestamp"
	ValueVideoSize  ValueType = "video_size"
	ValueVideoRate  ValueType = "video_rate"
	ValueRatio      ValueType = "ratio"
	ValuePixelFormat ValueType = "pixel_format"
	ValueSampleFmt  ValueType = "sample_format"
	ValueChannelLayout ValueType = "channel_layout"
	ValueCodec      ValueType = "codec"
	ValueFormat     ValueType = "format"
	ValueBoolean    ValueType = "boolean"
	ValueMapSpec    ValueType = "map_spec"
	ValueFilter     ValueType = "filter"
	ValueMetadata   ValueType = "metadata"
	ValueDisposition ValueType = "disposition"
	ValueBitrate    ValueType = "bitrate"
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
		{CanonicalName: "loglevel", ShortName: "v", Description: "set logging level", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "report", ShortName: "report", Description: "generate a report", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "hide_banner", ShortName: "hide_banner", Description: "suppress startup banner", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "benchmark", ShortName: "benchmark", Description: "show benchmark timing info", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "stats", ShortName: "stats", Description: "print statistics", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "cpuflags", ShortName: "cpuflags", Description: "set CPU flags", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "max_alloc", ShortName: "max_alloc", Description: "set maximum memory allocation", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "filter_complex", ShortName: "filter_complex", Description: "create a complex filtergraph", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFilter},
		{CanonicalName: "filter_complex_threads", ShortName: "filter_complex_threads", Description: "set threads for complex filtergraph", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "lavfi", ShortName: "lavfi", Description: "create a complex filtergraph (libavfilter)", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFilter},

		// Per-file options (input + output)
		{CanonicalName: "f", ShortName: "f", Description: "force container format", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueFormat},
		{CanonicalName: "t", ShortName: "t", Description: "limit duration of read/encode", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "ss", ShortName: "ss", Description: "seek to given time position", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "bitexact", ShortName: "bitexact", Description: "bitexact output", Scope: ScopePerFileOpt, Type: TypeBoolean},

		// Input-only options
		{CanonicalName: "sseof", ShortName: "sseof", Description: "seek relative to end of file", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "re", ShortName: "re", Description: "read input at native frame rate", Scope: ScopeInputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "readrate", ShortName: "readrate", Description: "read input at specified rate", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "isync", ShortName: "isync", Description: "sync input to reference stream", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "stream_loop", ShortName: "stream_loop", Description: "set number of stream loop iterations", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "i", ShortName: "i", Description: "input file", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueString},

		// Output-only options
		{CanonicalName: "to", ShortName: "to", Description: "record or transcode stop time", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "map", ShortName: "map", Description: "set input stream mapping", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueMapSpec},
		{CanonicalName: "map_metadata", ShortName: "map_metadata", Description: "set metadata stream mapping", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "map_chapters", ShortName: "map_chapters", Description: "set chapters input stream mapping", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "shortest", ShortName: "shortest", Description: "finish encoding when shortest input ends", Scope: ScopeOutputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "fs", ShortName: "fs", Description: "set file size limit in bytes", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueInt64},
		{CanonicalName: "dec", ShortName: "dec", Description: "force decoder", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString},

		// Per-stream options
		{CanonicalName: "codec", ShortName: "c", Aliases: []string{"vcodec", "acodec", "scodec", "dcodec"}, Description: "codec name", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueCodec, AcceptsSpec: true},
		{CanonicalName: "b", ShortName: "b", Aliases: []string{"ab"}, Description: "bitrate (bits/s)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "r", ShortName: "r", Description: "frame rate (Hz)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoRate, AcceptsSpec: true},
		{CanonicalName: "pix_fmt", ShortName: "pix_fmt", Description: "pixel format", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValuePixelFormat, AcceptsSpec: true},
		{CanonicalName: "ar", ShortName: "ar", Description: "audio sampling rate (Hz)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "ac", ShortName: "ac", Description: "number of audio channels", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "filter", ShortName: "filter", Description: "stream filter graph", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFilter, AcceptsSpec: true, Aliases: []string{"vf", "af"}},
		{CanonicalName: "frames", ShortName: "frames", Description: "limit number of frames to output", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "bsf", ShortName: "bsf", Description: "bitstream filter", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "disposition", ShortName: "disposition", Description: "stream disposition", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueDisposition, AcceptsSpec: true},
		{CanonicalName: "tag", ShortName: "tag", Description: "stream codec tag", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "metadata", ShortName: "metadata", Description: "stream metadata key=value", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueMetadata, AcceptsSpec: true},
		{CanonicalName: "threads", ShortName: "threads", Description: "number of encoding/decoding threads", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "aspect", ShortName: "aspect", Description: "set video aspect ratio", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueRatio},
		{CanonicalName: "s", ShortName: "s", Description: "frame size (WxH or abbreviation)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoSize},

		// Boolean stream flags (per-stream, no value)
		{CanonicalName: "vn", ShortName: "vn", Description: "disable video", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "an", ShortName: "an", Description: "disable audio", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "sn", ShortName: "sn", Description: "disable subtitle", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "dn", ShortName: "dn", Description: "disable data", Scope: ScopePerStreamOpt, Type: TypeBoolean},

		// Additional common options
		{CanonicalName: "probesize", ShortName: "probesize", Description: "set probing size in bytes", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueInt64},
		{CanonicalName: "analyzeduration", ShortName: "analyzeduration", Description: "set analysis duration", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "fpsmax", ShortName: "fpsmax", Description: "maximum frame rate (Hz)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoRate},
		{CanonicalName: "qscale", ShortName: "q", Description: "use fixed quality scale (VBR)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "profile", ShortName: "profile", Description: "set codec profile", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "level", ShortName: "level", Description: "set codec level", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "g", ShortName: "g", Description: "GOP size (keyframe interval)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "maxrate", ShortName: "maxrate", Description: "maximum bitrate (bits/s)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "minrate", ShortName: "minrate", Description: "minimum bitrate (bits/s)", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "bufsize", ShortName: "bufsize", Description: "set ratecontrol buffer size", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueBitrate, AcceptsSpec: true},
		{CanonicalName: "pass", ShortName: "pass", Description: "select encoding pass number", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "passlogfile", ShortName: "passlogfile", Description: "two-pass log file name prefix", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "deinterlace", ShortName: "deinterlace", Description: "deinterlace pictures", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "vstats_file", ShortName: "vstats_file", Description: "dump video statistics to file", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "nostdin", ShortName: "nostdin", Description: "disable reading from stdin", Scope: ScopeGlobalOpt, Type: TypeBoolean},
	}

	index := make(map[string]*OptionDef)
	for _, opt := range options {
		// Register by canonical name
		index[opt.CanonicalName] = opt
		// Register by short name (if different)
		if opt.ShortName != opt.CanonicalName {
			index[opt.ShortName] = opt
		}
		// Register aliases
		for _, alias := range opt.Aliases {
			index[alias] = opt
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
func ParseOptionName(raw string) (name string, specifier string) {
	for i, ch := range raw {
		if ch == ':' {
			return raw[:i], raw[i+1:]
		}
	}
	return raw, ""
}