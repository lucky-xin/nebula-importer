package config

import (
	_ "github.com/go-sql-driver/mysql"
	"os"
	"testing"
)

func TestLocalSource(t *testing.T) {
	content, err := os.ReadFile("sql.v3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	c, err := FromBytes(content)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Build()
	if err != nil {
		t.Fatal(err)
	}
}
