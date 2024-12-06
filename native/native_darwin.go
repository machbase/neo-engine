//go:build darwin
// +build darwin

package native

/*
#include <stdlib.h>
#include <stdio.h>
#include <signal.h>

sigset_t machengine_mask;

static inline void cliDarwinDisableSignalHandler() {
	sigemptyset(&machengine_mask);;
	sigaddset(&machengine_mask, SIGURG);
	sigprocmask(SIG_BLOCK, &machengine_mask, NULL);
}

static inline void cliDarwinEnableSignalHandler() {
	sigprocmask(SIG_UNBLOCK, &machengine_mask, NULL);
}
*/
import "C"

func InitSignalHandler() {
	C.cliDarwinDisableSignalHandler()
}

func DeinitSignalHandler() {
	C.cliDarwinEnableSignalHandler()
}
