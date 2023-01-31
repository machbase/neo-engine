//go:build linux && amd64 && edge_edition
// +build linux,amd64,edge_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_edge_linux_amd64.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_edge_linux_amd64"
