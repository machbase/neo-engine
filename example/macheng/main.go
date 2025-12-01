//go:build run

package main

import (
	_ "embed"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unsafe"

	mach "github.com/machbase/neo-engine/v8"
)

//go:embed machbase.conf
var machbase_conf []byte
var machPort = 5656
var machHome = "./tmp/machbase_home"
var machInit = uint32(0x0)

func main() {
	homePath, err := filepath.Abs(machHome)
	if err != nil {
		panic(err)
	}
	confPath := filepath.Join(homePath, "conf", "machbase.conf")

	os.RemoveAll(homePath)
	os.MkdirAll(homePath, 0755)
	os.MkdirAll(filepath.Join(homePath, "conf"), 0755)
	os.MkdirAll(filepath.Join(homePath, "trc"), 0755)
	os.MkdirAll(filepath.Join(homePath, "dbs"), 0755)
	os.WriteFile(confPath, machbase_conf, 0644)

	var svrEnvHandle unsafe.Pointer
	if err := mach.EngInitialize(homePath, machPort, machInit, &svrEnvHandle); err != nil {
		panic(err)
	}

	if !mach.EngExistsDatabase(svrEnvHandle) {
		mach.EngCreateDatabase(svrEnvHandle)
	}

	if err := mach.EngStartup(svrEnvHandle); err != nil {
		panic(err)
	}

	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptSignal

	if err := mach.EngShutdown(svrEnvHandle); err != nil {
		panic(err)
	}
	mach.EngFinalize(svrEnvHandle)
}
