package mapvalue

type ExpectedToken int

const (
	ExpectedFileIndex ExpectedToken = iota
	ExpectedSpecifier
	ExpectedViewSpecifier
	ExpectedLinkLabel
	ExpectedOptional
)

func (t ExpectedToken) String() string {
	switch t {
	case ExpectedFileIndex:
		return "FileIndex"
	case ExpectedSpecifier:
		return "Specifier"
	case ExpectedViewSpecifier:
		return "ViewSpecifier"
	case ExpectedLinkLabel:
		return "LinkLabel"
	case ExpectedOptional:
		return "Optional"
	}
	return "Unknown"
}

func (t ExpectedToken) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type CompletionContext struct {
	ExpectedTokens []ExpectedToken `json:"expectedTokens"`
	PartialIdent    string          `json:"partialIdent,omitempty"`
	Negate          bool            `json:"negate"`
	FileIndex       int             `json:"fileIndex"`
	HasSpecifier    bool            `json:"hasSpecifier"`
	Specifier       string          `json:"specifier,omitempty"`
	IsLinkLabel     bool            `json:"isLinkLabel"`
}