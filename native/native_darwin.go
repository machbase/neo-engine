//go:build darwin
// +build darwin

package native

/*
#include <stdlib.h>
#include <stdio.h>
#include <signal.h>

static inline void cliDarwinDisableSignalHandler() {
	sigset_t mask;
	sigemptyset(&mask);;
	sigaddset(&mask, SIGURG);
	sigprocmask(SIG_BLOCK, &mask, NULL);
}
*/
import "C"

func InitSignalHandler() {
	C.cliDarwinDisableSignalHandler()
}
