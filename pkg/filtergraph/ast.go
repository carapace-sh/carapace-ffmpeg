package filtergraph

// Filtergraph is the top-level structure: one or more chains separated by ';'.
type Filtergraph struct {
	Chains []*Chain
	Span   Span
}

// Chain is a sequence of filters connected by ',' with optional input/output labels.
type Chain struct {
	InputLabels  []string // [input_label]* before the chain
	Filters      []*Filter
	OutputLabels []string // [output_label]* after the chain
	Span        Span
}

// InputLabel returns the first input label (common case), or empty string.
func (c *Chain) InputLabel() string {
	if len(c.InputLabels) > 0 {
		return c.InputLabels[0]
	}
	return ""
}

// OutputLabel returns the first output label (common case), or empty string.
func (c *Chain) OutputLabel() string {
	if len(c.OutputLabels) > 0 {
		return c.OutputLabels[0]
	}
	return ""
}

// Filter is a single filter instance with name and options.
type Filter struct {
	Name    string
	Options []*FilterOption // positional and/or key=value options
	Span    Span
}

// FilterOption is either a positional value or a key=value pair.
type FilterOption struct {
	Key   string // empty for positional
	Value string
	Span  Span
}

// IsKeyed returns true if this option has a key (key=value form).
func (o *FilterOption) IsKeyed() bool {
	return o.Key != ""
}
