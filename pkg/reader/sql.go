package reader

import (
	"database/sql"
	"fmt"
	"github.com/lucky-xin/nebula-importer/pkg/source"
	"github.com/lucky-xin/nebula-importer/pkg/spec"
	"github.com/pkg/errors"
	"go/types"
	"io"
	"strings"
	"time"
)

type (
	sqlReader struct {
		*baseReader
		rows      *sql.Rows
		total     int64
		index     int64
		batchSize int64
		columns   []string
		lastId    string
	}
)

func NewSQLReader(s source.Source) RecordReader {
	size, err := s.Size()
	if err != nil {
		panic(err)
	}
	return &sqlReader{
		total:     size,
		batchSize: 10,
		baseReader: &baseReader{
			s: s,
		},
	}
}

func (r *sqlReader) Size() (int64, error) {
	return r.s.Size()
}

func (r *sqlReader) Read() (n int, record spec.Record, err error) {
	if r.rows == nil {
		err = r.initBatch()
		if err != nil {
			return
		}
	}
	if !r.rows.Next() {
		if r.rows != nil {
			_ = r.rows.Close()
		}
		if r.index >= r.total {
			err = io.EOF
			r.rows = nil
			return
		}
	}
	columnPointers := make([]interface{}, len(r.columns))
	for i := range columnPointers {
		var val interface{}
		columnPointers[i] = &val
	}
	err = r.rows.Scan(columnPointers...)
	if err != nil {
		return
	}
	for i := range r.columns {
		val := *(columnPointers[i].(*interface{}))
		switch val.(type) {
		case []uint8:
			val = string(val.([]uint8))
		case time.Time:
			t := val.(time.Time)
			val = t.Format("2006-01-02 15:04:05")
		case types.Nil:
			val = ""
		}
		if val == nil {
			val = ""
		}
		record = append(record, fmt.Sprintf("%v", val))
	}
	r.lastId = record[0]
	r.index++
	return
}

func (r *sqlReader) initBatch() error {
	var sqlSource *source.SqlSource
	if ss, ok := r.s.(*source.SqlSource); ok {
		sqlSource = ss
	} else {
		return errors.New("source is not sql source")
	}
	querySql := r.buildStatement(sqlSource)
	rows, err := sqlSource.Db.Query(querySql)
	if err != nil {
		return err
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	r.rows = rows
	r.columns = cols
	return nil
}

func (r *sqlReader) buildStatement(sqlSource *source.SqlSource) string {
	statement := "SELECT " + strings.Join(sqlSource.Config().SQL.DbTable.Fields, ",") + " FROM " +
		sqlSource.Config().SQL.DbTable.Name +
		" WHERE 1 = 1"
	if sqlSource.Config().SQL.DbTable.Filter != "" {
		statement += " AND " + sqlSource.Config().SQL.DbTable.Filter
	}
	if r.lastId != "" {
		statement += " AND " + sqlSource.Config().SQL.DbTable.PrimaryKey + " > '" + r.lastId + "'"
	}
	statement += " ORDER BY " + sqlSource.Config().SQL.DbTable.PrimaryKey + " ASC LIMIT " + fmt.Sprintf("%d", r.batchSize)
	println(statement)
	return statement
}
