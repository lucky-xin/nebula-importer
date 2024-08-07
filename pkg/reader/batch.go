//go:generate mockgen -source=batch.go -destination batch_mock.go -package reader BatchRecordReader
package reader

import (
	"database/sql"
	stderrors "errors"
	"fmt"
	"github.com/lucky-xin/nebula-importer/pkg/logger"
	"github.com/lucky-xin/nebula-importer/pkg/source"
	"github.com/lucky-xin/nebula-importer/pkg/spec"
	"io"
)

type (
	BatchRecordReader interface {
		Source() source.Source
		source.Sizer
		ReadBatch() (int, spec.Records, error)
	}

	Convertor interface {
		Apply(s source.Source, values []string) (spec.Records, error)
	}

	NoneConvertor struct {
	}

	continueError struct {
		Err error
	}

	defaultBatchReader struct {
		*options
		rr RecordReader
		c  Convertor
	}

	sqlBatchReader struct {
		*options
		s      *source.SQLSource
		total  int64
		lastId string
		c      Convertor
	}
)

var (
	converts map[string]Convertor
)

func init() {
	converts = map[string]Convertor{}
	converts["none"] = &NoneConvertor{}
}

func RegistryConvertor(name string, convert Convertor) {
	if converts == nil {
		converts = map[string]Convertor{}
	}
	converts[name] = convert
}

func GetConvertor(name string) Convertor {
	convertor := converts[name]
	if convertor == nil {
		return &NoneConvertor{}
	}
	return convertor
}

func NewBatchRecordReader(rr RecordReader, c string, opts ...Option) BatchRecordReader {
	brr := &defaultBatchReader{
		options: newOptions(opts...),
		rr:      rr,
		c:       GetConvertor(c),
	}
	brr.logger = brr.logger.With(logger.Field{Key: "source", Value: rr.Source().Name()})
	return brr
}

func NewSQLBatchRecordReader(s *source.SQLSource, c string, opts ...Option) BatchRecordReader {
	brr := &sqlBatchReader{
		options: newOptions(opts...),
		s:       s,
		c:       GetConvertor(c),
	}
	brr.logger = brr.logger.With(logger.Field{Key: "source", Value: s.Name()})
	return brr
}

func NewContinueError(err error) error {
	return &continueError{
		Err: err,
	}
}

func (*NoneConvertor) Apply(s source.Source, values []string) (spec.Records, error) {
	return spec.Records{values}, nil
}

func (r *defaultBatchReader) Source() source.Source {
	return r.rr.Source()
}

func (r *defaultBatchReader) Size() (int64, error) {
	return r.rr.Size()
}

func (r *defaultBatchReader) ReadBatch() (int, spec.Records, error) {
	var (
		totalBytes int
		records    = make(spec.Records, 0, r.batch)
	)

	for batch := 0; batch < r.batch; {
		n, record, err := r.rr.Read()
		totalBytes += n
		if err != nil {
			// case1: Read continue error.
			if ce := new(continueError); stderrors.As(err, &ce) {
				r.logger.WithError(ce.Err).Error("read source failed")
				continue
			}

			// case2: Read error and still have records.
			if totalBytes > 0 {
				break
			}

			// Read error and have no records.
			return 0, nil, err
		}
		batch++
		result, err := r.c.Apply(r.Source(), record)
		if err != nil {
			return 0, nil, err
		}
		records = append(records, result...)
	}
	return totalBytes, records, nil
}

func (ce *continueError) Error() string {
	return ce.Err.Error()
}

func (ce *continueError) Cause() error {
	return ce.Err
}

func (ce *continueError) Unwrap() error {
	return ce.Err
}

func (r *sqlBatchReader) Source() source.Source {
	return r.s
}

func (r *sqlBatchReader) Size() (int64, error) {
	return r.s.Size()
}

func (r *sqlBatchReader) ReadBatch() (n int, records spec.Records, err error) {
	querySql := r.s.BuildQuerySQL(r.lastId, r.batch)
	r.logger.Debug(fmt.Sprintf("query sql: %s", querySql))
	rows, err := r.s.Db.Query(querySql)
	if err != nil {
		return 0, nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		r.logger.Error(fmt.Sprintf("query error: %s", err.Error()))
		return 0, nil, err
	}
	var lastId string
	for rows.Next() {
		values := make([]interface{}, len(cols))
		for i := range values {
			values[i] = &sql.NullString{}
		}
		n++
		err = rows.Scan(values...)
		if err != nil {
			if ce := new(continueError); stderrors.As(err, &ce) {
				r.logger.WithError(ce.Err).Error("read source failed")
				continue
			}
			return 0, nil, err
		}
		vals := make([]string, 0, len(values))
		for i := range values {
			nullString := values[i].(*sql.NullString)
			if nullString.Valid {
				vals = append(vals, nullString.String)
			} else {
				vals = append(vals, "")
			}
		}
		lastId = vals[r.s.Config().SQL.DbTable.Id.Index]
		result, err := r.c.Apply(r.Source(), vals)
		if err != nil {
			return 0, nil, err
		}
		records = append(records, result...)
	}
	r.lastId = lastId
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	if n == 0 {
		r.logger.Debug("not found data")
		return n, nil, io.EOF
	}
	return n, records, nil
}
