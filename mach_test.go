package mach_test

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	mach "github.com/machbase/neo-engine"
	"github.com/stretchr/testify/require"
)

var machPort = 5656

//go:embed mach_test.conf
var machbase_conf []byte

var global = struct {
	Env unsafe.Pointer
}{}

func TestMain(m *testing.M) {
	homePath, err := filepath.Abs(filepath.Join(".", "tmp", "machbase"))
	if err != nil {
		panic(err)
	}
	confPath := filepath.Join(homePath, "conf", "machbase.conf")

	os.RemoveAll(homePath)
	os.MkdirAll(homePath, 0755)
	os.MkdirAll(filepath.Join(homePath, "conf"), 0755)
	os.MkdirAll(filepath.Join(homePath, "trc"), 0755)
	os.MkdirAll(filepath.Join(homePath, "dbs"), 0755)
	os.WriteFile(confPath, machbase_conf, 0644)

	var envHandle unsafe.Pointer
	err = mach.EngInitialize(homePath, machPort, 0x1, &envHandle)
	if err != nil {
		panic(err)
	}
	global.Env = envHandle

	if !mach.EngExistsDatabase(global.Env) {
		mach.EngCreateDatabase(global.Env)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)

	err = mach.EngStartup(global.Env)
	if err != nil {
		panic(err)
	}

	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	// create tag table simple_tag
	mach.EngConnectTrust(global.Env, "sys", &conn)
	mach.EngAllocStmt(conn, &stmt)
	mach.EngDirectExecute(stmt, `create tag table if not exists simple_tag (name varchar(100) primary key, time datetime basetime, value double)`)
	mach.EngFreeStmt(stmt)
	mach.EngDisconnect(conn)

	mach.EngConnectTrust(global.Env, "sys", &conn)
	mach.EngAllocStmt(conn, &stmt)
	mach.EngDirectExecute(stmt, `
		create tag table tag_data(
			name            varchar(100) primary key, 
			time            datetime basetime, 
			value           double summarized,
			short_value     short,
			ushort_value    ushort,
			int_value       integer,
			uint_value 	    uinteger,
			long_value      long,
			ulong_value 	ulong,
			str_value       varchar(400),
			json_value      json,
			ipv4_value      ipv4,
			ipv6_value      ipv6
		)
	`)
	mach.EngFreeStmt(stmt)
	mach.EngDisconnect(conn)

	m.Run()

	// drop table simple_tag
	mach.EngConnectTrust(global.Env, "sys", &conn)
	mach.EngAllocStmt(conn, &stmt)
	mach.EngDirectExecute(stmt, `drop table simple_tag`)
	mach.EngFreeStmt(stmt)
	mach.EngDisconnect(conn)

	// drop table tag_data
	mach.EngConnectTrust(global.Env, "sys", &conn)
	mach.EngAllocStmt(conn, &stmt)
	mach.EngDirectExecute(stmt, `drop table tag_data`)
	mach.EngFreeStmt(stmt)
	mach.EngDisconnect(conn)

	mach.EngShutdown(global.Env)
	mach.EngFinalize(global.Env)
	os.RemoveAll(homePath)
}

func BenchmarkSimpleTagInsertDirectExecute(b *testing.B) {
	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	err := mach.EngConnectTrust(global.Env, "sys", &conn)
	require.NoError(b, err)
	defer mach.EngDisconnect(conn)

	for i := 0; i < b.N; i++ {
		sqlText := fmt.Sprintf(`insert into simple_tag values('bench-insert', now, %f)`, 1.001*float64(i+1))
		err = mach.EngAllocStmt(conn, &stmt)
		require.NoError(b, err)
		err = mach.EngDirectExecute(stmt, sqlText)
		require.NoError(b, err)
		mach.EngFreeStmt(stmt)
	}
}

func BenchmarkSimpleTagInsertExecute(b *testing.B) {
	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	err := mach.EngConnectTrust(global.Env, "sys", &conn)
	require.NoError(b, err)
	defer mach.EngDisconnect(conn)

	sqlText := `insert into simple_tag values(?, ?, ?)`

	for i := 0; i < b.N; i++ {
		err = mach.EngAllocStmt(conn, &stmt)
		require.NoError(b, err)

		err = mach.EngPrepare(stmt, sqlText)
		require.NoError(b, err)
		err = mach.EngBindString(stmt, 0, "bench-insert")
		require.NoError(b, err)
		err = mach.EngBindInt64(stmt, 1, time.Now().UnixNano())
		require.NoError(b, err)
		err = mach.EngBindFloat64(stmt, 2, 1.001*float64(i+1))
		require.NoError(b, err)
		err = mach.EngExecute(stmt)
		require.NoError(b, err)

		mach.EngFreeStmt(stmt)
	}
}

