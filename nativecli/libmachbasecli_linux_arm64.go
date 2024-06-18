//go:build linux && arm64
// +build linux,arm64

package nativecli

// #cgo LDFLAGS: ${SRCDIR}/libmachbasecli_linux_arm64.a
import "C"

const LibMachLinkInfo = "static_cli_linux_arm64"
