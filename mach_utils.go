package mach

import (
	"fmt"
	"strconv"
	"time"
)

func convertInt16(v int16, c any) error {
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

func convertInt32(v int32, c any) error {
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

func convertInt64(v int64, c any) error {
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

func convertDateTime(v time.Time, c any) error {
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

func convertFloat32(v float32, c any) error {
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

func convertFloat64(v float64, c any) error {
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

func convertString(v string, c any) error {
	switch cv := c.(type) {
	case *string:
		*cv = v
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}

func convertBytes(v []byte, c any) error {
	switch cv := c.(type) {
	case *[]byte:
		copy(*cv, v)
	default:
		return fmt.Errorf("Scan convert from STRING to %T not supported", c)
	}
	return nil
}