func BenchmarkSimpleTagAppend(b *testing.B) {
	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	err := mach.EngConnectTrust(global.Env, "sys", &conn)
	require.NoError(b, err)
	defer mach.EngDisconnect(conn)

	err = mach.EngAllocStmt(conn, &stmt)
	require.NoError(b, err)

	err = mach.EngAppendOpen(stmt, "simple_tag")
	require.NoError(b, err)

	columnCount, err := mach.EngColumnCount(stmt)
	require.NoError(b, err)
	require.Equal(b, 3, columnCount)

	columnNames := make([]string, columnCount)
	columnTypes := make([]int, columnCount)
	for i := 0; i < columnCount; i++ {
		columnNames[i], err = mach.EngColumnName(stmt, i)
		require.NoError(b, err)
		columnTypes[i], _, err = mach.EngColumnType(stmt, i)
		require.NoError(b, err)
	}
	require.Equal(b, []string{"NAME", "TIME", "VALUE"}, columnNames)
	require.Equal(b, []int{
		int(mach.MACHCLI_SQL_TYPE_STRING),
		int(mach.MACHCLI_SQL_TYPE_DATETIME),
		int(mach.MACHCLI_SQL_TYPE_DOUBLE)}, columnTypes)

	buf := mach.EngMakeAppendBuffer(stmt, columnNames, []string{"string", "datetime", "double"})
	for i := 0; i < b.N; i++ {
		err := buf.Append("bench-append", time.Now().UnixNano(), 1.001*float64(i+1))
		require.NoError(b, err)
	}

	s, f, err := mach.EngAppendClose(stmt)
	require.NoError(b, err)
	require.Equal(b, int64(b.N), s)
	require.Equal(b, int64(0), f)
	mach.EngFreeStmt(stmt)
}

func TestSimpleTagSvrInsert(t *testing.T) {
	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	// connect
	err := mach.EngConnectTrust(global.Env, "sys", &conn)
	require.NoError(t, err)
	defer mach.EngDisconnect(conn)

	expectCount := 100_000

	// insert direct_execute
	for i := 0; i < expectCount; i++ {
		sqlText := fmt.Sprintf(`insert into simple_tag values('insert', now, %f)`, 1.001*float64(i+1))
		err = mach.EngAllocStmt(conn, &stmt)
		require.NoError(t, err)
		err = mach.EngDirectExecute(stmt, sqlText)
		require.NoError(t, err)
		mach.EngFreeStmt(stmt)
	}

	time.Sleep(500 * time.Millisecond)

	// select count(*) form simple_tag
	err = mach.EngAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.EngDirectExecute(stmt, `select count(*) from simple_tag where name = 'insert'`)
	require.NoError(t, err)

	// fetch
	next, err := mach.EngFetch(stmt)
	require.NoError(t, err)
	require.True(t, next)

	// get column
	count, valid, err := mach.EngColumnDataInt64(stmt, 0)
	require.NoError(t, err)
	require.True(t, valid)
	require.Equal(t, int64(expectCount), count)

	mach.EngFreeStmt(stmt)

	//
	// Issue: https://github.com/machbase/neo/issues/889
	//
	// // JOIN tag stat and meta
	// err = mach.EngAllocStmt(conn, &stmt)
	// require.NoError(t, err)

	// // SELECT m._ID, m.NAME, s.ROW_COUNT FROM _SIMPLE_TAG_META m, V$SIMPLE_TAG_STAT s WHERE s.NAME = m.NAME
	// err = mach.EngPrepare(stmt, `SELECT m._ID, m.NAME, s.ROW_COUNT FROM _SIMPLE_TAG_META m, V$SIMPLE_TAG_STAT s WHERE s.NAME = m.NAME`)
	// require.NoError(t, err)
	// err = mach.EngExecute(stmt)
	// require.NoError(t, err)

	// // fetch
	// next, err = mach.EngFetch(stmt)
	// require.NoError(t, err)
	// require.True(t, next)
	// mach.EngFreeStmt(stmt)
}

