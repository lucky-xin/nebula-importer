package configbase

import (
	"time"

	"github.com/lucky-xin/nebula-importer/pkg/manager"
)

type (
	Manager struct {
		Batch               int           `yaml:"batch,omitempty"`
		ReaderConcurrency   int           `yaml:"readerConcurrency,omitempty"`
		ImporterConcurrency int           `yaml:"importerConcurrency,omitempty"`
		StatsInterval       time.Duration `yaml:"statsInterval,omitempty"`
		Hooks               manager.Hooks `yaml:"hooks,omitempty"`
	}
)
