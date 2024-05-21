package picker

import (
	"strconv"
)

var (
	_ Converter = FunctionConverter{}
	_ Converter = FunctionStringConverter{}
)

type (
	FunctionConverter struct {
		Name string
	}
	FunctionStringConverter struct {
		Name string
	}

	FunctionDateTimeConverter struct {
		Name string
	}
)

func (fc FunctionConverter) Convert(v *Value) (*Value, error) {
	v.Val = getFuncValue(fc.Name, v.Val)
	return v, nil
}

func (fc FunctionStringConverter) Convert(v *Value) (*Value, error) {
	v.Val = getFuncValue(fc.Name, strconv.Quote(v.Val))
	return v, nil
}

func (fc FunctionDateTimeConverter) Convert(v *Value) (*Value, error) {
	if string(v.Val[len(v.Val)-1]) == "Z" {
		v.Val = v.Val[:len(v.Val)-1] + "+00:00"
	}
	v.Val = getFuncValue(fc.Name, strconv.Quote(v.Val))
	return v, nil
}

func getFuncValue(name, value string) string {
	return name + "(" + value + ")"
}
