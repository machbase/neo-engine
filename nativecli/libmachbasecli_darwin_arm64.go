//go:build darwin && arm64
// +build darwin,arm64

package nativecli

// #cgo LDFLAGS: ${SRCDIR}/libmachbasecli_darwin_arm64.a
import "C"

const LibMachLinkInfo = "static_cli_darwin_arm64"
