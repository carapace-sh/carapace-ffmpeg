package ffmpeg

import (
	"regexp"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/style"
)

// FilterOption describes a single option of an ffmpeg filter.
type FilterOption struct {
	Name        string
	Type        string
	Description string
	EnumValues  []FilterEnumValue
	IsFlags     bool
}

// FilterEnumValue describes a named constant value for a filter option.
type FilterEnumValue struct {
	Name        string
	Value       string
	Description string
}

var (
	// reOptionLine matches primary and secondary filter option lines.
	// Primary:  "   key               <type>        ..FV.....T. description"
	// Secondary: "  -key              <type>        ..FV.....T. description"  (2 spaces + dash prefix)
	reOptionLine = regexp.MustCompile(`^\s{2,3}(-?)(\S+)\s+<(\S+)>\s+\S+\s+(.*)$`)
	// reEnumLineWithNumericValue matches enum constant lines that have a numeric value.
	// "     name        value         FLAGS  description"  (description optional)
	reEnumLineWithNumericValue = regexp.MustCompile(`^\s{5,}(\S+)\s+(-?\d+)\s+\S+(?:\s+(.*))?$`)
	// reEnumLineFlags matches enum constant lines for <flags> type options (no numeric value).
	// "     name        FLAGS  description"  (description optional)
	// FLAGS starts with at least 2 uppercase/dot characters, e.g. "E..V......." or "..FV.....T."
	reEnumLineFlags = regexp.MustCompile(`^\s{5,}(\S+)\s+[A-Z.]{2}\S*(?:\s+(.*))?$`)
	// reSectionHeader matches AVOptions section headers like "scale AVOptions:" or "framesync AVOptions:".
	reSectionHeader = regexp.MustCompile(`^(\S+)\s+AVOptions:\s*$`)
)

// parseFilterHelp parses the output of `ffmpeg -h filter=<name>` into a list
// of FilterOption structs. It handles both primary options (under the filter's
// own AVOptions section) and secondary/inherited options (under other AVOptions
// sections like framesync, SWResampler, etc.).
func parseFilterHelp(output string) []FilterOption {
	var options []FilterOption

	lines := strings.Split(output, "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check for AVOptions section header
		if reSectionHeader.MatchString(line) {
			i++
			// Parse options in this section
			for i < len(lines) {
				opt, consumed, ok := parseOptionLine(lines, i)
				if !ok {
					break
				}
				options = append(options, opt)
				i = consumed
			}
			continue
		}

		// Check for "This filter has support for timeline" or empty end
		if strings.HasPrefix(line, "This filter has") || strings.HasPrefix(line, "Exiting") {
			break
		}

		i++
	}

	return options
}

// parseOptionLine parses an option line and its enum values starting at index i.
func parseOptionLine(lines []string, i int) (FilterOption, int, bool) {
	if i >= len(lines) {
		return FilterOption{}, i, false
	}

	matches := reOptionLine.FindStringSubmatch(lines[i])
	if matches == nil {
		return FilterOption{}, i, false
	}

	opt := FilterOption{
		Name:        matches[2],
		Type:        matches[3],
		Description: strings.TrimSpace(matches[4]),
		IsFlags:     matches[3] == "flags",
	}

	// If secondary option (has - prefix), prefix the name with -
	if matches[1] == "-" {
		opt.Name = "-" + opt.Name
	}

	i++

	// Parse enum values using the appropriate regex for the option type
	if opt.IsFlags {
		i = parseFlagsEnumValues(lines, i, &opt)
	} else {
		i = parseNumericEnumValues(lines, i, &opt)
	}

	return opt, i, true
}

// parseNumericEnumValues parses <int> type enum values (name + numeric value).
func parseNumericEnumValues(lines []string, i int, opt *FilterOption) int {
	for i < len(lines) {
		enumMatches := reEnumLineWithNumericValue.FindStringSubmatch(lines[i])
		if enumMatches == nil {
			break
		}
		opt.EnumValues = append(opt.EnumValues, FilterEnumValue{
			Name:        enumMatches[1],
			Value:       enumMatches[2],
			Description: strings.TrimSpace(enumMatches[3]),
		})
		i++
	}
	return i
}

