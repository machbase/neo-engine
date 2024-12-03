package mach

import "fmt"

var ErrDatabaseMach = func(code int, msg string) error {
	return fmt.Errorf("MACH-ERR %d %s", code, msg)
}

var ErrDatabaseCli = func(fn string, code int, msg string) error {
	if code == 0 {
		return fmt.Errorf("MACHCLI-ERR %s, %s()", msg, fn)
	} else {
		return fmt.Errorf("MACHCLI-ERR %d %s, %s", code, msg, fn)
	}
}

var ErrDatabaseReturns = func(fn string, rt int) error {
	return fmt.Errorf("%s returns %d", fn, rt)
}
var ErrDatabaseReturnsAtIdx = func(fn string, idx int, rt int) error {
	return fmt.Errorf("%s idx %d returns %d", fn, idx, rt)
}
var ErrDatabaseWrap = func(fn string, cause error) error {
	return fmt.Errorf("%s %s", fn, cause.Error())
}
var ErrDatabaseAppendUnknownType = func(typ string) error {
	return fmt.Errorf("MachAppendData unknown column type '%s'", typ)
}
var ErrDatabaseAppendWrongType = func(actual any, column string, typ string) error {
	return fmt.Errorf("MachAppendData cannot apply %T to %s (%s)", actual, column, typ)
}
var ErrDatabaseAppendWrongTimeValueType = func(actual string, column string, typ string) error {
	return fmt.Errorf("MachAppendData cannot apply %s to %s (%s)", actual, column, typ)
}
var ErrDatabaseAppendWrongTimeStringType = func(column string, typ string) error {
	return fmt.Errorf("MachAppendData cannot apply string without format to %s (%s)", column, typ)
}
var ErrDatabaseAppendWrongValueCount = func(expect int, actual int) error {
	return fmt.Errorf("MachAppendData required %d, but got %d", expect, actual)
}
