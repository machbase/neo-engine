//go:build darwin && amd64
// +build darwin,amd64

package mach

/*
#cgo CFLAGS: -I${SRCDIR}/native
#cgo LDFLAGS: -L${SRCDIR}/native -lmachengine.fog.MACOS.X86.64BIT.release -lm
#include "machEngine.h"
#include <stdlib.h>
*/
import "C"

const LibMachLinkInfo = "static_machengine_darwin_amd64_fog"
