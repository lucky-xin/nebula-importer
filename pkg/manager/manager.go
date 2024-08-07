//go:generate mockgen -source=manager.go -destination manager_mock.go -package manager Manager
package manager

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lucky-xin/nebula-importer/pkg/client"
	"github.com/lucky-xin/nebula-importer/pkg/errors"
	"github.com/lucky-xin/nebula-importer/pkg/importer"
	"github.com/lucky-xin/nebula-importer/pkg/logger"
	"github.com/lucky-xin/nebula-importer/pkg/reader"
	"github.com/lucky-xin/nebula-importer/pkg/source"
	"github.com/lucky-xin/nebula-importer/pkg/spec"
	"github.com/lucky-xin/nebula-importer/pkg/stats"

	"github.com/panjf2000/ants/v2"
)

const (
	DefaultReaderConcurrency   = 50
	DefaultImporterConcurrency = 512
	DefaultStatsInterval       = time.Second * 10
)

type (
	Manager interface {
		Import(s source.Source, brr reader.BatchRecordReader, importers ...importer.Importer) error
		Start() error
		Wait() error
		Stats() *stats.Stats
		Stop() error
	}

	defaultManager struct {
		graphName           string
		recordStats         bool
		pool                client.Pool
		getClientOptions    []client.Option
		stats               *stats.ConcurrencyStats
		batch               int
		readerConcurrency   int
		readerWaitGroup     sync.WaitGroup
		readerPool          *ants.Pool
		importerConcurrency int
		importerWaitGroup   sync.WaitGroup
		importerPool        *ants.Pool
		statsInterval       time.Duration
		hooks               *Hooks
		chStart             chan struct{}
		done                chan struct{}
		isStopped           atomic.Bool
		logger              logger.Logger
	}

	Option func(*defaultManager)
)

func New(pool client.Pool, opts ...Option) Manager {
	options := make([]Option, 0, 1+len(opts))
	options = append(options, WithClientPool(pool))
	options = append(options, opts...)
	return NewWithOpts(options...)
}

func NewWithOpts(opts ...Option) Manager {
	m := &defaultManager{
		readerConcurrency:   DefaultReaderConcurrency,
		importerConcurrency: DefaultImporterConcurrency,
		statsInterval:       DefaultStatsInterval,
		hooks:               &Hooks{},
		chStart:             make(chan struct{}),
		done:                make(chan struct{}),
	}

	for _, opt := range opts {
		opt(m)
	}
	m.stats = stats.NewConcurrencyStats(m.recordStats)
	m.readerPool, _ = ants.NewPool(m.readerConcurrency)
	m.importerPool, _ = ants.NewPool(m.importerConcurrency)

	if m.logger == nil {
		m.logger = logger.NopLogger
	}

	return m
}

func WithGraphName(graphName string) Option {
	return func(m *defaultManager) {
		m.graphName = graphName
	}
}

func WithClientPool(pool client.Pool) Option {
	return func(m *defaultManager) {
		m.pool = pool
	}
}

func WithRecordStats(recordStats bool) Option {
	return func(m *defaultManager) {
		m.recordStats = recordStats
	}
}

func WithGetClientOptions(opts ...client.Option) Option {
	return func(m *defaultManager) {
		m.getClientOptions = opts
	}
}

func WithBatch(batch int) Option {
	return func(m *defaultManager) {
		if batch > 0 {
			m.batch = batch
		}
	}
}
func WithReaderConcurrency(concurrency int) Option {
	return func(m *defaultManager) {
		if concurrency > 0 {
			m.readerConcurrency = concurrency
		}
	}
}

func WithImporterConcurrency(concurrency int) Option {
	return func(m *defaultManager) {
		if concurrency > 0 {
			m.importerConcurrency = concurrency
		}
	}
}

func WithStatsInterval(statsInterval time.Duration) Option {
	return func(m *defaultManager) {
		if statsInterval > 0 {
			m.statsInterval = statsInterval
		}
	}
}

