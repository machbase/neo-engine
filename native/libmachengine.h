/******************************************************************************
* Copyright of this product 2013-2023,
 * MACHBASE Corporation(or Inc.) or its subsidiaries.
 * All Rights reserved.
 ******************************************************************************/

#ifndef _LIB_MACH_ENGINE_H_
#define _LIB_MACH_ENGINE_H_

/**
 * function result code type
 */
typedef enum nbe_rc_t
{
    NBE_RC_SUCCESS = 0, /**< success */
    NBE_RC_FAILURE = -1 /**< failure */
} nbe_rc_t;

typedef char               nbp_char_t;
typedef signed char        nbp_sint8_t;
typedef unsigned char      nbp_uint8_t;
typedef signed short       nbp_sint16_t;
typedef unsigned short     nbp_uint16_t;
typedef signed int         nbp_sint32_t;
typedef unsigned int       nbp_uint32_t;
typedef nbp_uint8_t        nbp_bool_t;
typedef signed long long   nbp_sint64_t;
typedef unsigned long long nbp_uint64_t;
typedef float              nbp_float_t;
typedef double             nbp_double_t;

#define NBP_FALSE          ((nbp_bool_t)0)
#define NBP_TRUE           ((nbp_bool_t)1)


extern void MachLog(const char* aStr);

/**
 * @brief 서버 상태가 예상 상태가 일치하는지 확인
 * @param [in] aExpectedState NBP_TRUE : 서버 동작, NBP_FALSE : 서버 비동작
 * @return nbp_bool_t 일치 여부
 */
extern nbp_bool_t MachCheckEqualServerStatus(nbp_bool_t aExpectedStatus);

/**
 * @brief MachEngineConfig 초기화..
 * @param [in] aHomePath 설정할 Machbase Home 경로
 */
extern nbe_rc_t MachInitialize(nbp_char_t* aHomePath);
extern nbe_rc_t MachFinalize();

/**
 * @brief Machbase Database 생성 및 삭제
 */
extern nbe_rc_t MachCreateDB();
extern nbe_rc_t MachDestroyDB();

/**
 * @brief Machbase Thread 시작
 * @details Machbase Thread가 Startup 완료될 때 까지 기다린다
 * @param [in] aTimeoutSecond timeout 시간 (단위 :초)
 */
extern nbe_rc_t MachStartupDB(nbp_uint32_t aTimeoutSecond);

/**
 * @brief Machbase Thread 종료
 * @details cm protocol send를 통해 종료
 */
extern nbe_rc_t MachShutdownDB();


/*************************SQL Manage*********************************/


/**
 * @brief MachStmt를 할당 및 해제
 * @param [inout] aMachStmt 할당 및 해제할 MachStmt 주소
 */
extern nbe_rc_t MachAllocStmt(void** aMachStmt);
extern nbe_rc_t MachFreeStmt(void* aMachStmt);

/**
 * @brief 쿼리 Prepare 및 Prepare Clean
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 * @param [in] aQuery 실행 쿼리
 */
extern nbe_rc_t MachPrepare(void* aMachStmt, nbp_char_t* aQuery);
extern nbe_rc_t MachPrepareClean(void* aMachStmt);

/**
 * @brief 쿼리 Execute 및 Execute Clean
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 */
extern nbe_rc_t MachExecute(void* aMachStmt);
extern nbe_rc_t MachExecuteClean(void* aMachStmt);

/**
 * @brief MachStmt 할당받지 않고, 쿼리 즉시 실행
 * @param [in] aQuery 실행 쿼리
 */
extern nbe_rc_t MachDirectSQLExecute(nbp_char_t* aQuery);

/**
 * @brief Select 쿼리 결과 Fetch (가져오기)
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 * @param [out] aFetchEnd fetch할 데이터가 있는지 여부
 */
extern nbe_rc_t MachFetch(void* aMachStmt, nbp_bool_t* aFetchEnd);

/**
 * @brief Fetch row로 부터 각 컬럼의 결과를 가지고 온다.
 * @param [in] aMachStmt MachAllocStmt로 할당받은 stmt 
 * @param [in] aColumnIndex 가져올 column의 인덱스
 * @param [out] aDest column 데이터를 저장할 변수의 주소 (column 타입과 동일한 타입의 변수의 주소를 보내줘야한다)
 */
extern nbe_rc_t MachGetColumnData(void* aMachStmt, nbp_uint32_t aColumnIndex, void* aDest);

/**
 * @brief MachDirectSQLExecute 함수인데 별도의 Session 생성
 * @param [in] aQuery 실행 쿼리
 */
extern nbe_rc_t MachDirectSQLOnNewSession(nbp_char_t* aQuery);


#endif
