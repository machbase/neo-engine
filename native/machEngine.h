/******************************************************************************
* Copyright of this product 2013-2023,
 * MACHBASE Corporation(or Inc.) or its subsidiaries.
 * All Rights reserved.
 ******************************************************************************/

#ifndef _MMI_ENGINE_H_
#define _MMI_ENGINE_H_

#define MACH_DATA_TYPE_INT16    0
#define MACH_DATA_TYPE_INT32    1
#define MACH_DATA_TYPE_INT64    2
#define MACH_DATA_TYPE_DATETIME 3
#define MACH_DATA_TYPE_FLOAT    4
#define MACH_DATA_TYPE_DOUBLE   5
#define MACH_DATA_TYPE_IPV4     6
#define MACH_DATA_TYPE_IPV6     7
#define MACH_DATA_TYPE_STRING   8
#define MACH_DATA_TYPE_BINARY   9
#define MACH_DATA_TYPE_MAX      10

typedef struct MachEngineAppendVarStruct
{
    unsigned int    mLength;
    void*           mData;
} MachEngineAppendVarStruct;

typedef struct MachEngineAppendIPStruct
{
/* 어떤 타입으로 IP 값을 입력했는지를 가지고 있다. */
    unsigned char   mLength;
/* mLength 값이 MACH_ENGINE_APPEND_IP_IPV4 또는 MACH_ENGINE_APPEND_IP_IPV6 일 때 사용 */
    unsigned char   mAddr[16];
/* mLength 값이 MACH_ENGINE_APPEND_IP_STRING 일 때 사용 */
    char*           mAddrString;
} MachEngineAppendIPStruct;
#define MACH_ENGINE_APPEND_IP_NULL 0        /* IP 데이터에 Null 입력 */
#define MACH_ENGINE_APPEND_IP_IPV4 4        /* IPV4 형식 값을 가진다 */
#define MACH_ENGINE_APPEND_IP_IPV6 6        /* IPV6 형식 값을 가진다 */
#define MACH_ENGINE_APPEND_IP_STRING 255    /* IP 값을 String으로 입력 */

typedef struct MachEngineAppendDateTimeStruct
{
/* TimeStamp 또는 어떤 형식으로 입력할지를 가지고 있다. */
    long long   mTime;
/* mTime 값이 MACH_ENGINE_APPEND_DATETIME_STRING 일 때 사용 */
    char*       mDateStr;   /* Time String 값 */
    char*       mFormatStr; /* Time String 형식 */
} MachEngineAppendDateTimeStruct;
#define MACH_ENGINE_APPEND_DATETIME_DEFAULT 0
#define MACH_ENGINE_APPEND_DATETIME_NOW -1       /* 입력 당시 서버 시간으로 설정 */
#define MACH_ENGINE_APPEND_DATETIME_STRING -2    /* Time 값을 String으로 입력 */

typedef union MachEngineAppendParamData
{
    short                           mShort;
    unsigned short                  mUShort;
    int                             mInteger;
    unsigned int                    mUInteger;
    long long                       mLong;
    unsigned long long              mULong;
    float                           mFloat;
    double                          mDouble;
    MachEngineAppendIPStruct        mIP;
    MachEngineAppendVarStruct       mVar;
    MachEngineAppendVarStruct       mVarchar;
    MachEngineAppendVarStruct       mText;
    MachEngineAppendVarStruct       mJson;
    MachEngineAppendVarStruct       mBinary;
    MachEngineAppendVarStruct       mBlob;
    MachEngineAppendVarStruct       mClob;
    MachEngineAppendDateTimeStruct  mDateTime;
} MachEngineAppendParamData;

typedef struct MachEngineAppendParam
{
    int                         mIsNull; // 1 : NULL, 0: NOT NULL
    MachEngineAppendParamData   mData;
} MachEngineAppendParam;

/**
 * @brief MachEngineConfig 초기화..
 * @param [in] aHomePath 설정할 Machbase Home 경로
 */
int MachInitialize(char* aHomePath);
void MachFinalize();

/**
 * @brief Machbase Database 생성 및 삭제
 */
int MachCreateDB();
int MachDestroyDB();

/**
 * @brief Machbase Thread 시작
 * @details Machbase Thread가 Startup 완료될 때 까지 기다린다
 * @param [in] aTimeoutSecond timeout 시간 (단위 :초)
 */
int MachStartupDB(int aTimeoutSecond, void** aDBHandle);

/**
 * @brief Machbase Thread 종료
 * @details cm protocol send를 통해 종료
 */
int MachShutdownDB(void* aDBHandle);

/**
 * @brief Handle 또는 Stmt로 부터 설정된 에러 코드 및 메시지를 반환한다.
 * @param [in] aStmt (db handle / stmt)
 * @return error code / error msg
 */
int MachErrorCode(void* aHandle);
char* MachErrorMsg(void* aHandle);


/*************************SQL Function*********************************/

