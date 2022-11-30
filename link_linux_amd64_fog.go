//go:build linux && amd64
// +build linux,amd64

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine.fog.LINUX.X86.64BIT.release.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_machengine_linux_amd64_fog"
