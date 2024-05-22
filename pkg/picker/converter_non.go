package picker

var _ Converter = NonConverter{}

type (
	NonConverter            struct{}
	FunctionBoolConverter   struct{}
	FunctionIntConverter    struct{}
	FunctionFloatConverter  struct{}
	FunctionDoubleConverter struct{}
)

func (NonConverter) Convert(v *Value) (*Value, error) {
	return v, nil
}

func (FunctionBoolConverter) Convert(v *Value) (*Value, error) {
	if v.Val == "null" || v.Val == "NULL" || v.Val == "" {
		v.Val = "false"
	}
	return v, nil
}

func (FunctionIntConverter) Convert(v *Value) (*Value, error) {
	if v.Val == "null" || v.Val == "NULL" || v.Val == "" {
		v.Val = "0"
	}
	return v, nil
}

func (FunctionFloatConverter) Convert(v *Value) (*Value, error) {
	if v.Val == "null" || v.Val == "NULL" || v.Val == "" {
		v.Val = "0.0"
	}
	return v, nil
}

func (FunctionDoubleConverter) Convert(v *Value) (*Value, error) {
	if v.Val == "null" || v.Val == "NULL" || v.Val == "" {
		v.Val = "0.0"
	}
	return v, nil
}
