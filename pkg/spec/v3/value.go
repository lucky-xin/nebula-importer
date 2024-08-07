package specv3

import (
	"strings"
)

const (
	dbNULL = "NULL"

	ValueTypeBool          ValueType = "BOOL"
	ValueTypeInt           ValueType = "INT"
	ValueTypeInt8          ValueType = "INT8"
	ValueTypeInt16         ValueType = "INT16"
	ValueTypeInt32         ValueType = "INT32"
	ValueTypeInt64         ValueType = "INT64"
	ValueTypeString        ValueType = "STRING"
	ValueTypeFixedString   ValueType = "FIXED_STRING"
	ValueTypeFloat         ValueType = "FLOAT"
	ValueTypeDouble        ValueType = "DOUBLE"
	ValueTypeDate          ValueType = "DATE"
	ValueTypeTime          ValueType = "TIME"
	ValueTypeDateTime      ValueType = "DATETIME"
	ValueTypeTimestamp     ValueType = "TIMESTAMP"
	ValueTypeGeo           ValueType = "GEOGRAPHY"
	ValueTypeGeoPoint      ValueType = "GEOGRAPHY(POINT)"
	ValueTypeGeoLineString ValueType = "GEOGRAPHY(LINESTRING)"
	ValueTypeGeoPolygon    ValueType = "GEOGRAPHY(POLYGON)"

	ValueTypeDefault = ValueTypeString
)

var (
	supportedPropValueTypes = map[ValueType]struct{}{
		ValueTypeBool:          {},
		ValueTypeInt:           {},
		ValueTypeInt8:          {},
		ValueTypeInt16:         {},
		ValueTypeInt32:         {},
		ValueTypeInt64:         {},
		ValueTypeString:        {},
		ValueTypeFixedString:   {},
		ValueTypeFloat:         {},
		ValueTypeDouble:        {},
		ValueTypeDate:          {},
		ValueTypeTime:          {},
		ValueTypeDateTime:      {},
		ValueTypeTimestamp:     {},
		ValueTypeGeo:           {},
		ValueTypeGeoPoint:      {},
		ValueTypeGeoLineString: {},
		ValueTypeGeoPolygon:    {},
	}

	supportedNodeIDValueTypes = map[ValueType]struct{}{
		ValueTypeInt:         {},
		ValueTypeInt64:       {},
		ValueTypeString:      {},
		ValueTypeFixedString: {},
	}
)

type ValueType string

func IsSupportedPropValueType(t ValueType) bool {
	_, ok := supportedPropValueTypes[ValueType(strings.ToUpper(t.String()))]
	return ok
}

func IsSupportedNodeIDValueType(t ValueType) bool {
	_, ok := supportedNodeIDValueTypes[ValueType(strings.ToUpper(t.String()))]
	return ok
}

func (t ValueType) Equal(vt ValueType) bool {
	if !IsSupportedPropValueType(t) || !IsSupportedPropValueType(vt) {
		return false
	}
	return strings.EqualFold(t.String(), vt.String())
}

func (t ValueType) String() string {
	return string(t)
}

func ValueTypeOf(text string) ValueType {
	if strings.HasPrefix(text, ValueTypeFixedString.String()) {
		return ValueTypeFixedString
	}
	return ValueType(strings.ToUpper(text))
}
