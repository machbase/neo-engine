package mach

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

type AppenderOption func(*Appender)

// Appender creates a new Appender for the given table.
// Appender should be closed as soon as finshing work, otherwise it may cause server side resource leak.
//
//	ctx, cancelFunc := context.WithTimeout(5*time.Second)
//	defer cancelFunc()
//
//	app, _ := conn.Appender(ctx, "MYTABLE")
//	defer app.Close()
//	app.Append("name", time.Now(), 3.14)
func (conn *Conn) Appender(ctx context.Context, tableName string, opts ...AppenderOption) (*Appender, error) {
	appender := &Appender{}
	appender.conn = conn
	appender.tableName = strings.ToUpper(tableName)
	appender.timeformat = "ns"

	for _, opt := range opts {
		opt(appender)
	}

	// table type
	// make a new internal connection to avoid MACH-ERR 2118
	// MACH-ERR 2118 Lock object was already initialized. (Do not use select and append simultaneously in single session.)
	if qcon, err := conn.db.Connect(ctx, WithTrustUser("sys")); err != nil {
		return nil, err
	} else {
		defer qcon.Close()
		row := qcon.QueryRow(ctx, "select type from M$SYS_TABLES where name = ?", appender.tableName)
		var typ int32 = -1
		if err := row.Scan(&typ); err != nil {
			if err.Error() == "sql: no rows in result set" {
				return nil, fmt.Errorf("table '%s' not found", appender.tableName)
			} else {
				return nil, fmt.Errorf("table '%s' not found, %s", appender.tableName, err.Error())
			}
		}
		if typ < 0 || typ > 6 {
			return nil, fmt.Errorf("table '%s' not found", tableName)
		}
		appender.tableType = TableType(typ)
	}
	if err := machAllocStmt(appender.conn.handle, &appender.stmt); err != nil {
		return nil, err
	}
	if err := machAppendOpen(appender.stmt, tableName); err != nil {
		machFreeStmt(appender.stmt)
		return nil, err
	}
	statz.AllocAppender()

	colCount, err := machColumnCount(appender.stmt)
	if err != nil {
		machAppendClose(appender.stmt)
		machFreeStmt(appender.stmt)
		return nil, err
	}
	appender.columns = make([]*Column, colCount)
	for i := 0; i < colCount; i++ {
		nfo, err := machColumnInfo(appender.stmt, i)
		if err != nil {
			return nil, err
		}
		appender.columns[i] = nfo
	}
	return appender, nil
}

func AppenderTimeformat(timeformat string) AppenderOption {
	return func(a *Appender) {
		a.timeformat = timeformat
	}
}

type Appender struct {
	conn      *Conn
	stmt      unsafe.Pointer
	tableName string
	tableType TableType
	columns   []*Column
	closed    bool

	successCount int64
	failCount    int64

	timeformat string
}

func (ap *Appender) Close() (int64, int64, error) {
	if ap.closed {
		return ap.successCount, ap.failCount, nil
	}
	ap.closed = true
	var err error
	statz.FreeAppender()
	ap.successCount, ap.failCount, err = machAppendClose(ap.stmt)
	if err != nil {
		return ap.successCount, ap.failCount, err
	}

	if err := machFreeStmt(ap.stmt); err != nil {
		return ap.successCount, ap.failCount, err
	}
	return ap.successCount, ap.failCount, nil
}

func (ap *Appender) String() string {
	return fmt.Sprintf("appender %s %v", ap.tableName, ap.stmt)
}

func (ap *Appender) TableName() string {
	return ap.tableName
}

func (ap *Appender) Columns() ([]string, []string, error) {
	cols := ap.columns
	names := make([]string, len(cols))
	types := make([]string, len(cols))
	for i := range cols {
		names[i] = cols[i].Name
		types[i] = cols[i].Type
	}
	return names, types, nil
}

func (ap *Appender) TableType() TableType {
	return ap.tableType
}

func (ap *Appender) Append(values ...any) error {
	if ap.tableType == TagTableType {
		return ap.appendTable0(values)
	} else if ap.tableType == LogTableType {
		colsWithTime := append([]any{time.Time{}}, values...)
		return ap.appendTable0(colsWithTime)
	} else {
		return fmt.Errorf("%s is not appendable table", ap.tableName)
	}
}

func (ap *Appender) AppendWithTimestamp(ts time.Time, cols ...any) error {
	if ap.tableType == LogTableType {
		colsWithTime := append([]any{ts}, cols...)
		return ap.appendTable0(colsWithTime)
	} else if ap.tableType == TagTableType {
		colsWithTime := append([]any{cols[0], ts}, cols[1:]...)
		return ap.appendTable0(colsWithTime)
	} else {
		return fmt.Errorf("%s is not a log table, use Append() instead", ap.tableName)
	}
}