// parseFlagsEnumValues parses <flags> type enum values (name only, no numeric value).
func parseFlagsEnumValues(lines []string, i int, opt *FilterOption) int {
	for i < len(lines) {
		enumMatches := reEnumLineFlags.FindStringSubmatch(lines[i])
		if enumMatches == nil {
			break
		}
		opt.EnumValues = append(opt.EnumValues, FilterEnumValue{
			Name:        enumMatches[1],
			Description: strings.TrimSpace(enumMatches[2]),
		})
		i++
	}
	return i
}

// ActionFilterOptions completes option key names for the given ffmpeg filter.
// excludeKeys are option keys already set in the current filter instance
// (so they are not offered again).
//
//	scale=w=1280:  (completes remaining scale options like h, flags, etc.)
func ActionFilterOptions(filterName string, excludeKeys []string) carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-h", "filter="+filterName)(func(output []byte) carapace.Action {
		options := parseFilterHelp(string(output))
		if len(options) == 0 {
			return carapace.ActionValues()
		}

		exclude := make(map[string]bool, len(excludeKeys))
		for _, k := range excludeKeys {
			exclude[k] = true
		}

		vals := make([]string, 0)
		for _, opt := range options {
			if exclude[opt.Name] {
				continue
			}
			vals = append(vals, opt.Name, opt.Description, styleForFilterOptionType(opt.Type))
		}
		return carapace.ActionStyledValuesDescribed(vals...).NoSpace('=')
	}).Tag("filter options").UidF(Uid("filter-option", "filter", filterName))
}

// ActionFilterOptionValue completes values for a specific filter option key.
// For enum-type options (<int> with named constants, <boolean>, <flags>),
// it returns the named values. For other types it returns an empty action.
//
//	scale=w=<tab>  (no enum values for w, so returns empty)
//	scale=flags=<tab>  (returns flags enum values)
func ActionFilterOptionValue(filterName string, optionKey string) carapace.Action {
	return carapace.ActionExecCommand("ffmpeg", "-hide_banner", "-h", "filter="+filterName)(func(output []byte) carapace.Action {
		options := parseFilterHelp(string(output))

		for _, opt := range options {
			if opt.Name != optionKey {
				continue
			}

			switch opt.Type {
			case "boolean":
				return carapace.ActionValues("true", "false", "1", "0").
					StyleF(style.ForLogLevel).
					Tag("booleans").
					UidF(Uid("filter-option-value", "filter", filterName, "key", optionKey))
			case "flags":
				if len(opt.EnumValues) == 0 {
					return carapace.ActionValues()
				}
				vals := make([]string, 0)
				for _, ev := range opt.EnumValues {
					vals = append(vals, ev.Name, ev.Description)
				}
				return carapace.ActionValuesDescribed(vals...).
					Tag("filter option values").
					UidF(Uid("filter-option-value", "filter", filterName, "key", optionKey))
			case "int", "int64":
				if len(opt.EnumValues) == 0 {
					return carapace.ActionValues()
				}
				vals := make([]string, 0)
				for _, ev := range opt.EnumValues {
					desc := ev.Description
					if desc == "" {
						desc = ev.Value
					}
					vals = append(vals, ev.Name, desc)
				}
				return carapace.ActionValuesDescribed(vals...).
					Tag("filter option values").
					UidF(Uid("filter-option-value", "filter", filterName, "key", optionKey))
			default:
				// For <string>, <double>, <float>, <color>, <duration>,
				// <video_rate>, <image_size>, <rational>, <sample_fmt>,
				// <channel_layout>, <pix_fmt>, etc. — no enum completions.
				return carapace.ActionValues()
			}
		}

		return carapace.ActionValues()
	}).Tag("filter option values").UidF(Uid("filter-option-value", "filter", filterName, "key", optionKey))
}

// styleForFilterOptionType returns a style color based on the ffmpeg option type.
func styleForFilterOptionType(optType string) string {
	switch optType {
	case "boolean":
		return style.Green
	case "int", "int64", "float", "double":
		return style.Cyan
	case "flags":
		return style.Magenta
	case "string":
		return style.Yellow
	case "color":
		return style.Blue
	case "duration", "video_rate", "image_size", "rational":
		return style.Default
	default:
		return style.Default
	}
}