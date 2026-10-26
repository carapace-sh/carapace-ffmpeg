package filtergraph

import (
	"fmt"
	"strings"
)

func Format(fg *Filtergraph) string {
	chainStrs := make([]string, len(fg.Chains))
	for i, c := range fg.Chains {
		chainStrs[i] = formatChain(c)
	}
	return strings.Join(chainStrs, ";")
}

func formatChain(c *Chain) string {
	var sb strings.Builder
	for _, label := range c.InputLabels {
		fmt.Fprintf(&sb, "[%s]", label)
	}
	filterStrs := make([]string, len(c.Filters))
	for i, f := range c.Filters {
		filterStrs[i] = formatFilter(f)
	}
	sb.WriteString(strings.Join(filterStrs, ","))
	for _, label := range c.OutputLabels {
		fmt.Fprintf(&sb, "[%s]", label)
	}
	return sb.String()
}

func formatFilter(f *Filter) string {
	var sb strings.Builder
	sb.WriteString(f.Name)
	if len(f.Options) > 0 {
		sb.WriteString("=")
		optStrs := make([]string, len(f.Options))
		for i, o := range f.Options {
			optStrs[i] = formatOption(o)
		}
		sb.WriteString(strings.Join(optStrs, ":"))
	}
	return sb.String()
}

func formatOption(o *FilterOption) string {
	if o.IsKeyed() {
		return fmt.Sprintf("%s=%s", escapeFilterToken(o.Key), escapeFilterToken(o.Value))
	}
	return escapeFilterToken(o.Value)
}

// filterSpecials are characters that are structurally significant in a
// filtergraph option list. When they appear literally in a key or value
// they must be backslash-escaped so the formatted graph re-parses to the
// same AST.
const filterSpecials = "\\:,;'[]="

func escapeFilterToken(s string) string {
	if !strings.ContainsAny(s, filterSpecials) {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if strings.IndexByte(filterSpecials, ch) >= 0 {
			sb.WriteByte('\\')
		}
		sb.WriteByte(ch)
	}
	return sb.String()
}
