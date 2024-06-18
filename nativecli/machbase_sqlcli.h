/******************************************************************************
 * Copyright of this product 2013-2023,
 * MACHBASE Corporation(or Inc.) or its subsidiaries.
 * All Rights reserved.
 ******************************************************************************/

#ifndef _O_MACHBASE_SQL_CLI_H_
#define _O_MACHBASE_SQL_CLI_H_ 1

#define HAVE_LONG_LONG

#define SUPPORT_STRUCT_TM 1

////////////////////////////////////////////////////////////////////
// Check windows
// this comes from http://stackoverflow.com/questions/1505582/determining-32-vs-64-bit-in-c
#if _WIN32 || _WIN64
//#undef SUPPORT_STRUCT_TM

   #if _WIN64
     #define ENV64BIT
  #else
    #define ENV32BIT
  #endif
#endif

// Check GCC
#if __GNUC__
  #if __x86_64__ || __PPC64__ || __aarch64__
    #define ENV64BIT
  #else
    #define ENV32BIT
  #endif
#endif
////////////////////////////////////////////////////////////////////
#if defined(ENV64BIT)
    #if !defined(BUILD_REAL_64_BIT_MODE)
        #define BUILD_REAL_64_BIT_MODE
    #endif
#endif

#if defined(_MSC_VER)
#   include <windows.h>
#endif

#include <sqltypes.h>
#include <sql.h>
#include <sqlext.h>

#if defined(SUPPORT_STRUCT_TM)
# include <time.h>
#endif

