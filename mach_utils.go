package mach

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

func bind(stmt unsafe.Pointer, idx int, c any) error {
	switch cv := c.(type) {
	case int:
		if err := machBindInt32(stmt, idx, int32(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case int32:
		if err := machBindInt32(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case int64:
		if err := machBindInt64(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case float32:
		if err := machBindFloat64(stmt, idx, float64(cv)); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case float64:
		if err := machBindFloat64(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case string:
		if err := machBindString(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case []byte:
		if err := machBindBinary(stmt, idx, cv); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	default:
		return fmt.Errorf("bind supported idx %d type %T", idx, c)
	}
	return nil
}

func scanInt16(v int16, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *int16:
		*cv = int16(v)
	case *int32:
		*cv = int32(v)
	case *int64:
		*cv = int64(v)
	case *string:
		*cv = strconv.Itoa(int(v))
	default:
		return fmt.Errorf("Scan convert from INT16 to %T not supported", c)
	}
	return nil
}

func scanInt32(v int32, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *int16:
		*cv = int16(v)
	case *int32:
		*cv = int32(v)
	case *int64:
		*cv = int64(v)
	case *string:
		*cv = strconv.Itoa(int(v))
	default:
		return fmt.Errorf("Scan convert from INT32 to %T not supported", c)
	}
	return nil
}

func scanInt64(v int64, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *int16:
		*cv = int16(v)
	case *int32:
		*cv = int32(v)
	case *int64:
		*cv = int64(v)
	case *string:
		*cv = strconv.Itoa(int(v))
	default:
		return fmt.Errorf("Scan convert from INT64 to %T not supported", c)
	}
	return nil
}

func scanDateTime(v time.Time, c any) error {
	switch cv := c.(type) {
	case *int64:
		*cv = v.UnixNano()
	case *time.Time:
		*cv = v
	case *string:
		*cv = v.String()
	default:
		return fmt.Errorf("Scan convert from INT64 to %T not supported", c)
	}
	return nil
}

func scanFloat32(v float32, c any) error {
	switch cv := c.(type) {
	case *float32:
		*cv = v
	case *float64:
		*cv = float64(v)
	case *string:
		*cv = strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return fmt.Errorf("Scan convert from FLOAT32 to %T not supported", c)
	}
	return nil
}

func scanFloat64(v float64, c any) error {
	switch cv := c.(type) {
	case *float32:
		*cv = float32(v)
	case *float64:
		*cv = v
	case *string:
		*cv = strconv.FormatFloat(v, 'f', -1, 32)
	default:
		return fmt.Errorf("Scan convert from FLOAT64 to %T not supported", c)
	}
	return nil
}

func scanString(v string, c any) error {
	switch cv := c.(type) {
	case *string:
		*cv = v
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}

func scanBytes(v []byte, c any) error {
	switch cv := c.(type) {
	case *[]byte:
		copy(*cv, v)
	case *string:
		*cv = string(v)
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}
