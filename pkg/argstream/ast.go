package argstream

// Scope represents where we are in the ffmpeg argument stream.
type Scope int

const (
	ScopeGlobal Scope = iota
	ScopeInputFile
	ScopeOutputFile
)

func (s Scope) String() string {
	switch s {
	case ScopeGlobal:
		return "Global"
	case ScopeInputFile:
		return "InputFile"
	case ScopeOutputFile:
		return "OutputFile"
	}
	return "Unknown"
}

func (s Scope) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// TokenKind represents the type of argument token.
type TokenKind int

const (
	KindGlobalOption TokenKind = iota
	KindInputOption
	KindInputURL
	KindOutputOption
	KindOutputURL
	KindFilterComplex
)

func (k TokenKind) String() string {
	switch k {
	case KindGlobalOption:
		return "GlobalOption"
	case KindInputOption:
		return "InputOption"
	case KindInputURL:
		return "InputURL"
	case KindOutputOption:
		return "OutputOption"
	case KindOutputURL:
		return "OutputURL"
	case KindFilterComplex:
		return "FilterComplex"
	}
	return "Unknown"
}

func (k TokenKind) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

// Token represents a single parsed argument token.
type Token struct {
	Kind           TokenKind
	OptionName     string // e.g. "c" for -c:v:1
	StreamSpecifier string // e.g. "v:1" for -c:v:1
	Value          string // option value (empty for boolean flags)
	URL            string // for input/output URLs
	Span           Span
}

// InputFile tracks an input file declaration.
type InputFile struct {
	Index  int
	URL    string
	Span   Span
}

// OutputFile tracks an output file declaration.
type OutputFile struct {
	Index int
	URL   string
	Span  Span
}

// Program is the top-level AST for a parsed ffmpeg command line.
type Program struct {
	Tokens      []Token
	InputFiles  []InputFile
	OutputFiles []OutputFile
	Span        Span
}