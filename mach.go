package mach

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
	"unsafe"
)

/*
#cgo CFLAGS: -I${SRCDIR}/native
#include <machEngine.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"

func LinkInfo() string {
	return LibMachLinkInfo
}

func Edition() string {
	if strings.Contains(LibMachLinkInfo, "edge") {
		return "edge"
	} else if strings.Contains(LibMachLinkInfo, "fog") {
		return "fog"
	} else {
		return "none"
	}
}

type InitOption int

const (
	OPT_NONE               InitOption = 0x0
	OPT_SIGHANDLER_DISABLE InitOption = 0x1
)

func Initialize(homeDir string) error {
	return InitializeOption(homeDir, OPT_SIGHANDLER_DISABLE)
}

func InitializeOption(homeDir string, opt InitOption) error {
	var handle unsafe.Pointer
	err := initialize0(homeDir, int(opt), &handle)
	if err == nil {
		singleton.handle = handle
	}
	return err
}

func Finalize() {
	finalize0(singleton.handle)
}

func DestroyDatabase() error {
	return destroyDatabase0(singleton.handle)
}

func CreateDatabase() error {
	return createDatabase0(singleton.handle)
}

func ExistsDatabase() bool {
	return existsDatabase0(singleton.handle)
}

type Env struct {
	handle unsafe.Pointer
}

var singleton = Env{}

type Database struct {
	handle unsafe.Pointer
}

func New() *Database {
	return &Database{
		handle: singleton.handle,
	}
}

func (db *Database) Startup() error {
	// machbase startup 과정에서 현재 디렉터리를 HOME으로 변경하는데,
	// application의 Working directory를 유지하기 위해 chdir()을 호출한다.
	cwd, _ := os.Getwd()
	defer func() {
		os.Chdir(cwd)
	}()

	err := startup0(db.handle)
	return err
}

func (db *Database) Shutdown() error {
	return shutdown0(db.handle)
}

func (db *Database) Error() error {
	return machError0(db.handle)
}

func (db *Database) UserAuth(username, password string) (bool, error) {
	return machUserAuth(db.handle, username, password)
}

func (db *Database) SqlTidy(sqlText string) string {
	lines := strings.Split(sqlText, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimSpace(ln)
	}
	return strings.TrimSpace(strings.Join(lines, " "))
}

func (db *Database) Explain(sqlText string) (string, error) {
	var stmt unsafe.Pointer
	if err := machAllocStmt(db.handle, &stmt); err != nil {
		return "", err
	}
	defer machFreeStmt(db.handle, stmt)
	if err := machPrepare(stmt, sqlText); err != nil {
		return "", err
	}
	return machExplain(stmt)
}

func (db *Database) Exec(sqlText string, params ...any) *Result {
	var result = &Result{}

	var stmt unsafe.Pointer
	if err := machAllocStmt(db.handle, &stmt); err != nil {
		result.err = err
		return result
	}
	defer machFreeStmt(db.handle, stmt)
	if len(params) == 0 {
		if err := machDirectExecute(stmt, sqlText); err != nil {
			result.err = err
			return result
		}
	} else {
		err := machPrepare(stmt, sqlText)
		if err != nil {
			result.err = err
			return result
		}
		for i, p := range params {
			if err := bind(stmt, i, p); err != nil {
				result.err = err
				return result
			}
		}
		err = machExecute(stmt)
		if err != nil {
			result.err = err
			return result
		}
	}
	affectedRows, err := machEffectRows(stmt)
	if err != nil {
		result.err = err
		return result
	}
	stmtType, err := machStmtType(stmt)
	result.affectedRows = affectedRows
	result.stmtType = stmtType
	result.err = err
	return result
}

func (db *Database) Query(sqlText string, params ...any) (*Rows, error) {
	rows := &Rows{
		handle:  db.handle,
		sqlText: sqlText,
	}
	if err := machAllocStmt(db.handle, &rows.stmt); err != nil {
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
	if stmtType, err := machStmtType(rows.stmt); err != nil {
		return nil, err
	} else {
		rows.stmtType = stmtType
	}
	return rows, nil
}

func (db *Database) QueryRow(sqlText string, params ...any) *Row {
	var row = &Row{}

	var stmt unsafe.Pointer
	if row.err = machAllocStmt(db.handle, &stmt); row.err != nil {
		return row
	}
	defer func() {
		row.err = machFreeStmt(db.handle, stmt)
	}()

	if row.err = machPrepare(stmt, sqlText); row.err != nil {
		return row
	}
	for i, p := range params {
		if row.err = bind(stmt, i, p); row.err != nil {
			return row
		}
	}
	if row.err = machExecute(stmt); row.err != nil {
		return row
	}

	if typ, err := machStmtType(stmt); err != nil {
		row.err = err
		return row
	} else {
		row.stmtType = typ
	}

	row.ok = true

	// select 가 아니면 fetch를 진행하지 않는다.
	if !row.stmtType.IsSelect() {
		affectedRows, err := machEffectRows(stmt)
		if err != nil {
			row.err = err
			return row
		}
		row.affectedRows = affectedRows
		return row
	}

	if _, row.err = machFetch(stmt); row.err != nil {
		return row
	}

	var count int
	count, row.err = machColumnCount(stmt)
	if row.err != nil {
		return row
	}
	if count == 0 {
		row.err = sql.ErrNoRows
		return row
	}
	row.values = make([]any, count)
	for i := 0; i < count; i++ {
		typ, siz, err := machColumnType(stmt, i)
		if err != nil {
			row.err = err
			return row
		}
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			row.values[i] = new(int16)
		case 1: // MACH_DATA_TYPE_INT32
			row.values[i] = new(int32)
		case 2: // MACH_DATA_TYPE_INT64
			row.values[i] = new(int64)
		case 3: // MACH_DATA_TYPE_DATETIME
			row.values[i] = new(time.Time)
		case 4: // MACH_DATA_TYPE_FLOAT
			row.values[i] = new(float32)
		case 5: // MACH_DATA_TYPE_DOUBLE
			row.values[i] = new(float64)
		case 6: // MACH_DATA_TYPE_IPV4
			row.values[i] = new(net.IP)
		case 7: // MACH_DATA_TYPE_IPV6
			row.values[i] = new(net.IP)
		case 8: // MACH_DATA_TYPE_STRING
			row.values[i] = new(string)
		case 9: // MACH_DATA_TYPE_BINARY
			row.values[i] = make([]byte, siz)
		default:
			row.err = fmt.Errorf("QueryRow unsupported type %d", typ)
		}
	}
	row.err = scan(stmt, row.values...)
	return row
}

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

func (ap *Appender) Close() (uint64, uint64, error) {
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

func (ap *Appender) Table() string {
	return ap.tableName
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
