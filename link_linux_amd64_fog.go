//go:build linux && amd64
// +build linux,amd64

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_fog_linux_amd64.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_linux_amd64_fog"
