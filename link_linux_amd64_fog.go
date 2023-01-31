//go:build linux && amd64 && fog_edition
// +build linux,amd64,fog_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_fog_linux_amd64.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_fog_linux_amd64"
