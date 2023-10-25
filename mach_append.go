package mach

/*
#cgo CFLAGS: -I${SRCDIR}/native
#include <machEngine.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"unsafe"

	spi "github.com/machbase/neo-spi"
)

func (conn *connection) Appender(ctx context.Context, tableName string, opts ...spi.AppendOption) (spi.Appender, error) {
	appender := &Appender{}
	appender.tableName = strings.ToUpper(tableName)
	appender.timeformat = "ns"

	for _, opt := range opts {
		switch v := opt.(type) {
		case spi.AppendTimeformatOption:
			appender.timeformat = string(v)
		default:
			return nil, fmt.Errorf("unknown appender option %T", v)
		}
	}

	var stmt unsafe.Pointer
	if err := machAllocStmt(conn.handle, &stmt); err != nil {
		return nil, err
	}
	defer machFreeStmt(stmt)

	row := conn.QueryRow(ctx, "select type from M$SYS_TABLES where name = ?", appender.tableName)
	var typ int32 = -1
	if err := row.Scan(&typ); err != nil {
		return nil, err
	}
	if typ < 0 || typ > 6 {
		return nil, fmt.Errorf("table '%s' not found", tableName)
	}
	appender.tableType = spi.TableType(typ)

	if err := machAllocStmt(conn.handle, &appender.stmt); err != nil {
		return nil, err
	}
	if err := machAppendOpen(appender.stmt, tableName); err != nil {
		return nil, err
	}

	colCount, err := machColumnCount(appender.stmt)
	if err != nil {
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

type Appender struct {
	mutex     sync.Mutex
	stmt      unsafe.Pointer
	tableName string
	tableType spi.TableType
	columns   []*spi.Column
	closed    bool

	timeformat string
}

func (ap *Appender) Close() (int64, int64, error) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.closed {
		return 0, 0, nil
	}
	ap.closed = true
	s, f, err := machAppendClose(ap.stmt)
	if err != nil {
		return s, f, err
	}

	if err := machFreeStmt(ap.stmt); err != nil {
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
		case spi.ColumnBufferTypeInt16:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case uint16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *uint16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case int16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *int16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			}
		case spi.ColumnBufferTypeInt32:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case int16:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case *int16:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			case uint16:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case *uint16:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			case int32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case *int32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			case uint32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case *uint32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			case int:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case *int:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			case uint:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case *uint:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			}
		case spi.ColumnBufferTypeInt64:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case int16:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *int16:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case uint16:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *uint16:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case int32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *int32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case uint32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *uint32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case int:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *int:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case uint:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *uint:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case int64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *int64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case uint64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *uint64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			}
		case spi.ColumnBufferTypeFloat:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case int:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(v)
			case *int:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(*v)
			case int16:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(v)
			case *int16:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(*v)
			case int32:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(v)
			case *int32:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(*v)
			case int64:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(v)
			case *int64:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(*v)
			case float32:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(v)
			case *float32:
				*(*C.float)(unsafe.Pointer(&buffer[i].mData[0])) = C.float(*v)
			}
		case spi.ColumnBufferTypeDouble:
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case int:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			case *int:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(*v)
			case int16:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			case *int16:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(*v)
			case int32:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			case *int32:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(*v)
			case int64:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			case *int64:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(*v)
			case float32:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			case *float32:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(*v)
			case float64:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(v)
			case *float64:
				*(*C.double)(unsafe.Pointer(&buffer[i].mData[0])) = C.double(*v)
			}
		case spi.ColumnBufferTypeDatetime:
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = nil
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = nil
			switch v := val.(type) {
			default:
				return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", v, c.Name, c.Type)
			case time.Time:
				tv := v.UnixNano()
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case *time.Time:
				tv := v.UnixNano()
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int:
				tv := int64(v)
				switch ap.timeformat {
				case "s":
					tv = tv * 1000000000
				case "ms":
					tv = tv * 1000000
				case "us":
					tv = tv * 1000
				case "ns":
				default:
					return fmt.Errorf("MachAppendData cannot apply int with %s to %s (%s)", ap.timeformat, c.Name, c.Type)
				}
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int16:
				tv := int64(v)
				switch ap.timeformat {
				case "s":
					tv = tv * 1000000000
				case "ms":
					tv = tv * 1000000
				case "us":
					tv = tv * 1000
				case "ns":
				default:
					return fmt.Errorf("MachAppendData cannot apply int16 with %s to %s (%s)", ap.timeformat, c.Name, c.Type)
				}
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int32:
				tv := int64(v)
				switch ap.timeformat {
				case "s":
					tv = tv * 1000000000
				case "ms":
					tv = tv * 1000000
				case "us":
					tv = tv * 1000
				case "ns":
				default:
					return fmt.Errorf("MachAppendData cannot apply int32 with %s to %s (%s)", ap.timeformat, c.Name, c.Type)
				}
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int64:
				tv := int64(v)
				switch ap.timeformat {
				case "s":
					tv = tv * 1000000000
				case "ms":
					tv = tv * 1000000
				case "us":
					tv = tv * 1000
				case "ns":
				default:
					return fmt.Errorf("MachAppendData cannot apply int64 with %s to %s (%s)", ap.timeformat, c.Name, c.Type)
				}
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case string:
				if len(ap.timeformat) > 0 {
					cstr := C.CString(v)
					defer C.free(unsafe.Pointer(cstr))
					cfmt := C.CString(ap.timeformat)
					defer C.free(unsafe.Pointer(cfmt))
					(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = -2 // MACH_ENGINE_APPEND_DATETIME_STRING
					(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = cstr
					(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = cfmt
				} else {
					return fmt.Errorf("MachAppendData cannot apply string without format to %s (%s)", c.Name, c.Type)
				}
			case *string:
				if len(ap.timeformat) > 0 {
					cstr := C.CString(*v)
					defer C.free(unsafe.Pointer(cstr))
					cfmt := C.CString(ap.timeformat)
					defer C.free(unsafe.Pointer(cfmt))
					(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = -2 // MACH_ENGINE_APPEND_DATETIME_STRING
					(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = cstr
					(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = cfmt
				} else {
					return fmt.Errorf("MachAppendData cannot apply string without format to %s (%s)", c.Name, c.Type)
				}
			}
		case spi.ColumnBufferTypeIPv4:
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
		case spi.ColumnBufferTypeIPv6:
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
		case spi.ColumnBufferTypeString:
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
			case *string:
				if len(*v) == 0 {
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(0)
				} else {
					cstr := C.CString(*v)
					defer C.free(unsafe.Pointer(cstr))
					cstrlen := C.strlen(cstr)
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(cstrlen)
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mData = unsafe.Pointer(cstr)
				}
			}
		case spi.ColumnBufferTypeBinary:
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
			case *string:
				if len(*v) == 0 {
					(*C.MachEngineAppendVarStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uint(0)
				} else {
					cstr := C.CString(*v)
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

	ap.mutex.Lock()
	err := machAppendData(ap.stmt, &buffer[0])
	ap.mutex.Unlock()
	return err
}
