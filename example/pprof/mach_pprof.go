package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	_ "runtime/pprof"
	"syscall"
	"unsafe"

	mach "github.com/machbase/neo-engine/v8"
)

var global = struct {
	CliEnv unsafe.Pointer
}{}

func checkErr(err error, msgAndArgs ...any) {
	if err != nil {
		if len(msgAndArgs) == 0 {
			panic(err)
		} else if len(msgAndArgs) == 1 {
			panic(fmt.Sprintf("%s %v", err.Error(), msgAndArgs[0]))
		} else {
			panic(fmt.Sprintf(fmt.Sprintf("%s %v", err.Error(), msgAndArgs[0]), msgAndArgs[1:]...))
		}
	}
}

func main() {
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGURG)
	go func() {
		for sig := range c {
			if s, ok := sig.(syscall.Signal); ok {
				fmt.Println("signal:", sig.String(), fmt.Sprintf("sig(0x%X)", int(s)))
			} else {
				fmt.Println("signal:", sig.String(), sig)
			}
		}
	}()

	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
		http.HandleFunc("/debug/pprof/", pprof.Index)
		http.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		http.HandleFunc("/debug/pprof/profile", pprof.Profile)
		http.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		http.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}()

	var cliEnvHandler unsafe.Pointer
	if err := mach.CliInitialize(&cliEnvHandler); err != nil {
		panic(err)
	}
	global.CliEnv = cliEnvHandler

	var conn unsafe.Pointer
	err := mach.CliConnect(global.CliEnv, "SERVER=127.0.0.1;UID=SYS;PWD=MANAGER;CONNTYPE=1;PORT_NO=5656", &conn)
	checkErr(err)

	for i := 0; true; i++ {
		var stmt unsafe.Pointer
		err = mach.CliAllocStmt(conn, &stmt)
		checkErr(err, i)

		err = mach.CliPrepare(stmt, `select count(*) from example`)
		checkErr(err, "iter=%d", i)

		err = mach.CliExecute(stmt)
		checkErr(err, "iter=%d", i)

		_, err := mach.CliFetch(stmt)
		checkErr(err, "iter=%d", i)

		resultCount := int64(-1)
		_, err = mach.CliGetData(stmt, 0, mach.MACHCLI_C_TYPE_INT64, unsafe.Pointer(&resultCount), 8)
		checkErr(err)

		err = mach.CliFreeStmt(stmt)
		checkErr(err, "iter=%d", i)

		if i > 0 && i%100_000 == 0 {
			fmt.Println("iter:", i)
		}
	}
	err = mach.CliDisconnect(conn)
	checkErr(err)
	os.Exit(0)
}
