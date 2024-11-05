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
#include <stdio.h>
#include <string.h>
#include <time.h>
#include <machcli.h>

extern void cliDefaultAppendErrorCallback(void* aStmtHandle, int aErrorCode, char* aErrorMessage, long aErrorBufLen, char* aRowBuf, long aRowBufLen);

static inline void cliAppendErrorCallback(void* aStmtHandle,
										  int   aErrorCode,
										  char* aErrorMessage,
										  long  aErrorBufLen,
										  char* aRowBuf,
										 long  aRowBufLen) {
	cliDefaultAppendErrorCallback(aStmtHandle, aErrorCode, aErrorMessage, aErrorBufLen, aRowBuf, aRowBufLen);
}
*/
import "C"

func LinkInfo() string {
	return LibMachLinkInfo
}

func EngInitialize(homeDir string, machPort int, flag int, envHandle *unsafe.Pointer) error {
	cstr := C.CString(homeDir)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachInitialize(cstr, C.int(machPort), C.int(flag), envHandle); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachInitialize", int(rt))
	}
}

func EngFinalize(envHandle unsafe.Pointer) {
	C.MachFinalize(envHandle)
}

func EngCreateDatabase(envHandle unsafe.Pointer) error {
	if rt := C.MachCreateDB(envHandle); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCreateDB", int(rt))
	}
}

func EngDestroyDatabase(envHandle unsafe.Pointer) error {
	if rt := C.MachDestroyDB(envHandle); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachDestroyDB", int(rt))
	}
}

func EngExistsDatabase(envHandle unsafe.Pointer) bool {
	rt := C.MachIsDBCreated(envHandle)
	return rt == 1
}

func EngRestoreDatabase(envHandle unsafe.Pointer, dbPath string) error {
	cstr := C.CString(dbPath)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachRestoreDB(envHandle, cstr); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachRestoreDB", int(rt))
	}
}

func EngStartup(envHandle unsafe.Pointer) error {
	if rt := C.MachStartupDB(envHandle); rt != 0 {
		dbErr := EngError(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachStartupDB", int(rt))
		}
	}
	return nil
}

func EngShutdown(envHandle unsafe.Pointer) error {
	if rt := C.MachShutdownDB(envHandle); rt == 0 {
		return nil
	} else {
		dbErr := EngError(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachShutdownDB", int(rt))
		}
	}
}

func EngConnectionCount(envHandle unsafe.Pointer) int {
	ret := C.MachGetConnectionCount(envHandle)
	return int(ret)
}

func EngConnect(envHandle unsafe.Pointer, username string, password string, conn *unsafe.Pointer) error {
	cusername := C.CString(username)
	cpassword := C.CString(password)
	defer func() {
		C.free(unsafe.Pointer(cusername))
		C.free(unsafe.Pointer(cpassword))
	}()
	if rt := C.MachConnect(envHandle, cusername, cpassword, conn); rt == 0 {
		return nil
	} else {
		dbErr := EngError(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachConnect", int(rt))
		}
	}
}

func EngConnectTrust(envHandle unsafe.Pointer, username string, conn *unsafe.Pointer) error {
	cusername := C.CString(username)
	defer func() {
		C.free(unsafe.Pointer(cusername))
	}()
	if rt := C.MachConnectNoAuth(envHandle, cusername, conn); rt == 0 {
		return nil
	} else {
		dbErr := EngError(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachConnect", int(rt))
		}
	}
}

func EngDisconnect(conn unsafe.Pointer) error {
	if rt := C.MachDisconnect(conn); rt == 0 {
		return nil
	} else {
		dbErr := EngError(conn)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachDisconnect", int(rt))
		}
	}
}

func EngCancel(conn unsafe.Pointer) error {
	if rt := C.MachCancel(conn); rt == 0 {
		return nil
	} else {
		dbErr := EngError(conn)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachCancel", int(rt))
		}
	}
}

func EngSessionID(conn unsafe.Pointer) (uint64, error) {
	rt := C.MachSessionID(conn)
	return uint64(rt), nil
}

func EngError(handle unsafe.Pointer) error {
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
func EngUserAuth(envHandle unsafe.Pointer, username string, password string) (bool, error) {
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

func EngExplain(stmt unsafe.Pointer, full bool) (string, error) {
	var cstr = [1024 * 16]C.char{}
	var mode = 0
	if full {
		mode = 1
	}
	if rt := C.MachExplain(stmt, &cstr[0], C.int(len(cstr)), C.int(mode)); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return "", stmtErr
		} else {
			return "", ErrDatabaseReturns("MachExplain", int(rt))
		}
	}
	return C.GoString(&cstr[0]), nil
}

