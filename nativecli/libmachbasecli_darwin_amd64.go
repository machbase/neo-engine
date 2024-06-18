//go:build darwin && amd64
// +build darwin,amd64

package nativecli

// #cgo LDFLAGS: ${SRCDIR}/libmachbasecli_darwin_amd64.a
import "C"

const LibMachLinkInfo = "static_cli_darwin_amd64"