func WithBeforeHooks(hooks ...*Hook) Option {
	return func(m *defaultManager) {
		m.hooks.Before = hooks
	}
}

func WithAfterHooks(hooks ...*Hook) Option {
	return func(m *defaultManager) {
		m.hooks.After = hooks
	}
}

func WithLogger(l logger.Logger) Option {
	return func(m *defaultManager) {
		m.logger = l
	}
}

func (m *defaultManager) Import(s source.Source, brr reader.BatchRecordReader, importers ...importer.Importer) error {
	if len(importers) == 0 {
		return nil
	}
	logSourceField := logger.Field{Key: "source", Value: s.Name()}
	if err := s.Open(); err != nil {
		err = errors.NewImportError(err, "manager: open import source failed").SetGraphName(m.graphName)
		m.logError(err, "", logSourceField)
		return err
	}

	n, err := s.Size()
	if err != nil {
		_ = s.Close()
		err = errors.NewImportError(err, "manager: get size of import source failed").SetGraphName(m.graphName)
		m.logError(err, "", logSourceField)
		return err
	}
	m.stats.AddTotal(n)

	m.readerWaitGroup.Add(1)
	for _, i := range importers {
		i.Add(1) // Add 1 for start, will call Done after i.Import finish
	}

	cleanup := func() {
		for _, i := range importers {
			i.Done() // Done 1 for finish, corresponds to start
		}
		m.readerWaitGroup.Done()
		_ = s.Close()
	}

	go func() {
		err = m.readerPool.Submit(func() {
			<-m.chStart
			defer cleanup()

			for _, i := range importers {
				i.Wait()
			}
			_ = m.loopImport(s, brr, importers...)
		})
		if err != nil {
			cleanup()
			m.logError(err, "manager: submit reader failed", logSourceField)
		}
	}()

	m.logger.Info("manager: add import source successfully", logSourceField)
	return nil
}

func (m *defaultManager) Start() error {
	m.logger.Info("manager: starting")

	if err := m.Before(); err != nil {
		return err
	}

	m.stats.Init()

	if err := m.pool.Open(); err != nil {
		m.logger.WithError(err).Error("manager: start client pool failed")
		return err
	}

	close(m.chStart)

	go m.loopPrintStats()
	m.logger.Info("manager: start successfully")
	return nil
}

func (m *defaultManager) Wait() error {
	m.logger.Info("manager: wait")

	m.readerWaitGroup.Wait()
	m.importerWaitGroup.Wait()

	m.logger.Info("manager: wait successfully")
	return m.Stop()
}

func (m *defaultManager) Stats() *stats.Stats {
	return m.stats.Stats()
}

func (m *defaultManager) Stop() (err error) {
	if m.isStopped.Load() {
		return nil
	}
	m.isStopped.Store(true)

	m.logger.Info("manager: stop")
	defer func() {
		if err != nil {
			err = errors.NewImportError(err, "manager: stop failed")
			m.logError(err, "")
		} else {
			m.logger.Info("manager: stop successfully")
		}
	}()
	close(m.done)

	m.readerWaitGroup.Wait()
	m.importerWaitGroup.Wait()

	m.logStats()
	return m.After()
}

func (m *defaultManager) Before() error {
	m.logger.Info("manager: exec before hook")
	return m.execHooks(BeforeHook)
}

func (m *defaultManager) After() error {
	m.logger.Info("manager: exec after hook")
	return m.execHooks(AfterHook)
}

func (m *defaultManager) execHooks(name HookName) error {
	var hooks []*Hook
	switch name {
	case BeforeHook:
		hooks = m.hooks.Before
	case AfterHook:
		hooks = m.hooks.After
	}
	if len(hooks) == 0 {
		return nil
	}

	var cli client.Client
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		for _, statement := range hook.Statements {
			if statement == "" {
				continue
			}

			if cli == nil {
				var err error
				cli, err = m.pool.GetClient(m.getClientOptions...)
				if err != nil {
					return err
				}
			}
			resp, err := cli.Execute(statement)
			if err != nil {
				err = errors.NewImportError(err,
					"manager: exec failed in %s hook", name,
				).SetStatement(statement)
				m.logError(err, "")
				return err
			}
			if !resp.IsSucceed() {
				err = errors.NewImportError(err,
					"manager: exec failed in %s hook, %s", name, resp.GetError(),
				).SetStatement(statement)
				m.logError(err, "")
				return err
			}
		}
		if hook.Wait != 0 {
			m.logger.Info(fmt.Sprintf("manager: waiting %s", hook.Wait))
			time.Sleep(hook.Wait)
		}
	}
	return nil
}

