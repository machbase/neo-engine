package mach

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"net"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

func bind(stmt unsafe.Pointer, idx int, c any) error {
	if c == nil {
		if err := machBindNull(stmt, idx); err != nil {
			return errors.Wrapf(err, "bind error idx %d with NULL", idx)
		}
		return nil
	}
	switch cv := c.(type) {
	case int:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *int:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case uint:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *uint:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case int16:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *int16:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case uint16:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *uint16:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case int32:
		if err := machBindInt32(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *int32:
		if err := machBindInt32(stmt, idx, *cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case uint32:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *uint32:
		if err := machBindInt32(stmt, idx, int32(*cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case int64:
		if err := machBindInt64(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *int64:
		if err := machBindInt64(stmt, idx, *cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case uint64:
		if err := machBindInt64(stmt, idx, int64(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *uint64:
		if err := machBindInt64(stmt, idx, int64(*cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case float32:
		if err := machBindFloat64(stmt, idx, float64(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *float32:
		if err := machBindFloat64(stmt, idx, float64(*cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case float64:
		if err := machBindFloat64(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *float64:
		if err := machBindFloat64(stmt, idx, *cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case string:
		if err := machBindString(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *string:
		if err := machBindString(stmt, idx, *cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case []byte:
		if err := machBindBinary(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case net.IP:
		if err := machBindString(stmt, idx, cv.String()); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case time.Time:
		if err := machBindInt64(stmt, idx, cv.UnixNano()); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case *time.Time:
		if err := machBindInt64(stmt, idx, cv.UnixNano()); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	default:
		return fmt.Errorf("bind unsupported idx %d type %T", idx, c)
	}
	return nil
}
