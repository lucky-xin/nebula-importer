package reader

import "github.com/lucky-xin/nebula-importer/pkg/source"

type (
	baseReader struct {
		s source.Source
	}
)

func (r *baseReader) Source() source.Source {
	return r.s
}
