package mach

import (
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"

	_ "github.com/machbase/neo-engine/native"
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
	nfo := LinkInfo()
	if strings.Contains(nfo, "standard") {
		return "standard"
	} else if strings.Contains(nfo, "edge") {
		return "edge"
	} else if strings.Contains(nfo, "fog") {
		return "fog"
	} else {
		return "none"
	}
}

func initialize0(homeDir string, machPort int, flag int, envHandle *unsafe.Pointer) error {
	cstr := C.CString(homeDir)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachInitialize(cstr, C.int(machPort), C.int(flag), envHandle); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachInitialize", int(rt))
	}
}

func finalize0(envHandle unsafe.Pointer) {
	C.MachFinalize(envHandle)
}

func createDatabase0(envHandle unsafe.Pointer) error {
	if rt := C.MachCreateDB(envHandle); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCreateDB", int(rt))
	}
}

func destroyDatabase0(envHandle unsafe.Pointer) error {
	if rt := C.MachDestroyDB(envHandle); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachDestroyDB", int(rt))
	}
}

func existsDatabase0(envHandle unsafe.Pointer) bool {
	rt := C.MachIsDBCreated(envHandle)
	return rt == 1
}

func startup0(envHandle unsafe.Pointer) error {
	if rt := C.MachStartupDB(envHandle); rt != 0 {
		dbErr := machError0(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachStartupDB", int(rt))
		}
	}
	return nil
}

func shutdown0(envHandle unsafe.Pointer) error {
	if rt := C.MachShutdownDB(envHandle); rt == 0 {
		return nil
	} else {
		dbErr := machError0(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachShutdownDB", int(rt))
		}
	}
}

func machConnectionCount(envHandle unsafe.Pointer) int {
	ret := C.MachGetConnectionCount(envHandle)
	return int(ret)
}

func machConnect(envHandle unsafe.Pointer, username string, password string, conn *unsafe.Pointer) error {
	cusername := C.CString(username)
	cpassword := C.CString(password)
	defer func() {
		C.free(unsafe.Pointer(cusername))
		C.free(unsafe.Pointer(cpassword))
	}()
	if rt := C.MachConnect(envHandle, cusername, cpassword, conn); rt == 0 {
		return nil
	} else {
		dbErr := machError0(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachConnect", int(rt))
		}
	}
}

func machConnectTrust(envHandle unsafe.Pointer, username string, conn *unsafe.Pointer) error {
	cusername := C.CString(username)
	defer func() {
		C.free(unsafe.Pointer(cusername))
	}()
	if rt := C.MachConnectNoAuth(envHandle, cusername, conn); rt == 0 {
		return nil
	} else {
		dbErr := machError0(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachConnect", int(rt))
		}
	}
}

func machDisconnect(conn unsafe.Pointer) error {
	if rt := C.MachDisconnect(conn); rt == 0 {
		return nil
	} else {
		dbErr := machError0(conn)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachDisconnect", int(rt))
		}
	}
}

func machCancel(conn unsafe.Pointer) error {
	if rt := C.MachCancel(conn); rt == 0 {
		return nil
	} else {
		dbErr := machError0(conn)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachCancel", int(rt))
		}
	}
}

func machSessionID(conn unsafe.Pointer) (uint64, error) {
	rt := C.MachSessionID(conn)
	return uint64(rt), nil
}

func machError0(handle unsafe.Pointer) error {
	code := C.MachErrorCode(handle)
	msg := C.MachErrorMsg(handle)
	if code != 0 && msg != nil {
		return ErrDatabaseMach(int(code), C.GoString(msg))
	}
	return nil
}

// 0: id and password are correct
// 2080: user does not exist
// 2081: password is not correct
// int MachUserAuth(void* aEnvHandle, char* aUserName, char* aPassword);
func machUserAuth(envHandle unsafe.Pointer, username string, password string) (bool, error) {
	cusername := C.CString(username)
	cpassword := C.CString(password)
	defer func() {
		C.free(unsafe.Pointer(cusername))
		C.free(unsafe.Pointer(cpassword))
	}()

	rt := C.MachUserAuth(envHandle, cusername, cpassword)
	switch rt {
	case 0:
		return true, nil
	case 2080:
		return false, nil
	case 2081:
		return false, nil
	default:
		return false, ErrDatabaseReturns("MachUserAuth", int(rt))
	}
}

