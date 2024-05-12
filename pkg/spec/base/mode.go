package specbase

import "strings"

const (
	DefaultMode      = UpsertMode
	InsertMode  Mode = "INSERT"
	UpsertMode  Mode = "UPSERT"
	UpdateMode  Mode = "UPDATE"
	DeleteMode  Mode = "DELETE"
)

type Mode string

func (m Mode) Convert() Mode {
	if m == "" {
		return DefaultMode
	}
	return Mode(strings.ToUpper(string(m)))
}

func (m Mode) IsSupport() bool {
	return m == InsertMode || m == UpdateMode || m == DeleteMode || m == UpsertMode
}
