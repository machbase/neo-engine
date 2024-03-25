package mach

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/machbase/neo-server/spi"
)

func (conn *connection) Appender(ctx context.Context, tableName string, opts ...spi.AppenderOption) (spi.Appender, error) {
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
			return nil, err
		}
		if typ < 0 || typ > 6 {
			return nil, fmt.Errorf("table '%s' not found", tableName)
		}
		appender.tableType = spi.TableType(typ)
	}
	if err := machAllocStmt(conn.handle, &appender.stmt); err != nil {
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
	appender.columns = make([]*spi.Column, colCount)
	for i := 0; i < colCount; i++ {
		nfo, err := machColumnInfo(appender.stmt, i)
		if err != nil {
			return nil, err
		}
		appender.columns[i] = nfo
	}
	return appender, nil
}

func AppenderTimeformat(timeformat string) spi.AppenderOption {
	return func(a spi.Appender) {
		if apd, ok := a.(*Appender); ok {
			apd.timeformat = timeformat
		}
	}
}

type Appender struct {
	conn      *connection
	stmt      unsafe.Pointer
	tableName string
	tableType spi.TableType
	columns   []*spi.Column
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

func (ap *Appender) Columns() (spi.Columns, error) {
	cols := ap.columns
	result := make([]*spi.Column, len(cols))
	for i := range cols {
		result[i] = &spi.Column{
			Name:   cols[i].Name,
			Type:   cols[i].Type,
			Size:   cols[i].Size,
			Length: cols[i].Length,
		}
	}
	return result, nil
}

func (ap *Appender) TableType() spi.TableType {
	return ap.tableType
}

func (ap *Appender) Append(values ...any) error {
	if ap.tableType == spi.TagTableType {
		return ap.appendTable0(values)
	} else if ap.tableType == spi.LogTableType {
		colsWithTime := append([]any{time.Time{}}, values...)
		return ap.appendTable0(colsWithTime)
	} else {
		return fmt.Errorf("%s is not appendable table", ap.tableName)
	}
}

func (ap *Appender) AppendWithTimestamp(ts time.Time, cols ...any) error {
	if ap.tableType == spi.LogTableType {
		colsWithTime := append([]any{ts}, cols...)
		return ap.appendTable0(colsWithTime)
	} else if ap.tableType == spi.TagTableType {
		colsWithTime := append([]any{cols[0], ts}, cols[1:]...)
		return ap.appendTable0(colsWithTime)
	} else {
		return fmt.Errorf("%s is not a log table, use Append() instead", ap.tableName)
	}
}
