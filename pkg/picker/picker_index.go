package picker

import (
	"github.com/lucky-xin/nebula-importer/pkg/errors"
)

var (
	_ Picker = IndexPicker(0)
)

type IndexPicker int

func (ip IndexPicker) Pick(record []string) (*Value, error) {
	index := int(ip)
	if index < 0 || index >= len(record) {
		return nil, errors.ErrNoRecord
	}
	return NewValue(record[index]), nil
}