func machExplain(stmt unsafe.Pointer, full bool) (string, error) {
	var cstr = [1024 * 16]C.char{}
	var mode = 0
	if full {
		mode = 1
	}
	if rt := C.MachExplain(stmt, &cstr[0], C.int(len(cstr)), C.int(mode)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return "", stmtErr
		} else {
			return "", ErrDatabaseReturns("MachExplain", int(rt))
		}
	}
	return C.GoString(&cstr[0]), nil
}

func machAllocStmt(conn unsafe.Pointer, stmt *unsafe.Pointer) error {
	var ptr unsafe.Pointer
	if rt := C.MachAllocStmt(conn, &ptr); rt != 0 {
		dbErr := machError0(conn)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachAllocStmt", int(rt))
		}
	}
	*stmt = ptr
	return nil
}

func machFreeStmt(stmt unsafe.Pointer) error {
	if rt := C.MachFreeStmt(stmt); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachFreeStmt", int(rt))
		}
	}
	return nil
}

func machPrepare(stmt unsafe.Pointer, sqlText string) error {
	cstr := C.CString(sqlText)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachPrepare(stmt, cstr); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachPrepare", int(rt))
		}
	}
	return nil
}

func machExecute(stmt unsafe.Pointer) error {
	if rt := C.MachExecute(stmt); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachExecute", int(rt))
		}
	}
	return nil
}

func machExecuteClean(stmt unsafe.Pointer) error {
	if rt := C.MachExecuteClean(stmt); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachExecuteClean", int(rt))
		}
	}
	return nil
}

func machDirectExecute(stmt unsafe.Pointer, sqlText string) error {
	cstr := C.CString(sqlText)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachDirectExecute(stmt, cstr); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachDirectExecute", int(rt))
		}
	}
	return nil
}

func machStmtType(stmt unsafe.Pointer) (StmtType, error) {
	var typ C.int
	if rt := C.MachStmtType(stmt, &typ); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturns("MachStmtType", int(rt))
		}
	}
	return StmtType(typ), nil
}

func machEffectRows(stmt unsafe.Pointer) (int64, error) {
	var rn C.ulonglong
	if rt := C.MachEffectRows(stmt, &rn); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturns("MachEffectRows", int(rt))
		}
	}
	return int64(rn), nil
}

// return true if fetch success(record exists), otherwise false
func machFetch(stmt unsafe.Pointer) (bool, error) {
	var fetchEnd C.int // 0 if record exists, otherwise 1
	if rt := C.MachFetch(stmt, &fetchEnd); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return false, stmtErr
		} else {
			return false, ErrDatabaseReturns("MachFetch", int(rt))
		}
	}
	return fetchEnd == 0, nil
}

func machBindInt32(stmt unsafe.Pointer, idx int, val int32) error {
	if rt := C.MachBindInt32(stmt, C.int(idx), C.int(val)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindInt32", idx, int(rt))
		}
	}
	return nil
}

func machBindInt64(stmt unsafe.Pointer, idx int, val int64) error {
	if rt := C.MachBindInt64(stmt, C.int(idx), C.longlong(val)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindInt64", idx, int(rt))
		}
	}
	return nil
}

func machBindFloat64(stmt unsafe.Pointer, idx int, val float64) error {
	if rt := C.MachBindDouble(stmt, C.int(idx), C.double(val)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindDouble", idx, int(rt))
		}
	}
	return nil
}

func machBindString(stmt unsafe.Pointer, idx int, val string) error {
	cstr := C.CString(val)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachBindString(stmt, C.int(idx), cstr, C.int(len(val))); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindString", idx, int(rt))
		}
	}
	return nil
}

func machBindBinary(stmt unsafe.Pointer, idx int, data []byte) error {
	ptr := unsafe.Pointer(&data[0])
	if rt := C.MachBindBinary(stmt, C.int(idx), ptr, C.int(len(data))); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindBinary", idx, int(rt))
		}
	}
	return nil
}

