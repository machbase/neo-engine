package mach

import (
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"

	_ "github.com/machbase/neo-engine/native"
	"github.com/pkg/errors"
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
	if strings.Contains(nfo, "edge") {
		return "edge"
	} else if strings.Contains(nfo, "fog") {
		return "fog"
	} else {
		return "none"
	}
}

func initialize0(homeDir string, flag int, envHandle *unsafe.Pointer) error {
	cstr := C.CString(homeDir)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachInitialize(cstr, C.int(flag), envHandle); rt == 0 {
		return nil
	} else {
		return fmt.Errorf("MachInitialize returns %d", rt)
	}
}

func finalize0(envHandle unsafe.Pointer) {
	C.MachFinalize(envHandle)
}

func createDatabase0(envHandle unsafe.Pointer) error {
	if rt := C.MachCreateDB(envHandle); rt == 0 {
		return nil
	} else {
		return fmt.Errorf("MachCreateDB returns %d", rt)
	}
}

func destroyDatabase0(envHandle unsafe.Pointer) error {
	if rt := C.MachDestroyDB(envHandle); rt == 0 {
		return nil
	} else {
		return fmt.Errorf("MachDestroyDB returns %d", rt)
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
			fmt.Println("========>", dbErr.Error())
			return dbErr
		} else {
			return fmt.Errorf("MachStartupDB returns %d", rt)
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
			return fmt.Errorf("MachShutdownDB returns %d", rt)
		}
	}
}

func machError0(handle unsafe.Pointer) error {
	code := C.MachErrorCode(handle)
	msg := C.MachErrorMsg(handle)
	if code != 0 && msg != nil {
		return fmt.Errorf("MACH-ERR %d %s", code, C.GoString(msg))
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
		return false, fmt.Errorf("MachUserAuth returns %d", rt)
	}
}

func machExplain(stmt unsafe.Pointer) (string, error) {
	var cstr = [1024]C.char{}
	if rt := C.MachExplain(stmt, &cstr[0], C.int(len(cstr))); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return "", stmtErr
		} else {
			return "", fmt.Errorf("MachExplain returns %d", rt)
		}
	}
	return C.GoString(&cstr[0]), nil
}

func machAllocStmt(envHandle unsafe.Pointer, stmt *unsafe.Pointer) error {
	var ptr unsafe.Pointer
	if rt := C.MachAllocStmt(envHandle, &ptr); rt != 0 {
		dbErr := machError0(envHandle)
		if dbErr != nil {
			return dbErr
		} else {
			return fmt.Errorf("MachAllocStmt returns %d", rt)
		}
	}
	*stmt = ptr
	return nil
}

func machFreeStmt(envHandle unsafe.Pointer, stmt unsafe.Pointer) error {
	if rt := C.MachFreeStmt(envHandle, stmt); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return fmt.Errorf("MachFreeStmt returns %d", rt)
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
			return fmt.Errorf("MachPrepare returns %d", rt)
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
			return fmt.Errorf("MachExecute returns %d", rt)
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
			return fmt.Errorf("MachExecuteClean returns %d", rt)
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
			return fmt.Errorf("MachDirectExecute returns %d", rt)
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
			return 0, fmt.Errorf("MachStmtType returns %d", rt)
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
			return 0, fmt.Errorf("MachEffectRows returns %d", rt)
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
			return false, fmt.Errorf("MachFetch returns %d", rt)
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
			return fmt.Errorf("MachBindInt32 idx %d returns %d", idx, rt)
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
			return fmt.Errorf("MachBindInt64 idx %d returns %d", idx, rt)
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
			return fmt.Errorf("MachBindDouble idx %d returns %d", idx, rt)
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
			return fmt.Errorf("MachBindString idx %d returns %d", idx, rt)
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
			return fmt.Errorf("MachBindBinary idx %d returns %d", idx, rt)
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
			return fmt.Errorf("MachBindNull returns %d", rt)
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
			return 0, fmt.Errorf("MachColumnCount returns %d", rt)
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
			return nil, fmt.Errorf("MachColumnInfo returns %d", rt)
		}
	}

	typ, err := ColumnTypeString(ColumnType(nfo.mColumnType))
	if err != nil {
		return nil, fmt.Errorf("MachColumnInfo %s", err.Error())
	}

	return &Column{
		Name: C.GoString(&nfo.mColumnName[0]),
		Type: typ,
		Size: int(nfo.mColumnSize),
		Len:  int(nfo.mColumnLength),
	}, nil
}

