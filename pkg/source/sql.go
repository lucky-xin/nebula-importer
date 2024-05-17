package source

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var _ Source = (*s3Source)(nil)

type (
	SqlConfig struct {
		Endpoint         string   `yaml:"endpoint,omitempty"`
		DbName           string   `yaml:"dbName,omitempty"`
		DbTable          SqlTable `yaml:"dbTable"`
		Username         string   `yaml:"username,omitempty"`
		Password         string   `yaml:"password,omitempty"`
		UrlQuery         string   `yaml:"urlQuery,omitempty"`
		DriverName       string   `yaml:"driverName,omitempty"`
		MaxConnections   int      `yaml:"maxConnections,omitempty"`
		MaxIdleConns     int      `yaml:"maxIdleConns,omitempty"`
		MaxLifetimeMills int      `yaml:"maxLifetimeMills,omitempty"`
		MaxIdleTimeMills int      `yaml:"maxIdleTimeMills,omitempty"`
	}

	SqlTable struct {
		PrimaryKey string            `yaml:"primaryKey"`
		Name       string            `yaml:"name"`
		Fields     []string          `yaml:"fields"`
		FieldMap   map[string]string `yaml:"fieldMap,omitempty"`
		Filter     string            `yaml:"filter,omitempty"`
	}

	SqlSource struct {
		c    *Config
		size *int64
		Db   *sql.DB
	}
)

func newSQLSource(c *Config) Source {
	return &SqlSource{
		c: c,
	}
}

func (s *SqlSource) Name() string {
	return s.c.SQL.String()
}

func (s *SqlSource) Open() (err error) {
	if s.Db != nil {
		return
	}
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

	db, err := sql.Open(s.c.SQL.DriverName,
		fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", s.c.SQL.Username, s.c.SQL.Password, s.c.SQL.Endpoint, s.c.SQL.DbName, s.c.SQL.UrlQuery))
	if err != nil {
		return
	}
	db.SetMaxOpenConns(s.c.SQL.MaxConnections)
	db.SetMaxIdleConns(s.c.SQL.MaxIdleConns)
	// mysql default conn timeout=8h, should < mysql_timeout
	db.SetConnMaxIdleTime(time.Duration(s.c.SQL.MaxIdleTimeMills) * time.Millisecond)
	db.SetConnMaxLifetime(time.Duration(s.c.SQL.MaxLifetimeMills) * time.Millisecond)
	err = db.Ping()
	if err != nil {
		return
	}
	s.Db = db
	return
}

func (s *SqlSource) Config() *Config {
	return s.c
}

func (s *SqlSource) Size() (total int64, err error) {
	if s.Db == nil {
		err = s.Open()
		if err != nil {
			return
		}
	}
	if s.size != nil {
		total = *s.size
		return
	}
	totalSql := "SELECT count(1) total FROM " + s.Config().SQL.DbTable.Name + " WHERE 1 = 1"
	if s.Config().SQL.DbTable.Filter != "" {
		totalSql += " AND " + s.Config().SQL.DbTable.Filter
	}
	rows, err := s.Db.Query(totalSql)
	if err != nil {
		return
	}
	if rows.Next() {
		err = rows.Scan(&total)
		if err != nil {
			return
		}
	}
	s.size = &total
	return
}

func (s *SqlSource) Read(p []byte) (int, error) {
	return 2, nil
}

func (s *SqlSource) Close() error {
	return s.Db.Close()
}

func (s *SqlSource) String() string {
	return s.c.SQL.String()
}

func (c *SqlConfig) String() string {
	return fmt.Sprintf("sql %s/%s", c.Endpoint, c.DbTable.Name)
}