func machBindNull(stmt unsafe.Pointer, idx int) error {
	if rt := C.MachBindNull(stmt, C.int(idx)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindNull", idx, int(rt))
		}
	}
	return nil
}

func machColumnCount(stmt unsafe.Pointer) (int, error) {
	var count C.int = 0
	if rt := C.MachColumnCount(stmt, &count); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturns("MachColumnCount", int(rt))
		}
	}
	return int(count), nil
}

func machColumnInfo(stmt unsafe.Pointer, idx int) (*Column, error) {
	var nfo C.MachEngineColumnInfo
	if rt := C.MachColumnInfo(stmt, C.int(idx), &nfo); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return nil, stmtErr
		} else {
			return nil, ErrDatabaseReturns("MachColumnInfo", int(rt))
		}
	}

	typ, err := NativeColumnTypeString(NativeColumnType(nfo.mColumnType))
	if err != nil {
		return nil, ErrDatabaseWrap("MachColumnInfo %s", err)
	}

	return &Column{
		Name:   C.GoString(&nfo.mColumnName[0]),
		Type:   typ,
		Size:   int(nfo.mColumnSize),
		Length: int(nfo.mColumnLength),
	}, nil
}

func machColumnName(stmt unsafe.Pointer, idx int) (string, error) {
	var cstr = [100]C.char{}
	if rt := C.MachColumnName(stmt, C.int(idx), &cstr[0], C.int(len(cstr))); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return fmt.Sprintf("col-%d", idx), stmtErr
		} else {
			return fmt.Sprintf("col-%d", idx), ErrDatabaseReturns("MachColumnName", int(rt))
		}
	}
	return C.GoString(&cstr[0]), nil
}

func machColumnType(stmt unsafe.Pointer, idx int) (NativeColumnType, NativeColumnSize, error) {
	var typ C.int = 0
	var siz C.int = 0
	if rt := C.MachColumnType(stmt, C.int(idx), &typ, &siz); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, 0, stmtErr
		} else {
			return 0, 0, ErrDatabaseReturnsAtIdx("MachColumnType", idx, int(rt))
		}
	}
	return NativeColumnType(typ), NativeColumnSize(siz), nil
}

func machColumnLength(stmt unsafe.Pointer, idx int) (int, error) {
	var length C.int = 0
	if rt := C.MachColumnLength(stmt, C.int(idx), &length); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturnsAtIdx("MachColumnLength", idx, int(rt))
		}
	}
	return int(length), nil
}

// returns true if not null
func machColumnData(stmt unsafe.Pointer, idx int, buf unsafe.Pointer, bufLen int) (bool, error) {
	var isNull C.char
	if rt := C.MachColumnData(stmt, C.int(idx), buf, C.int(bufLen), &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return false, stmtErr
		} else {
			return false, ErrDatabaseReturnsAtIdx("MachColumnData", idx, int(rt))
		}
	}
	return isNull == 0, nil
}

// returns int16 and true if NOT NULL, false if NULL
func machColumnDataInt16(stmt unsafe.Pointer, idx int) (int16, bool, error) {
	var val C.short
	var isNull C.char
	if rt := C.MachColumnDataInt16(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataInt16", idx, int(rt))
		}
	}
	return int16(val), isNull == 0, nil
}

// returns int32 and true if NOT NULL, false if NULL
func machColumnDataInt32(stmt unsafe.Pointer, idx int) (int32, bool, error) {
	var val C.int
	var isNull C.char
	if rt := C.MachColumnDataInt32(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataInt32", idx, int(rt))
		}
	}
	return int32(val), isNull == 0, nil
}

// returns int64 and true if NOT NULL, false if NULL
func machColumnDataInt64(stmt unsafe.Pointer, idx int) (int64, bool, error) {
	var val C.longlong
	var isNull C.char
	if rt := C.MachColumnDataInt64(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataInt64", idx, int(rt))
		}
	}
	return int64(val), isNull == 0, nil
}

// returns Time and true if NOT NULL, false if NULL
func machColumnDataDateTime(stmt unsafe.Pointer, idx int) (time.Time, bool, error) {
	var val C.longlong
	var isNull C.char
	if rt := C.MachColumnDataDateTime(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return time.Time{}, false, stmtErr
		} else {
			return time.Time{}, false, ErrDatabaseReturnsAtIdx("MachColumnDataDateTime", idx, int(rt))
		}
	}
	return time.Unix(0, int64(val)), isNull == 0, nil
}

