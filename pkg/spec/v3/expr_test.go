package specv3

import (
	"github.com/expr-lang/expr"
	"testing"
)

func TestExpr(t *testing.T) {
	// 检查所有 tweets 的 Content 长度是否小于 240
	code := `(Record[0] != '' && Record[1] != '')`
	var env = map[string]any{
		"Record": Record{"", ""},
	}
	program, err := expr.Compile(code, expr.Env(env), expr.AsBool())
	out, err := expr.Run(program, env)
	if err != nil {
		panic(err)
	}
	println(out)
}
