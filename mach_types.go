package mach

import (
	"fmt"
)

// 0: Log Table, 1: Fixed Table, 3: Volatile Table,
// 4: Lookup Table, 5: KeyValue Table, 6: Tag Table
type TableType int

const (
	LogTableType      TableType = iota + 0
	FixedTableType              = 1
	VolatileTableType           = 3
	LookupTableType             = 4
	KeyValueTableType           = 5
	TagTableType                = 6
)

func (t TableType) String() string {
	switch t {
	case LogTableType:
		return "LogTable"
	case FixedTableType:
		return "FixedTable"
	case VolatileTableType:
		return "VolatileTable"
	case LookupTableType:
		return "LookupTable"
	case KeyValueTableType:
		return "KeyValueTable"
	case TagTableType:
		return "TagTable"
	default:
		return "Undefined"
	}
}

type IndexType int

func IndexTypeString(typ IndexType) string {
	switch typ {
	case 6:
		return "BITMAP (LSM)"
	case 8:
		return "REDBLACK"
	case 9:
		return "KEYWORD (LSM)"
	case 11:
		return "TAG"
	default:
		return fmt.Sprintf("undef-%d", typ)
	}
}

type ColumnType int

const (
	Int16ColumnType    ColumnType = iota + 4
	Uint16ColumnType              = 104
	Int32ColumnType               = 8
	Uint32ColumnType              = 108
	Int64ColumnType               = 12
	Uint64ColumnType              = 112
	Float32ColumnType             = 16
	Float64ColumnType             = 20
	VarcharColumnType             = 5
	TextColumnType                = 49
	ClobColumnType                = 53
	BlobColumnType                = 57
	BinaryColumnType              = 97
	DatetimeColumnType            = 6
	IpV4ColumnType                = 32
	IpV6ColumnType                = 36
	JsonColumnType                = 61
)

// ColumnTypeString converts ColumnType into string.
func ColumnTypeString(typ ColumnType) string {
	switch typ {
	case Int16ColumnType:
		return "int16"
	case Uint16ColumnType:
		return "uint16"
	case Int32ColumnType:
		return "int32"
	case Uint32ColumnType:
		return "uint32"
	case Int64ColumnType:
		return "int64"
	case Uint64ColumnType:
		return "uint64"
	case Float32ColumnType:
		return "float"
	case Float64ColumnType:
		return "double"
	case VarcharColumnType:
		return "varchar"
	case TextColumnType:
		return "text"
	case ClobColumnType:
		return "clob"
	case BlobColumnType:
		return "blob"
	case BinaryColumnType:
		return "binary"
	case DatetimeColumnType:
		return "datetime"
	case IpV4ColumnType:
		return "ipv4"
	case IpV6ColumnType:
		return "ipv6"
	case JsonColumnType:
		return "json"
	default:
		return fmt.Sprintf("undef-%d", typ)
	}
}

const (
	DB_COLUMN_TYPE_SHORT    = "short"
	DB_COLUMN_TYPE_USHORT   = "ushort"
	DB_COLUMN_TYPE_INTEGER  = "integer"
	DB_COLUMN_TYPE_UINTEGER = "uinteger"
	DB_COLUMN_TYPE_LONG     = "long"
	DB_COLUMN_TYPE_ULONG    = "ulong"
	DB_COLUMN_TYPE_FLOAT    = "float"
	DB_COLUMN_TYPE_DOUBLE   = "double"
	DB_COLUMN_TYPE_DATETIME = "datetime"
	DB_COLUMN_TYPE_VARCHAR  = "varchar"
	DB_COLUMN_TYPE_IPV4     = "ipv4"
	DB_COLUMN_TYPE_IPV6     = "ipv6"
	DB_COLUMN_TYPE_TEXT     = "text"
	DB_COLUMN_TYPE_CLOB     = "clob"
	DB_COLUMN_TYPE_BLOB     = "blob"
	DB_COLUMN_TYPE_BINARY   = "binary"
	DB_COLUMN_TYPE_JSON     = "json"
)

// ColumnTypeStringNative converts ColumnType into native type string
func ColumnTypeStringNative(typ ColumnType) string {
	switch typ {
	case Int16ColumnType:
		return DB_COLUMN_TYPE_SHORT
	case Uint16ColumnType:
		return DB_COLUMN_TYPE_USHORT
	case Int32ColumnType:
		return DB_COLUMN_TYPE_INTEGER
	case Uint32ColumnType:
		return DB_COLUMN_TYPE_UINTEGER
	case Int64ColumnType:
		return DB_COLUMN_TYPE_LONG
	case Uint64ColumnType:
		return DB_COLUMN_TYPE_ULONG
	case Float32ColumnType:
		return DB_COLUMN_TYPE_FLOAT
	case Float64ColumnType:
		return DB_COLUMN_TYPE_DOUBLE
	case VarcharColumnType:
		return DB_COLUMN_TYPE_VARCHAR
	case TextColumnType:
		return DB_COLUMN_TYPE_TEXT
	case ClobColumnType:
		return DB_COLUMN_TYPE_CLOB
	case BlobColumnType:
		return DB_COLUMN_TYPE_BLOB
	case BinaryColumnType:
		return DB_COLUMN_TYPE_BINARY
	case DatetimeColumnType:
		return DB_COLUMN_TYPE_DATETIME
	case IpV4ColumnType:
		return DB_COLUMN_TYPE_IPV4
	case IpV6ColumnType:
		return DB_COLUMN_TYPE_IPV6
	case JsonColumnType:
		return DB_COLUMN_TYPE_JSON
	default:
		return fmt.Sprintf("undef-%d", typ)
	}
}

