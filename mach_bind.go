package mach

import (
	"net"
	"time"
	"unsafe"
)

func bind(stmt unsafe.Pointer, idx int, c any) error {
	if c == nil {
		if err := machBindNull(stmt, idx); err != nil {
			return ErrDatabaseBindNull(idx, err)
		}
		return nil
	}
	switch cv := c.(type) {
	case int:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *int:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case uint:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *uint:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case int16:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *int16:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case uint16:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *uint16:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case int32:
		if err := machBindInt32(stmt, idx, cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *int32:
		if err := machBindInt32(stmt, idx, *cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case uint32:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *uint32:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case int64:
		if err := machBindInt64(stmt, idx, cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *int64:
		if err := machBindInt64(stmt, idx, *cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case uint64:
		if err := machBindInt64(stmt, idx, int64(cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *uint64:
		if err := machBindInt64(stmt, idx, int64(*cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case float32:
		if err := machBindFloat64(stmt, idx, float64(cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *float32:
		if err := machBindFloat64(stmt, idx, float64(*cv)); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case float64:
		if err := machBindFloat64(stmt, idx, cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *float64:
		if err := machBindFloat64(stmt, idx, *cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case string:
		if err := machBindString(stmt, idx, cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *string:
		if err := machBindString(stmt, idx, *cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case []byte:
		if err := machBindBinary(stmt, idx, cv); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case net.IP:
		if err := machBindString(stmt, idx, cv.String()); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case time.Time:
		if err := machBindInt64(stmt, idx, cv.UnixNano()); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	case *time.Time:
		if err := machBindInt64(stmt, idx, cv.UnixNano()); err != nil {
			return ErrDatabaseBind(idx, c, err)
		}
	default:
		return ErrDatabaseBindType(idx, c)
	}
	return nil
}
