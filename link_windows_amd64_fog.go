//go:build windows && amd64 && fog_edition
// +build windows,amd64,fog_edition

package mach

/*
#cgo LDFLAGS: ${SRCDIR}/native/libmachengine_fog_windows_amd64.a -lm -lws2_32 -lnetapi32 -ladvapi32 -liphlpapi -ldbghelp -lshell32 -luser32
*/
import "C"

const LibMachLinkInfo = "static_fog_windows_amd64"
