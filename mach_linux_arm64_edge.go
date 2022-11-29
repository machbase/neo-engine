//go:build linux && arm64
// +build linux,arm64

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine.edge.LINUX.ARM.64BIT.release.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_machengine_linux_arm64_edge"
