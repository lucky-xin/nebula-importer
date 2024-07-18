package source

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
	"testing"
	"time"
)

func TestSQLSource(t *testing.T) {
	println(time.Hour)
	println(time.Second * 30)
	sql := "SELECT count(1) total FROM t_vehicle_series_odx_address addr LEFT JOIN t_vehicle_series_ecu_las las ON las.vehicle_series_code = addr.vehicle_series_code WHERE 1 = 1"
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		// Do something with the err
	}

	countSql := sqlparser.AliasedExpr{
		Expr: &sqlparser.FuncExpr{
			Name: sqlparser.NewColIdent("count"),
			Exprs: sqlparser.SelectExprs{
				&sqlparser.AliasedExpr{
					Expr: sqlparser.NewStrVal([]byte("*")),
				},
			},
		},
		As: sqlparser.NewColIdent("total"),
	}

	// Otherwise do something with stmt
	selectStmt := stmt.(*sqlparser.Select)
	where := selectStmt.Where
	selectStmt.SelectExprs = sqlparser.SelectExprs{
		&countSql,
	}
	buf := sqlparser.NewTrackedBuffer(nil)
	selectStmt.Format(buf)
	countSQL := buf.String()
	println(countSQL)

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
