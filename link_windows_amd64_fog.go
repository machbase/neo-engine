//go:build windows && amd64 && fog_edition
// +build windows,amd64,fog_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_fog_windows_amd64.lib
import "C"

const LibMachLinkInfo = "static_fog_windows_amd64"
