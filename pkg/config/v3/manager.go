package configv3

import (
	"github.com/lucky-xin/nebula-importer/pkg/client"
	configbase "github.com/lucky-xin/nebula-importer/pkg/config/base"
	"github.com/lucky-xin/nebula-importer/pkg/logger"
	"github.com/lucky-xin/nebula-importer/pkg/manager"
	"github.com/lucky-xin/nebula-importer/pkg/reader"
)

type (
	Manager struct {
		GraphName          string `yaml:"spaceName" json:"spaceName"`
		configbase.Manager `yaml:",inline" json:",inline"`
	}
)

func (m *Manager) BuildManager(
	l logger.Logger,
	pool client.Pool,
	sources Sources,
	opts ...manager.Option,
) (manager.Manager, error) {
	options := make([]manager.Option, 0, 8+len(opts))
	options = append(options,
		manager.WithClientPool(pool),
		manager.WithBatch(m.Batch),
		manager.WithReaderConcurrency(m.ReaderConcurrency),
		manager.WithImporterConcurrency(m.ImporterConcurrency),
		manager.WithStatsInterval(m.StatsInterval),
		manager.WithBeforeHooks(m.Hooks.Before...),
		manager.WithAfterHooks(m.Hooks.After...),
		manager.WithLogger(l),
		manager.WithRecordStats(m.RecordStats),
	)
	options = append(options, opts...)

	mgr := manager.NewWithOpts(options...)

	for i := range sources {
		s := sources[i]
		src, brr, err := s.BuildSourceAndReader(reader.WithBatch(m.Batch), reader.WithLogger(l))
		if err != nil {
			return nil, err
		}
		importers, err := s.BuildImporters(m.GraphName, pool)
		if err != nil {
			return nil, err
		}
		if err = mgr.Import(src, brr, importers...); err != nil {
			return nil, err
		}
	}

	return mgr, nil
}
