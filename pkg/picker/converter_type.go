package picker

import (
	"strconv"
	"strings"

	"github.com/lucky-xin/nebula-importer/pkg/errors"
	"github.com/lucky-xin/nebula-importer/pkg/utils"
)

var (
	_ Converter = TypeBoolConverter{}
	_ Converter = TypeIntConverter{}
	_ Converter = TypeFloatConverter{}
	_ Converter = TypeDoubleConverter{}
	_ Converter = TypeStringConverter{}
	_ Converter = TypeDateConverter{}
	_ Converter = TypeTimeConverter{}
	_ Converter = TypeDatetimeConverter{}
	_ Converter = TypeTimestampConverter{}
	_ Converter = TypeGeoConverter{}
	_ Converter = TypeGeoPointConverter{}
	_ Converter = TypeGeoLineStringConverter{}
	_ Converter = TypeGeoPolygonConverter{}
)

type (
	TypeBoolConverter = FunctionBoolConverter

	TypeIntConverter = FunctionIntConverter

	TypeFloatConverter = FunctionFloatConverter

	TypeDoubleConverter = FunctionDoubleConverter

	TypeStringConverter struct {
	}

	TypeDateConverter = FunctionDateConverter

	TypeTimeConverter = FunctionTimeConverter

	TypeDatetimeConverter = FunctionDateTimeConverter

	TypeTimestampConverter struct {
		fc  FunctionConverter
		fsc FunctionStringConverter
	}

	TypeGeoConverter = FunctionStringConverter

	TypeGeoPointConverter = FunctionStringConverter

	TypeGeoLineStringConverter = FunctionStringConverter

	TypeGeoPolygonConverter = FunctionStringConverter
)

func NewTypeConverter(t string) (Converter, error) {
	switch strings.ToUpper(t) {
	case "BOOL":
		return TypeBoolConverter{}, nil
	case "INT":
		return TypeIntConverter{}, nil
	case "INT8":
		return TypeIntConverter{}, nil
	case "INT16":
		return TypeIntConverter{}, nil
	case "INT32":
		return TypeIntConverter{}, nil
	case "INT64":
		return TypeIntConverter{}, nil
	case "FLOAT":
		return TypeFloatConverter{}, nil
	case "DOUBLE":
		return TypeDoubleConverter{}, nil
	case "STRING":
		return TypeStringConverter{}, nil
	case "FIXED_STRING":
		return TypeStringConverter{}, nil
	case "DATE":
		return TypeDateConverter{
			Name: "DATE",
		}, nil
	case "TIME":
		return TypeTimeConverter{
			Name: "TIME",
		}, nil
	case "DATETIME":
		return TypeDatetimeConverter{
			Name: "DATETIME",
		}, nil
	case "TIMESTAMP":
		return TypeTimestampConverter{
			fc: FunctionConverter{
				Name: "TIMESTAMP",
			},
			fsc: FunctionStringConverter{
				Name: "TIMESTAMP",
			},
		}, nil
	case "GEOGRAPHY":
		return TypeGeoConverter{
			Name: "ST_GeogFromText",
		}, nil
	case "GEOGRAPHY(POINT)":
		return TypeGeoPointConverter{
			Name: "ST_GeogFromText",
		}, nil
	case "GEOGRAPHY(LINESTRING)":
		return TypeGeoLineStringConverter{
			Name: "ST_GeogFromText",
		}, nil
	case "GEOGRAPHY(POLYGON)":
		return TypeGeoPolygonConverter{
			Name: "ST_GeogFromText",
		}, nil
	}
	return nil, errors.ErrUnsupportedValueType
}

func (TypeStringConverter) Convert(v *Value) (*Value, error) {
	v.Val = strconv.Quote(v.Val)
	return v, nil
}

func (tc TypeTimestampConverter) Convert(v *Value) (*Value, error) {
	if utils.IsUnsignedInteger(v.Val) {
		return tc.fc.Convert(v)
	}
	return tc.fsc.Convert(v)
}
