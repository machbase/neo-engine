//go:build (darwin && amd64) || (darwin && arm64)
// +build darwin,amd64 darwin,arm64

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_dummy_darwin_amd64.a
import "C"

const LibMachLinkInfo = "static_machengine_darwin_amd64_fog"
