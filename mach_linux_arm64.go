//go:build linux && arm64
// +build linux,arm64

package mach

/*
#cgo CFLAGS: -I./native -I.
#cgo LDFLAGS: -L./native -lmachengine.edge.LINUX.ARM.64BIT.release -lm
#include "machEngine.h"
#include <stdlib.h>
*/
import "C"

const LibMachLinkInfo = "static_machengine_edge_linux_arm64"