/**
 * @brief MachStmt를 할당 및 해제
 * @param [inout] aMachStmt 할당 및 해제할 MachStmt 주소
 */
int MachAllocStmt(void* aDBHandle, void** aMachStmt);
int MachFreeStmt(void* aMachStmt);

/**
 * @brief 쿼리 Prepare 및 Prepare Clean
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 * @param [in] aQuery 실행 쿼리
 */
int MachPrepare(void* aMachStmt, char* aSQL);

/**
 * @brief 쿼리 Execute 및 Execute Clean
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 */
int MachExecute(void* aMachStmt);
int MachExecuteClean(void* aMachStmt);

/**
 * @brief 쿼리 즉시 실행
 * @param [in] aQuery 실행 쿼리
 */
int MachDirectExecute(void* aMachStmt, char* aSQL);

/**
 * @brief 실행된 쿼리의 결과 개수를 가져온다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt
 * @param [out] aEffectRows 결과 개수를 저장할 변수의 주소
 */
int MachEffectRows(void* aMachStmt, unsigned long long* aEffectRows);

/**
 * @brief Select 쿼리 결과 Fetch (가져오기)
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 * @param [out] aFetchEnd fetch할 데이터가 있는지 여부 1: 없음, 0: 있음
 */
int MachFetch(void* aMachStmt, int* aFetchEnd);

/**
 * Ferch 결과의 컬럼 개수를 가져온다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt
 * @param [out] aColumnCount 결과 컬럼 개수
 */
int MachColumnCount(void* aMachStmt, int* aColumnCount);

/**
 * @brief Fetch row로 부터 각 컬럼의 결과를 가지고 온다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 * @param [in] aColumnIndex 가져올 column의 인덱스
 * @param [out] aColumnType 컬럼의 데이터 타입
 * @param [out] aColumnLength 컬럼의 데이터 크기
 */
int MachColumnType(void* aMachStmt, int aColumnIndex, int* aColumnType);
int MachColumnLength(void* aMachStmt, int aColumnIndex, int* aColumnLength);

/**
 * @brief Fetch row로 부터 각 컬럼의 결과를 가지고 온다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 * @param [in] aColumnIndex 가져올 column의 인덱스
 * @param [out] aDest column 데이터를 저장할 변수의 주소 (column 타입과 동일한 타입의 변수의 주소를 보내줘야한다)
 */
int MachColumnData(void* aMachStmt, int aColumnIndex, void* aDest, int aBufSize);
/**
 * @brief MachColumnData 함수를 컬럼 타입에 맞게 호출하는 함수이다.
 */
int MachColumnDataInt16(void* aMachStmt, int aColumnIndex, short* aDest);
int MachColumnDataInt32(void* aMachStmt, int aColumnIndex, int* aDest);
int MachColumnDataInt64(void* aMachStmt, int aColumnIndex, long long* aDest);
int MachColumnDataDateTime(void* aMachStmt, int aColumnIndex, long long* aDest);
int MachColumnDataFloat(void* aMachStmt, int aColumnIndex, float* aDest);
int MachColumnDataDouble(void* aMachStmt, int aColumnIndex, double* aDest);
int MachColumnDataIPV4(void* aMachStmt, int aColumnIndex, void* aDest);
int MachColumnDataIPV6(void* aMachStmt, int aColumnIndex, void* aDest);
int MachColumnDataString(void* aMachStmt, int aColumnIndex, char* aDest, int aBufSize);
int MachColumnDataBinary(void* aMachStmt, int aColumnIndex, void* aDest, int aBufSize);

/**
 * @brief 쿼리 실행시에 바인드 변수에 바인드 값을 설정한다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt
 * @param [in] aParamNo 바인드 변수 인덱스
 * @param [in] aValue 바인드 값
 */
int MachBindInt32(void* aMachStmt, int aParamNo, int aValue);
int MachBindInt64(void* aMachStmt, int aParamNo, long long aValue);
int MachBindDouble(void* aMachStmt, int aParamNo, double aValue);
int MachBindString(void* aMachStmt, int aParamNo, char* aValue, int aLength);
int MachBindBinary(void* aMachStmt, int aParamNo, void* aValue, int aLength);

/*************************Append Function*********************************/

/**
 * @brief 데이터 Append Stmt를 관리하는 함수이다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt
 * @param [in] aTableName Append 대상 테이블 이름
 * @param [out] aAppendSuccessCount Append 입력 성공 횟수
 * @param [out] aAppendFailureCount Append 입력 실패 횟수
 */
int MachAppendOpen(void* aMachStmt, char* aTableName);
int MachAppendClose(void* aMachStmt,
                    unsigned long long* aAppendSuccessCount,
                    unsigned long long* aAppendFailureCount);

/**
 * @brief 테이블에 데이터를 Append 한다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt
 * @param [in] aAppendParamArr Append 할 데이터 값
 */
int MachAppendData(void* aMachStmt, MachEngineAppendParam* aAppendParamArr);

#endif

