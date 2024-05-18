package reader

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/lucky-xin/nebula-importer/pkg/source"
	"testing"
)

func TestSQLSource(t *testing.T) {
	c := &source.Config{
		SQL: &source.SQLConfig{
			DriverName: "mysql",
			Endpoint:   "",
			DbName:     "",
			Username:   "",
			Password:   "cCFzQHQkbyRuLmkubi50I0AhbXlzc6Ww",
			DbTable: source.SQLTable{
				PrimaryKey: "user_id",
				Name:       "sys_user",
				Fields: []string{
					"user_id",
					"username",
					"alias",
					"first_letter",
					"password",
					"phone",
					"dept_id",
					"lock_flag",
					"del_flag",
					"tenant_id",
					"email",
					"status",
					"effective_date",
					"expiration_date",
					"sex",
					"creator_id",
					"job",
					"avatar",
					"create_time",
					"update_time",
				},
			},
		},
	}
	sou, err := source.New(c)
	if err != nil {
		t.Fatal(err)
	}
	err = sou.Open()
	if err != nil {
		t.Fatal(err)
	}
	reader := NewSQLBatchRecordReader(sou.(*source.SQLSource), WithBatch(10))
	n, record, err := reader.ReadBatch()
	n, record, err = reader.ReadBatch()
	if err != nil {
		t.Fatal(err)
	}
	println(n)
	println(record)

}
