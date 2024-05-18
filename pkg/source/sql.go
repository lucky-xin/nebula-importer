package source

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var _ Source = (*s3Source)(nil)

type (
	SQLConfig struct {
		Endpoint         string   `yaml:"endpoint,omitempty"`
		DbName           string   `yaml:"dbName,omitempty"`
		DbTable          SQLTable `yaml:"dbTable"`
		Username         string   `yaml:"username,omitempty"`
		Password         string   `yaml:"password,omitempty"`
		UrlQuery         string   `yaml:"urlQuery,omitempty"`
		DriverName       string   `yaml:"driverName,omitempty"`
		MaxConnections   int      `yaml:"maxConnections,omitempty"`
		MaxIdleConns     int      `yaml:"maxIdleConns,omitempty"`
		MaxLifetimeMills int      `yaml:"maxLifetimeMills,omitempty"`
		MaxIdleTimeMills int      `yaml:"maxIdleTimeMills,omitempty"`
	}

	SQLTable struct {
		Name       string            `yaml:"name"`
		PrimaryKey string            `yaml:"primaryKey"`
		Fields     []string          `yaml:"fields"`
		FieldMap   map[string]string `yaml:"fieldMap,omitempty"`
		Filter     string            `yaml:"filter,omitempty"`
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
	c := s.Config().SQL
	if c.DbTable.Fields[0] != c.DbTable.PrimaryKey {
		return errors.New("primary key must be first field")
	}
	return nil
}

func (c *SQLConfig) String() string {
	return fmt.Sprintf("sql %s/%s", c.Endpoint, c.DbTable.Name)
}