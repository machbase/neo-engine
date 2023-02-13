package mach

/*
#cgo CFLAGS: -I${SRCDIR}/native
#include <machEngine.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"
)

func (db *Database) Appender(tableName string) (*Appender, error) {
	appender := &Appender{}
	appender.handle = db.handle
	appender.tableName = strings.ToUpper(tableName)

	row := db.QueryRow("select type from M$SYS_TABLES where name = ?", appender.tableName)
	var typ int32 = -1
	if err := row.Scan(&typ); err != nil {
		return nil, err
	}
	if typ < 0 || typ > 6 {
		return nil, fmt.Errorf("table '%s' not found", tableName)
	}
	appender.tableType = TableType(typ)

	if err := machAllocStmt(db.handle, &appender.stmt); err != nil {
		return nil, err
	}
	if err := machAppendOpen(appender.stmt, tableName); err != nil {
		return nil, err
	}

	colCount, err := machColumnCount(appender.stmt)
	if err != nil {
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

type Appender struct {
	handle    unsafe.Pointer
	stmt      unsafe.Pointer
	tableName string
	tableType TableType
	columns   []*Column
	closed    bool
}

func (ap *Appender) Close() (int64, int64, error) {
	if ap.closed {
		return 0, 0, nil
	}
	ap.closed = true
	s, f, err := machAppendClose(ap.stmt)
	if err != nil {
		return s, f, err
	}

	if err := machFreeStmt(ap.handle, ap.stmt); err != nil {
		return s, f, err
	}
	return s, f, nil
}

func (ap *Appender) String() string {
	return fmt.Sprintf("appender %s %v", ap.tableName, ap.stmt)
}

func (ap *Appender) TableName() string {
	return ap.tableName
}

func (ap *Appender) Columns() []*Column {
	return ap.columns
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

func (ap *Appender) appendTable0(vals []any) error {
	if len(ap.columns) == 0 {
		return fmt.Errorf("table '%s' has no columns", ap.tableName)
	}
	if len(ap.columns) != len(vals) {
		return fmt.Errorf("value count %d, table '%s' requres %d columns for appeding", len(vals), ap.tableName, len(ap.columns))
	}

	buffer := make([]C.MachEngineAppendParam, len(ap.columns))

	for i, val := range vals {
		if val == nil {
			buffer[i].mIsNull = C.int(1)
			continue
		}
		c := ap.columns[i]
		switch c.Type {
		default:
			return fmt.Errorf("machAppendData unknown column type '%s'", c.Type)
		case ColumnTypeNameInt16:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case uint16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case int16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			}
		case ColumnTypeNameInt32:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case int16:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case uint16:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case int32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case uint32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case int:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case uint:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			}
		case ColumnTypeNameInt64:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case int16:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case uint16:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case int32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case uint32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case int:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case uint:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case int64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case uint64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			}
		case ColumnTypeNameFloat:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case float32:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(v)
			}
		case ColumnTypeNameDouble:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case float32:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			case float64:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			}
		case ColumnTypeNameDatetime:
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = nil
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = nil
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case time.Time:
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(v.UnixNano())
			case int:
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(v)
			case int32:
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(v)
			case int64:
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(v)
			}
		case ColumnTypeNameIPv4:
			ip, ok := val.(net.IP)
			if !ok {
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", val, c.Name, c.Type)
			}
			if ipv4 := ip.To4(); ipv4 == nil {
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", val, c.Name, c.Type)
			} else {
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uchar(net.IPv4len)
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddrString = nil
				for i := 0; i < net.IPv4len; i++ {
					(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddr[i] = C.uchar(ipv4[i])
				}
			}
		case ColumnTypeNameIPv6:
			ip, ok := val.(net.IP)
			if !ok {
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", val, c.Name, c.Type)
			}
			if ipv6 := ip.To16(); ipv6 == nil {
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", val, c.Name, c.Type)
			} else {
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uchar(net.IPv6len)
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddrString = nil
				for i := 0; i < net.IPv6len; i++ {
					(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddr[i] = C.uchar(ipv6[i])
				}
			}
		case ColumnTypeNameString:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case string:
				if len(v) == 0 {
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(0)
				} else {
					cstr := C.CString(v)
					defer C.free(unsafe.Pointer(cstr))
					cstrlen := C.strlen(cstr)
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(cstrlen)
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mData = unsafe.Pointer(cstr)
				}
			}
		case ColumnTypeNameBinary:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case string:
				if len(v) == 0 {
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(0)
				} else {
					cstr := C.CString(v)
					defer C.free(unsafe.Pointer(cstr))
					cstrlen := C.strlen(cstr)
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(cstrlen)
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mData = unsafe.Pointer(cstr)
				}
			case []byte:
				(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(len(v))
				if len(v) > 0 {
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mData = unsafe.Pointer(&v[0])
				}
			}
		}
	}

	if err := machAppendData(ap.stmt, &buffer[0]); err != nil {
		return err
	}
	return nil
}
