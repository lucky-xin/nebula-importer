//go:generate mockgen -package idx -destination id_mock.go -source id.go Generator
package idx

import (
	"context"
	"github.com/redis/go-redis/v9"
	"sync"

	"github.com/rs/xid"
)

const (
	// idPrefix is used to avoid the first character may be a number.
	idPrefix = "n"
)

var (
	gGenerator     Generator = (*xidGenerator)(nil)
	gGeneratorInit sync.Once
)

type (
	Generator interface {
		String() string
		Int64(label string) int64
	}

	xidGenerator struct {
		redis *redis.Client
	}
)

func New(redis *redis.Client) Generator {
	return &xidGenerator{redis: redis}
}

func (*xidGenerator) String() string {
	return idPrefix + xid.New().String()
}

func (xid *xidGenerator) Int64(label string) int64 {
	return xid.redis.Incr(context.Background(), label).Val()
}
