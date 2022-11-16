package mach

import "time"

func Initialize(homeDir string) {
	initialize0(homeDir)
}

func DestroyDatabase() {
	destroyDatabase0()
}

func CreateDatabase() {
	createDatabase0()
}

func Startup(timeout time.Duration) {
	startup0(timeout)
}

func Shutdown() bool {
	return shutdown0()
}

func IsRunning() bool {
	return isRunning0()
}

func Execute(sqlText string) {
	execute0(sqlText)
}

func ExecuteNewSession(sqlText string) {
	executeNewSession0(sqlText)
}

func Query(sqlText string, cols ...any) (Rows, error) {
	return query0(sqlText, cols...)
}

type Rows interface {
	Close()
	Next() bool
	Scan(cols ...any) error
}
