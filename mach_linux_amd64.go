//go:build linux && amd64
// +build linux,amd64

package mach

/*
#cgo CFLAGS: -I./native -I.
#cgo LDFLAGS: -L./native -lmachengine.LINUX.X86.64BIT.release -lpthread -ljemalloc -ldl -lm -lcrypto -Wl,-rpath=./lib
#include "libmachengine.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"time"
	"unsafe"
)

func initialize0(homeDir string) {
	C.MachInitialize(C.CString(homeDir))
}

func destroyDatabase0() {
	C.MachDestroyDB()
}

func createDatabase0() {
	C.MachCreateDB()
}

func startup0(timeout time.Duration) {
	i := C.nbp_uint32_t(uint32(timeout.Seconds()))
	C.MachStartupDB(i)
}

func shutdown0() bool {
	rt := C.MachShutdownDB()
	return rt == 0
}

func isRunning0() bool {
	rt := C.MachCheckEqualServerStatus(C.nbp_bool_t(1))
	return rt == 1
}

func execute0(sqlText string) {
	C.MachDirectSQLExecute(C.CString(sqlText))
}

func executeNewSession0(sqlText string) {
	C.MachDirectSQLOnNewSession(C.CString(sqlText))
}

func query0(sqlText string, args ...any) (Rows, error) {
	rt := rows{
		sqlText: sqlText,
	}
	rc := C.MachAllocStmt(&rt.stmt)
	if rc != 0 {
		return nil, errors.New("MachAllocStmt")
	}
	rc = C.MachPrepare(rt.stmt, C.CString(sqlText))
	if rc != 0 {
		return nil, errors.New("MachPrepare")
	}
	rc = C.MachExecute(rt.stmt)
	if rc != 0 {
		return nil, errors.New("MachExecute")
	}

	return &rt, nil
}

type rows struct {
	sqlText string
	stmt    unsafe.Pointer
	eor     C.nbp_bool_t // end of resultset
}

func (rows *rows) Close() {
	C.MachExecuteClean(rows.stmt)
	C.MachPrepareClean(rows.stmt)
	C.MachFreeStmt(rows.stmt)
}

func (rows *rows) Next() bool {
	if rows.eor != 0 {
		return false
	}
	C.MachFetch(rows.stmt, &rows.eor)
	return rows.eor == 0
}

func (rows *rows) Scan(cols ...any) error {
	var buff [100]byte
	for i, c := range cols {
		var ptr unsafe.Pointer
		switch col := c.(type) {
		case *uint:
			ptr = (unsafe.Pointer)(col)
		case *int:
			ptr = (unsafe.Pointer)(col)
		case *uint32:
			ptr = (unsafe.Pointer)(col)
		case *int32:
			ptr = (unsafe.Pointer)(col)
		case *uint64:
			ptr = (unsafe.Pointer)(col)
		case *int64:
			ptr = (unsafe.Pointer)(col)
		case *string:
			// ptr = (unsafe.Pointer)(col)
			ptr = (unsafe.Pointer)(&buff)
		case *float32:
			ptr = (unsafe.Pointer)(col)
		case *float64:
			ptr = (unsafe.Pointer)(col)
		default:
			return fmt.Errorf("MachGetColumnData unsupported type %T", c)
		}
		rc := C.MachGetColumnData(rows.stmt, C.uint(i), ptr)
		if rc != 0 {
			return errors.New("MachGetColumnData")
		}
		// fmt.Printf("[%d] %T %p\n", i, c, c)

		switch col := c.(type) {
		case *string:
			*col = string(buff[:])
		default:
		}
	}
	return nil
}