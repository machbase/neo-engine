#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <signal.h>
#include <machcli.h>

void printError(void *handle, int handleType, char *fn);

static void cliDisableSignal() {
	int return_value;
	sigset_t newset;
	sigset_t *newset_p;

	newset_p = &newset;

	return_value = sigemptyset(newset_p);
	printf("---> cliDisableSignal sigemptyset: %d\n", return_value);

	return_value = sigaddset(newset_p, SIGURG);
	printf("---> cliDisableSignal sigaddset: %d\n", return_value);

	return_value = sigprocmask(SIG_SETMASK, newset_p, NULL);
	printf("---> cliDisableSignal sigprocmask: %d\n", return_value);
}

int main() {
    cliDisableSignal();

    void *env = NULL;
    int ret = 0;
    ret = MachCLIInitialize(&env);
    if (ret != 0) {
        printf("MachCLIInitialize failed\n");
        return -1;
    }

    void *conn = NULL;
    char *CONNSTR = "SERVER=127.0.0.1;UID=SYS;PWD=MANAGER;CONNTYPE=1;PORT_NO=5656";
    ret = MachCLIConnect(env, CONNSTR, &conn);
    if (ret != 0) {
        printError(env, MACHCLI_HANDLE_ENV, "MachCLIConnect");
        return -1;
    }

    for (int i = 0; i < 1000000; i++) {
        void *stmt = NULL;
        ret = MachCLIAllocStmt(conn, &stmt);
        if (ret != 0) {
            printError(conn, MACHCLI_HANDLE_DBC, "MachCLIAllocStmt");
            return -1;
        }

        ret = MachCLIPrepare(stmt, "select count(*) from example");
        if (ret != 0) {
            printError(stmt, MACHCLI_HANDLE_STMT, "MachCLIPrepare");
            return -1;
        }

        ret = MachCLIExecute(stmt);
        if (ret != 0) {
            printError(stmt, MACHCLI_HANDLE_STMT, "MachCLIExecute");
            return -1;
        }

        int eof = 0;
        ret = MachCLIFetch(stmt, &eof);
        if (ret != 0) {
            printError(stmt, MACHCLI_HANDLE_STMT, "MachCLIFetch");
            return -1;
        }

        long long resultCount = -1;
        long dataLen = 0;
        ret = MachCLIGetData(stmt, 0, MACHCLI_C_TYPE_INT64, &resultCount, 8, &dataLen);
        if (ret != 0) {
            printError(stmt, MACHCLI_HANDLE_STMT, "MachCLIGetData");
            return -1;
        }

        ret = MachCLIFreeStmt(stmt);
        if (ret != 0) {
            printError(stmt, MACHCLI_HANDLE_STMT, "MachCLIFreeStmt");
            return -1;
        }
        if (i % 10000 == 0) {
            printf("resultCount: %lld, iter=%d\n", resultCount, i);
        }
    }

    ret = MachCLIDisconnect(conn);
    if (ret != 0) {
        printError(conn, MACHCLI_HANDLE_DBC, "MachCLIDisconnect");
        return -1;
    }

    ret = MachCLIFinalize(env);
    if (ret != 0) {
        printError(env, MACHCLI_HANDLE_ENV, "MachCLIFinalize");
        return -1;
    }
    return 0;
}

void printError(void *handle, int handleType, char *fn) {
    int errCode = 0;
    char *errMsg = malloc(1024);
    if (MachCLIError(handle, handleType, &errCode, errMsg, 1024) == 0) {
        free(errMsg);
        printf("%s failed, error code: %d, error message: %s\n", fn, errCode, errMsg);
    } else {
        free(errMsg);
        printf("%s failed\n", fn);
    }
}
