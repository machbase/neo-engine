#include <stdlib.h>
#include "machEngine.h"

int MachInitialize(char* aHomePath) {return -1;}
void MachFinalize(){}

int MachCreateDB(){ return -1; }
int MachDestroyDB() { return -1; }
int MachIsDBCreated() { return 0; }

int MachStartupDB(int aTimeoutSecond, void** aDBHandle){ return -1; }

int MachShutdownDB(void* aDBHandle) { return -1; }

int MachErrorCode(void* aHandle){ return -1; }
char* MachErrorMsg(void* aHandle){ return NULL; }


int MachAllocStmt(void* aDBHandle, void** aMachStmt){ return -1; }
int MachFreeStmt(void* aMachStmt){ return -1; }

int MachPrepare(void* aMachStmt, char* aSQL){ return -1; }

int MachExecute(void* aMachStmt){ return -1; }
int MachExecuteClean(void* aMachStmt){ return -1; }

int MachDirectExecute(void* aMachStmt, char* aSQL){ return -1; }

int MachEffectRows(void* aMachStmt, unsigned long long* aEffectRows){ return -1; }

int MachFetch(void* aMachStmt, int* aFetchEnd){ return -1; }

int MachColumnCount(void* aMachStmt, int* aColumnCount){ return -1; }

int MachColumnType(void* aMachStmt, int aColumnIndex, int* aType, int* aSize){ return -1; }
int MachColumnLength(void* aMachStmt, int aColumnIndex, int* aColumnLength){ return -1; }

int MachColumnData(void* aMachStmt, int aColumnIndex, void* aDest, int aBufSize){ return -1; }
int MachColumnDataInt16(void* aMachStmt, int aColumnIndex, short* aDest){ return -1; }
int MachColumnDataInt32(void* aMachStmt, int aColumnIndex, int* aDest){ return -1; }
int MachColumnDataInt64(void* aMachStmt, int aColumnIndex, long long* aDest){ return -1; }
int MachColumnDataDateTime(void* aMachStmt, int aColumnIndex, long long* aDest){ return -1; }
int MachColumnDataFloat(void* aMachStmt, int aColumnIndex, float* aDest){ return -1; }
int MachColumnDataDouble(void* aMachStmt, int aColumnIndex, double* aDest){ return -1; }
int MachColumnDataIPV4(void* aMachStmt, int aColumnIndex, void* aDest){ return -1; }
int MachColumnDataIPV6(void* aMachStmt, int aColumnIndex, void* aDest){ return -1; }
int MachColumnDataString(void* aMachStmt, int aColumnIndex, char* aDest, int aBufSize){ return -1; }
int MachColumnDataBinary(void* aMachStmt, int aColumnIndex, void* aDest, int aBufSize){ return -1; }

int MachBindInt32(void* aMachStmt, int aParamNo, int aValue){ return -1; }
int MachBindInt64(void* aMachStmt, int aParamNo, long long aValue){ return -1; }
int MachBindDouble(void* aMachStmt, int aParamNo, double aValue){ return -1; }
int MachBindString(void* aMachStmt, int aParamNo, char* aValue, int aLength){ return -1; }
int MachBindBinary(void* aMachStmt, int aParamNo, void* aValue, int aLength){ return -1; }

int MachAppendOpen(void* aMachStmt, char* aTableName){ return -1; }
int MachAppendClose(void* aMachStmt,
                    unsigned long long* aAppendSuccessCount,
                    unsigned long long* aAppendFailureCount){ return -1; }

int MachAppendData(void* aMachStmt, MachEngineAppendParam* aAppendParamArr){ return -1; }
