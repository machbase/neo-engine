//go:build linux && arm64 && fog_edition
// +build linux,arm64,fog_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_fog_linux_arm64.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_fog_linux_arm64"
