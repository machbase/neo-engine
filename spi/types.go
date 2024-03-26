package spi

import (
	"fmt"
	"net"
	"strconv"
	"time"
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

func ColumnTypeStringNative(typ ColumnType) string {
	switch typ {
	case Int16ColumnType:
		return "short"
	case Uint16ColumnType:
		return "ushort"
	case Int32ColumnType:
		return "integer"
	case Uint32ColumnType:
		return "uinteger"
	case Int64ColumnType:
		return "long"
	case Uint64ColumnType:
		return "ulong"
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

func ParseColumnValue(str string, ctype ColumnType, tz *time.Location, timeformat string) (any, error) {
	switch ctype {
	case Int16ColumnType:
		return strconv.ParseInt(str, 10, 16)
	case Uint16ColumnType:
		return strconv.ParseUint(str, 10, 16)
	case Int32ColumnType:
		return strconv.ParseInt(str, 10, 32)
	case Uint32ColumnType:
		return strconv.ParseUint(str, 10, 32)
	case Int64ColumnType:
		return strconv.ParseInt(str, 10, 64)
	case Uint64ColumnType:
		return strconv.ParseUint(str, 10, 64)
	case Float32ColumnType:
		return strconv.ParseFloat(str, 32)
	case Float64ColumnType:
		return strconv.ParseFloat(str, 64)
	case VarcharColumnType:
		return str, nil
	case TextColumnType:
		return str, nil
	case ClobColumnType:
		return str, nil
	case BlobColumnType:
		return str, nil
	case BinaryColumnType:
		return str, nil
	case DatetimeColumnType:
		switch timeformat {
		case "ns":
			v, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return nil, err
			}
			return time.Unix(0, v), nil
		case "ms":
			v, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return nil, err
			}
			return time.Unix(0, v*int64(time.Millisecond)), nil
		case "us":
			v, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return nil, err
			}
			return time.Unix(0, v*int64(time.Microsecond)), nil
		case "s":
			v, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return nil, err
			}
			return time.Unix(v, 0), nil
		default:
			return time.ParseInLocation(timeformat, str, tz)
		}
	case IpV4ColumnType:
		if ip := net.ParseIP(str); ip != nil {
			return ip, nil
		} else {
			return nil, fmt.Errorf("unable to parse as ip address %s", str)
		}
	case IpV6ColumnType:
		if ip := net.ParseIP(str); ip != nil {
			return ip, nil
		} else {
			return nil, fmt.Errorf("unable to parse as ip address %s", str)
		}
	default:
		return nil, fmt.Errorf("unknown column type %d", ctype)
	}
}