func EngAllocStmt(conn unsafe.Pointer, stmt *unsafe.Pointer) error {
	var ptr unsafe.Pointer
	if rt := C.MachAllocStmt(conn, &ptr); rt != 0 {
		dbErr := EngError(conn)
		if dbErr != nil {
			return dbErr
		} else {
			return ErrDatabaseReturns("MachAllocStmt", int(rt))
		}
	}
	*stmt = ptr
	return nil
}

func EngFreeStmt(stmt unsafe.Pointer) error {
	if rt := C.MachFreeStmt(stmt); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachFreeStmt", int(rt))
		}
	}
	return nil
}

func EngPrepare(stmt unsafe.Pointer, sqlText string) error {
	cstr := C.CString(sqlText)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachPrepare(stmt, cstr); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachPrepare", int(rt))
		}
	}
	return nil
}

func EngExecute(stmt unsafe.Pointer) error {
	if rt := C.MachExecute(stmt); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachExecute", int(rt))
		}
	}
	return nil
}

func EngExecuteClean(stmt unsafe.Pointer) error {
	if rt := C.MachExecuteClean(stmt); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachExecuteClean", int(rt))
		}
	}
	return nil
}

func EngDirectExecute(stmt unsafe.Pointer, sqlText string) error {
	cstr := C.CString(sqlText)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachDirectExecute(stmt, cstr); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachDirectExecute", int(rt))
		}
	}
	return nil
}

func EngStmtType(stmt unsafe.Pointer) (int, error) {
	var typ C.int
	if rt := C.MachStmtType(stmt, &typ); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturns("MachStmtType", int(rt))
		}
	}
	return int(typ), nil
}

func EngEffectRows(stmt unsafe.Pointer) (int64, error) {
	var rn C.ulonglong
	if rt := C.MachEffectRows(stmt, &rn); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturns("MachEffectRows", int(rt))
		}
	}
	return int64(rn), nil
}

// return true if fetch success(record exists), otherwise false
func EngFetch(stmt unsafe.Pointer) (bool, error) {
	var fetchEnd C.int // 0 if record exists, otherwise 1
	if rt := C.MachFetch(stmt, &fetchEnd); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return false, stmtErr
		} else {
			return false, ErrDatabaseReturns("MachFetch", int(rt))
		}
	}
	return fetchEnd == 0, nil
}

func EngBindInt32(stmt unsafe.Pointer, idx int, val int32) error {
	if rt := C.MachBindInt32(stmt, C.int(idx), C.int(val)); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindInt32", idx, int(rt))
		}
	}
	return nil
}

func EngBindInt64(stmt unsafe.Pointer, idx int, val int64) error {
	if rt := C.MachBindInt64(stmt, C.int(idx), C.longlong(val)); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindInt64", idx, int(rt))
		}
	}
	return nil
}

func EngBindFloat64(stmt unsafe.Pointer, idx int, val float64) error {
	if rt := C.MachBindDouble(stmt, C.int(idx), C.double(val)); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindDouble", idx, int(rt))
		}
	}
	return nil
}

func EngBindString(stmt unsafe.Pointer, idx int, val string) error {
	cstr := C.CString(val)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachBindString(stmt, C.int(idx), cstr, C.int(len(val))); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindString", idx, int(rt))
		}
	}
	return nil
}

func EngBindBinary(stmt unsafe.Pointer, idx int, data []byte) error {
	ptr := unsafe.Pointer(&data[0])
	if rt := C.MachBindBinary(stmt, C.int(idx), ptr, C.int(len(data))); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindBinary", idx, int(rt))
		}
	}
	return nil
}

func EngBindNull(stmt unsafe.Pointer, idx int) error {
	if rt := C.MachBindNull(stmt, C.int(idx)); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturnsAtIdx("MachBindNull", idx, int(rt))
		}
	}
	return nil
}

func EngColumnCount(stmt unsafe.Pointer) (int, error) {
	var count C.int = 0
	if rt := C.MachColumnCount(stmt, &count); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturns("MachColumnCount", int(rt))
		}
	}
	return int(count), nil
}

func EngColumnInfo(stmt unsafe.Pointer, idx int, pName *string, pType *int, pSize *int, pLength *int) error {
	var nfo C.MachEngineColumnInfo
	if rt := C.MachColumnInfo(stmt, C.int(idx), &nfo); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachColumnInfo", int(rt))
		}
	}

	*pName = C.GoString(&nfo.mColumnName[0])
	*pType = int(nfo.mColumnType)
	*pSize = int(nfo.mColumnSize)
	*pLength = int(nfo.mColumnLength)
	return nil
}

func EngColumnName(stmt unsafe.Pointer, idx int) (string, error) {
	var cstr = [100]C.char{}
	if rt := C.MachColumnName(stmt, C.int(idx), &cstr[0], C.int(len(cstr))); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return fmt.Sprintf("col-%d", idx), stmtErr
		} else {
			return fmt.Sprintf("col-%d", idx), ErrDatabaseReturns("MachColumnName", int(rt))
		}
	}
	return C.GoString(&cstr[0]), nil
}

