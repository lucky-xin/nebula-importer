package source

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var _ Source = (*s3Source)(nil)

type (
	MySQLConfig struct {
		Endpoint   string `yaml:"endpoint,omitempty"`
		DbName     string `yaml:"dbName,omitempty"`
		Username   string `yaml:"username,omitempty"`
		Password   string `yaml:"password,omitempty"`
		UrlQuery   string `yaml:"urlQuery,omitempty,default=charset=utf8mb4&parseTime=true&loc=Local"`
		DriverName string `yaml:"driverName,omitempty"`
	}

	mySQLSource struct {
		c *Config
	}
)

func newMySQLSource(c *Config) Source {
	return &mySQLSource{
		c: c,
	}
}

func (s *mySQLSource) Name() string {
	return s.c.S3.String()
}

func (s *mySQLSource) Open() (err error) {
	db, err := sql.Open(s.c.SQL.DriverName,
		fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", s.c.SQL.Username, s.c.SQL.Password, s.c.SQL.Endpoint, s.c.SQL.DbName, s.c.SQL.UrlQuery))
	if err != nil {
		return
	}
	query, err := db.Query("")
	if err != nil {
		return
	}
	println(query)
	return nil
}

func (s *mySQLSource) Config() *Config {
	return s.c
}

func (s *mySQLSource) Size() (int64, error) {
	return 2, nil
}

func (s *mySQLSource) Read(p []byte) (int, error) {
	return 2, nil
}

func (s *mySQLSource) Close() error {
	return nil
}

func (s *mySQLSource) String() string {
	return fmt.Sprintf("")
}
