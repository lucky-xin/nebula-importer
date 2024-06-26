//go:generate mockgen -source=source.go -destination source_mock.go -package source Source,Sizer,Globber
package source

import (
	"io"

	"github.com/lucky-xin/nebula-importer/pkg/errors"
)

type (
	Source interface {
		Config() *Config
		Name() string
		Open() error
		Sizer
		io.Reader
		io.Closer
	}

	Sizer interface {
		Size() (int64, error)
	}

	Globber interface {
		Glob() ([]*Config, error)
	}
)

func New(c *Config) (Source, error) {
	// TODO: support blob and so on
	switch {
	case c.SQL != nil:
		return newSQLSource(c), nil
	case c.S3 != nil:
		return newS3Source(c), nil
	case c.OSS != nil:
		return newOSSSource(c), nil
	case c.FTP != nil:
		return newFTPSource(c), nil
	case c.SFTP != nil:
		return newSFTPSource(c), nil
	case c.HDFS != nil:
		return newHDFSSource(c), nil
	case c.GCS != nil:
		return newGCSSource(c), nil
	case c.Local != nil:
		return newLocalSource(c), nil
	}
	return nil, errors.ErrUnsetSource
}
