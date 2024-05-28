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
	"strings"
)

type (
	BatchRecordReader interface {
		Source() source.Source
		source.Sizer
		ReadBatch() (int, spec.Records, error)
	}

	Convertor interface {
		Apply(values []string) (spec.Records, error)
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

func (*NoneConvertor) Apply(values []string) (spec.Records, error) {
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
		result, err := r.c.Apply(record)
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

func (r *sqlBatchReader) ReadBatch() (int, spec.Records, error) {
	querySql := r.buildStatement(r.s)
	rows, err := r.s.Db.Query(querySql)
	if err != nil {
		return 0, nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return 0, nil, err
	}
	records := make(spec.Records, 0, r.batch)
	for rows.Next() {
		values := make([]interface{}, len(cols))
		for i := range values {
			values[i] = &sql.NullString{}
		}
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
		result, err := r.c.Apply(vals)
		if err != nil {
			return 0, nil, err
		}
		records = append(records, result...)

	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	n := len(records)
	if n == 0 {
		return 0, nil, io.EOF
	}
	r.lastId = records[n-1][0]
	return n, records, nil
}

func (r *sqlBatchReader) buildStatement(sqlSource *source.SQLSource) string {
	table := sqlSource.Config().SQL.DbTable
	var statement string
	if table.SQL != "" {
		statement = table.SQL
	} else {
		statement = "SELECT `" + strings.Join(table.Fields, "`,`") + "` FROM " + table.Name + " WHERE 1 = 1"
	}

	if table.Filter != "" {
		statement += " AND " + table.Filter
	}
	if r.lastId != "" {
		statement += " AND " + table.PrimaryKey + " > '" + r.lastId + "'"
	}
	statement += " ORDER BY " + table.PrimaryKey + " ASC LIMIT " + fmt.Sprintf("%d", r.batch)
	return statement
}