func TestTagTableSVRInsertAndSelect(t *testing.T) {
	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	// connect
	err := mach.EngConnectTrust(global.Env, "sys", &conn)
	require.NoError(t, err)
	defer mach.EngDisconnect(conn)

	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2021-01-01 00:00:00", time.UTC)

	// Because INSERT statement uses '2021-01-01 00:00:00' as time value which was parsed in Local timezone,
	// the time value should be converted to UTC timezone to compare
	// TODO: improve this behavior
	nowStrInLocal := now.In(time.Local).Format("2006-01-02 15:04:05")

	// insert
	err = mach.EngAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.EngPrepare(stmt,
		`insert into tag_data values('insert-once', '`+nowStrInLocal+`', 1.23, `+ // name, time, value
			`?, ?, ?, ?,`+ // short_value, ushort_value, int_value, uint_value
			`?, ?, ?, ?,`+ // long_value, ulong_value, str_value, json_value
			`?, ?)`, // ipv4_value, ipv6_value
	)
	require.NoError(t, err, "insert prepare fail")

	err = mach.EngBindInt32(stmt, 0, 1) // short_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindInt32(stmt, 1, 2) // ushort_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindInt32(stmt, 2, 3) // int_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindInt32(stmt, 3, 4) // uint_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindInt64(stmt, 4, 5) // long_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindInt64(stmt, 5, 6) // ulong_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindString(stmt, 6, "str1") // str_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindString(stmt, 7, `{"key1": "value1"}`) // json_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindString(stmt, 8, net.IPv4(192, 168, 0, 1).String()) // ipv4_value
	require.NoError(t, err, "bind fail")
	err = mach.EngBindString(stmt, 9, net.IPv6loopback.String()) // ipv6_value
	require.NoError(t, err, "bind fail")

	err = mach.EngExecute(stmt)
	require.NoError(t, err, "execute fail")
	err = mach.EngFreeStmt(stmt)
	require.NoError(t, err, "close fail")

	// flush
	err = mach.EngAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.EngDirectExecute(stmt, `EXEC table_flush(tag_data)`)
	require.NoError(t, err, "table_flush fail")
	err = mach.EngFreeStmt(stmt)
	require.NoError(t, err, "close fail")

	// select
	err = mach.EngAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.EngPrepare(stmt, `select * from tag_data where name = 'insert-once'`)
	require.NoError(t, err, "select fail")
	err = mach.EngExecute(stmt)
	require.NoError(t, err, "execute fail")
	stmtType, err := mach.EngStmtType(stmt)
	require.NoError(t, err, "stmt type fail")
	require.Equal(t, 512, stmtType)

	next, err := mach.EngFetch(stmt)
	require.NoError(t, err, "fetch fail")
	require.True(t, next, "fetch fail")

	// name
	if v, isValid, err := mach.EngColumnDataString(stmt, 0); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, "insert-once", v)
	}

	// time
	if v, isValid, err := mach.EngColumnDataDateTime(stmt, 1); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, now.In(time.UTC), v.In(time.UTC))
	}

	// value
	if v, isValid, err := mach.EngColumnDataFloat64(stmt, 2); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, 1.23, v)
	}

	// short_value
	if v, isValid, err := mach.EngColumnDataInt16(stmt, 3); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, int16(1), v)
	}

	// ushort_value
	if v, isValid, err := mach.EngColumnDataInt16(stmt, 4); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, uint16(2), uint16(v))
	}

	// int_value
	if v, isValid, err := mach.EngColumnDataInt32(stmt, 5); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, int32(3), v)
	}

	// uint_value
	if v, isValid, err := mach.EngColumnDataInt32(stmt, 6); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, uint32(4), uint32(v))
	}

	// long_value
	if v, isValid, err := mach.EngColumnDataInt64(stmt, 7); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, int64(5), v)
	}

	// ulong_value
	if v, isValid, err := mach.EngColumnDataInt64(stmt, 8); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, uint64(6), uint64(v))
	}

	// str_value
	if v, isValid, err := mach.EngColumnDataString(stmt, 9); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, "str1", v)
	}

	// json_value
	if v, isValid, err := mach.EngColumnDataString(stmt, 10); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, `{"key1": "value1"}`, v)
	}

	// ipv4_value
	if v, isValid, err := mach.EngColumnDataIPv4(stmt, 11); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, net.IPv4(192, 168, 0, 1).To4(), v)
	}

	// ipv6_value
	if v, isValid, err := mach.EngColumnDataIPv6(stmt, 12); err != nil || !isValid {
		require.NoError(t, err, "column data fail")
	} else {
		require.True(t, isValid, "column data fail")
		require.Equal(t, net.IPv6loopback, v)
	}
	err = mach.EngFreeStmt(stmt)
	require.NoError(t, err, "close fail")
}