func EngColumnType(stmt unsafe.Pointer, idx int) (int, int, error) {
	var typ C.int = 0
	var siz C.int = 0
	if rt := C.MachColumnType(stmt, C.int(idx), &typ, &siz); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, 0, stmtErr
		} else {
			return 0, 0, ErrDatabaseReturnsAtIdx("MachColumnType", idx, int(rt))
		}
	}
	return int(typ), int(siz), nil
}

func EngColumnLength(stmt unsafe.Pointer, idx int) (int, error) {
	var length C.int = 0
	if rt := C.MachColumnLength(stmt, C.int(idx), &length); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, ErrDatabaseReturnsAtIdx("MachColumnLength", idx, int(rt))
		}
	}
	return int(length), nil
}

// returns true if not null
func EngColumnData(stmt unsafe.Pointer, idx int, buf unsafe.Pointer, bufLen int) (bool, error) {
	var isNull C.char
	if rt := C.MachColumnData(stmt, C.int(idx), buf, C.int(bufLen), &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return false, stmtErr
		} else {
			return false, ErrDatabaseReturnsAtIdx("MachColumnData", idx, int(rt))
		}
	}
	return isNull == 0, nil
}

// returns int16 and true if NOT NULL, false if NULL
func EngColumnDataInt16(stmt unsafe.Pointer, idx int) (int16, bool, error) {
	var val C.short
	var isNull C.char
	if rt := C.MachColumnDataInt16(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataInt16", idx, int(rt))
		}
	}
	return int16(val), isNull == 0, nil
}

// returns int32 and true if NOT NULL, false if NULL
func EngColumnDataInt32(stmt unsafe.Pointer, idx int) (int32, bool, error) {
	var val C.int
	var isNull C.char
	if rt := C.MachColumnDataInt32(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataInt32", idx, int(rt))
		}
	}
	return int32(val), isNull == 0, nil
}

// returns int64 and true if NOT NULL, false if NULL
func EngColumnDataInt64(stmt unsafe.Pointer, idx int) (int64, bool, error) {
	var val C.longlong
	var isNull C.char
	if rt := C.MachColumnDataInt64(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataInt64", idx, int(rt))
		}
	}
	return int64(val), isNull == 0, nil
}

// returns Time and true if NOT NULL, false if NULL
func EngColumnDataDateTime(stmt unsafe.Pointer, idx int) (time.Time, bool, error) {
	var val C.longlong
	var isNull C.char
	if rt := C.MachColumnDataDateTime(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return time.Time{}, false, stmtErr
		} else {
			return time.Time{}, false, ErrDatabaseReturnsAtIdx("MachColumnDataDateTime", idx, int(rt))
		}
	}
	return time.Unix(0, int64(val)), isNull == 0, nil
}

// returns float32 and true if NOT NULL, false if NULL
func EngColumnDataFloat32(stmt unsafe.Pointer, idx int) (float32, bool, error) {
	var val C.float
	var isNull C.char
	if rt := C.MachColumnDataFloat(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataFloat", idx, int(rt))
		}
	}
	return float32(val), isNull == 0, nil
}

// returns float64 and true if NOT NULL, false if NULL
func EngColumnDataFloat64(stmt unsafe.Pointer, idx int) (float64, bool, error) {
	var val C.double
	var isNull C.char
	if rt := C.MachColumnDataDouble(stmt, C.int(idx), &val, &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, false, stmtErr
		} else {
			return 0, false, ErrDatabaseReturnsAtIdx("MachColumnDataDouble", idx, int(rt))
		}
	}
	return float64(val), isNull == 0, nil
}

// returns net.IP (v4) and true if NOT NULL, false if NULL
func EngColumnDataIPv4(stmt unsafe.Pointer, idx int) (net.IP, bool, error) {
	var val [net.IPv4len + 1]byte
	var isNull C.char
	// 주의) val[0]는 IP version
	if rt := C.MachColumnDataIPV4(stmt, C.int(idx), unsafe.Pointer(&val), &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return net.IPv6zero, false, stmtErr
		} else {
			return net.IPv4zero, false, ErrDatabaseReturnsAtIdx("MachColumnDataIPv4", idx, int(rt))
		}
	}
	return net.IP(val[1:]), isNull == 0, nil
}

// returns net.IP (v6) and true if NOT NULL, false if NULL
func EngColumnDataIPv6(stmt unsafe.Pointer, idx int) (net.IP, bool, error) {
	var val [net.IPv6len + 1]byte
	var isNull C.char
	// 주의) val[0]는 IP version
	if rt := C.MachColumnDataIPV6(stmt, C.int(idx), unsafe.Pointer(&val), &isNull); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return net.IPv6zero, false, stmtErr
		} else {
			return net.IPv6zero, false, ErrDatabaseReturnsAtIdx("MachColumnDataIPv6", idx, int(rt))
		}
	}
	return net.IP(val[1:]), isNull == 0, nil
}

