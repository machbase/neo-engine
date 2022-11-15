package mach

/*
#cgo CFLAGS: -I./native -I.
#cgo LDFLAGS: -L./native -lmachengine -lpthread -ljemalloc -ldl -lm -lcrypto -Wl,-rpath=./lib
#include "libmachengine.h"
#include <stdlib.h>
#include <stdio.h>

void FetchAllAgain() {
    void* stmt;
    nbp_bool_t      fetchEnd = 0;

    MachAllocStmt(&stmt);
    MachPrepare(stmt, "select * from log");
    MachExecute(stmt);

    MachFetch(stmt, &fetchEnd);

    while(fetchEnd == NBP_FALSE)
    {
        nbp_sint32_t    id;
        nbp_char_t      name[20] = {0, };
        nbp_double_t    pre;

        MachGetColumnData(stmt, 0, (void*)&id);
        MachGetColumnData(stmt, 1, (void*)name);
        MachGetColumnData(stmt, 2, (void*)&pre);

        printf("id[%d] name[%s] pre[%f]\n", id, name, pre);
        MachFetch(stmt, &fetchEnd);
    }

    MachExecuteClean(stmt);
    MachPrepareClean(stmt);
    MachFreeStmt(stmt);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"time"
	"unsafe"
)

func Initialize(homeDir string) {
	C.MachInitialize(C.CString(homeDir))
}

func DestroyDatabase() {
	C.MachDestroyDB()
}

func CreateDatabase() {
	C.MachCreateDB()
}

func Startup(timeout time.Duration) {
	i := C.nbp_uint32_t(uint32(timeout.Seconds()))
	C.MachStartupDB(i)
}

func Shutdown() {
	C.MachShutdownDB()
}

func IsRunning() bool {
	rt := C.MachCheckEqualServerStatus(C.nbp_bool_t(1))
	return rt == 1
}

func Execute(sqlText string) {
	C.MachDirectSQLExecute(C.CString(sqlText))
}

func ExecuteNewSession(sqlText string) {
	C.MachDirectSQLOnNewSession(C.CString(sqlText))
}

func Query(sqlText string, args ...any) (*Rows, error) {
	rt := Rows{}
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

type Rows struct {
	stmt unsafe.Pointer
	eor  C.nbp_bool_t // end of resultset
}

func (rows *Rows) Close() {
	C.MachExecuteClean(rows.stmt)
	C.MachPrepareClean(rows.stmt)
	C.MachFreeStmt(rows.stmt)
}

func (rows *Rows) Next() bool {
	if rows.eor != 0 {
		return false
	}
	C.MachFetch(rows.stmt, &rows.eor)
	return rows.eor == 0
}

func (rows *Rows) Scan(cols ...any) error {
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

/*
func main2() {

	fmt.Println("-------------------------------")

	// Inline C
	C.MachInitialize(C.CString("/home/eirny/Developer/sample-machdb/tmp/home"))

	C.MachDestroyDB()
	C.MachCreateDB()

	C.MachStartupDB(10)

	C.MachDirectSQLExecute(C.CString("alter system set trace_log_level=1023"))
	C.MachDirectSQLExecute(C.CString("create log table log(id int, name varchar(20), pre double)"))
	C.MachDirectSQLOnNewSession(C.CString("insert into log values(0, 'zero', 1.01)"))
	C.MachDirectSQLOnNewSession(C.CString("insert into log values(1, 'one', 2.0002)"))
	C.MachDirectSQLOnNewSession(C.CString("insert into log select id + 20, name, pre *4 from log"))

	C.FetchAll()

	time.Sleep(5 * time.Second)

	C.MachShutdownDB()

	//time.Sleep(10 * time.Second)

	fmt.Println("-------------------------------")
}
*/
