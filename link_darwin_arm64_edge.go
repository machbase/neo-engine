//go:build darwin && arm64 && edge_edition
// +build darwin,arm64,edge_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_edge_darwin_arm64.a
import "C"

const LibMachLinkInfo = "static_darwin_arm64_edge"