// returns string and true if NOT NULL, false if NULL
func EngColumnDataString(stmt unsafe.Pointer, idx int) (string, bool, error) {
	length, err := EngColumnLength(stmt, idx)
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
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return "", false, stmtErr
		} else {
			return "", false, ErrDatabaseReturnsAtIdx("MachColumnDataString", idx, int(rt))
		}
	}
	return string(buf), isNull == 0, nil
}

// returns []byte and true if NOT NULL, false if NULL
func EngColumnDataBinary(stmt unsafe.Pointer, idx int) ([]byte, bool, error) {
	length, err := EngColumnLength(stmt, idx)
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
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return nil, false, stmtErr
		} else {
			return nil, false, ErrDatabaseReturnsAtIdx("MachColumnDataString", idx, int(rt))
		}
	}
	return buf, isNull == 0, nil
}

func EngAppendOpen(stmt unsafe.Pointer, tableName string) error {
	cstr := C.CString(strings.ToUpper(tableName))
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachAppendOpen(stmt, cstr); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachAppendOpen", int(rt))
		}
	}
	return nil
}

func EngAppendClose(stmt unsafe.Pointer) (int64, int64, error) {
	var successCount C.ulonglong
	var failureCount C.ulonglong
	if rt := C.MachAppendClose(stmt, &successCount, &failureCount); rt != 0 {
		stmtErr := EngError(stmt)
		if stmtErr != nil {
			return 0, 0, stmtErr
		} else {
			return 0, 0, ErrDatabaseReturns("MachAppendClose", int(rt))
		}
	}
	return int64(successCount), int64(failureCount), nil
}

type AppendBuffer struct {
	stmt        unsafe.Pointer
	columnTypes []string
	columnNames []string
	buffer      []C.MachEngineAppendParam
}

func EngMakeAppendBuffer(stmt unsafe.Pointer, columnNames []string, columnTypes []string) *AppendBuffer {
	ret := &AppendBuffer{}
	ret.stmt = stmt
	ret.columnNames = columnNames
	ret.columnTypes = columnTypes
	ret.buffer = make([]C.MachEngineAppendParam, len(columnNames))
	return ret
}

