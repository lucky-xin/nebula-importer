package source

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
	"testing"
)

func TestSQLSource(t *testing.T) {
	sql := "SELECT CONCAT(addr.id,'-',las.id) id, addr.odx_ecu_name, addr.request_address, addr.response_address, addr.frame_type, addr.ecu_cn, las.las_code FROM t_vehicle_series_odx_address addr LEFT JOIN t_vehicle_series_ecu_las las ON las.vehicle_series_code = addr.vehicle_series_code WHERE 1 = 1"
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		// Do something with the err
	}

	// Otherwise do something with stmt
	selectStmt := stmt.(*sqlparser.Select)
	where := selectStmt.Where
	if where != nil {
		buffer := sqlparser.NewTrackedBuffer(func(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
			node.Format(buf)
		})
		where.Expr = &sqlparser.RangeCond{
			Left: &sqlparser.ColName{
				Name: sqlparser.NewColIdent("id"),
			},
			Operator: sqlparser.GreaterThanStr,
			From: &sqlparser.SQLVal{
				Type: sqlparser.StrVal,
				Val:  []byte("1"),
			},
			To: where.Expr,
		}
		where.Format(buffer)
		println(buffer.String())
	} else {
		sqlparser.NewWhere("where", &sqlparser.RangeCond{})
	}

}
