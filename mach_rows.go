package mach

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"
	"unsafe"

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
			scanInt16(*v, cols[i])
		case *int32:
			scanInt32(*v, cols[i])
		case *int64:
			scanInt64(*v, cols[i])
		case *time.Time:
			scanDateTime(*v, cols[i])
		case *float32:
			scanFloat32(*v, cols[i])
		case *float64:
			scanFloat64(*v, cols[i])
		case *net.IP:
			scanIP(*v, cols[i])
		case *string:
			scanString(*v, cols[i])
		case []byte:
			scanBytes(v, cols[i])
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
				if err = scanInt16(v, c); err != nil {
					return err
				}
			}
		case 1: // MACH_DATA_TYPE_INT32
			if v, err := machColumnDataInt32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = scanInt32(v, c); err != nil {
					return err
				}
			}
		case 2: // MACH_DATA_TYPE_INT64
			if v, err := machColumnDataInt64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = scanInt64(v, c); err != nil {
					return err
				}
			}
		case 3: // MACH_DATA_TYPE_DATETIME
			if v, err := machColumnDataDateTime(stmt, i); err != nil {
				return errors.Wrap(err, "Scan datetime")
			} else {
				if err = scanDateTime(v, c); err != nil {
					return err
				}
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			if v, err := machColumnDataFloat32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = scanFloat32(v, c); err != nil {
					return err
				}
			}
		case 5: // MACH_DATA_TYPE_DOUBLE
			if v, err := machColumnDataFloat64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = scanFloat64(v, c); err != nil {
					return err
				}
			}
		case 6: // MACH_DATA_TYPE_IPV4
			if v, err := machColumnDataIPv4(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else {
				if err = scanIP(v, c); err != nil {
					return err
				}
			}
		case 7: // MACH_DATA_TYPE_IPV6
			if v, err := machColumnDataIPv6(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else {
				if err = scanIP(v, c); err != nil {
					return err
				}
			}
		case 8: // MACH_DATA_TYPE_STRING
			if v, err := machColumnDataString(stmt, i); err != nil {
				return errors.Wrap(err, "Scan string")
			} else {
				if err = scanString(v, c); err != nil {
					return err
				}
			}
		case 9: // MACH_DATA_TYPE_BINARY
			if v, err := machColumnDataBinary(stmt, i); err != nil {
				return errors.Wrap(err, "Scan binary")
			} else {
				if err = scanBytes(v, c); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("MachGetColumnData unsupported type %T", c)
		}
	}
	return nil
}

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
