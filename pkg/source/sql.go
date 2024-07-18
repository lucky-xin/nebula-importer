package source

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
	"strings"
	"time"
)

var _ Source = (*s3Source)(nil)

type (
	SQLConfig struct {
		Endpoint         string   `yaml:"endpoint,omitempty" json:"endpoint"`
		DbName           string   `yaml:"dbName,omitempty" json:"dbName"`
		DbTable          SQLTable `yaml:"dbTable" json:"dbTable"`
		Username         string   `yaml:"username,omitempty" json:"username"`
		Password         string   `yaml:"password,omitempty" json:"password"`
		UrlQuery         string   `yaml:"urlQuery,omitempty" json:"urlQuery,omitempty,optional"`
		DriverName       string   `yaml:"driverName,omitempty" json:"driverName,omitempty,optional,default=mysql"`
		MaxConnections   int      `yaml:"maxConnections,omitempty" json:"maxConnections,omitempty,optional,default=100"`
		MaxIdleConns     int      `yaml:"maxIdleConns,omitempty" json:"maxIdleConns,omitempty,optional,default=20"`
		MaxLifetimeMills int      `yaml:"maxLifetimeMills,omitempty" json:"maxLifetimeMills,omitempty,optional,default=3600000000000"`
		MaxIdleTimeMills int      `yaml:"maxIdleTimeMills,omitempty" json:"maxIdleTimeMills,omitempty,optional,default=30000000000"`
	}

	SQLTable struct {
		Id     SQLId    `yaml:"id" json:"id"`
		Name   string   `yaml:"name,omitempty" json:"name,omitempty,optional"`
		Fields []string `yaml:"fields,omitempty" json:"fields,omitempty,optional"`
		Query  string   `yaml:"query,omitempty" json:"query,omitempty,optional"`
		Count  string   `yaml:"count,omitempty" json:"count,omitempty,optional"`
		Filter string   `yaml:"filter,omitempty" json:"filter,omitempty,optional"`
	}

	SQLSource struct {
		c  *Config
		Db *sql.DB
	}

	SQLId struct {
		Name  string `yaml:"name,omitempty" json:"name,omitempty,optional,default=id"`
		Index int    `yaml:"index,omitempty" json:"index,omitempty,optional,default=0"`
		Alias string `yaml:"alias,omitempty" json:"alias,omitempty,optional"`
	}
)

func newSQLSource(c *Config) Source {
	return &SQLSource{
		c: c,
	}
}

func (s *SQLSource) Name() string {
	return s.c.SQL.String()
}

func (s *SQLSource) Open() (err error) {
	db, err := s.Connect(s.c.SQL.DbName)
	if err != nil {
		return
	}
	s.Db = db
	return
}

