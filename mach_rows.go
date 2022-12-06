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
		switch v := row.values[i].(type) {
		case *int16:
			valconv.Int16ToAny(*v, cols[i])
		case *int32:
			valconv.Int32ToAny(*v, cols[i])
		case *int64:
			valconv.Int64ToAny(*v, cols[i])
		case *time.Time:
			valconv.DateTimeToAny(*v, cols[i])
		case *float32:
			valconv.Float32ToAny(*v, cols[i])
		case *float64:
			valconv.Float64ToAny(*v, cols[i])
		case *net.IP:
			valconv.IPToAny(*v, cols[i])
		case *string:
			valconv.StringToAny(*v, cols[i])
		case []byte:
			valconv.BytesToAny(v, cols[i])
		}
	}
	return nil
}

type Rows struct {
	stmt    unsafe.Pointer
	sqlText string
}

func (this *Rows) Close() {
	if this.stmt != nil {
		machFreeStmt(this.stmt)
		this.stmt = nil
	}
	this.sqlText = ""
}

// internal use only from machrpcserver
func (this *Rows) Fetch() ([]any, bool, error) {
	next, err := machFetch(this.stmt)
	if err != nil {
		return nil, next, errors.Wrap(err, "Fetch")
	}
	if !next {
		return nil, false, nil
	}

	colCount, err := machColumnCount(this.stmt)
	if err != nil {
		return nil, next, err
	}

	values := make([]any, colCount)
	for i := range values {
		typ, _, err := machColumnType(this.stmt, i)
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
			values[i] = new(time.Time)
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
	err = scan(this.stmt, values...)
	if err != nil {
		return nil, next, errors.Wrap(err, "Fetch")
	}
	return values, next, nil
}

func (this *Rows) Next() bool {
	next, err := machFetch(this.stmt)
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
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			if v, err := machColumnDataInt16(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = valconv.Int16ToAny(v, c); err != nil {
					return err
				}
			}
		case 1: // MACH_DATA_TYPE_INT32
			if v, err := machColumnDataInt32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = valconv.Int32ToAny(v, c); err != nil {
					return err
				}
			}
		case 2: // MACH_DATA_TYPE_INT64
			if v, err := machColumnDataInt64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = valconv.Int64ToAny(v, c); err != nil {
					return err
				}
			}
		case 3: // MACH_DATA_TYPE_DATETIME
			if v, err := machColumnDataDateTime(stmt, i); err != nil {
				return errors.Wrap(err, "Scan datetime")
			} else {
				if err = valconv.DateTimeToAny(v, c); err != nil {
					return err
				}
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			if v, err := machColumnDataFloat32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = valconv.Float32ToAny(v, c); err != nil {
					return err
				}
			}
		case 5: // MACH_DATA_TYPE_DOUBLE
			if v, err := machColumnDataFloat64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = valconv.Float64ToAny(v, c); err != nil {
					return err
				}
			}
		case 6: // MACH_DATA_TYPE_IPV4
			if v, err := machColumnDataIPv4(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else {
				if err = valconv.IPToAny(v, c); err != nil {
					return err
				}
			}
		case 7: // MACH_DATA_TYPE_IPV6
			if v, err := machColumnDataIPv6(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else {
				if err = valconv.IPToAny(v, c); err != nil {
					return err
				}
			}
		case 8: // MACH_DATA_TYPE_STRING
			if v, err := machColumnDataString(stmt, i); err != nil {
				return errors.Wrap(err, "Scan string")
			} else {
				if err = valconv.StringToAny(v, c); err != nil {
					return err
				}
			}
		case 9: // MACH_DATA_TYPE_BINARY
			if v, err := machColumnDataBinary(stmt, i); err != nil {
				return errors.Wrap(err, "Scan binary")
			} else {
				if err = valconv.BytesToAny(v, c); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("MachGetColumnData unsupported type %T", c)
		}
	}
	return nil
}

/*
func scanInt16(v int16, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *int16:
		*cv = int16(v)
	case *int32:
		*cv = int32(v)
	case *int64:
		*cv = int64(v)
	case *string:
		*cv = strconv.Itoa(int(v))
	default:
		return fmt.Errorf("Scan convert from INT16 to %T not supported", c)
	}
	return nil
}

func scanInt32(v int32, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *int16:
		*cv = int16(v)
	case *int32:
		*cv = int32(v)
	case *int64:
		*cv = int64(v)
	case *string:
		*cv = strconv.Itoa(int(v))
	default:
		return fmt.Errorf("Scan convert from INT32 to %T not supported", c)
	}
	return nil
}

func scanInt64(v int64, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *int16:
		*cv = int16(v)
	case *int32:
		*cv = int32(v)
	case *int64:
		*cv = int64(v)
	case *string:
		*cv = strconv.Itoa(int(v))
	default:
		return fmt.Errorf("Scan convert from INT64 to %T not supported", c)
	}
	return nil
}

func scanDateTime(v time.Time, c any) error {
	switch cv := c.(type) {
	case *int64:
		*cv = v.UnixNano()
	case *time.Time:
		*cv = v
	case *string:
		*cv = v.String()
	default:
		return fmt.Errorf("Scan convert from INT64 to %T not supported", c)
	}
	return nil
}

func scanFloat32(v float32, c any) error {
	switch cv := c.(type) {
	case *float32:
		*cv = v
	case *float64:
		*cv = float64(v)
	case *string:
		*cv = strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return fmt.Errorf("Scan convert from FLOAT32 to %T not supported", c)
	}
	return nil
}

func scanFloat64(v float64, c any) error {
	switch cv := c.(type) {
	case *float32:
		*cv = float32(v)
	case *float64:
		*cv = v
	case *string:
		*cv = strconv.FormatFloat(v, 'f', -1, 32)
	default:
		return fmt.Errorf("Scan convert from FLOAT64 to %T not supported", c)
	}
	return nil
}

func scanString(v string, c any) error {
	switch cv := c.(type) {
	case *string:
		*cv = v
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}

func scanBytes(v []byte, c any) error {
	switch cv := c.(type) {
	case *[]uint8:
		*cv = v
	case *string:
		*cv = string(v)
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}

func scanIP(v net.IP, c any) error {
	switch cv := c.(type) {
	case *net.IP:
		*cv = v
	case *string:
		*cv = v.String()
	default:
		return fmt.Errorf("Scan convert from IPv4 to %T not supported", c)
	}
	return nil
}
*/
