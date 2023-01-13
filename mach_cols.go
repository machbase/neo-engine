package mach

import "fmt"

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
	Len  int    `json:"length"`
}

type Columns []*Column

func ColumnTypeString(typ ColumnType) (string, error) {
	switch typ {
	case 0: // MACH_DATA_TYPE_INT16
		return "int16", nil
	case 1: // MACH_DATA_TYPE_INT32
		return "int32", nil
	case 2: // MACH_DATA_TYPE_INT64
		return "int64", nil
	case 3: // MACH_DATA_TYPE_DATETIME
		return "datetime", nil
	case 4: // MACH_DATA_TYPE_FLOAT
		return "float", nil
	case 5: // MACH_DATA_TYPE_DOUBLE
		return "double", nil
	case 6: // MACH_DATA_TYPE_IPV4
		return "ipv4", nil
	case 7: // MACH_DATA_TYPE_IPV6
		return "ipv6", nil
	case 8: // MACH_DATA_TYPE_STRING
		return "string", nil
	case 9: // MACH_DATA_TYPE_BINARY
		return "binary", nil
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
