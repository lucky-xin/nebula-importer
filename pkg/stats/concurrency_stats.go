package stats

import (
	"sync"
	"time"
)

type (
	ConcurrencyStats struct {
		s       Stats
		mu      sync.Mutex
		initOne sync.Once
	}
)

func NewConcurrencyStats(byteStats bool) *ConcurrencyStats {
	return &ConcurrencyStats{
		s: Stats{
			RecordStats: byteStats,
		},
	}
}

func (s *ConcurrencyStats) Init() {
	s.initOne.Do(func() {
		s.s.StartTime = time.Now()
	})
}

func (s *ConcurrencyStats) AddTotalBytes(nBytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Total += nBytes
}

func (s *ConcurrencyStats) Failed(n, nRecords int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Processed += n
	s.s.FailedRecords += nRecords
	s.s.TotalRecords += nRecords
}

func (s *ConcurrencyStats) Succeeded(n, nRecords int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Processed += n
	s.s.TotalRecords += nRecords
}

func (s *ConcurrencyStats) RequestFailed(nRecords int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.FailedRequest++
	s.s.TotalRequest++
	s.s.FailedProcessed += nRecords
	s.s.TotalProcessed += nRecords
}

func (s *ConcurrencyStats) RequestSucceeded(nRecords int64, latency, respTime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.TotalRequest++
	s.s.TotalLatency += latency
	s.s.TotalRespTime += respTime
	s.s.TotalProcessed += nRecords
}

func (s *ConcurrencyStats) Stats() *Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	cpy := s.s
	return &cpy
}

func (s *ConcurrencyStats) String() string {
	return s.Stats().String()
}
