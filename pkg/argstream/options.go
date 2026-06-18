package argstream

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
)

// OptionDef defines a single ffmpeg option.
type OptionDef struct {
	CanonicalName string      // primary name (e.g. "codec")
	ShortName     string      // short name (e.g. "c")
	Aliases       []string    // other aliases
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
		{CanonicalName: "y", ShortName: "y", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "n", ShortName: "n", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "loglevel", ShortName: "v", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "report", ShortName: "report", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "hide_banner", ShortName: "hide_banner", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "benchmark", ShortName: "benchmark", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "stats", ShortName: "stats", Scope: ScopeGlobalOpt, Type: TypeBoolean},
		{CanonicalName: "cpuflags", ShortName: "cpuflags", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "max_alloc", ShortName: "max_alloc", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "filter_complex", ShortName: "filter_complex", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFilter},
		{CanonicalName: "filter_complex_threads", ShortName: "filter_complex_threads", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "lavfi", ShortName: "lavfi", Scope: ScopeGlobalOpt, Type: TypeValue, ValueType: ValueFilter},

		// Per-file options (input + output)
		{CanonicalName: "f", ShortName: "f", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueFormat},
		{CanonicalName: "t", ShortName: "t", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "ss", ShortName: "ss", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "bitexact", ShortName: "bitexact", Scope: ScopePerFileOpt, Type: TypeBoolean},

		// Input-only options
		{CanonicalName: "sseof", ShortName: "sseof", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "re", ShortName: "re", Scope: ScopeInputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "readrate", ShortName: "readrate", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueFloat},
		{CanonicalName: "isync", ShortName: "isync", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "stream_loop", ShortName: "stream_loop", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "i", ShortName: "i", Scope: ScopeInputOnlyOpt, Type: TypeValue, ValueType: ValueString},

		// Output-only options
		{CanonicalName: "to", ShortName: "to", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "map", ShortName: "map", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueMapSpec},
		{CanonicalName: "map_metadata", ShortName: "map_metadata", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "map_chapters", ShortName: "map_chapters", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueInt},
		{CanonicalName: "shortest", ShortName: "shortest", Scope: ScopeOutputOnlyOpt, Type: TypeBoolean},
		{CanonicalName: "fs", ShortName: "fs", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueInt64},
		{CanonicalName: "dec", ShortName: "dec", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString},

		// Per-stream options
		{CanonicalName: "codec", ShortName: "c", Aliases: []string{"vcodec", "acodec", "scodec", "dcodec"}, Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueCodec, AcceptsSpec: true},
		{CanonicalName: "b", ShortName: "b", Aliases: []string{"ab"}, Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "r", ShortName: "r", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoRate, AcceptsSpec: true},
		{CanonicalName: "pix_fmt", ShortName: "pix_fmt", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValuePixelFormat, AcceptsSpec: true},
		{CanonicalName: "ar", ShortName: "ar", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "ac", ShortName: "ac", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "filter", ShortName: "filter", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFilter, AcceptsSpec: true},
		{CanonicalName: "vf", ShortName: "vf", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFilter},
		{CanonicalName: "af", ShortName: "af", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFilter},
		{CanonicalName: "frames", ShortName: "frames", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "bsf", ShortName: "bsf", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "disposition", ShortName: "disposition", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueDisposition, AcceptsSpec: true},
		{CanonicalName: "tag", ShortName: "tag", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "metadata", ShortName: "metadata", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueMetadata, AcceptsSpec: true},
		{CanonicalName: "threads", ShortName: "threads", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "aspect", ShortName: "aspect", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueRatio},
		{CanonicalName: "s", ShortName: "s", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoSize},

		// Boolean stream flags (per-stream, no value)
		{CanonicalName: "vn", ShortName: "vn", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "an", ShortName: "an", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "sn", ShortName: "sn", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "dn", ShortName: "dn", Scope: ScopePerStreamOpt, Type: TypeBoolean},

		// Additional common options
		{CanonicalName: "probesize", ShortName: "probesize", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueInt64},
		{CanonicalName: "analyzeduration", ShortName: "analyzeduration", Scope: ScopePerFileOpt, Type: TypeValue, ValueType: ValueDuration},
		{CanonicalName: "fpsmax", ShortName: "fpsmax", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueVideoRate},
		{CanonicalName: "qscale", ShortName: "q", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "profile", ShortName: "profile", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "level", ShortName: "level", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFloat, AcceptsSpec: true},
		{CanonicalName: "g", ShortName: "g", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "maxrate", ShortName: "maxrate", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "minrate", ShortName: "minrate", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "bufsize", ShortName: "bufsize", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt64, AcceptsSpec: true},
		{CanonicalName: "pass", ShortName: "pass", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueInt, AcceptsSpec: true},
		{CanonicalName: "passlogfile", ShortName: "passlogfile", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueString, AcceptsSpec: true},
		{CanonicalName: "deinterlace", ShortName: "deinterlace", Scope: ScopePerStreamOpt, Type: TypeBoolean},
		{CanonicalName: "vstats_file", ShortName: "vstats_file", Scope: ScopeOutputOnlyOpt, Type: TypeValue, ValueType: ValueString},
		{CanonicalName: "nostdin", ShortName: "nostdin", Scope: ScopeGlobalOpt, Type: TypeBoolean},
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

	// Special derived aliases
	index["vf"] = &OptionDef{CanonicalName: "filter", ShortName: "vf", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFilter, AcceptsSpec: false}
	index["af"] = &OptionDef{CanonicalName: "filter", ShortName: "af", Scope: ScopePerStreamOpt, Type: TypeValue, ValueType: ValueFilter, AcceptsSpec: false}

	return index
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