#ifdef __cplusplus
extern "C" {
#endif

#if defined(MACHBASE_BUILD_CLI_DLL) && defined(NBP_CFG_OS_WINDOWS)
#  ifdef __cplusplus
#   define SQL_EXTERN NBP_EXPORT extern "C"
#  else
#   define SQL_EXTERN NBP_EXPORT
#  endif
#else
#  ifdef __cplusplus
#   define SQL_EXTERN extern "C"
#  else
#   define SQL_EXTERN
#  endif
#endif

/*
 *  For WINDOWS old compilers such as VS60 and so forth,
 *  which do not have any information about SQLLEN series types.
 */
#if defined(_MSC_VER)
# if (_MSC_VER <= 1200)
#   if !defined(SQLLEN)
#       define  SQLLEN           SQLINTEGER
#       define  SQLULEN          SQLUINTEGER
#       define  SQLSETPOSIROW    SQLUSMALLINT
typedef SQLULEN          SQLROWCOUNT;
typedef SQLULEN          SQLROWSETSIZE;
typedef SQLULEN          SQLTRANSID;
typedef SQLLEN           SQLROWOFFSET;
#   endif /* if !defined(SQLLEN) */
# endif /* if (_MSC_VER <= 1200) */
#endif /* if defined(_MSC_VER) */

#if defined(_MSC_VER)

#define MACHBASE_UINT64_LITERAL(aVal)   ((unsigned __int64)(aVal ## ui64))
#define MACHBASE_SINT64_LITERAL(aVal)   ((signed __int64)(aVal ## i64))

#elif defined(ENV64BIT)

#define MACHBASE_UINT64_LITERAL(aVal)   ((unsigned long)(aVal ## UL))
#define MACHBASE_SINT64_LITERAL(aVal)   ((signed long)(aVal ## L))

#else

#define MACHBASE_UINT64_LITERAL(aVal)   ((unsigned long long)(aVal ## ULL))
#define MACHBASE_SINT64_LITERAL(aVal)   ((signed long long)(aVal ## LL))

#endif


/* -----------------------------
 * Machbase specific attributes
 * ----------------------------- */
#define SQL_ATTR_PORT_NO            2005
#define SQL_ATTR_SESSION_ID         2006
#define SQL_ATTR_CONNECT_COUNT      2007       /** MACHBASECONN으로부터 mConnectCount를 가져오기 위한 attribute */
#define SQL_C_BIGINT                SQL_BIGINT
#define SQL_NULL_TYPE               0

#define SQL_ALREADY_OPENED          (-3)

#define SQL_CLOB                    (2004)
#define SQL_BLOB                    (2005)
#define SQL_TEXT                    (2100)
#define SQL_JSON                    (2101)
#define SQL_IPV4                    (2104)
#define SQL_IPV6                    (2106)

#define SQL_USMALLINT               (2201)
#define SQL_UINTEGER                (2202)
#define SQL_UBIGINT                 (2203)


/*------------------------------------------------------------------
 *  APPEND TYPE MACRO
 *------------------------------------------------------------------*/

/* FIXED TYPE */
#define SQL_APPEND_SHORT_NULL       0x8000
#define SQL_APPEND_USHORT_NULL      0xFFFF
#define SQL_APPEND_INTEGER_NULL     0x80000000
#define SQL_APPEND_UINTEGER_NULL    0xFFFFFFFF
#define SQL_APPEND_LONG_NULL        MACHBASE_SINT64_LITERAL(0x8000000000000000)
#define SQL_APPEND_ULONG_NULL       MACHBASE_UINT64_LITERAL(0xFFFFFFFFFFFFFFFF)
#define SQL_APPEND_FLOAT_NULL       3.402823466e+38F
#define SQL_APPEND_DOUBLE_NULL      1.7976931348623158e+308

/* IP TYPE */
#define SQL_APPEND_IP_NULL          0
#define SQL_APPEND_IP_IPV4          4
#define SQL_APPEND_IP_IPV6          6
#define SQL_APPEND_IP_STRING        255

/* DATETIME TYPE */
#define SQL_APPEND_DATETIME_NOW            MACHBASE_UINT64_LITERAL(0xFFFFFFFFFFFFFFFC)
#define SQL_APPEND_DATETIME_STRING         MACHBASE_UINT64_LITERAL(0xFFFFFFFFFFFFFFFE)
#define SQL_APPEND_DATETIME_NULL           MACHBASE_UINT64_LITERAL(0xFFFFFFFFFFFFFFFF)
#if defined(SUPPORT_STRUCT_TM)
#define SQL_APPEND_DATETIME_STRUCT_TM      MACHBASE_UINT64_LITERAL(0xFFFFFFFFFFFFFFFD)
#endif

/* VARYING TYPE : VARCHAR, TEXT, CLOB, BLOB */
#define SQL_APPEND_VARCHAR_NULL                   0
#define SQL_APPEND_TEXT_NULL                      0
#define SQL_APPEND_CLOB_NULL                      0
#define SQL_APPEND_BLOB_NULL                      0
#define SQL_APPEND_BINARY_NULL                    0
#define SQL_APPEND_JSON_NULL                      0

typedef struct machbaseAppendVarStruct
{
    unsigned int mLength;
    void        *mData;
} machbaseAppendVarStruct;

/* for IPv4, IPv6 as bin or string representation */
typedef struct machbaseAppendIPStruct
{
    unsigned char   mLength; /* 0:null, 4:ipv4, 6:ipv6, 255:string representation */
    unsigned char   mAddr[16];
    char           *mAddrString;
} machbaseAppendIPStruct;

/* Date time*/
typedef struct machbaseAppendDateTimeStruct
{
    long long       mTime;
#if defined(SUPPORT_STRUCT_TM)
    struct tm       mTM;
#endif
    char           *mDateStr;
    char           *mFormatStr;
} machbaseAppendDateTimeStruct;

typedef union machbaseAppendParam
{
    short                        mShort;
    unsigned short               mUShort;
    int                          mInteger;
    unsigned int                 mUInteger;
    long long                    mLong;
    unsigned long long           mULong;
    float                        mFloat;
    double                       mDouble;
    machbaseAppendIPStruct       mIP;
    machbaseAppendVarStruct      mVar;     /* for all varying type */
    machbaseAppendVarStruct      mVarchar; /* alias */
    machbaseAppendVarStruct      mText;    /* alias */
    machbaseAppendVarStruct      mJson;    /* alias */
    machbaseAppendVarStruct      mBinary;  /* binary */
    machbaseAppendVarStruct      mBlob;    /* reserved alias */
    machbaseAppendVarStruct      mClob;    /* reserved alias */
    machbaseAppendDateTimeStruct mDateTime;
} machbaseAppendParam;

#define SQL_APPEND_PARAM machbaseAppendParam

/* machbase type definitiion */
typedef enum
{
    SQL_APPEND_TYPE_NULL   = 0,
    SQL_APPEND_TYPE_INT16  = 1,
    SQL_APPEND_TYPE_INT32  = 2,
    SQL_APPEND_TYPE_INT64  = 3,
    SQL_APPEND_TYPE_UINT16 = 4,
    SQL_APPEND_TYPE_UINT32 = 5,
    SQL_APPEND_TYPE_UINT64 = 6,
    SQL_APPEND_TYPE_FLT32  = 7,
    SQL_APPEND_TYPE_FLT64  = 8,
    SQL_APPEND_TYPE_IPV4   = 9,
    SQL_APPEND_TYPE_IPV6   = 10,
    SQL_APPEND_TYPE_DATE   = 11,
    SQL_APPEND_TYPE_VARCHAR = 12,
    SQL_APPEND_TYPE_CLOB    = 13,
    SQL_APPEND_TYPE_BLOB    = 14,
    SQL_APPEND_TYPE_JSON    = 15,
} SQL_APPEND_TYPES;

typedef void (*SQLAppendErrorCallback)(SQLHSTMT    aStmtHandle,
                                       SQLINTEGER  aErrorCode,
                                       SQLPOINTER  aErrorMessage,
                                       SQLLEN      aErrorBufLen,
                                       SQLPOINTER  aRowBuf,
                                       SQLLEN      aRowBufLen);

SQLRETURN SQL_API SQLAppendOpen(SQLHSTMT   aStmtHandle,
                                SQLCHAR   *aTableName,
                                SQLINTEGER aErrorCheckCount );

SQLRETURN SQL_API SQLAppendData(SQLHSTMT aStmtHandle, void *aData[] );
SQLRETURN SQL_API SQLAppendDataByTime(SQLHSTMT      aStmtHandle,
                                      SQLBIGINT     aTime,
                                      void         *aData[] );

/* New API for Machbase 2.0, support TEXT */
SQLRETURN SQL_API SQLAppendDataV2(SQLHSTMT          aStmtHandle,
                                  SQL_APPEND_PARAM*aData );

SQLRETURN SQL_API SQLAppendDataByTimeV2(SQLHSTMT          aStmtHandle,
                                        SQLBIGINT         aTime,
                                        SQL_APPEND_PARAM  *aData );

SQLRETURN SQL_API SQLAppendBatchByTime(SQLHSTMT           aStmtHandle,
                                       SQLCHAR           *aTableName,
                                       SQLBIGINT          aTime,
                                       SQLINTEGER         aRowCount,
                                       SQLINTEGER         aColCount,
                                       SQL_APPEND_TYPES  *aTypes,
                                       SQL_APPEND_PARAM  *aData );

SQLRETURN SQL_API SQLAppendBatch(SQLHSTMT           aStmtHandle,
                                 SQLCHAR           *aTableName,
                                 SQLINTEGER         aRowCount,
                                 SQLINTEGER         aColCount,
                                 SQL_APPEND_TYPES  *aTypes,
                                 SQL_APPEND_PARAM  *aData );

SQLRETURN SQL_API SQLAppendGetDateTimeFromDateString(SQLHSTMT      aStmtHandle,
                                                     SQLCHAR      *aDateString,
                                                     SQLCHAR      *aFormatString,
                                                     SQLBIGINT    *aHrTime);

SQLRETURN SQL_API SQLSetConnectAppendFlush(SQLHDBC hdbc, SQLINTEGER option);

SQLRETURN SQL_API SQLSetStmtAppendInterval(SQLHSTMT aStmtHandle, SQLINTEGER aMSec);

#if defined(SUPPORT_STRUCT_TM)
SQLRETURN SQL_API SQLAppendGetDateTimeFromTM(struct tm    *aTM,
                                             SQLBIGINT    *aHrTime);
#endif
SQLRETURN SQL_API SQLAppendFlush(SQLHSTMT aStmtHandle);
SQLRETURN SQL_API SQLAppendClose(SQLHSTMT    aStmtHandle,
                                 SQLBIGINT  *aSuccessCount,
                                 SQLBIGINT  *aFailureCount );
SQLRETURN SQL_API SQLAppendSetErrorCallback(SQLHSTMT               aStmtHandle,
                                            SQLAppendErrorCallback aFunc);

SQLRETURN SQL_API SQLLoadErrorRowCount (SQLHSTMT  aStmt, SQLLEN   *aRowCount);

/* Direct Execute Message Callback */
typedef void (*SQLMsgCallback)(SQLHSTMT aStmtHandle, SQLPOINTER aMessage, SQLLEN aMessageLen);

SQLRETURN SQL_API SQLSetMsgCallback (SQLHDBC aStmt, SQLMsgCallback aFunc);

/* for utf-16 */
SQLRETURN SQL_API SQLAppendOpenW(SQLHSTMT    aStmtHandle,
                                 SQLWCHAR    *TableName,
                                 SQLINTEGER  aErrorCheckCount );

SQLRETURN SQL_API SQLAppendGetDateTimeFromDateStringW(SQLHSTMT      aStmtHandle,
                                                      SQLWCHAR      *aDateString,
                                                      SQLWCHAR      *aFormatString,
                                                      SQLBIGINT     *aHrTime);

#ifdef __cplusplus
}  /* extern "C" */
#endif

#endif
