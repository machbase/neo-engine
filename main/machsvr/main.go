package main

import (
	"github.com/machbase/booter"
	_ "github.com/machbase/cemlib/banner"
	mach "github.com/machbase/dbms-mach-go"
	_ "github.com/machbase/dbms-mach-go/server"
)

func main() {
	booter.SetVersionString(mach.VersionString())
	booter.Startup()
	booter.WaitSignal()
	booter.ShutdownAndExit(0)
}
