package filtergraph

type ExpectedToken int

const (
	ExpectedFilterName ExpectedToken = iota
	ExpectedFilterOption
	ExpectedFilterOptionKey
	ExpectedFilterOptionValue
	ExpectedChainSeparator
	ExpectedLinkLabel
	ExpectedInputLabel
	ExpectedOutputLabel
)

func (t ExpectedToken) String() string {
	switch t {
	case ExpectedFilterName:
		return "FilterName"
	case ExpectedFilterOption:
		return "FilterOption"
	case ExpectedFilterOptionKey:
		return "FilterOptionKey"
	case ExpectedFilterOptionValue:
		return "FilterOptionValue"
	case ExpectedChainSeparator:
		return "ChainSeparator"
	case ExpectedLinkLabel:
		return "LinkLabel"
	case ExpectedInputLabel:
		return "InputLabel"
	case ExpectedOutputLabel:
		return "OutputLabel"
	}
	return "Unknown"
}

func (t ExpectedToken) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type FilterContext struct {
	Name       string   `json:"name"`
	Options    []string `json:"options,omitempty"`
	OptionKeys []string `json:"optionKeys,omitempty"`
	ArgIndex   int      `json:"argIndex"`
	InKey      bool     `json:"inKey"`
	InValue    bool     `json:"inValue"`
}

type CompletionContext struct {
	ExpectedTokens []ExpectedToken `json:"expectedTokens"`
	PartialIdent   string          `json:"partialIdent,omitempty"`
	Filter         *FilterContext  `json:"filter,omitempty"`
	InLabel        bool            `json:"inLabel"`
	LabelContent   string          `json:"labelContent,omitempty"`
	ChainIndex     int             `json:"chainIndex"`
	FilterIndex    int             `json:"filterIndex"`
	IsComplex      bool            `json:"isComplex"`
}
