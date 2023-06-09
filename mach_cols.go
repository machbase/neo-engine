package mach

import (
	"fmt"

	spi "github.com/machbase/neo-spi"
)

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
		return spi.ColumnBufferTypeInt16, nil
	case 1: // MACH_DATA_TYPE_INT32
		return spi.ColumnBufferTypeInt32, nil
	case 2: // MACH_DATA_TYPE_INT64
		return spi.ColumnBufferTypeInt64, nil
	case 3: // MACH_DATA_TYPE_DATETIME
		return spi.ColumnBufferTypeDatetime, nil
	case 4: // MACH_DATA_TYPE_FLOAT
		return spi.ColumnBufferTypeFloat, nil
	case 5: // MACH_DATA_TYPE_DOUBLE
		return spi.ColumnBufferTypeDouble, nil
	case 6: // MACH_DATA_TYPE_IPV4
		return spi.ColumnBufferTypeIPv4, nil
	case 7: // MACH_DATA_TYPE_IPV6
		return spi.ColumnBufferTypeIPv6, nil
	case 8: // MACH_DATA_TYPE_STRING
		return spi.ColumnBufferTypeString, nil
	case 9: // MACH_DATA_TYPE_BINARY
		return spi.ColumnBufferTypeBinary, nil
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
