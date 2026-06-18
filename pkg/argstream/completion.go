package argstream

type ExpectedToken int

const (
	ExpectedGlobalOption ExpectedToken = iota
	ExpectedInputOption
	ExpectedInputURL
	ExpectedOutputOption
	ExpectedOutputURL
	ExpectedOptionValue
	ExpectedStreamSpecifier
	ExpectedFilterValue
	ExpectedMapValue
)

func (t ExpectedToken) String() string {
	switch t {
	case ExpectedGlobalOption:
		return "GlobalOption"
	case ExpectedInputOption:
		return "InputOption"
	case ExpectedInputURL:
		return "InputURL"
	case ExpectedOutputOption:
		return "OutputOption"
	case ExpectedOutputURL:
		return "OutputURL"
	case ExpectedOptionValue:
		return "OptionValue"
	case ExpectedStreamSpecifier:
		return "StreamSpecifier"
	case ExpectedFilterValue:
		return "FilterValue"
	case ExpectedMapValue:
		return "MapValue"
	}
	return "Unknown"
}

func (t ExpectedToken) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// OptionContext provides details about the option being completed.
type OptionContext struct {
	Name           string    `json:"name"`
	CanonicalName  string    `json:"canonicalName"`
	StreamSpecifier string   `json:"streamSpecifier,omitempty"`
	ValueType      ValueType `json:"valueType"`
	AcceptsSpec    bool      `json:"acceptsSpec"`
	IsBoolean      bool      `json:"isBoolean"`
	Style          string    `json:"style"`
}

// CompletionContext describes what is expected at the completion position.
type CompletionContext struct {
	ExpectedTokens []ExpectedToken `json:"expectedTokens"`
	Scope          Scope           `json:"scope"`
	InputCount     int             `json:"inputCount"`
	OutputCount    int             `json:"outputCount"`

	// Current option being completed (if any)
	CurrentOption *OptionContext `json:"currentOption,omitempty"`

	// Partial text for filtering
	PartialOption string `json:"partialOption,omitempty"`
	PartialValue  string `json:"partialValue,omitempty"`
	PartialSpec   string `json:"partialSpec,omitempty"`
}
