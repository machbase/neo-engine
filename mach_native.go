package mach

import (
	"fmt"
	"net"
	"time"
	"unsafe"

	_ "github.com/machbase/dbms-mach-go/native"
	"github.com/pkg/errors"
)

/*
#cgo CFLAGS: -I${SRCDIR}/native
#include "machEngine.h"
#include <stdlib.h>
*/
import "C"

func initialize0(homeDir string) error {
	cstr := C.CString(homeDir)
	defer C.free(unsafe.Pointer(cstr))
	if rt := C.MachInitialize(cstr); rt == 0 {
		return nil
	} else {
		return fmt.Errorf("MachInitialize returns %d", rt)
	}
}

func createDatabase0() error {
	if rt := C.MachCreateDB(); rt == 0 {
		return nil
	} else {
		return fmt.Errorf("MachCreateDB returns %d", rt)
	}
}

func destroyDatabase0() error {
	if rt := C.MachDestroyDB(); rt == 0 {
		return nil
	} else {
		return fmt.Errorf("MachDestroyDB returns %d", rt)
	}
}

func existsDatabase0() bool {
	rt := C.MachIsDBCreated()
	return rt == 1
}

func startup0(handle *unsafe.Pointer, timeout time.Duration) error {
	timeoutSec := C.int(timeout.Seconds())
	if rt := C.MachStartupDB(timeoutSec, handle); rt != 0 {
		dbErr := machError0(*handle)
		if dbErr != nil {
			return dbErr
		} else {
			return fmt.Errorf("MachStartupDB returns %d", rt)
		}
	}
	return nil
}

func shutdown0(handle unsafe.Pointer) error {
	if rt := C.MachShutdownDB(handle); rt == 0 {
		return nil
	} else {
		dbErr := machError0(handle)
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
		return fmt.Errorf("MachError %d %s", code, C.GoString(msg))
	}
	return nil
}

func machAllocStmt(handle unsafe.Pointer, stmt *unsafe.Pointer) error {
	var ptr unsafe.Pointer
	if rt := C.MachAllocStmt(handle, &ptr); rt != 0 {
		dbErr := machError0(handle)
		if dbErr != nil {
			return dbErr
		} else {
			return fmt.Errorf("MachAllocStmt returns %d", rt)
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

func machStmtType(stmt unsafe.Pointer) (int, error) {
	var typ C.int
	if rt := C.MachStmtType(stmt, &typ); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, fmt.Errorf("MachStmtType returns %d", rt)
		}
	}
	return int(typ), nil
}

func isFetchableStmtType(typ int) bool {
	return typ == 512 || typ == 516
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

func machFetch(stmt unsafe.Pointer) (bool, error) {
	var fetchEnd C.int
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

func machColumnType(stmt unsafe.Pointer, idx int) (int, int, error) {
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
	return int(typ), int(siz), nil
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

func machColumnData(stmt unsafe.Pointer, idx int, buf unsafe.Pointer, bufLen int) error {
	if rt := C.MachColumnData(stmt, C.int(idx), buf, C.int(bufLen)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return fmt.Errorf("MachColumnData idx %d returns %d", idx, rt)
		}
	}
	return nil
}

func machColumnDataInt16(stmt unsafe.Pointer, idx int) (int16, error) {
	var val C.short
	if rt := C.MachColumnDataInt16(stmt, C.int(idx), &val); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, fmt.Errorf("MachColumnDataInt16 idx %d returns %d", idx, rt)
		}
	}
	return int16(val), nil
}

func machColumnDataInt32(stmt unsafe.Pointer, idx int) (int32, error) {
	var val C.int
	if rt := C.MachColumnDataInt32(stmt, C.int(idx), &val); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, fmt.Errorf("MachColumnDataInt32 idx %d returns %d", idx, rt)
		}
	}
	return int32(val), nil
}

func machColumnDataInt64(stmt unsafe.Pointer, idx int) (int64, error) {
	var val C.longlong
	if rt := C.MachColumnDataInt64(stmt, C.int(idx), &val); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, fmt.Errorf("MachColumnDataInt64 idx %d returns %d", idx, rt)
		}
	}
	return int64(val), nil
}

func machColumnDataDateTime(stmt unsafe.Pointer, idx int) (time.Time, error) {
	var val C.longlong
	if rt := C.MachColumnDataDateTime(stmt, C.int(idx), &val); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return time.Time{}, stmtErr
		} else {
			return time.Time{}, fmt.Errorf("MachColumnDataDateTime idx %d returns %d", idx, rt)
		}
	}
	return time.Unix(0, int64(val)), nil
}