func cliErrorStmt(stmt unsafe.Pointer) string {
	var errCode int
	var errMsg string
	mach.CliError(stmt, mach.MACHCLI_HANDLE_STMT, &errCode, &errMsg)
	if errMsg != "" {
		return fmt.Sprintf("error code: %d, error message: %s", errCode, errMsg)
	} else {
		return ""
	}
}

func TestSimpleTagCliInsert(t *testing.T) {
	// TODO: fix this test
	t.Skip("Skip CLI test")
	env := new(unsafe.Pointer)
	if err := mach.CliInitialize(env); err != nil {
		t.Fatal(err)
	}
	defer mach.CliFinalize(*env)

	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	// connect
	err := mach.CliConnect(*env, fmt.Sprintf("SERVER=127.0.0.1;UID=SYS;PWD=MANAGER;CONNTYPE=1;PORT_NO=%d", machPort), &conn)
	require.NoError(t, err)
	defer mach.CliDisconnect(conn)

	expectCount := 100_000

	// insert direct_execute
	for i := 0; i < expectCount; i++ {
		sqlText := fmt.Sprintf(`insert into simple_tag values('insert-cli', now, %.6f)`, 1.001*float64(i+1))
		err = mach.CliAllocStmt(conn, &stmt)
		require.NoError(t, err)
		err = mach.CliExecDirect(stmt, sqlText)
		if err != nil {
			t.Logf("i=%d error %s, %s", i, err.Error(), cliErrorStmt(stmt))
			t.Fail()
		}
		require.NoError(t, err)
		mach.CliFreeStmt(stmt)
	}

	time.Sleep(500 * time.Millisecond)

	// select count(*) form simple_tag
	err = mach.CliAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.CliExecDirect(stmt, `select count(*) from simple_tag where name = 'insert-cli'`)
	require.NoError(t, err)

	// fetch
	endOfResult, err := mach.CliFetch(stmt)
	require.NoError(t, err)
	require.False(t, endOfResult)

	// get column
	resultCount := int64(0)
	_, err = mach.CliGetData(stmt, 0, mach.MACHCLI_C_TYPE_INT64, unsafe.Pointer(&resultCount), 8)
	require.NoError(t, err)
	require.Equal(t, int64(expectCount), resultCount)

	mach.CliFreeStmt(stmt)

	// JOIN tag stat and meta
	err = mach.CliAllocStmt(conn, &stmt)
	require.NoError(t, err)

	// SELECT m._ID, m.NAME, s.ROW_COUNT FROM _SIMPLE_TAG_META m, V$SIMPLE_TAG_STAT s WHERE s.NAME = m.NAME
	err = mach.CliPrepare(stmt, `SELECT m._ID, m.NAME, s.ROW_COUNT FROM _SIMPLE_TAG_META m, V$SIMPLE_TAG_STAT s WHERE s.NAME = m.NAME`)
	require.NoError(t, err)
	err = mach.CliExecute(stmt)
	require.NoError(t, err)

	// fetch
	endOfResult, err = mach.EngFetch(stmt)
	require.NoError(t, err)
	require.False(t, endOfResult)

	mach.CliFreeStmt(stmt)
}

