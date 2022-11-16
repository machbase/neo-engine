//go:build darwin && amd64
// +build darwin,amd64

package mach

// /*
// #cgo CFLAGS: -I./native -I.
// #cgo LDFLAGS: -L./native -lmachengine.LINUX.X86.64BIT.release -lpthread -ljemalloc -ldl -lm -lcrypto -Wl,-rpath=./lib
// #include "libmachengine.h"
// */
// import "C"

import (
	"time"
	"unsafe"
)

func initialize0(homeDir string) {
}

func destroyDatabase0() {
}

func createDatabase0() {
}

func startup0(timeout time.Duration) {
}

func shutdown0() bool {
	return false
}

func isRunning0() bool {
	return false
}

func execute0(sqlText string) {
}

func executeNewSession0(sqlText string) {
}

func query0(sqlText string, args ...any) (*rows, error) {
	rt := rows{
		sqlText: sqlText,
	}
	return &rt, nil
}

type rows struct {
	sqlText string
	stmt    unsafe.Pointer
}

func (rows *rows) Close() {
}

func (rows *rows) Next() bool {
	return false
}

func (rows *rows) Scan(cols ...any) error {
	return nil
}
