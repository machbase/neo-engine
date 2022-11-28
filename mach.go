package mach

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

func Initialize(homeDir string) error {
	return initialize0(homeDir)
}

func DestroyDatabase() error {
	return destroyDatabase0()
}

func CreateDatabase() error {
	return createDatabase0()
}

type Database struct {
	handle unsafe.Pointer
}

func NewDatabase() *Database {
	return &Database{}
}

func (this *Database) Startup(timeout time.Duration) error {
	return startup0(&this.handle, timeout)
}

func (this *Database) Shutdown() error {
	return shutdown0(this.handle)
}

func (this *Database) Error() error {
	return db_error0(this.handle)
}

func (this *Database) Exec(sqlText string, params ...any) error {
	var stmt unsafe.Pointer
	if err := machAllocStmt(this.handle, &stmt); err != nil {
		return err
	}
	defer machFreeStmt(stmt)
	if len(params) == 0 {
		if err := machDirectExecute(stmt, sqlText); err != nil {
			return err
		}
	} else {
		err := machPrepare(stmt, sqlText)
		if err != nil {
			return err
		}
		for i, p := range params {
			if err := bind(stmt, i, p); err != nil {
				return err
			}
		}
		err = machExecute(stmt)
	}
	return nil
}

func (this *Database) Query(sqlText string, params ...any) (*Rows, error) {
	rows := &Rows{
		sqlText: sqlText,
	}
	if err := machAllocStmt(this.handle, &rows.stmt); err != nil {
		return nil, err
	}
	if err := machPrepare(rows.stmt, sqlText); err != nil {
		return nil, err
	}
	for i, p := range params {
		if err := bind(rows.stmt, i, p); err != nil {
			return nil, err
		}
	}
	if err := machExecute(rows.stmt); err != nil {
		return nil, err
	}
	return rows, nil
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
	for i, c := range cols {
		typ, err := machColumnType(rows.stmt, i)
		if err != nil {
			return errors.Wrap(err, "Scan")
		}
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			if v, err := machColumnDataInt16(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = convertInt16(v, c); err != nil {
					return err
				}
			}
		case 1: // MACH_DATA_TYPE_INT32
			if v, err := machColumnDataInt32(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = convertInt32(v, c); err != nil {
					return err
				}
			}
		case 2: // MACH_DATA_TYPE_INT64
			if v, err := machColumnDataInt64(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = convertInt64(v, c); err != nil {
					return err
				}
			}
		case 3: // MACH_DATA_TYPE_DATETIME
			if v, err := machColumnDataDateTime(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan datetime")
			} else {
				if err = convertDateTime(v, c); err != nil {
					return err
				}
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			if v, err := machColumnDataFloat32(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = convertFloat32(v, c); err != nil {
					return err
				}
			}
		case 5: // MACH_DATA_TYPE_DOUBLE
			if v, err := machColumnDataFloat64(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = convertFloat64(v, c); err != nil {
					return err
				}
			}
		case 6: // MACH_DATA_TYPE_IPV4
			return fmt.Errorf("not (yet) implemented ipv4")
		case 7: // MACH_DATA_TYPE_IPV6
			return fmt.Errorf("not (yet) implemented ipv6")
		case 8: // MACH_DATA_TYPE_STRING
			if v, err := machColumnDataString(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan string")
			} else {
				if err = convertString(v, c); err != nil {
					return err
				}
			}
		case 9: // MACH_DATA_TYPE_BINARY
			if v, err := machColumnDataBinary(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan binary")
			} else {
				if err = convertBytes(v, c); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("MachGetColumnData unsupported type %T", c)
		}
	}
	return nil
}

func bind(stmt unsafe.Pointer, idx int, c any) error {
	switch cv := c.(type) {
	case int:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case int32:
		if err := machBindInt32(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case int64:
		if err := machBindInt64(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case float32:
		if err := machBindFloat64(stmt, idx, float64(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case float64:
		if err := machBindFloat64(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case string:
		if err := machBindString(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case []byte:
		if err := machBindBinary(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	default:
		return fmt.Errorf("bind supported idx %d type %T", idx, c)
	}
	return nil
}

func convertInt16(v int16, c any) error {
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

func convertInt32(v int32, c any) error {
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

func convertInt64(v int64, c any) error {
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

func convertDateTime(v time.Time, c any) error {
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

func convertFloat32(v float32, c any) error {
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

func convertFloat64(v float64, c any) error {
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

func convertString(v string, c any) error {
	switch cv := c.(type) {
	case *string:
		*cv = v
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}

func convertBytes(v []byte, c any) error {
	switch cv := c.(type) {
	case *[]byte:
		copy(*cv, v)
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}
