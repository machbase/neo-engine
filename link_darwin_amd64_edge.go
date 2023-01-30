//go:build darwin && amd64 && edge_edition
// +build darwin,amd64,edge_edition

package mach

// #cgo LDFLAGS: ${SRCDIR}/native/libmachengine_edge_darwin_amd64.a
import "C"

const LibMachLinkInfo = "static_darwin_amd64_edge"