func (ab *AppendBuffer) Append(vals ...any) error {
	if len(vals) != len(ab.columnNames) {
		return ErrDatabaseAppendWrongValueCount(len(ab.columnNames), len(vals))
	}
	for i, val := range vals {
		if val == nil {
			ab.buffer[i].mIsNull = 1
			continue
		} else {
			ab.buffer[i].mIsNull = 0
		}
		cName := ab.columnNames[i]
		cType := ab.columnTypes[i]
		buffer := ab.buffer

		switch cType {
		default:
			return ErrDatabaseAppendUnknownType(cType)
		case "short", "int16":
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, cName, cType)
			case uint16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *uint16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case int16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *int16:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case uint32:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *uint32:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(*v)
			case int32:
				*(*C.short)(unsafe.Pointer(&buffer[i].mData[0])) = C.short(v)
			case *int32:
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
		case "integer", "int32":
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, cName, cType)
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
		case "long", "int64":
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, cName, cType)
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
		case "float", "float32":
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, cName, cType)
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
		case "double", "float64":
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, cName, cType)
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
		case "datetime":
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = nil
			(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = nil
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongTimeValueType(fmt.Sprintf("%T", v), cName, cType)
			case time.Time:
				tv := v.UnixNano()
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case *time.Time:
				tv := v.UnixNano()
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int:
				tv := int64(v)
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int16:
				tv := int64(v)
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int32:
				tv := int64(v)
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case int64:
				tv := int64(v)
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
			case float64:
				tv := int64(v)
				(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = C.longlong(tv)
				// case string:
				// 	if len(timeformat) > 0 {
				// 		cstr := C.CString(v)
				// 		defer C.free(unsafe.Pointer(cstr))
				// 		cfmt := C.CString(timeformat)
				// 		defer C.free(unsafe.Pointer(cfmt))
				// 		(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = -2 // MACH_ENGINE_APPEND_DATETIME_STRING
				// 		(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = cstr
				// 		(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = cfmt
				// 	} else {
				// 		return ErrDatabaseAppendWrongTimeStringType(cName, cType)
				// 	}
				// case *string:
				// 	if len(timeformat) > 0 {
				// 		cstr := C.CString(*v)
				// 		defer C.free(unsafe.Pointer(cstr))
				// 		cfmt := C.CString(timeformat)
				// 		defer C.free(unsafe.Pointer(cfmt))
				// 		(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mTime = -2 // MACH_ENGINE_APPEND_DATETIME_STRING
				// 		(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mDateStr = cstr
				// 		(*C.MachEngineAppendDateTimeStruct)(unsafe.Pointer(&buffer[i].mData[0])).mFormatStr = cfmt
				// 	} else {
				// 		return ErrDatabaseAppendWrongTimeStringType(cName, cType)
				// 	}
			}
		case "ipv4":
			var ipv4 net.IP
			switch ip := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(val, cName, cType)
			case net.IP:
				if ipv4 = ip.To4(); ipv4 == nil {
					return ErrDatabaseAppendWrongType(val, cName, cType)
				}
			case string:
				if ipv4 = net.ParseIP(ip).To4(); ipv4 == nil {
					return ErrDatabaseAppendWrongType(val, cName, cType)
				}
			}
			(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uchar(net.IPv4len)
			(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddrString = nil
			for n := 0; n < net.IPv4len; n++ {
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddr[n] = C.uchar(ipv4[n])
			}
		case "ipv6":
			var ipv6 net.IP
			switch ip := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(val, cName, cType)
			case net.IP:
				if ipv6 = ip.To16(); ipv6 == nil {
					return ErrDatabaseAppendWrongType(val, cName, cType)
				}
			case string:
				if ipv6 = net.ParseIP(ip).To16(); ipv6 == nil {
					return ErrDatabaseAppendWrongType(val, cName, cType)
				}
			}
			(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mLength = C.uchar(net.IPv6len)
			(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddrString = nil
			for n := 0; n < net.IPv6len; n++ {
				(*C.MachEngineAppendIPStruct)(unsafe.Pointer(&buffer[i].mData[0])).mAddr[n] = C.uchar(ipv6[n])
			}
		case "varchar", "string", "json", "text":
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, cName, cType)
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
		case "binary":
			switch v := val.(type) {
			default:
				return ErrDatabaseAppendWrongType(v, cName, cType)
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

	if rt := C.MachAppendData(ab.stmt, &ab.buffer[0]); rt != 0 {
		stmtErr := EngError(ab.stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return ErrDatabaseReturns("MachAppendBuffer", int(rt))
		}
	}
	return nil
}

func CliInitialize(env *unsafe.Pointer) error {
	if rt := C.MachCLIInitialize(env); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIInitialize", int(rt))
	}
}

func CliFinalize(env unsafe.Pointer) error {
	if rt := C.MachCLIFinalize(env); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIFinalize", int(rt))
	}
}

func CliConnect(env unsafe.Pointer, connStr string, conn *unsafe.Pointer) error {
	cstr := C.CString(connStr)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachCLIConnect(env, cstr, conn); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIConnect", int(rt))
	}
}

func CliDisconnect(conn unsafe.Pointer) error {
	if rt := C.MachCLIDisconnect(conn); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIDisconnect", int(rt))
	}
}

type HandleType int

const (
	MACHCLI_HANDLE_ENV  HandleType = 1
	MACHCLI_HANDLE_DBC  HandleType = 2
	MACHCLI_HANDLE_STMT HandleType = 3
)

func CliError(handle unsafe.Pointer, handleType HandleType, code *int, msg *string) error {
	var ccode C.int
	var cmsg = [500]C.char{}
	if rt := C.MachCLIError(handle, C.int(handleType), &ccode, &cmsg[0], C.int(len(cmsg))); rt != 0 {
		return ErrDatabaseReturns("MachCLIError", int(rt))
	}
	*code = int(ccode)
	*msg = C.GoString(&cmsg[0])
	return nil
}

func CliAllocStmt(conn unsafe.Pointer, stmt *unsafe.Pointer) error {
	if rt := C.MachCLIAllocStmt(conn, stmt); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIAllocStmt", int(rt))
	}
}

func CliFreeStmt(stmt unsafe.Pointer) error {
	if rt := C.MachCLIFreeStmt(stmt); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIFreeStmt", int(rt))
	}
}

func CliPrepare(stmt unsafe.Pointer, query string) error {
	sqlCString := C.CString(query)
	defer C.free(unsafe.Pointer(sqlCString))
	if rt := C.MachCLIPrepare(stmt, sqlCString); rt != 0 {
		return ErrDatabaseReturns("MachCLIPrepare", int(rt))
	}
	return nil
}

func CliExecute(stmt unsafe.Pointer) error {
	if rt := C.MachCLIExecute(stmt); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIExecute", int(rt))
	}
}

func CliExecDirect(stmt unsafe.Pointer, query string) error {
	sqlCString := C.CString(query)
	defer C.free(unsafe.Pointer(sqlCString))
	if rt := C.MachCLIExecDirect(stmt, sqlCString); rt != 0 {
		return ErrDatabaseReturns("MachCLIExecDirect", int(rt))
	}
	return nil
}