func machColumnDataFloat32(stmt unsafe.Pointer, idx int) (float32, error) {
	var val C.float
	if rt := C.MachColumnDataFloat(stmt, C.int(idx), &val); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, fmt.Errorf("MachColumnDataFloat idx %d returns %d", idx, rt)
		}
	}
	return float32(val), nil
}

func machColumnDataFloat64(stmt unsafe.Pointer, idx int) (float64, error) {
	var val C.double
	if rt := C.MachColumnDataDouble(stmt, C.int(idx), &val); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return 0, stmtErr
		} else {
			return 0, fmt.Errorf("MachColumnDataDouble idx %d returns %d", idx, rt)
		}
	}
	return float64(val), nil
}

func machColumnDataIPv4(stmt unsafe.Pointer, idx int) (net.IP, error) {
	var val [net.IPv4len + 1]byte
	// 주의) val[0]는 IP version
	if rt := C.MachColumnDataIPV4(stmt, C.int(idx), unsafe.Pointer(&val)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return net.IPv6zero, stmtErr
		} else {
			return net.IPv4zero, fmt.Errorf("MachColumnDataIPv4 idx %d returns %d", idx, rt)
		}
	}
	return net.IP(val[1:]), nil
}

func machColumnDataIPv6(stmt unsafe.Pointer, idx int) (net.IP, error) {
	var val [net.IPv6len + 1]byte
	// 주의) val[0]는 IP version
	if rt := C.MachColumnDataIPV6(stmt, C.int(idx), unsafe.Pointer(&val)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return net.IPv6zero, stmtErr
		} else {
			return net.IPv6zero, fmt.Errorf("MachColumnDataIPv6 idx %d returns %d", idx, rt)
		}
	}
	return net.IP(val[1:]), nil
}

func machColumnDataString(stmt unsafe.Pointer, idx int) (string, error) {
	length, err := machColumnLength(stmt, idx)
	if err != nil {
		return "", errors.Wrap(err, "machColumnDataString")
	}
	if length == 0 {
		return "", nil
	}
	buf := make([]byte, length)
	val := (*C.char)(unsafe.Pointer(&buf[0]))
	if rt := C.MachColumnDataString(stmt, C.int(idx), val, C.int(length)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return "", stmtErr
		} else {
			return "", fmt.Errorf("MachColumnDataString idx %d returns %d", idx, rt)
		}
	}
	return string(buf), nil
}

func machColumnDataBinary(stmt unsafe.Pointer, idx int) ([]byte, error) {
	length, err := machColumnLength(stmt, idx)
	if err != nil {
		return nil, errors.Wrap(err, "machColumnDataString")
	}
	if length == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, length)
	val := (*C.char)(unsafe.Pointer(&buf[0]))
	if rt := C.MachColumnDataString(stmt, C.int(idx), val, C.int(length)); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return nil, stmtErr
		} else {
			return nil, fmt.Errorf("MachColumnDataString idx %d returns %d", idx, rt)
		}
	}
	return buf, nil
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

func machAppendClose(stmt unsafe.Pointer) (uint64, uint64, error) {
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
	return uint64(successCount), uint64(failureCount), nil
}

type machAppendDataValue [32]byte

type machAppendDataNullValue struct {
	IsValid bool
	Value   machAppendDataValue
	cstr    *C.char
}

func (v *machAppendDataNullValue) Free() {
	if v.cstr != nil {
		C.free(unsafe.Pointer(v.cstr))
		v.cstr = nil
	}
}

func machAppendData(stmt unsafe.Pointer, valueArr []*machAppendDataNullValue) error {
	values := make([]C.MachEngineAppendParam, len(valueArr))
	for i, v := range valueArr {
		isNull := 0 // NOT NULL
		if !v.IsValid {
			isNull = 1 // NULL
		}
		values[i] = C.MachEngineAppendParam{
			mIsNull: C.int(isNull),
			mData:   C.MachEngineAppendParamData(v.Value),
		}
	}

	if rt := C.MachAppendData(stmt, &values[0]); rt != 0 {
		stmtErr := machError0(stmt)
		if stmtErr != nil {
			return stmtErr
		} else {
			return fmt.Errorf("MachAppendData returns %d", rt)
		}
	}
	return nil
}