func TestTagTableCLIInsertAndSelect(t *testing.T) {
	env := new(unsafe.Pointer)
	if err := mach.CliInitialize(env); err != nil {
		t.Fatal(err)
	}
	defer mach.CliFinalize(*env)

	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	// connect
	err := mach.CliConnect(*env, fmt.Sprintf("SERVER=127.0.0.1;UID=SYS;PWD=MANAGER;CONNTYPE=1;PORT_NO=%d", machPort), &conn)
	require.NoError(t, err)
	defer mach.CliDisconnect(conn)

	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2021-01-01 00:00:00", time.UTC)

	// Because INSERT statement uses '2021-01-01 00:00:00' as time value which was parsed in Local timezone,
	// the time value should be converted to UTC timezone to compare
	// TODO: improve this behavior
	nowStrInLocal := now.In(time.Local).Format("2006-01-02 15:04:05")

	// insert
	err = mach.CliAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.CliPrepare(stmt,
		`insert into tag_data values('insert-cli', '`+nowStrInLocal+`', 1.23, `+ // name, time, value
			`?, ?, ?, ?,`+ // short_value, ushort_value, int_value, uint_value
			`?, ?, ?, ?,`+ // long_value, ulong_value, str_value, json_value
			`?, ?)`, // ipv4_value, ipv6_value
	)
	require.NoError(t, err, "insert prepare fail")

	numParam, err := mach.CliNumParam(stmt)
	require.NoError(t, err, "num param fail")
	require.Equal(t, 10, numParam)

	for i := 0; i < 10; i++ {
		desc, err := mach.CliDescribeParam(stmt, i)
		require.NoError(t, err, "describe param fail")
		require.NotNil(t, desc)
	}

	// bind
	shortValue := int16(1) // short_value
	err = mach.CliBindParam(stmt, 0, mach.MACHCLI_C_TYPE_INT16, mach.MACHCLI_SQL_TYPE_INT16, unsafe.Pointer(&shortValue), 2)
	require.NoError(t, err, "bind fail")
	ushortValue := uint16(2) // ushort_value
	err = mach.CliBindParam(stmt, 1, mach.MACHCLI_C_TYPE_INT16, mach.MACHCLI_SQL_TYPE_INT16, unsafe.Pointer(&ushortValue), 2)
	require.NoError(t, err, "bind fail")
	intValue := int(3) // int_value
	err = mach.CliBindParam(stmt, 2, mach.MACHCLI_C_TYPE_INT32, mach.MACHCLI_SQL_TYPE_INT32, unsafe.Pointer(&intValue), 4)
	require.NoError(t, err, "bind fail")
	uintValue := uint(4) // uint_value
	err = mach.CliBindParam(stmt, 3, mach.MACHCLI_C_TYPE_INT32, mach.MACHCLI_SQL_TYPE_INT32, unsafe.Pointer(&uintValue), 4)
	require.NoError(t, err, "bind fail")
	longValue := int64(5) // long_value
	err = mach.CliBindParam(stmt, 4, mach.MACHCLI_C_TYPE_INT64, mach.MACHCLI_SQL_TYPE_INT64, unsafe.Pointer(&longValue), 8)
	require.NoError(t, err, "bind fail")
	ulongValue := uint64(6) // ulong_value
	err = mach.CliBindParam(stmt, 5, mach.MACHCLI_C_TYPE_INT64, mach.MACHCLI_SQL_TYPE_INT64, unsafe.Pointer(&ulongValue), 8)
	require.NoError(t, err, "bind fail")
	strValue := "str1" // str_value
	err = mach.CliBindParam(stmt, 6, mach.MACHCLI_C_TYPE_CHAR, mach.MACHCLI_SQL_TYPE_STRING, unsafe.Pointer(&[]byte(strValue)[0]), len(strValue))
	require.NoError(t, err, "bind fail")
	jsonValue := `{"key1": "value1"}` // json_value
	err = mach.CliBindParam(stmt, 7, mach.MACHCLI_C_TYPE_CHAR, mach.MACHCLI_SQL_TYPE_STRING, unsafe.Pointer(&[]byte(jsonValue)[0]), len(jsonValue))
	require.NoError(t, err, "bind fail")
	ipv4Value := net.IPv4(192, 168, 0, 1).To4().String() // ipv4_value
	err = mach.CliBindParam(stmt, 8, mach.MACHCLI_C_TYPE_CHAR, mach.MACHCLI_SQL_TYPE_IPV4, unsafe.Pointer(&[]byte(ipv4Value)[0]), len(ipv4Value))
	require.NoError(t, err, "bind fail")
	ipv6Value := net.IPv6loopback.String() // ipv6_value
	err = mach.CliBindParam(stmt, 9, mach.MACHCLI_C_TYPE_CHAR, mach.MACHCLI_SQL_TYPE_IPV6, unsafe.Pointer(&[]byte(ipv6Value)[0]), len(ipv6Value))
	require.NoError(t, err, "bind fail")

	// execute
	err = mach.CliExecute(stmt)
	require.NoError(t, err, "execute fail; %s", cliErrorStmt(stmt))
	err = mach.CliFreeStmt(stmt)
	require.NoError(t, err, "close fail; %s", cliErrorStmt(stmt))

	// flush
	err = mach.CliAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.CliExecDirect(stmt, `EXEC table_flush(tag_data)`)
	require.NoError(t, err, "table_flush fail")
	err = mach.CliFreeStmt(stmt)
	require.NoError(t, err, "close fail")

	// select
	err = mach.CliAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.CliPrepare(stmt, `select * from tag_data where name = 'insert-cli'`)
	require.NoError(t, err, "select fail")
	err = mach.CliExecute(stmt)
	require.NoError(t, err, "execute fail")

	// stmt type (e.g. INSERT, SELECT) is not supported in CLI

	// fetch
	endOfResult, err := mach.CliFetch(stmt)
	require.NoError(t, err, "fetch fail")
	require.False(t, endOfResult, "fetch fail")

	// name
	nameData := make([]byte, 100)
	if len, err := mach.CliGetData(stmt, 0, mach.MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&nameData[0]), len(nameData)); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, "insert-cli", string(nameData[:len]))
	}

	// time
	timeData := int64(0)
	if len, err := mach.CliGetData(stmt, 1, mach.MACHCLI_C_TYPE_INT64, unsafe.Pointer(&timeData), 8); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, now.UnixNano(), timeData)
	}

	// value
	valueData := float64(0)
	if len, err := mach.CliGetData(stmt, 2, mach.MACHCLI_C_TYPE_DOUBLE, unsafe.Pointer(&valueData), 8); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, 1.23, valueData)
	}

	// short_value
	shortData := int16(0)
	if len, err := mach.CliGetData(stmt, 3, mach.MACHCLI_C_TYPE_INT16, unsafe.Pointer(&shortData), 2); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, int16(1), shortData)
	}

	// ushort_value
	ushortData := uint16(0)
	if len, err := mach.CliGetData(stmt, 4, mach.MACHCLI_C_TYPE_INT16, unsafe.Pointer(&ushortData), 2); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, uint16(2), ushortData)
	}

	// int_value
	intData := int32(0)
	if len, err := mach.CliGetData(stmt, 5, mach.MACHCLI_C_TYPE_INT32, unsafe.Pointer(&intData), 4); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, int32(3), intData)
	}

	// uint_value
	uintData := uint32(0)
	if len, err := mach.CliGetData(stmt, 6, mach.MACHCLI_C_TYPE_INT32, unsafe.Pointer(&uintData), 4); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, uint32(4), uintData)
	}

	// long_value
	longData := int64(0)
	if len, err := mach.CliGetData(stmt, 7, mach.MACHCLI_C_TYPE_INT64, unsafe.Pointer(&longData), 8); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, int64(5), longData)
	}

	// ulong_value
	ulongData := uint64(0)
	if len, err := mach.CliGetData(stmt, 8, mach.MACHCLI_C_TYPE_INT64, unsafe.Pointer(&ulongData), 8); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, uint64(6), ulongData)
	}

	// str_value
	strData := make([]byte, 400)
	if len, err := mach.CliGetData(stmt, 9, mach.MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&strData[0]), len(strData)); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, "str1", string(strData[:len]))
	}

	// json_value
	jsonData := make([]byte, 400)
	if len, err := mach.CliGetData(stmt, 10, mach.MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&jsonData[0]), len(jsonData)); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, `{"key1": "value1"}`, string(jsonData[:len]))
	}

	// ipv4_value
	ipv4Data := make([]byte, 100)
	if len, err := mach.CliGetData(stmt, 11, mach.MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&ipv4Data[0]), len(ipv4Data)); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, net.IPv4(192, 168, 0, 1).To4().String(), string(ipv4Data[:len]))
	}

	// ipv6_value
	ipv6Data := make([]byte, 100)
	if len, err := mach.CliGetData(stmt, 12, mach.MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&ipv6Data[0]), len(ipv6Data)); err != nil || len < 0 {
		require.NoError(t, err, "column data fail")
	} else {
		require.Equal(t, net.IPv6loopback.String(), string(ipv6Data[:len]))
	}
	err = mach.CliFreeStmt(stmt)
	require.NoError(t, err, "close fail")
}
