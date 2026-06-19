package argstream

// ToolProfile defines the option set and behavior for an ff* tool (ffmpeg, ffplay, ffprobe).
type ToolProfile struct {
	Name             string
	HasOutputSection bool
	OptionIndex      map[string]*OptionDef
}

// lookupOption looks up an option by name in the profile's option index.
// Falls back to the default FFmpeg OptionIndex if the profile is nil or has no index.
func (p *ToolProfile) lookupOption(name string) *OptionDef {
	if p != nil && p.OptionIndex != nil {
		return p.OptionIndex[name]
	}
	return OptionIndex[name]
}

// DefaultFFmpegProfile is the profile for the ffmpeg command.
var DefaultFFmpegProfile = &ToolProfile{
	Name:             "ffmpeg",
	HasOutputSection: true,
	OptionIndex:      nil, // set in init()
}

// DefaultFFplayProfile is the profile for the ffplay command.
var DefaultFFplayProfile = &ToolProfile{
	Name:             "ffplay",
	HasOutputSection: false,
	OptionIndex:      nil, // set in init()
}

// DefaultFFprobeProfile is the profile for the ffprobe command.
var DefaultFFprobeProfile = &ToolProfile{
	Name:             "ffprobe",
	HasOutputSection: false,
	OptionIndex:      nil, // set in init()
}

func init() {
	DefaultFFmpegProfile.OptionIndex = OptionIndex
	DefaultFFplayProfile.OptionIndex = buildFFplayOptionIndex()
	DefaultFFprobeProfile.OptionIndex = buildFFprobeOptionIndex()
}
