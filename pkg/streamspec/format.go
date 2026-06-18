package streamspec

import (
	"fmt"
	"strings"
)

func Format(spec *Specifier) string {
	return formatSpecifier(spec)
}

func formatSpecifier(spec *Specifier) string {
	var s string
	switch spec.Kind {
	case KindStreamIndex:
		s = fmt.Sprintf("%d", spec.StreamIndex())
	case KindStreamType:
		st := spec.StreamType()
		s = st.Type.String()
	case KindGroup:
		g := spec.Group()
		switch g.Kind {
		case GroupByIndex:
			s = fmt.Sprintf("g:%d", g.Index)
		case GroupByID:
			if g.ID != "" {
				s = fmt.Sprintf("g:#%s", g.ID)
			}
		}
	case KindProgram:
		p := spec.Program()
		s = fmt.Sprintf("p:%s", p.ID)
	case KindStreamID:
		id := spec.StreamID()
		if id.Alt {
			s = fmt.Sprintf("i:%s", id.ID)
		} else {
			s = fmt.Sprintf("#%s", id.ID)
		}
	case KindMetadata:
		m := spec.Metadata()
		if m.Value != "" {
			s = fmt.Sprintf("m:%s:%s", m.Key, m.Value)
		} else {
			s = fmt.Sprintf("m:%s", m.Key)
		}
	case KindDisposition:
		d := spec.Disposition()
		s = "disp:" + strings.Join(d.Dispositions, "+")
	case KindUsable:
		s = "u"
	default:
		s = "?"
	}
	if spec.Additional != nil {
		s += ":" + formatSpecifier(spec.Additional)
	}
	return s
}
