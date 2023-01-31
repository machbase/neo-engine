//go:build darwin && arm64 && fog_edition
// +build darwin,arm64,fog_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_fog_darwin_arm64.a
import "C"

const LibMachLinkInfo = "static_fog_darwin_arm64"
