//go:build darwin && arm64
// +build darwin,arm64

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_fog_darwin_arm64.a
import "C"

const LibMachLinkInfo = "static_darwin_arm64_fog"
