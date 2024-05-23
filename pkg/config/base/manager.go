package configbase

import (
	"time"

	"github.com/lucky-xin/nebula-importer/pkg/manager"
)

type (
	Manager struct {
		Batch               int           `yaml:"batch,omitempty" json:"batch,omitempty,optional"`
		ReaderConcurrency   int           `yaml:"readerConcurrency,omitempty" json:"readerConcurrency,omitempty,optional"`
		ImporterConcurrency int           `yaml:"importerConcurrency,omitempty" json:"importerConcurrency,omitempty,optional"`
		StatsInterval       time.Duration `yaml:"statsInterval,omitempty" json:"statsInterval,omitempty,optional"`
		Hooks               manager.Hooks `yaml:"hooks,omitempty" json:"hooks,omitempty,optional"`
	}
)
