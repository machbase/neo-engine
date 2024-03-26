package mach

import (
	"github.com/machbase/neo-engine/spi"
)

type ColumnType int

const (
	Int16ColumnType   ColumnType = iota + 0
	Int32ColumnType              = 1
	Int64ColumnType              = 2
	TimeColumnType               = 3
	Float32ColumnType            = 4
	Float64ColumnType            = 5
	IPv4ColumnType               = 6
	Ipv6ColumnType               = 7
	StringColumnType             = 8
	BinaryColumnType             = 9
)

type ColumnSize int

// * DDL: 1-255
// * ALTER SYSTEM: 256-511
// * SELECT: 512
// * INSERT: 513
// * DELETE: 514-517
// * INSERT_SELECT: 518
// * UPDATE: 519
// * EXEC_ROLLUP: 521-523
type StmtType int

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
		return "", spi.ErrDatabaseUnsupportedType("ColumnTypeString", int(typ))
	}
}

func (typ StmtType) IsSelect() bool {
	return typ == 512
}

func (typ StmtType) IsDDL() bool {
	return typ >= 1 && typ <= 255
}

func (typ StmtType) IsAlterSystem() bool {
	return typ >= 256 && typ <= 511
}

func (typ StmtType) IsInsert() bool {
	return typ == 513 || typ == 516
}

func (typ StmtType) IsDelete() bool {
	return typ >= 514 && typ <= 515
}

func (typ StmtType) IsUpdate() bool {
	return typ == 517
}
