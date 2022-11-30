package mach

import (
	"fmt"
	"strings"
	"time"
	"unsafe"
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

func (this *Database) SqlTidy(sqlText string) string {
	lines := strings.Split(sqlText, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimSpace(ln)
	}
	return strings.TrimSpace(strings.Join(lines, " "))
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
		vals[i] = bindValue(c)
	}
	if err := machAppendData(this.stmt, vals); err != nil {
		fmt.Printf("%v", err)
		return err
	}
	return nil
}
