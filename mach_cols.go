package mach

import "fmt"

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
	Len  int    `json:"length"`
}

type Columns []*Column

const (
	ColumnTypeNameInt16    = "int16"
	ColumnTypeNameInt32    = "int32"
	ColumnTypeNameInt64    = "int64"
	ColumnTypeNameDatetime = "datetime"
	ColumnTypeNameFloat    = "float"
	ColumnTypeNameDouble   = "double"
	ColumnTypeNameIPv4     = "ipv4"
	ColumnTypeNameIPv6     = "ipv6"
	ColumnTypeNameString   = "string"
	ColumnTypeNameBinary   = "binary"
)

func ColumnTypeString(typ ColumnType) (string, error) {
	switch typ {
	case 0: // MACH_DATA_TYPE_INT16
		return ColumnTypeNameInt16, nil
	case 1: // MACH_DATA_TYPE_INT32
		return ColumnTypeNameInt32, nil
	case 2: // MACH_DATA_TYPE_INT64
		return ColumnTypeNameInt64, nil
	case 3: // MACH_DATA_TYPE_DATETIME
		return ColumnTypeNameDatetime, nil
	case 4: // MACH_DATA_TYPE_FLOAT
		return ColumnTypeNameFloat, nil
	case 5: // MACH_DATA_TYPE_DOUBLE
		return ColumnTypeNameDouble, nil
	case 6: // MACH_DATA_TYPE_IPV4
		return ColumnTypeNameIPv4, nil
	case 7: // MACH_DATA_TYPE_IPV6
		return ColumnTypeNameIPv6, nil
	case 8: // MACH_DATA_TYPE_STRING
		return ColumnTypeNameString, nil
	case 9: // MACH_DATA_TYPE_BINARY
		return ColumnTypeNameBinary, nil
	default:
		return "", fmt.Errorf("unknown type %T", typ)
	}
}

func (cols Columns) Names() []string {
	rt := make([]string, len(cols))
	for i, c := range cols {
		rt[i] = c.Name
	}
	return rt
}

func (cols Columns) Types() []string {
	rt := make([]string, len(cols))
	for i, c := range cols {
		rt[i] = c.Type
	}
	return rt
}

func (cols Columns) Sizes() []int {
	rt := make([]int, len(cols))
	for i, c := range cols {
		rt[i] = c.Size
	}
	return rt
}

func (cols Columns) Lengths() []int {
	rt := make([]int, len(cols))
	for i, c := range cols {
		rt[i] = c.Len
	}
	return rt
}
