//go:build linux && arm && edge_edition
// +build linux,arm,edge_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_edge_linux_arm32.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_edge_linux_armv7l"