func (m *defaultManager) loopImport(s source.Source, r reader.BatchRecordReader, importers ...importer.Importer) error {
	logSourceField := logger.Field{Key: "source", Value: s.Name()}
	for {
		select {
		case <-m.done:
			return nil
		default:
			n, records, err := r.ReadBatch()
			if err != nil {
				if err != io.EOF {
					err = errors.NewImportError(err, "manager: read batch failed").SetGraphName(m.graphName)
					m.logError(err, "", logSourceField)
					return err
				}
				return nil
			}
			m.submitImporterTask(n, records, importers...)
		}
	}
}

func (m *defaultManager) submitImporterTask(n int, records spec.Records, importers ...importer.Importer) {
	importersDone := func() {
		for _, i := range importers {
			i.Done() // Done 1 for batch
		}
	}

	for _, i := range importers {
		i.Add(1) // Add 1 for batch
	}
	m.importerWaitGroup.Add(1)
	if err := m.importerPool.Submit(func() {
		defer m.importerWaitGroup.Done()
		defer importersDone()

		size := len(records)
		var faileds []spec.Record
		var succeededs []spec.Record
		if size > 0 {
			for _, i := range importers {
				times := size / m.batch
				if size%m.batch != 0 || times == 0 {
					times++
				}
				for j := range times {
					start := j * m.batch
					end := (j + 1) * m.batch
					if end > size {
						end = size
					}
					subs := records[start:end]
					result, err := i.Import(subs...)
					if err != nil {
						m.logError(err, "manager: import failed")
						m.onRequestFailed(subs)
						faileds = append(faileds, subs...)
						// do not return, continue the subsequent importer.
					} else {
						if result.RecordNum > 0 {
							m.onRequestSucceeded(result)
						}
						succeededs = append(succeededs, subs...)
					}
				}
			}
		}
		m.logger.Debug(fmt.Sprintf("manager: import %d records, n:%d successfully", size, n))
		m.onFailed(0, faileds)
		m.onSucceeded(n, succeededs)
	}); err != nil {
		importersDone()
		m.importerWaitGroup.Done()
		m.logError(err, "manager: submit importer failed")
	}
}

func (m *defaultManager) loopPrintStats() {
	if m.statsInterval <= 0 {
		return
	}
	ticker := time.NewTicker(m.statsInterval)
	m.logStats()
	for {
		select {
		case <-ticker.C:
			m.logStats()
		case <-m.done:
			return
		}
	}
}

func (m *defaultManager) logStats() {
	m.logger.Info(m.Stats().String())
}

func (m *defaultManager) onFailed(nBytes int, records spec.Records) {
	m.stats.Failed(int64(nBytes), int64(len(records)))
}

func (m *defaultManager) onSucceeded(nBytes int, records spec.Records) {
	m.stats.Succeeded(int64(nBytes), int64(len(records)))
}

func (m *defaultManager) onRequestFailed(records spec.Records) {
	m.stats.RequestFailed(int64(len(records)))
}

func (m *defaultManager) onRequestSucceeded(result *importer.ImportResp) {
	m.stats.RequestSucceeded(int64(result.RecordNum), result.Latency, result.RespTime)
}

func (m *defaultManager) logError(err error, msg string, fields ...logger.Field) {
	e := errors.AsOrNewImportError(err)
	fields = append(fields, logger.MapToFields(e.Fields())...)
	m.logger.SkipCaller(1).WithError(e.Cause()).Error(msg, fields...)
}
