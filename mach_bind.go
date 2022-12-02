package mach

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"math/bits"
	"net"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

func bind(stmt unsafe.Pointer, idx int, c any) error {
	if c == nil {
		return fmt.Errorf("bind nil idx %d type %T", idx, c)
	}
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
	case net.IP:
		if err := machBindString(stmt, idx, cv.String()); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	case time.Time:
		if err := machBindInt64(stmt, idx, cv.UnixNano()); err != nil {
			return errors.Wrapf(err, "bind error idx %d type %T", idx, c)
		}
	default:
		return fmt.Errorf("bind unsupported idx %d type %T", idx, c)
	}
	return nil
}

func bindValue(c any) *machAppendDataNullValue {
	nv := &machAppendDataNullValue{
		IsValid: c != nil,
		Value:   machAppendDataValue{},
	}
	if !nv.IsValid {
		return nv
	}
	switch cv := c.(type) {
	case int16:
		*(*int16)(unsafe.Pointer(&nv.Value[0])) = cv
	case uint16:
		*(*uint16)(unsafe.Pointer(&nv.Value[0])) = cv
	case int:
		*(*int)(unsafe.Pointer(&nv.Value[0])) = cv
	case int32:
		*(*int32)(unsafe.Pointer(&nv.Value[0])) = cv
	case uint32:
		*(*uint32)(unsafe.Pointer(&nv.Value[0])) = cv
	case int64:
		*(*int64)(unsafe.Pointer(&nv.Value[0])) = cv
	case uint64:
		*(*uint64)(unsafe.Pointer(&nv.Value[0])) = cv
	case float32:
		*(*float32)(unsafe.Pointer(&nv.Value[0])) = cv
	case float64:
		*(*float64)(unsafe.Pointer(&nv.Value[0])) = cv
	case net.IP:
		if ipv4 := cv.To4(); ipv4 != nil { // ip v4
			*(*C.char)(unsafe.Pointer(&nv.Value[0])) = C.char(int8(4))
			for i := 0; i < net.IPv4len; i++ {
				nv.Value[1+i] = ipv4[i]
			}
		} else { // ip v6
			*(*C.char)(unsafe.Pointer(&nv.Value[0])) = C.char(int8(6))
			for i := 0; i < net.IPv6len; i++ {
				nv.Value[1+i] = cv[i]
			}
		}
	case string:
		nv.cstr = C.CString(cv)
		*(*uint)(unsafe.Pointer(&nv.Value[0])) = uint(len(cv))
		*(**C.char)(unsafe.Pointer(&nv.Value[bits.UintSize/8])) = nv.cstr
	case []byte:
		*(*uint)(unsafe.Pointer(&nv.Value[0])) = uint(len(cv))
		*(**C.char)(unsafe.Pointer(&nv.Value[bits.UintSize/8])) = (*C.char)(unsafe.Pointer(&cv[0]))
	case time.Time:
		*(*int64)(unsafe.Pointer(&nv.Value[0])) = cv.UnixNano()
	}

	return nv
}
