package mach

import (
	"database/sql"
	"fmt"
	"net"
	"time"
	"unsafe"

	"github.com/machbase/dbms-mach-go/valconv"
	"github.com/pkg/errors"
)

type Row struct {
	ok     bool
	err    error
	values []any
}

func (row *Row) Err() error {
	return row.err
}

func (row *Row) Values() []any {
	return row.values
}

func (row *Row) Scan(cols ...any) error {
	if row.err != nil {
		return row.err
	}
	if !row.ok {
		return sql.ErrNoRows
	}
	for i := range cols {
		if i >= len(row.values) {
			return fmt.Errorf("column %d is out of range %d", i, len(row.values))
		}
		var isNull bool
		switch v := row.values[i].(type) {
		case *int16:
			valconv.Int16ToAny(*v, cols[i], &isNull)
		case *int32:
			valconv.Int32ToAny(*v, cols[i], &isNull)
		case *int64:
			valconv.Int64ToAny(*v, cols[i], &isNull)
		case *time.Time:
			valconv.DateTimeToAny(*v, cols[i])
		case *float32:
			valconv.Float32ToAny(*v, cols[i])
		case *float64:
			valconv.Float64ToAny(*v, cols[i])
		case *net.IP:
			valconv.IPToAny(*v, cols[i])
		case *string:
			valconv.StringToAny(*v, cols[i], &isNull)
		case []byte:
			valconv.BytesToAny(v, cols[i])
		}
		if isNull {
			cols[i] = nil
		}
	}
	return nil
}

type Rows struct {
	handle     unsafe.Pointer
	stmt       unsafe.Pointer
	stmtType   int
	sqlText    string
	timeFormat string
}

func (rows *Rows) Close() {
	if rows.stmt != nil {
		machFreeStmt(rows.handle, rows.stmt)
		rows.stmt = nil
	}
	rows.sqlText = ""
}

func (rows *Rows) IsFetchable() bool {
	return isFetchableStmtType(rows.stmtType)
}

func (rows *Rows) ResultString(nrows int64) string {
	var verb = ""

	if rows.stmtType >= 1 && rows.stmtType <= 255 {
		return "DDL executed"
	} else if rows.stmtType >= 256 && rows.stmtType <= 511 {
		// "ALTER SYSTEM"
		return "system altered"
	} else if rows.stmtType == 512 {
		verb = "selected"
	} else if rows.stmtType == 513 {
		verb = "inserted"
	} else if rows.stmtType == 514 || rows.stmtType == 515 {
		verb = "deleted"
	} else if rows.stmtType == 516 {
		verb = "inserted and selected"
	} else if rows.stmtType == 517 {
		verb = "updated"
	} else {
		return "unknown"
	}
	if nrows == 0 {
		return fmt.Sprintf("no row %s", verb)
	} else if nrows == 1 {
		return fmt.Sprintf("1 row %s", verb)
	} else {
		return fmt.Sprintf("%d rows %s", nrows, verb)
	}
}

func (rows *Rows) AffectedRows() (int64, error) {
	if rows.IsFetchable() {
		return 0, nil
	}
	return machEffectRows(rows.stmt)
}

func (rows *Rows) SetTimeFormat(format string) {
	rows.timeFormat = format
}

func (rows *Rows) ColumnNames() ([]string, error) {
	count, err := machColumnCount(rows.stmt)
	if err != nil {
		return nil, err
	}
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i], err = machColumnName(rows.stmt, i)
		if err != nil {
			names[i] = fmt.Sprintf("col%02d", i)
		}
	}
	return names, nil
}

func (rows *Rows) ColumnTypes() ([]string, error) {
	count, err := machColumnCount(rows.stmt)
	if err != nil {
		return nil, err
	}
	types := make([]string, count)
	for i := 0; i < count; i++ {
		typ, _, err := machColumnType(rows.stmt, i)
		if err != nil {
			return nil, errors.Wrap(err, "ColumnTypes")
		}
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			types[i] = "int16"
		case 1: // MACH_DATA_TYPE_INT32
			types[i] = "int32"
		case 2: // MACH_DATA_TYPE_INT64
			types[i] = "int64"
		case 3: // MACH_DATA_TYPE_DATETIME
			types[i] = "time"
		case 4: // MACH_DATA_TYPE_FLOAT
			types[i] = "float32"
		case 5: // MACH_DATA_TYPE_DOUBLE
			types[i] = "float64"
		case 6: // MACH_DATA_TYPE_IPV4
			types[i] = "ipv4"
		case 7: // MACH_DATA_TYPE_IPV6
			types[i] = "ipv6"
		case 8: // MACH_DATA_TYPE_STRING
			types[i] = "string"
		case 9: // MACH_DATA_TYPE_BINARY
			types[i] = "binary"
		default:
			return nil, fmt.Errorf("Fetch unsupported type %T", typ)
		}
	}
	return types, nil
}

