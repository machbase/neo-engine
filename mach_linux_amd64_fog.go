//go:build linux && amd64
// +build linux,amd64

package mach

/*
#cgo CFLAGS: -I${SRCDIR}/native
#cgo LDFLAGS: -L${SRCDIR}/native -lmachengine.fog.LINUX.X86.64BIT.release -lm
#include "machEngine.h"
#include <stdlib.h>
*/
import "C"

const LibMachLinkInfo = "static_machengine_linux_amd64_fog"
