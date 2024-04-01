package mach

import (
	"errors"
	"fmt"

	we "github.com/pkg/errors"
)

var ErrDatabaseNoFactory = errors.New("no database factory found")

var ErrDatabaseFactoryNotFound = func(name string) error {
	return fmt.Errorf("database factory '%s' not found", name)
}
var ErrDatabaseMach = func(code int, msg string) error {
	return fmt.Errorf("MACH-ERR %d %s", code, msg)
}
var ErrDatabaseReturns = func(fn string, rt int) error {
	return fmt.Errorf("%s returns %d", fn, rt)
}
var ErrDatabaseReturnsAtIdx = func(fn string, idx int, rt int) error {
	return fmt.Errorf("%s idx %d returns %d", fn, idx, rt)
}
var ErrDatabaseConnectID = func(cause string) error {
	return fmt.Errorf("connection id fail, %s", cause)
}
var ErrDatabaseUnsupportedType = func(fn string, typ int) error {
	return fmt.Errorf("%s unsupported type %d", fn, typ)
}
var ErrDatabaseWrap = func(fn string, cause error) error {
	return fmt.Errorf("%s %s", fn, cause.Error())
}
var ErrDatabaseNoColumns = func(table string) error {
	return fmt.Errorf("table '%s' has no columns", table)
}
var ErrDatabaseLengthOfColumns = func(table string, expectColumns int, actualColumns int) error {
	return fmt.Errorf("value count %d, table '%s' requres %d columns to append", actualColumns, table, expectColumns)
}
var ErrDatabaseAppendUnknownType = func(typ string) error {
	return fmt.Errorf("MachAppendData unknown column type '%s'", typ)
}
var ErrDatabaseAppendWrongType = func(actual any, column string, typ string) error {
	return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", actual, column, typ)
}
var ErrDatabaseAppendWrongTimeStringType = func(column string, typ string) error {
	return fmt.Errorf("MachAppendData cannot apply string without format to %s (%s)", column, typ)
}
var ErrDatabaseAppendWrongTimeValueType = func(actual string, timeformat string, column string, typ string) error {
	return fmt.Errorf("MachAppendData cannot apply %s with %s to %s (%s)", actual, timeformat, column, typ)
}
var ErrDatabaseClosedAppender = errors.New("closed appender")

var ErrDatabaseNoConnection = errors.New("invalid connection")

var ErrDatabaseBindNull = func(idx int, err error) error {
	return fmt.Errorf("bind error idx %d with NULL, %q", idx, err.Error())
}
var ErrDatabaseBindType = func(idx int, val any) error {
	return fmt.Errorf("bind unsupported idx %d type %T", idx, val)
}
var ErrDatabaseBind = func(idx int, val any, err error) error {
	return we.Wrapf(err, "bind error idx %d with %T", idx, val)
}
var ErrDatabaseScan = func(err error) error {
	return we.Wrap(err, "scan")
}
var ErrDatabaseScanTypeName = func(typ string, err error) error {
	return we.Wrapf(err, "scan %s", typ)
}
var ErrDatabaseScanType = func(from string, to any) error {
	return fmt.Errorf("scan convert from %s to %T not supported", from, to)
}
var ErrDatabaseScanUnsupportedType = func(to any) error {
	return fmt.Errorf("scan unsupported type %T", to)
}
var ErrDatabaseScanIndex = func(idx int, len int) error {
	return fmt.Errorf("scan column %d is out of range %d", idx, len)
}
var ErrDatabaseFetch = func(err error) error {
	return we.Wrap(err, "fetch")
}
