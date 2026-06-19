package streamspec

type SpecifierKind int

const (
	KindStreamIndex SpecifierKind = iota
	KindStreamType
	KindGroup
	KindProgram
	KindStreamID
	KindMetadata
	KindDisposition
	KindUsable
)

func (k SpecifierKind) String() string {
	switch k {
	case KindStreamIndex:
		return "StreamIndex"
	case KindStreamType:
		return "StreamType"
	case KindGroup:
		return "Group"
	case KindProgram:
		return "Program"
	case KindStreamID:
		return "StreamID"
	case KindMetadata:
		return "Metadata"
	case KindDisposition:
		return "Disposition"
	case KindUsable:
		return "Usable"
	}
	return "Unknown"
}

type StreamType int

const (
	TypeVideo           StreamType = iota // v
	TypeVideoNoAttached                   // V
	TypeAudio                             // a
	TypeSubtitle                          // s
	TypeData                              // d
	TypeAttachment                        // t
)

func (t StreamType) String() string {
	switch t {
	case TypeVideo:
		return "v"
	case TypeVideoNoAttached:
		return "V"
	case TypeAudio:
		return "a"
	case TypeSubtitle:
		return "s"
	case TypeData:
		return "d"
	case TypeAttachment:
		return "t"
	}
	return ""
}

type StreamTypeExpr struct {
	Type StreamType
}

type GroupKind int

const (
	GroupByIndex GroupKind = iota
	GroupByID
)

type GroupExpr struct {
	Kind  GroupKind
	Index int
	ID    string
}

type ProgramExpr struct {
	ID string
}

type StreamIDExpr struct {
	ID  string
	Alt bool // true for i: prefix, false for # prefix
}

type MetadataExpr struct {
	Key   string
	Value string
}

type DispositionExpr struct {
	Dispositions []string
}

type Specifier struct {
	Kind       SpecifierKind
	Span       Span
	payload    any
	Additional *Specifier
}

func (s *Specifier) StreamIndex() int {
	if s.Kind != KindStreamIndex {
		return -1
	}
	return s.payload.(int)
}

func (s *Specifier) StreamType() *StreamTypeExpr {
	if s.Kind != KindStreamType {
		return nil
	}
	return s.payload.(*StreamTypeExpr)
}

func (s *Specifier) Group() *GroupExpr {
	if s.Kind != KindGroup {
		return nil
	}
	return s.payload.(*GroupExpr)
}

func (s *Specifier) Program() *ProgramExpr {
	if s.Kind != KindProgram {
		return nil
	}
	return s.payload.(*ProgramExpr)
}

func (s *Specifier) StreamID() *StreamIDExpr {
	if s.Kind != KindStreamID {
		return nil
	}
	return s.payload.(*StreamIDExpr)
}

func (s *Specifier) Metadata() *MetadataExpr {
	if s.Kind != KindMetadata {
		return nil
	}
	return s.payload.(*MetadataExpr)
}

func (s *Specifier) Disposition() *DispositionExpr {
	if s.Kind != KindDisposition {
		return nil
	}
	return s.payload.(*DispositionExpr)
}
