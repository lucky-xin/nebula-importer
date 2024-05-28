package configbase

import (
	"time"

	"github.com/lucky-xin/nebula-importer/pkg/manager"
)

type (
	Manager struct {
		Batch               int           `yaml:"batch,omitempty" json:"batch,omitempty,optional,default=100"`
		ReaderConcurrency   int           `yaml:"readerConcurrency,omitempty" json:"readerConcurrency,omitempty,optional,default=100"`
		ImporterConcurrency int           `yaml:"importerConcurrency,omitempty" json:"importerConcurrency,omitempty,optional,default=200"`
		StatsInterval       time.Duration `yaml:"statsInterval,omitempty" json:"statsInterval,omitempty,optional,default=10000000000"`
		Hooks               manager.Hooks `yaml:"hooks,omitempty" json:"hooks,omitempty,optional"`
	}
)
