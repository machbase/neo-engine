package mach

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"time"
)

func ScanInt16(v int16, c any, isNull *bool) error {
	if v == math.MinInt16 {
		*isNull = true
		return nil
	}
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
		return fmt.Errorf("scan convert from INT16 to %T not supported", c)
	}
	return nil
}

func ScanInt32(v int32, c any, isNull *bool) error {
	if v == math.MinInt32 {
		*isNull = true
		return nil
	}
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
		return fmt.Errorf("scan convert from INT32 to %T not supported", c)
	}
	return nil
}

func ScanInt64(v int64, c any, isNull *bool) error {
	if v == math.MinInt64 {
		*isNull = true
		return nil
	}
	*isNull = false
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
	case *time.Time:
		*cv = time.Unix(0, v)
	default:
		return fmt.Errorf("scan convert from INT64 to %T not supported", c)
	}
	return nil
}

func ScanDateTime(v time.Time, c any) error {
	switch cv := c.(type) {
	case *int64:
		*cv = v.UnixNano()
	case *time.Time:
		*cv = v
	case *string:
		*cv = v.String()
	default:
		return fmt.Errorf("scan convert from INT64 to %T not supported", c)
	}
	return nil
}

func ScanFloat32(v float32, c any) error {
	switch cv := c.(type) {
	case *float32:
		*cv = v
	case *float64:
		*cv = float64(v)
	case *string:
		*cv = strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return fmt.Errorf("scan convert from FLOAT32 to %T not supported", c)
	}
	return nil
}

func ScanFloat64(v float64, c any) error {
	switch cv := c.(type) {
	case *float32:
		*cv = float32(v)
	case *float64:
		*cv = v
	case *string:
		*cv = strconv.FormatFloat(v, 'f', -1, 32)
	default:
		return fmt.Errorf("scan convert from FLOAT64 to %T not supported", c)
	}
	return nil
}

func ScanString(v string, c any, isNull *bool) error {
	if len(v) == 0 {
		*isNull = true
		return nil
	}
	switch cv := c.(type) {
	case *string:
		*cv = v
	default:
		return fmt.Errorf("scan convert from STRING to %T not supported", c)
	}
	return nil
}

func ScanBytes(v []byte, c any) error {
	switch cv := c.(type) {
	case *[]uint8:
		*cv = v
	case *string:
		*cv = string(v)
	default:
		return fmt.Errorf("scan convert from STRING to %T not supported", c)
	}
	return nil
}

func ScanIP(v net.IP, c any) error {
	switch cv := c.(type) {
	case *net.IP:
		*cv = v
	case *string:
		*cv = v.String()
	default:
		return fmt.Errorf("scan convert from IPv4 to %T not supported", c)
	}
	return nil
}
