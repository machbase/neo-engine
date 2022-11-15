package main

/*

#cgo CFLAGS: -I ./native
#cgo LDFLAGS: -L ./native -lmachengine -lpthread -ljemalloc -ldl -lm -lcrypto -Wl,-rpath=./lib
#include "libmachengine.h"
#include <stdlib.h>
#include <stdio.h>

void FetchAll() {
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
	"fmt"
	"time"
)

func main() {

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
