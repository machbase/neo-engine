//go:build darwin
// +build darwin

package native

/*
#include <stdlib.h>
#include <stdio.h>
#include <signal.h>
#include <string.h>
#include <time.h>

static void inline cliDarwinSignalHandler(int sig) {
	// ignore
}

static inline void cliDarwinDisableSignalHandler() {
     signal(SIGURG, cliDarwinSignalHandler);
}
*/
import "C"

func InitSignalHandler() {
	C.cliDarwinDisableSignalHandler()
}
