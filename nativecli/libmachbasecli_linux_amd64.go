//go:build linux && amd64
// +build linux,amd64

package nativecli

// #cgo LDFLAGS: ${SRCDIR}/libmachbasecli_linux_amd64.a
import "C"

const LibMachLinkInfo = "static_cli_linux_amd64"
