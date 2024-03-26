package spi

import (
	"errors"
	"fmt"
)

var ErrDatabaseNoFactory = errors.New("no database factory found")
var ErrDatabaseFactoryNotFound = func(name string) error { return fmt.Errorf("database factory '%s' not found", name) }
var ErrDatabaseMach = func(code int, msg string) error { return fmt.Errorf("MACH-ERR %d %s", code, msg) }
var ErrDatabaseReturns = func(fn string, rt int) error { return fmt.Errorf("%s returns %d", fn, rt) }
var ErrDatabaseReturnsAtIdx = func(fn string, idx int, rt int) error { return fmt.Errorf("%s idx %d returns %d", fn, idx, rt) }
var ErrDatabaseConnectID = func(cause string) error { return fmt.Errorf("connection id fail, %s", cause) }
var ErrDatabaseUnsupportedType = func(fn string, typ int) error { return fmt.Errorf("%s unsupported type %d", fn, typ) }
var ErrDatabaseWrap = func(fn string, cause error) error { return fmt.Errorf("%s %s", fn, cause.Error()) }
var ErrDatabaseScanType = func(from string, to any) error {
	return fmt.Errorf("scan convert from %s to %T not supported", from, to)
}
var ErrDatabaseScanIndex = func(idx int, len int) error { return fmt.Errorf("column %d is out of range %d", idx, len) }