func CliCancel(stmt unsafe.Pointer) error {
	if rt := C.MachCLICancel(stmt); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLICancel", int(rt))
	}
}

func CliRowCount(stmt unsafe.Pointer) (int64, error) {
	var count C.longlong
	if rt := C.MachCLIRowCount(stmt, &count); rt == 0 {
		return int64(count), nil
	} else {
		return 0, ErrDatabaseReturns("MachCLIRowCount", int(rt))
	}
}

// returns true if it reaches the end of fetch
func CliFetch(stmt unsafe.Pointer) (bool, error) {
	var end C.int
	if rt := C.MachCLIFetch(stmt, &end); rt == 0 {
		return end == 1, nil
	} else {
		return true, ErrDatabaseReturns("MachCLIFetch", int(rt))
	}
}

type SqlType int

const (
	MACHCLI_SQL_TYPE_INT16    SqlType = 0
	MACHCLI_SQL_TYPE_INT32    SqlType = 1
	MACHCLI_SQL_TYPE_INT64    SqlType = 2
	MACHCLI_SQL_TYPE_DATETIME SqlType = 3
	MACHCLI_SQL_TYPE_FLOAT    SqlType = 4
	MACHCLI_SQL_TYPE_DOUBLE   SqlType = 5
	MACHCLI_SQL_TYPE_IPV4     SqlType = 6
	MACHCLI_SQL_TYPE_IPV6     SqlType = 7
	MACHCLI_SQL_TYPE_STRING   SqlType = 8
	MACHCLI_SQL_TYPE_BINARY   SqlType = 9
)

type CType int

const (
	MACHCLI_C_TYPE_INT16  CType = 101
	MACHCLI_C_TYPE_INT32  CType = 102
	MACHCLI_C_TYPE_INT64  CType = 103
	MACHCLI_C_TYPE_FLOAT  CType = 104
	MACHCLI_C_TYPE_DOUBLE CType = 105
	MACHCLI_C_TYPE_CHAR   CType = 106
)

// returns the length of the actual data
func CliGetData(stmt unsafe.Pointer, columnNo int, cType CType, buf unsafe.Pointer, bufLen int) (int64, error) {
	var resultLen C.long
	if rt := C.MachCLIGetData(stmt, C.int(columnNo), C.int(cType), buf, C.int(bufLen), &resultLen); rt != 0 {
		return 0, ErrDatabaseReturnsAtIdx("MachCLIGetData", columnNo, int(rt))
	}
	return int64(resultLen), nil
}

type CliBindColData struct {
	Type      CType
	Buf       unsafe.Pointer
	BufLen    int
	resultLen C.long // can not use int64 because arm32 can not convert int64 to C.long
}

// returns the length of the actual data
func CliBindCol(stmt unsafe.Pointer, columnNo int, data *CliBindColData) error {
	if rt := C.MachCLIBindCol(stmt, C.int(columnNo), C.int(data.Type), data.Buf, C.int(data.BufLen), &data.resultLen); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturnsAtIdx("MachCLIBindCol", columnNo, int(rt))
	}
}

func CliBindParam(stmt unsafe.Pointer, paramNo int, cType CType, sqlType SqlType, value unsafe.Pointer, valueLen int) error {
	if rt := C.MachCLIBindParam(stmt, C.int(paramNo), C.int(cType), C.int(sqlType), value, C.int(valueLen)); rt != 0 {
		return ErrDatabaseReturnsAtIdx("MachCLIBindParam", paramNo, int(rt))
	}
	return nil
}

type CliParamDesc struct {
	Type      SqlType
	Precision int
	Scale     int
	Nullable  bool
}

func CliDescribeParam(stmt unsafe.Pointer, paramNo int) (CliParamDesc, error) {
	ret := CliParamDesc{}
	var typ, prec, scale, nullable C.int
	if rt := C.MachCLIDescribeParam(stmt, C.int(paramNo), &typ, &prec, &scale, &nullable); rt == 0 {
		ret.Type = SqlType(typ)
		ret.Precision = int(prec)
		ret.Scale = int(scale)
		ret.Nullable = nullable == 1
		return ret, nil
	} else {
		return ret, ErrDatabaseReturnsAtIdx("MachCLIDescribeParam", paramNo, int(rt))
	}
}

func CliNumParam(stmt unsafe.Pointer) (int, error) {
	var num C.int
	if rt := C.MachCLINumParam(stmt, &num); rt == 0 {
		return int(num), nil
	} else {
		return 0, ErrDatabaseReturns("MachCLINumParam", int(rt))
	}
}

