//go:build linux && arm64 && fog_edition
// +build linux,arm64,fog_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_edge_linux_amd64.a -lm -ldl
import "C"

const LibMachLinkInfo = "static_linux_amd64_edge"
