//go:build linux && arm64 && edge_edition
// +build linux,arm64,edge_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_edge_linux_arm64.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_edge_linux_arm64"
