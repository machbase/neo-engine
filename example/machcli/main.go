//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"unsafe"

	mach "github.com/machbase/neo-engine/v8"
)

var global struct {
	CliEnv unsafe.Pointer
}

func main() {
	mach.CliDarwinSignalHandler()

	var cliEnvHandler unsafe.Pointer
	if err := mach.CliInitialize(&cliEnvHandler); err != nil {
		panic(err)
	}
	global.CliEnv = cliEnvHandler

	var conn unsafe.Pointer
	err := mach.CliConnect(global.CliEnv, "SERVER=127.0.0.1;UID=SYS;PWD=MANAGER;CONNTYPE=1;PORT_NO=5656", &conn)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 1_000_000; i++ {
		var stmt unsafe.Pointer
		err = mach.CliAllocStmt(conn, &stmt)
		if err != nil {
			panic(err)
		}

		err = mach.CliPrepare(stmt, `select count(*) from example`)
		if err != nil {
			panic(err)
		}

		err = mach.CliExecute(stmt)
		if err != nil {
			panic(err)
		}

		_, err := mach.CliFetch(stmt)
		if err != nil {
			panic(err)
		}

		resultCount := int64(-1)
		_, err = mach.CliGetData(stmt, 0, mach.MACHCLI_C_TYPE_INT64, unsafe.Pointer(&resultCount), 8)
		if err != nil {
			panic(err)
		}

		err = mach.CliFreeStmt(stmt)
		if err != nil {
			panic(err)
		}
		if i > 0 && i%10_000 == 0 {
			fmt.Println(i, resultCount)
		}
	}
	err = mach.CliDisconnect(conn)
	if err != nil {
		panic(err)
	}
}