// internal use only from machrpcserver
func (rows *Rows) Fetch() ([]any, bool, error) {
	// select 가 아니면 fetch를 진행하지 않는다.
	if !rows.IsFetchable() {
		return nil, false, sql.ErrNoRows
	}

	next, err := machFetch(rows.stmt)
	if err != nil {
		return nil, next, errors.Wrap(err, "Fetch")
	}
	if !next {
		return nil, false, nil
	}

	colCount, err := machColumnCount(rows.stmt)
	if err != nil {
		return nil, next, err
	}

	values := make([]any, colCount)
	for i := range values {
		typ, _, err := machColumnType(rows.stmt, i)
		if err != nil {
			return nil, next, errors.Wrap(err, "Fetch")
		}
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			values[i] = new(int)
		case 1: // MACH_DATA_TYPE_INT32
			values[i] = new(int32)
		case 2: // MACH_DATA_TYPE_INT64
			values[i] = new(int64)
		case 3: // MACH_DATA_TYPE_DATETIME
			switch rows.timeFormat {
			case "epoch":
				values[i] = new(int64)
			case "":
				values[i] = new(time.Time)
			default:
				values[i] = new(string)
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			values[i] = new(float32)
		case 5: // MACH_DATA_TYPE_DOUBLE
			values[i] = new(float64)
		case 6: // MACH_DATA_TYPE_IPV4
			values[i] = new(net.IP)
		case 7: // MACH_DATA_TYPE_IPV6
			values[i] = new(net.IP)
		case 8: // MACH_DATA_TYPE_STRING
			values[i] = new(string)
		case 9: // MACH_DATA_TYPE_BINARY
			values[i] = []byte{}
		default:
			return nil, next, fmt.Errorf("Fetch unsupported type %T", typ)
		}
	}
	err = scan(rows.stmt, values...)
	if err != nil {
		return nil, next, errors.Wrap(err, "Fetch")
	}
	return values, next, nil
}

func (rows *Rows) Next() bool {
	// select 가 아니면
	if !rows.IsFetchable() {
		return false
	}

	next, err := machFetch(rows.stmt)
	if err != nil {
		return false
	}
	return next
}

func (rows *Rows) Scan(cols ...any) error {
	return scan(rows.stmt, cols...)
}

func scan(stmt unsafe.Pointer, cols ...any) error {
	for i, c := range cols {
		typ, _ /*size*/, err := machColumnType(stmt, i)
		if err != nil {
			return errors.Wrap(err, "Scan")
		}
		var isNull bool
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			if v, _ /*nonNull*/, err := machColumnDataInt16(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = valconv.Int16ToAny(v, c, &isNull); err != nil {
					return err
				}
			}
		case 1: // MACH_DATA_TYPE_INT32
			if v, _ /*nonNull*/, err := machColumnDataInt32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = valconv.Int32ToAny(v, c, &isNull); err != nil {
					return err
				}
			}
		case 2: // MACH_DATA_TYPE_INT64
			if v, _ /*nonNull*/, err := machColumnDataInt64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = valconv.Int64ToAny(v, c, &isNull); err != nil {
					return err
				}
			}
		case 3: // MACH_DATA_TYPE_DATETIME
			if v, _ /*nonNull*/, err := machColumnDataDateTime(stmt, i); err != nil {
				return errors.Wrap(err, "Scan datetime")
			} else {
				if err = valconv.DateTimeToAny(v, c); err != nil {
					return err
				}
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			if v, _ /*nonNull*/, err := machColumnDataFloat32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = valconv.Float32ToAny(v, c); err != nil {
					return err
				}
			}
		case 5: // MACH_DATA_TYPE_DOUBLE
			if v, _ /*nonNull*/, err := machColumnDataFloat64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = valconv.Float64ToAny(v, c); err != nil {
					return err
				}
			}
		case 6: // MACH_DATA_TYPE_IPV4
			if v, _ /*nonNull*/, err := machColumnDataIPv4(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else {
				if err = valconv.IPToAny(v, c); err != nil {
					return err
				}
			}
		case 7: // MACH_DATA_TYPE_IPV6
			if v, _ /*nonNull*/, err := machColumnDataIPv6(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else {
				if err = valconv.IPToAny(v, c); err != nil {
					return err
				}
			}
		case 8: // MACH_DATA_TYPE_STRING
			if v, _ /*nonNull*/, err := machColumnDataString(stmt, i); err != nil {
				return errors.Wrap(err, "Scan string")
			} else {
				if err = valconv.StringToAny(v, c, &isNull); err != nil {
					return err
				}
			}
		case 9: // MACH_DATA_TYPE_BINARY
			if v, _ /*nonNull*/, err := machColumnDataBinary(stmt, i); err != nil {
				return errors.Wrap(err, "Scan binary")
			} else {
				if err = valconv.BytesToAny(v, c); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("MachGetColumnData unsupported type %T", c)
		}
		if isNull {
			cols[i] = nil
		}
	}
	return nil
}
