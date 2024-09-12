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
	var env1 = map[string]string{
		"val": "",
	}
	program, err = expr.Compile("val != ''", expr.Env(env1), expr.AsBool())
	if err != nil {
		panic(err)
	}

	var envExec = map[string]string{
		"val": "1",
	}
	if out, err := expr.Run(program, envExec); err == nil {
		if b, ok := out.(bool); ok && !b {
			println(b)
		}
	}
}
