package mach

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

func LinkInfo() string {
	return LibMachLinkInfo
}

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

func New() *Database {
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

func (this *Database) Appender(tableName string) (*Appender, error) {
	appender := &Appender{}
	if err := machAllocStmt(this.handle, &appender.stmt); err != nil {
		return nil, err
	}
	if err := machAppendOpen(appender.stmt, tableName); err != nil {
		return nil, err
	}
	// MachColumnCount returns -1
	//
	// colCount, err := machColumnCount(appender.stmt)
	// if err != nil {
	// 	return nil, err
	// }
	//fmt.Printf("======> colCount: %d\n", colCount)
	return appender, nil
}

type Appender struct {
	stmt         unsafe.Pointer
	SuccessCount uint64
	FailureCount uint64
}

func (this *Appender) Close() error {
	if this.stmt == nil {
		return nil
	}
	s, f, err := machAppendClose(this.stmt)
	if err != nil {
		return err
	}
	this.SuccessCount = s
	this.FailureCount = f

	if err := machFreeStmt(this.stmt); err != nil {
		return err
	}
	this.stmt = nil
	return nil
}

func (this *Appender) Append(cols ...any) error {
	vals := make([]*machAppendDataNullValue, len(cols))
	for i, c := range cols {
		vals[i] = makeAppendDataNullValue(c)
	}
	if err := machAppendData(this.stmt, vals); err != nil {
		fmt.Printf("%v", err)
		return err
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
				if err = scanInt16(v, c); err != nil {
					return err
				}
			}
		case 1: // MACH_DATA_TYPE_INT32
			if v, err := machColumnDataInt32(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = scanInt32(v, c); err != nil {
					return err
				}
			}
		case 2: // MACH_DATA_TYPE_INT64
			if v, err := machColumnDataInt64(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else {
				if err = scanInt64(v, c); err != nil {
					return err
				}
			}
		case 3: // MACH_DATA_TYPE_DATETIME
			if v, err := machColumnDataDateTime(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan datetime")
			} else {
				if err = scanDateTime(v, c); err != nil {
					return err
				}
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			if v, err := machColumnDataFloat32(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = scanFloat32(v, c); err != nil {
					return err
				}
			}
		case 5: // MACH_DATA_TYPE_DOUBLE
			if v, err := machColumnDataFloat64(rows.stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else {
				if err = scanFloat64(v, c); err != nil {
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
				if err = scanString(v, c); err != nil {
					return err
				}
			}
		case 9: // MACH_DATA_TYPE_BINARY
			if v, err := machColumnDataBinary(rows.stmt, i); err != nil {
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
