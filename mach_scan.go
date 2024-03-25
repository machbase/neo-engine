package mach

import (
	"net"
	"strconv"
	"time"

	"github.com/machbase/neo-server/spi"
)

func ScanInt16(v int16, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *uint:
		*cv = uint(v)
	case *int16:
		*cv = int16(v)
	case *uint16:
		*cv = uint16(v)
	case *int32:
		*cv = int32(v)
	case *uint32:
		*cv = uint32(v)
	case *int64:
		*cv = int64(v)
	case *uint64:
		*cv = uint64(v)
	case *string:
		*cv = strconv.Itoa(int(v))
	default:
		return spi.ErrDatabaseScanType("INT16", c)
	}
	return nil
}

func ScanInt32(v int32, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *uint:
		*cv = uint(v)
	case *int16:
		*cv = int16(v)
	case *uint16:
		*cv = uint16(v)
	case *int32:
		*cv = int32(v)
	case *uint32:
		*cv = uint32(v)
	case *int64:
		*cv = int64(v)
	case *uint64:
		*cv = uint64(v)
	case *string:
		*cv = strconv.FormatInt(int64(v), 10)
	default:
		return spi.ErrDatabaseScanType("INT32", c)
	}
	return nil
}

func ScanInt64(v int64, c any) error {
	switch cv := c.(type) {
	case *int:
		*cv = int(v)
	case *uint:
		*cv = uint(v)
	case *int16:
		*cv = int16(v)
	case *uint16:
		*cv = uint16(v)
	case *int32:
		*cv = int32(v)
	case *uint32:
		*cv = uint32(v)
	case *int64:
		*cv = int64(v)
	case *uint64:
		*cv = uint64(v)
	case *string:
		*cv = strconv.FormatInt(v, 10)
	case *time.Time:
		*cv = time.Unix(0, v)
	default:
		return spi.ErrDatabaseScanType("INT64", c)
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
		return spi.ErrDatabaseScanType("DATETIME", c)
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
		return spi.ErrDatabaseScanType("FLOAT32", c)
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
		*cv = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return spi.ErrDatabaseScanType("FLOAT64", c)
	}
	return nil
}

func ScanString(v string, c any) error {
	switch cv := c.(type) {
	case *string:
		*cv = v
	case *[]uint8:
		*cv = []uint8(v)
	case *net.IP:
		*cv = net.ParseIP(v)
	default:
		return spi.ErrDatabaseScanType("STRING", c)
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
		return spi.ErrDatabaseScanType("BYTES", c)
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
		return spi.ErrDatabaseScanType("IPv4", c)
	}
	return nil
}