func machColumnName(stmt unsafe.Pointer, idx int) (string, error) {
	var cstr = [100]C.char{}
	if rt := C.MachColumnName(stmt, C.int(idx), &cstr[0], C.int(len(cstr))); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return fmt.Sprintf("col-%d", idx), stmtErr
		} else {
			return fmt.Sprintf("col-%d", idx), fmt.Errorf("MachColumnName returns %d", rt)
		}
	}
	return C.GoString(&cstr[0]), nil
}

func machColumnType(stmt unsafe.Pointer, idx int) (ColumnType, ColumnSize, error) {
	var typ C.int = 0
	var siz C.int = 0
	if rt := C.MachColumnType(stmt, C.int(idx), &typ, &siz); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, 0, stmtErr
		} else {
			return 0, 0, fmt.Errorf("MachColumnType idx %d returns %d", idx, rt)
		}
	}
	return ColumnType(typ), ColumnSize(siz), nil
}

func machColumnLength(stmt unsafe.Pointer, idx int) (int, error) {
	var length C.int = 0
	if rt := C.MachColumnLength(stmt, C.int(idx), &length); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, fmt.Errorf("MachColumnLength idx %d returns %d", idx, rt)
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
			return false, fmt.Errorf("MachColumnData idx %d returns %d", idx, rt)
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
			return 0, false, fmt.Errorf("MachColumnDataInt16 idx %d returns %d", idx, rt)
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
			return 0, false, fmt.Errorf("MachColumnDataInt32 idx %d returns %d", idx, rt)
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
			return 0, false, fmt.Errorf("MachColumnDataInt64 idx %d returns %d", idx, rt)
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
			return time.Time{}, false, fmt.Errorf("MachColumnDataDateTime idx %d returns %d", idx, rt)
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
			return 0, false, fmt.Errorf("MachColumnDataFloat idx %d returns %d", idx, rt)
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
			return 0, false, fmt.Errorf("MachColumnDataDouble idx %d returns %d", idx, rt)
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
			return net.IPv4zero, false, fmt.Errorf("MachColumnDataIPv4 idx %d returns %d", idx, rt)
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
			return net.IPv6zero, false, fmt.Errorf("MachColumnDataIPv6 idx %d returns %d", idx, rt)
		}
	}
	return net.IP(val[1:]), isNull == 0, nil
}

// returns string and true if NOT NULL, false if NULL
func machColumnDataString(stmt unsafe.Pointer, idx int) (string, bool, error) {
	length, err := machColumnLength(stmt, idx)
	if err != nil {
		return "", false, errors.Wrap(err, "machColumnDataString")
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
			return "", false, fmt.Errorf("MachColumnDataString idx %d returns %d", idx, rt)
		}
	}
	return string(buf), isNull == 0, nil
}

// returns []byte and true if NOT NULL, false if NULL
func machColumnDataBinary(stmt unsafe.Pointer, idx int) ([]byte, bool, error) {
	length, err := machColumnLength(stmt, idx)
	if err != nil {
		return nil, false, errors.Wrap(err, "machColumnDataString")
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
			return nil, false, fmt.Errorf("MachColumnDataString idx %d returns %d", idx, rt)
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
			return fmt.Errorf("MachAppendOpen %s returns %d", tableName, rt)
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
			return 0, 0, fmt.Errorf("MachAppendClose returns %d", rt)
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
			return fmt.Errorf("MachAppendData returns %d", rt)
		}
	}
	return nil
}