func CliDescribeCol(stmt unsafe.Pointer, columnNo int, pName *string, pType *SqlType, pSize *int, pScale *int, pNullable *bool) error {
	var name = [200]C.char{}
	var nameSize = C.int(len(name))
	var nameLen, dataType, colSize, scale, nullable C.int
	if rt := C.MachCLIDescribeCol(stmt, C.int(columnNo), &name[0], nameSize, &nameLen, &dataType, &colSize, &scale, &nullable); rt == 0 {
		*pName = C.GoStringN(&name[0], nameLen)
		*pType = SqlType(dataType)
		*pSize = int(colSize)
		*pScale = int(scale)
		*pNullable = nullable == 1
		return nil
	} else {
		return ErrDatabaseReturnsAtIdx("MachCLIDescribeCol", columnNo, int(rt))
	}
}

func CliNumResultCol(stmt unsafe.Pointer) (int, error) {
	var num C.int
	if rt := C.MachCLINumResultCol(stmt, &num); rt == 0 {
		return int(num), nil
	} else {
		return 0, ErrDatabaseReturns("MachCLINumResultCol", int(rt))
	}
}

func CliAppendOpen(stmt unsafe.Pointer, tableName string, errCheckCount int) error {
	cstr := C.CString(tableName)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachCLIAppendOpen(stmt, cstr, C.int(errCheckCount)); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIAppendOpen", int(rt))
	}
}

func CliAppendData(stmt unsafe.Pointer, types []SqlType, names []string, args []any) error {
	if len(types) != len(args) || len(types) != len(names) {
		return ErrDatabaseAppendWrongValueCount(len(types), len(args))
	}

	data := make([]C.MachCLIAppendParam, len(args))
	for i, typ := range types {
		name := names[i]
		switch typ {
		case MACHCLI_SQL_TYPE_INT16:
			if args[i] == nil {
				x := 0x8000 // MACHCLI_APPEND_SHORT_NULL
				*(*C.short)(unsafe.Pointer(&data[i])) = C.short(x)
			} else {
				switch value := args[i].(type) {
				case int16:
					*(*C.short)(unsafe.Pointer(&data[i])) = C.short(value)
				default:
					return ErrDatabaseAppendWrongType(value, name, "MACHCLI_SQL_TYPE_INT16")
				}
			}
		case MACHCLI_SQL_TYPE_INT32:
			if args[i] == nil {
				x := uint32(0x80000000) // MACHCLI_APPEND_INTEGER_NULL
				*(*C.int)(unsafe.Pointer(&data[i])) = C.int(x)
			} else {
				switch value := args[i].(type) {
				case int32:
					*(*C.int)(unsafe.Pointer(&data[i])) = C.int(value)
				case int:
					*(*C.int)(unsafe.Pointer(&data[i])) = C.int(value)
				default:
					return ErrDatabaseAppendWrongType(value, name, "MACHCLI_SQL_TYPE_INT32")
				}
			}
		case MACHCLI_SQL_TYPE_INT64:
			if args[i] == nil {
				x := int64(-9223372036854775808) // MACHCLI_APPEND_LONG_NULL 0x8000000000000000
				*(*C.longlong)(unsafe.Pointer(&data[i])) = C.longlong(x)
			} else {
				switch value := args[i].(type) {
				case int:
					*(*C.longlong)(unsafe.Pointer(&data[i])) = C.longlong(value)
				case int32:
					*(*C.longlong)(unsafe.Pointer(&data[i])) = C.longlong(value)
				case int64:
					*(*C.longlong)(unsafe.Pointer(&data[i])) = C.longlong(value)
				default:
					return ErrDatabaseAppendWrongType(value, name, "MACHCLI_SQL_TYPE_INT64")
				}
			}
		case MACHCLI_SQL_TYPE_FLOAT:
			if args[i] == nil {
				x := float32(3.402823466e+38) // MACHCLI_APPEND_FLOAT_NULL
				*(*C.float)(unsafe.Pointer(&data[i])) = C.float(x)
			} else {
				switch value := args[i].(type) {
				case float32:
					*(*C.float)(unsafe.Pointer(&data[i])) = C.float(value)
				case float64:
					*(*C.float)(unsafe.Pointer(&data[i])) = C.float(value)
				case int:
					*(*C.float)(unsafe.Pointer(&data[i])) = C.float(value)
				case int32:
					*(*C.float)(unsafe.Pointer(&data[i])) = C.float(value)
				case int64:
					*(*C.float)(unsafe.Pointer(&data[i])) = C.float(value)
				default:
					return ErrDatabaseAppendWrongType(value, name, "MACHCLI_SQL_TYPE_FLOAT")
				}
			}
		case MACHCLI_SQL_TYPE_DOUBLE:
			if args[i] == nil {
				x := float64(1.7976931348623158e+308) // MACHCLI_APPEND_DOUBLE_NULL
				*(*C.double)(unsafe.Pointer(&data[i])) = C.double(x)
			} else {
				switch value := args[i].(type) {
				case float64:
					*(*C.double)(unsafe.Pointer(&data[i])) = C.double(value)
				case float32:
					*(*C.double)(unsafe.Pointer(&data[i])) = C.double(value)
				case int:
					*(*C.double)(unsafe.Pointer(&data[i])) = C.double(value)
				case int32:
					*(*C.double)(unsafe.Pointer(&data[i])) = C.double(value)
				case int64:
					*(*C.double)(unsafe.Pointer(&data[i])) = C.double(value)
				default:
					return ErrDatabaseAppendWrongType(value, name, "MACHCLI_SQL_TYPE_DOUBLE")
				}
			}
		case MACHCLI_SQL_TYPE_DATETIME:
			if args[i] == nil {
				x := int64(0) // MACHCLI_APPEND_DATETIME_NULL
				*(*C.longlong)(unsafe.Pointer(&data[i])) = C.longlong(x)
			} else {
				switch value := args[i].(type) {
				case time.Time:
					tv := value.UnixNano()
					*(*C.longlong)(unsafe.Pointer(&data[i])) = C.longlong(tv)
				case int16:
					tv := int64(value)
					*(*C.longlong)(unsafe.Pointer(&data[i])) = C.longlong(tv)
				default:
					return ErrDatabaseAppendWrongType(value, name, "MACHCLI_SQL_TYPE_DATETIME")
				}
			}
		case MACHCLI_SQL_TYPE_IPV4:
		case MACHCLI_SQL_TYPE_IPV6:
		case MACHCLI_SQL_TYPE_STRING:
			switch value := args[i].(type) {
			case string:
				cstr := []byte(value)
				(*C.MachCLIAppendVarStruct)(unsafe.Pointer(&data[i])).mLength = C.uint(len(cstr))
				(*C.MachCLIAppendVarStruct)(unsafe.Pointer(&data[i])).mData = unsafe.Pointer(&cstr[0])
			default:
				return ErrDatabaseAppendWrongType(value, name, "MACHCLI_SQL_TYPE_STRING")
			}
		case MACHCLI_SQL_TYPE_BINARY:
		}
	}

	if rt := C.MachCLIAppendData(stmt, (*C.MachCLIAppendParam)(&data[0])); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIAppendData", int(rt))
	}
}