// returns float32 and true if NOT NULL, false if NULL
func machColumnDataFloat32(stmt unsafe.Pointer, idx int) (float32, bool, error) {
	var val C.float
	var isNull C.char
	if rt := C.MachColumnDataFloat(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataFloat", idx, int(rt))
		}
	}
	return float32(val), isNull == 0, nil
}

// returns float64 and true if NOT NULL, false if NULL
func machColumnDataFloat64(stmt unsafe.Pointer, idx int) (float64, bool, error) {
	var val C.double
	var isNull C.char
	if rt := C.MachColumnDataDouble(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataDouble", idx, int(rt))
		}
	}
	return float64(val), isNull == 0, nil
}

// returns net.IP (v4) and true if NOT NULL, false if NULL
func machColumnDataIPv4(stmt unsafe.Pointer, idx int) (net.IP, bool, error) {
	var val [net.IPv4len + 1]byte
	var isNull C.char
	// 주의) val[0]는 IP version
	if rt := C.MachColumnDataIPV4(stmt, C.int(idx), unsafe.Pointer(&val), &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return net.IPv6zero, false, stmtErr
		} else {
			return net.IPv4zero, false, ErrDatabaseReturnsAtIdx("MachColumnDataIPv4", idx, int(rt))
		}
	}
	return net.IP(val[1:]), isNull == 0, nil
}

// returns net.IP (v6) and true if NOT NULL, false if NULL
func machColumnDataIPv6(stmt unsafe.Pointer, idx int) (net.IP, bool, error) {
	var val [net.IPv6len + 1]byte
	var isNull C.char
	// 주의) val[0]는 IP version
	if rt := C.MachColumnDataIPV6(stmt, C.int(idx), unsafe.Pointer(&val), &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return net.IPv6zero, false, stmtErr
		} else {
			return net.IPv6zero, false, ErrDatabaseReturnsAtIdx("MachColumnDataIPv6", idx, int(rt))
		}
	}
	return net.IP(val[1:]), isNull == 0, nil
}

// returns string and true if NOT NULL, false if NULL
func machColumnDataString(stmt unsafe.Pointer, idx int) (string, bool, error) {
	length, err := machColumnLength(stmt, idx)
	if err != nil {
		return "", false, ErrDatabaseWrap("machColumnDataString", err)
	}
	if length == 0 {
		return "", false, nil
	}
	buf := make([]byte, length)
	val := (*C.char)(unsafe.Pointer(&buf[0]))
	var isNull C.char
	if rt := C.MachColumnDataString(stmt, C.int(idx), val, C.int(length), &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return "", false, stmtErr
		} else {
			return "", false, ErrDatabaseReturnsAtIdx("MachColumnDataString", idx, int(rt))
		}
	}
	return string(buf), isNull == 0, nil
}

// returns []byte and true if NOT NULL, false if NULL
func machColumnDataBinary(stmt unsafe.Pointer, idx int) ([]byte, bool, error) {
	length, err := machColumnLength(stmt, idx)
	if err != nil {
		return nil, false, ErrDatabaseWrap("machColumnDataString", err)
	}
	if length == 0 {
		return []byte{}, false, nil
	}
	buf := make([]byte, length)
	var isNull C.char
	val := (*C.char)(unsafe.Pointer(&buf[0]))
	if rt := C.MachColumnDataString(stmt, C.int(idx), val, C.int(length), &isNull); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return nil, false, stmtErr
		} else {
			return nil, false, ErrDatabaseReturnsAtIdx("MachColumnDataString", idx, int(rt))
		}
	}
	return buf, isNull == 0, nil
}

func machAppendOpen(stmt unsafe.Pointer, tableName string) error {
	cstr := C.CString(tableName)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachAppendOpen(stmt, cstr); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachAppendOpen", int(rt))
		}
	}
	return nil
}