func (s *SQLSource) Connect(dbname string) (db *sql.DB, err error) {
	s.setDefaultConfig()
	err = s.validate()
	if err != nil {
		return
	}
	conf := s.c.SQL
	db, err = sql.Open(conf.DriverName,
		fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", conf.Username, conf.Password, conf.Endpoint, dbname, conf.UrlQuery))
	if err != nil {
		return
	}
	db.SetMaxOpenConns(conf.MaxConnections)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	// mysql default conn timeout=8h, should < mysql_timeout
	db.SetConnMaxIdleTime(time.Duration(conf.MaxIdleTimeMills))
	db.SetConnMaxLifetime(time.Duration(conf.MaxLifetimeMills))
	err = db.Ping()
	if err != nil {
		return
	}
	return
}

func (s *SQLSource) setDefaultConfig() {
	if s.c.SQL.UrlQuery == "" {
		s.c.SQL.UrlQuery = "charset=utf8mb4&parseTime=true&loc=Local"
	}
	if s.c.SQL.DriverName == "" {
		s.c.SQL.DriverName = "mysql"
	}
	if s.c.SQL.MaxConnections == 0 {
		s.c.SQL.MaxConnections = 100
	}
	if s.c.SQL.MaxIdleConns == 0 {
		s.c.SQL.MaxIdleConns = 10
	}
	if s.c.SQL.MaxLifetimeMills == 0 {
		s.c.SQL.MaxLifetimeMills = 3600000000000
	}
	if s.c.SQL.MaxLifetimeMills == 0 {
		s.c.SQL.MaxIdleTimeMills = 30000000000
	}
}

func (s *SQLSource) Config() *Config {
	return s.c
}

func (s *SQLSource) Size() (total int64, err error) {
	countSQL, err := s.BuildCountSQL()
	if err != nil {
		return
	}
	rows, err := s.Db.Query(countSQL)
	if err != nil {
		return
	}
	if rows.Next() {
		err = rows.Scan(&total)
		if err != nil {
			return
		}
	}
	return
}

func (s *SQLSource) Read(p []byte) (int, error) {
	return 2, nil
}

func (s *SQLSource) Close() error {
	return s.Db.Close()
}

func (s *SQLSource) String() string {
	return s.c.SQL.String()
}

func (s *SQLSource) validate() error {
	if s.Config().SQL.DbTable.Name == "" {
		return errors.New("dbTable.name is required")
	}
	c := s.Config().SQL
	if len(c.DbTable.Fields) != 0 && c.DbTable.Fields[c.DbTable.Id.Index] != c.DbTable.Id.Name {
		return errors.New("must contains primary key field")
	}
	if c.DbTable.Query == "" {
		if c.DbTable.Name == "" || len(c.DbTable.Fields) == 0 {
			return errors.New("name and fields must not be empty, when sql is empty")
		}
	}
	return nil
}

func (s *SQLSource) BuildCountSQL() (countSQL string, err error) {
	table := s.Config().SQL.DbTable
	if table.Count != "" {
		return table.Count, nil
	} else if table.Query != "" {
		stmt, err := sqlparser.Parse(table.Query)
		if err != nil {
			return "", err
		}
		expr := sqlparser.AliasedExpr{
			Expr: &sqlparser.FuncExpr{
				Name:  sqlparser.NewColIdent("count"),
				Exprs: sqlparser.SelectExprs{&sqlparser.AliasedExpr{Expr: sqlparser.NewStrVal([]byte("*"))}},
			},
			As: sqlparser.NewColIdent("total"),
		}
		selectStmt := stmt.(*sqlparser.Select)
		selectStmt.SelectExprs = sqlparser.SelectExprs{
			&expr,
		}
		buf := sqlparser.NewTrackedBuffer(nil)
		selectStmt.Format(buf)
		countSQL = buf.String()
	} else if table.Filter != "" {
		countSQL = fmt.Sprintf("SELECT COUNT(1) total FROM `%s` WHERE %s", table.Name, table.Filter)
	} else {
		countSQL = fmt.Sprintf("SELECT COUNT(1) total FROM `%s` WHERE 1 = 1", table.Name)
	}
	return
}

func (s *SQLSource) BuildQuerySQL(lastId string, batch int) string {
	t := s.Config().SQL.DbTable
	var stmt string
	if t.Query != "" {
		stmt = t.Query
	} else if t.Filter != "" {
		stmt = fmt.Sprintf("SELECT `%s` FROM `%s` WHERE %s", strings.Join(t.Fields, "`,`"), t.Name, t.Filter)
	} else {
		stmt = fmt.Sprintf("SELECT `%s` FROM `%s` WHERE 1 = 1", strings.Join(t.Fields, "`,`"), t.Name)
	}

	key := t.PrimaryKey()
	if lastId != "" {
		stmt += fmt.Sprintf(" AND %s > '%s'", key, lastId)
	}
	stmt += fmt.Sprintf(" ORDER BY %s ASC LIMIT %d", key, batch)
	return stmt
}

func (c *SQLConfig) String() string {
	return fmt.Sprintf("sql %s/%s", c.Endpoint, c.DbTable.Name)
}

func (t *SQLTable) PrimaryKey() string {
	if t.Id.Alias != "" {
		return t.Id.Alias
	}
	return fmt.Sprintf("`%s`", t.Id.Name)
}