func CliAppendClose(stmt unsafe.Pointer) (int64, int64, error) {
	var successCount C.longlong
	var failureCount C.longlong
	defer func() {
		delete(cliAppendErrorCallbacks, fmt.Sprintf("%X", stmt))
	}()
	if rt := C.MachCLIAppendClose(stmt, &successCount, &failureCount); rt == 0 {
		return int64(successCount), int64(failureCount), nil
	} else {
		return 0, 0, ErrDatabaseReturns("MachCLIAppendClose", int(rt))
	}
}

func CliAppendFlush(stmt unsafe.Pointer) error {
	if rt := C.MachCLIAppendFlush(stmt); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIAppendFlush", int(rt))
	}
}

type CLIAppendErrorCallback func(stmt unsafe.Pointer, errCode int, errMsg string, buf []byte)

var cliAppendErrorCallbacks map[string]CLIAppendErrorCallback

//export CliDefaultAppendErrorCallback
func CliDefaultAppendErrorCallback(stmt unsafe.Pointer, errCode C.int, errMsg *C.char, errMsgLen C.long, rowBuf *C.char, rowBufLen C.long) {
	msgLen := C.int(int64(errMsgLen))
	msg := C.GoStringN(errMsg, msgLen)
	buf := C.GoBytes(unsafe.Pointer(rowBuf), C.int(rowBufLen))
	if cb, ok := cliAppendErrorCallbacks[fmt.Sprintf("%X", stmt)]; ok {
		cb(stmt, int(errCode), msg, buf)
	}
}

func CliAppendSetErrorCallback(stmt unsafe.Pointer, cb CLIAppendErrorCallback) error {
	if rt := C.MachCLIAppendSetErrorCallback(stmt, (*[0]byte)(C.cliAppendErrorCallback)); rt == 0 {
		if cb != nil {
			cliAppendErrorCallbacks[fmt.Sprintf("%X", stmt)] = cb
		}
		return nil
	} else {
		return ErrDatabaseReturns("MachCLIAppendSetErrorCallback", int(rt))
	}
}

func CliSetConnectAppendFlush(conn unsafe.Pointer, opt int) error {
	if rt := C.MachCLISetConnectAppendFlush(conn, C.int(opt)); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLISetConnectAppendFlush", int(rt))
	}
}

func CliSetStmtAppendInterval(stmt unsafe.Pointer, intervalMilliseconds int) error {
	if rt := C.MachCLISetStmtAppendInterval(stmt, C.int(intervalMilliseconds)); rt == 0 {
		return nil
	} else {
		return ErrDatabaseReturns("MachCLISetStmtAppendInterval", int(rt))
	}
}
