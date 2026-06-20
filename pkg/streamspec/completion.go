package streamspec

type ExpectedToken int

const (
	ExpectedSpecifierType ExpectedToken = iota
	ExpectedStreamIndex
	ExpectedStreamTypeLetter
	ExpectedGroupSpecifier
	ExpectedProgramID
	ExpectedGroupIndex
	ExpectedGroupID
	ExpectedStreamIDValue
	ExpectedMetadataKey
	ExpectedMetadataValue
	ExpectedDispositionName
)

func (t ExpectedToken) String() string {
	switch t {
	case ExpectedSpecifierType:
		return "SpecifierType"
	case ExpectedStreamIndex:
		return "StreamIndex"
	case ExpectedStreamTypeLetter:
		return "StreamTypeLetter"
	case ExpectedGroupSpecifier:
		return "GroupSpecifier"
	case ExpectedProgramID:
		return "ProgramID"
	case ExpectedGroupIndex:
		return "GroupIndex"
	case ExpectedGroupID:
		return "GroupID"
	case ExpectedStreamIDValue:
		return "StreamIDValue"
	case ExpectedMetadataKey:
		return "MetadataKey"
	case ExpectedMetadataValue:
		return "MetadataValue"
	case ExpectedDispositionName:
		return "DispositionName"
	}
	return "Unknown"
}

func (t ExpectedToken) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type SpecifierForm struct {
	Prefix      string `json:"prefix"`
	Description string `json:"description"`
	Suffix      string `json:"suffix,omitempty"`
	NeedsValue  bool   `json:"needsValue,omitempty"`
}

type CompletionContext struct {
	ExpectedTokens  []ExpectedToken `json:"expectedTokens"`
	ValidForms      []SpecifierForm `json:"validForms,omitempty"`
	PartialIdent    string          `json:"partialIdent,omitempty"`
	CurrentKind     SpecifierKind   `json:"currentKind,omitempty"`
	InMetadataValue bool            `json:"inMetadataValue"`
	MetadataKey     string          `json:"metadataKey,omitempty"`
	Dispositions    []string        `json:"dispositions,omitempty"`
}