const (
	ColumnFlagTagName    = 0x08000000
	ColumnFlagBasetime   = 0x01000000
	ColumnFlagSummarized = 0x02000000
	ColumnFlagMetaColumn = 0x04000000
)

func ColumnFlagString(flag int) string {
	switch flag {
	case ColumnFlagTagName:
		return "tag name"
	case ColumnFlagBasetime:
		return "basetime"
	case ColumnFlagSummarized:
		return "summarized"
	case ColumnFlagMetaColumn:
		return "meta"
	default:
		return ""
	}
}

const (
	ColumnBufferTypeInt16    = "int16"
	ColumnBufferTypeInt32    = "int32"
	ColumnBufferTypeInt64    = "int64"
	ColumnBufferTypeDatetime = "datetime"
	ColumnBufferTypeFloat    = "float"
	ColumnBufferTypeDouble   = "double"
	ColumnBufferTypeIPv4     = "ipv4"
	ColumnBufferTypeIPv6     = "ipv6"
	ColumnBufferTypeString   = "string"
	ColumnBufferTypeBinary   = "binary"
	ColumnBufferTypeBoolean  = "bool"
	ColumnBufferTypeByte     = "int8"
)

func ColumnBufferType(typ ColumnType) string {
	switch typ {
	case Int16ColumnType:
		return ColumnBufferTypeInt16
	case Uint16ColumnType:
		return ColumnBufferTypeInt16
	case Int32ColumnType:
		return ColumnBufferTypeInt32
	case Uint32ColumnType:
		return ColumnBufferTypeInt32
	case Int64ColumnType:
		return ColumnBufferTypeInt64
	case Uint64ColumnType:
		return ColumnBufferTypeInt64
	case Float32ColumnType:
		return ColumnBufferTypeFloat
	case Float64ColumnType:
		return ColumnBufferTypeDouble
	case VarcharColumnType:
		return ColumnBufferTypeString
	case TextColumnType:
		return ColumnBufferTypeString
	case ClobColumnType:
		return ColumnBufferTypeBinary
	case BlobColumnType:
		return ColumnBufferTypeBinary
	case BinaryColumnType:
		return ColumnBufferTypeBinary
	case DatetimeColumnType:
		return ColumnBufferTypeDatetime
	case IpV4ColumnType:
		return ColumnBufferTypeIPv4
	case IpV6ColumnType:
		return ColumnBufferTypeIPv6
	case JsonColumnType:
		return ColumnBufferTypeString
	default:
		return "undef-buffer"
	}
}

type NativeColumnType int

const (
	Int16NativeColumnType   NativeColumnType = iota + 0
	Int32NativeColumnType                    = 1
	Int64NativeColumnType                    = 2
	TimeNativeColumnType                     = 3
	Float32NativeColumnType                  = 4
	Float64NativeColumnType                  = 5
	IPv4NativeColumnType                     = 6
	Ipv6NativeColumnType                     = 7
	StringNativeColumnType                   = 8
	BinaryNativeColumnType                   = 9
)

type NativeColumnSize int

func NativeColumnTypeString(typ NativeColumnType) (string, error) {
	switch typ {
	case 0: // MACH_DATA_TYPE_INT16
		return ColumnBufferTypeInt16, nil
	case 1: // MACH_DATA_TYPE_INT32
		return ColumnBufferTypeInt32, nil
	case 2: // MACH_DATA_TYPE_INT64
		return ColumnBufferTypeInt64, nil
	case 3: // MACH_DATA_TYPE_DATETIME
		return ColumnBufferTypeDatetime, nil
	case 4: // MACH_DATA_TYPE_FLOAT
		return ColumnBufferTypeFloat, nil
	case 5: // MACH_DATA_TYPE_DOUBLE
		return ColumnBufferTypeDouble, nil
	case 6: // MACH_DATA_TYPE_IPV4
		return ColumnBufferTypeIPv4, nil
	case 7: // MACH_DATA_TYPE_IPV6
		return ColumnBufferTypeIPv6, nil
	case 8: // MACH_DATA_TYPE_STRING
		return ColumnBufferTypeString, nil
	case 9: // MACH_DATA_TYPE_BINARY
		return ColumnBufferTypeBinary, nil
	default:
		return "", ErrDatabaseUnsupportedType("ColumnTypeString", int(typ))
	}
}

// * DDL: 1-255
// * ALTER SYSTEM: 256-511
// * SELECT: 512
// * INSERT: 513
// * DELETE: 514-518
// * INSERT_SELECT: 519
// * UPDATE: 520
// * EXEC_ROLLUP: 522-524

type StmtType int

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
	return typ == 513
}

func (typ StmtType) IsDelete() bool {
	return typ >= 514 && typ <= 518
}

func (typ StmtType) IsInsertSelect() bool {
	return typ == 519
}

func (typ StmtType) IsUpdate() bool {
	return typ == 520
}

func (typ StmtType) IsExecRollup() bool {
	return typ >= 522 && typ <= 524
}
