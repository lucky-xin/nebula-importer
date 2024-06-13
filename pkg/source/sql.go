package source

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
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
		MaxConnections   int      `yaml:"maxConnections,omitempty" json:"maxConnections,omitempty,optional,default=2000"`
		MaxIdleConns     int      `yaml:"maxIdleConns,omitempty" json:"maxIdleConns,omitempty,optional,default=50"`
		MaxLifetimeMills int      `yaml:"maxLifetimeMills,omitempty" json:"maxLifetimeMills,omitempty,optional,default=3600000"`
		MaxIdleTimeMills int      `yaml:"maxIdleTimeMills,omitempty" json:"maxIdleTimeMills,omitempty,optional,default=600000"`
	}

	SQLTable struct {
		PrimaryKey string   `yaml:"primaryKey" json:"primaryKey"`
		Name       string   `yaml:"name,omitempty" json:"name,omitempty,optional"`
		Fields     []string `yaml:"fields,omitempty" json:"fields,omitempty,optional"`
		SQL        string   `yaml:"sql,omitempty" json:"sql,omitempty,optional"`
		Filter     string   `yaml:"filter,omitempty" json:"filter,omitempty,optional"`
	}

	SQLSource struct {
		c  *Config
		Db *sql.DB
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
	s.setDefaultConfig()
	err = s.validate()
	if err != nil {
		return
	}
	conf := s.c.SQL
	db, err := sql.Open(conf.DriverName,
		fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", conf.Username, conf.Password, conf.Endpoint, conf.DbName, conf.UrlQuery))
	if err != nil {
		return
	}
	db.SetMaxOpenConns(conf.MaxConnections)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	// mysql default conn timeout=8h, should < mysql_timeout
	db.SetConnMaxIdleTime(time.Duration(conf.MaxIdleTimeMills) * time.Millisecond)
	db.SetConnMaxLifetime(time.Duration(conf.MaxLifetimeMills) * time.Millisecond)
	err = db.Ping()
	if err != nil {
		return
	}
	s.Db = db
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
		s.c.SQL.MaxConnections = 2000
	}
	if s.c.SQL.MaxIdleConns == 0 {
		s.c.SQL.MaxIdleConns = 2000
	}
	if s.c.SQL.MaxLifetimeMills == 0 {
		s.c.SQL.MaxLifetimeMills = 3600000
	}
	if s.c.SQL.MaxLifetimeMills == 0 {
		s.c.SQL.MaxIdleTimeMills = 6000
	}
}

func (s *SQLSource) Config() *Config {
	return s.c
}

func (s *SQLSource) Size() (total int64, err error) {
	table := s.Config().SQL.DbTable
	var countSQL string
	if table.SQL != "" {
		stmt, err := sqlparser.Parse(table.SQL)
		if err != nil {
			return 0, err
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
	} else {
		countSQL = "SELECT count(1) total FROM " + s.Config().SQL.DbTable.Name + " WHERE 1 = 1"
		if s.Config().SQL.DbTable.Filter != "" {
			countSQL += " AND " + s.Config().SQL.DbTable.Filter
		}
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
	if len(c.DbTable.Fields) != 0 && c.DbTable.Fields[0] != c.DbTable.PrimaryKey {
		return errors.New("primary key must be first field")
	}
	if c.DbTable.SQL == "" {
		if c.DbTable.Name == "" || len(c.DbTable.Fields) == 0 {
			return errors.New("name and fields must not be empty,when sql is empty")
		}
	}
	return nil
}

func (c *SQLConfig) String() string {
	return fmt.Sprintf("sql %s/%s", c.Endpoint, c.DbTable.Name)
}
