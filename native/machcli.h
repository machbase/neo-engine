/******************************************************************************
* Copyright of this product 2013-2023,
 * MACHBASE Corporation(or Inc.) or its subsidiaries.
 * All Rights reserved.
 ******************************************************************************/

#ifndef _MACHCLI_H_
#define _MACHCLI_H_

#ifdef __cplusplus
extern "C" {
#endif

#if defined(_WIN64)
typedef long long sqllen_t;
#else
typedef long sqllen_t;
#endif

#define MACHCLI_HANDLE_ENV  (1)
#define MACHCLI_HANDLE_DBC  (2)
#define MACHCLI_HANDLE_STMT (3)

#define MACHCLI_C_TYPE_INT16     (101)
#define MACHCLI_C_TYPE_INT32     (102)
#define MACHCLI_C_TYPE_INT64     (103)
#define MACHCLI_C_TYPE_FLOAT     (104)
#define MACHCLI_C_TYPE_DOUBLE    (105)
#define MACHCLI_C_TYPE_CHAR      (106)

#define MACHCLI_SQL_TYPE_INT16    (0)     
#define MACHCLI_SQL_TYPE_INT32    (1)
#define MACHCLI_SQL_TYPE_INT64    (2)
#define MACHCLI_SQL_TYPE_DATETIME (3)
#define MACHCLI_SQL_TYPE_FLOAT    (4)
#define MACHCLI_SQL_TYPE_DOUBLE   (5)    
#define MACHCLI_SQL_TYPE_IPV4     (6)   
#define MACHCLI_SQL_TYPE_IPV6     (7)  
#define MACHCLI_SQL_TYPE_STRING   (8) 
#define MACHCLI_SQL_TYPE_BINARY   (9)

/* FIXED TYPE */
#define MACHCLI_APPEND_SHORT_NULL       (0x8000)
#define MACHCLI_APPEND_USHORT_NULL      (0xFFFF)
#define MACHCLI_APPEND_INTEGER_NULL     (0x80000000)
#define MACHCLI_APPEND_UINTEGER_NULL    (0xFFFFFFFF)
#define MACHCLI_APPEND_LONG_NULL        (0x8000000000000000LL)
#define MACHCLI_APPEND_ULONG_NULL       (0xFFFFFFFFFFFFFFFFULL)
#define MACHCLI_APPEND_FLOAT_NULL       (3.402823466e+38F)
#define MACHCLI_APPEND_DOUBLE_NULL      (1.7976931348623158e+308)

typedef struct MachCLIAppendVarStruct
{
    unsigned int mLength;  /* 0: null */
    void        *mData;
} MachCLIAppendVarStruct;

/* for IPv4, IPv6 as bin or string representation */
typedef struct MachCLIAppendIPStruct
{
    unsigned char   mLength; /* 0:null, 4:ipv4, 6:ipv6, 255:string representation */
    unsigned char   mAddr[16];
    char           *mAddrString;
} MachCLIAppendIPStruct;
/* Date time*/

typedef struct MachCLIAppendDateTimeStruct
{
    long long   mTime;       /* -1: null, -2: string, -3: TM, -4: now */
    struct tm   mTM;         /* affect only when mTime = -3 */
    char      * mDateStr;    /* affect only when mTime = -2 */
    char      * mFormatStr;  /* affect only when mTime = -2 */
} MachCLIAppendDateTimeStruct;

typedef union MachCLIAppendParam
{
    short                        mShort;
    unsigned short               mUShort;
    int                          mInteger;
    unsigned int                 mUInteger;
    long long                    mLong;
    unsigned long long           mULong;
    float                        mFloat;
    double                       mDouble;
    MachCLIAppendIPStruct        mIP;
    MachCLIAppendVarStruct       mVar;     /* for all varying type */
    MachCLIAppendVarStruct       mVarchar; /* alias */
    MachCLIAppendVarStruct       mText;    /* alias */
    MachCLIAppendVarStruct       mJson;    /* alias */
    MachCLIAppendVarStruct       mBinary;  /* binary */
    MachCLIAppendVarStruct       mBlob;    /* reserved alias */
    MachCLIAppendVarStruct       mClob;    /* reserved alias */
    MachCLIAppendDateTimeStruct  mDateTime;
} MachCLIAppendParam;

int MachCLIInitialize(void** aEnv);
int MachCLIFinalize(void* aEnv);

int MachCLIConnect(void * aEnv,
                   char * aConString,
                   void** aCon);
int MachCLIDisconnect(void* aCon);

int MachCLIError(void* aHandle,
                 int   aHandleType,
                 int * aErrorCode,
                 char* aErrorMsg,
                 int   aErrorMsgSize);

int MachCLIAllocStmt(void* aCon, void** aStmt);
int MachCLIFreeStmt(void* aStmt);

int MachCLIPrepare(void* aStmt, char* aSQL);
int MachCLIExecute(void* aStmt);
int MachCLIExecDirect(void* aStmt, char* aSQL);
int MachCLICancel(void* aStmt);

int MachCLIFetch(void* aStmt, int* aFetchEnd);

int MachCLIGetData(void* aStmt,
                   int   aColumnNo,
                   int   aCType,
                   void* aValuePtr,
                   int   aBufferSize,
                   long* aResultLen);

int MachCLIRowCount(void     * aStmt,
                    long long* aRowCount);

int MachCLIBindParam(void* aStmt,
                     int   aParamNo,
                     int   aCType,
                     int   aSQLType,
                     void* aValuePtr,
                     int   aValueLength);
int MachCLIDescribeParam(void* aStmt,
                         int   aParamNo,
                         int * aType,
                         int * aPrecision,
                         int * aScale,
                         int * aNullable);
int MachCLINumParam(void* aStmt,
                    int * aParamCount);

int MachCLIBindCol(void    * aStmt,
                   int       aColumnNo,
                   int       aCType,
                   void    * aValuePtr,
                   int       aBufferSize,
                   sqllen_t* aResultLen);
int MachCLIDescribeCol(void* aStmt,
                       int   aColumnNo,
                       char* aColumnName,
                       int   aBufferSize,
                       int * aColumnNameLength,
                       int * aDataType,
                       int * aColumnSize,
                       int * aScale,
                       int * aNullable);
int MachCLINumResultCol(void* aStmt,
                        int * aColumnCount);

typedef void (*MachCLIAppendErrorCallback)(void* aStmtHandle,
                                           int   aErrorCode,
                                           char* aErrorMessage,
                                           long  aErrorBufLen,
                                           char* aRowBuf,
                                           long  aRowBufLen);
int MachCLIAppendOpen(void* aStmt,
                      char* aTableName,
                      int   aErrorCheckCount);
int MachCLIAppendData(void              * aStmt,
                      MachCLIAppendParam* aData);
int MachCLIAppendDataByTime(void              * aStmt,
                            long long           aTime,
                            MachCLIAppendParam* aData);
int MachCLIAppendClose(void     * aStmt,
                       long long* aSuccessCount,
                       long long* aFailureCount);

int MachCLIAppendFlush(void* aStmt);
int MachCLIAppendSetErrorCallback(void                      * aStmt,
                                  MachCLIAppendErrorCallback  aFunc);
int MachCLISetConnectAppendFlush(void* aCon, int aOpt);
int MachCLISetStmtAppendInterval(void* aStmt, int aMSec);

/*
 * unknown: -1
 * DDL: 1-255
 * ALTER SYSTEM: 256-511
 * SELECT: 512
 * INSERT: 513
 * DELETE: 514-518
 * INSERT_SELECT: 519
 * UPDATE: 520
 * EXEC_ROLLUP: 1000-1002
 */
int MachCLIGetStmtType(void* aStmt, int* aStmtType);

#ifdef __cplusplus
}  /* extern "C" */
#endif
#endif