func machAppendClose(stmt unsafe.Pointer) (int64, int64, error) {
	var successCount C.ulonglong
	var failureCount C.ulonglong
	if rt := C.MachAppendClose(stmt, &successCount, &failureCount); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, 0, stmtErr
		} else {
			return 0, 0, ErrDatabaseReturns("MachAppendClose", int(rt))
		}
	}
	return int64(successCount), int64(failureCount), nil
}

func machAppendData(stmt unsafe.Pointer, values *C.MachEngineAppendParam) error {
	if rt := C.MachAppendData(stmt, values); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachAppendData", int(rt))
		}
	}
	return nil
}

func (ap *Appender) appendTable0(vals []any) error {
	if len(ap.columns) == 0 {
		return ErrDatabaseNoColumns(ap.tableName)
	}
	if len(ap.columns) != len(vals) {
		return ErrDatabaseLengthOfColumns(ap.tableName, len(ap.columns), len(vals))
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
			return ErrDatabaseAppendUnknownType(c.Type)
		case ColumnBufferTypeInt16:
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, c.Name, c.Type)
			case uint16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *uint16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case int16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *int16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case *float64:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case float64:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *float32:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case float32:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			}
		case ColumnBufferTypeInt32:
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, c.Name, c.Type)
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
			case *float64:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			case float64:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			case *float32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(*v)
			case float32:
				*(*C.int)(unsafe.Pointer(&buffer[i].mData[0])) = C.int(v)
			}
		case ColumnBufferTypeInt64:
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, c.Name, c.Type)
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
			case *float64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case float64:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			case *float32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(*v)
			case float32:
				*(*C.longlong)(unsafe.Pointer(&buffer[i].mData[0])) = C.longlong(v)
			}
		case ColumnBufferTypeFloat:
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, c.Name, c.Type)
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
		case ColumnBufferTypeDouble:
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, c.Name, c.Type)
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
		case ColumnBufferTypeDatetime:
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = nil
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = nil
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongTimeValueType(fmt.Sprintf("%T", v), ap.timeformat, c.Name, c.Type)
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
					return ErrDatabaseAppendWrongTimeValueType("int", ap.timeformat, c.Name, c.Type)
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
					return ErrDatabaseAppendWrongTimeValueType("int16", ap.timeformat, c.Name, c.Type)
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
					return ErrDatabaseAppendWrongTimeValueType("int32", ap.timeformat, c.Name, c.Type)
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
					return ErrDatabaseAppendWrongTimeValueType("int64", ap.timeformat, c.Name, c.Type)
				}
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case float64:
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
					return ErrDatabaseAppendWrongTimeValueType("float64", ap.timeformat, c.Name, c.Type)
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
					return ErrDatabaseAppendWrongTimeStringType(c.Name, c.Type)
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
					return ErrDatabaseAppendWrongTimeStringType(c.Name, c.Type)
				}
			}
		case ColumnBufferTypeIPv4:
			ip, ok := val.(net.IP)
			if !ok {
				return ErrDatabaseAppendWrongType(val, c.Name, c.Type)
			}
			if ipv4 := ip.To4(); ipv4 == nil {
				return ErrDatabaseAppendWrongType(val, c.Name, c.Type)
			} else {
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uchar(net.IPv4len)
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddrString = nil
				for i := 0; i < net.IPv4len; i++ {
					(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddr[i] = C.uchar(ipv4[i])
				}
			}
		case ColumnBufferTypeIPv6:
			ip, ok := val.(net.IP)
			if !ok {
				return ErrDatabaseAppendWrongType(val, c.Name, c.Type)
			}
			if ipv6 := ip.To16(); ipv6 == nil {
				return ErrDatabaseAppendWrongType(val, c.Name, c.Type)
			} else {
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uchar(net.IPv6len)
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddrString = nil
				for i := 0; i < net.IPv6len; i++ {
					(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddr[i] = C.uchar(ipv6[i])
				}
			}
		case ColumnBufferTypeString:
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, c.Name, c.Type)
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
		case ColumnBufferTypeBinary:
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, c.Name, c.Type)
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

	if ap.closed {
		return ErrDatabaseClosedAppender
	}
	if ap.conn == nil || !ap.conn.Connected() {
		return ErrDatabaseNoConnection
	}
	err := machAppendData(ap.stmt, &buffer[0])
	return err
}
