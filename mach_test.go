package mach_test

import (
	_ "embed"
	"fmt"
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

	m.Run()

	// drop table simple_tag
	mach.EngConnectTrust(global.Env, "sys", &conn)
	mach.EngAllocStmt(conn, &stmt)
	mach.EngDirectExecute(stmt, `drop table simple_tag`)
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

func TestSimpleTagInsert(t *testing.T) {
	var conn unsafe.Pointer
	var stmt unsafe.Pointer

	// connect
	err := mach.EngConnectTrust(global.Env, "sys", &conn)
	require.NoError(t, err)
	defer mach.EngDisconnect(conn)

	// create tag table simple_tag
	err = mach.EngAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.EngDirectExecute(stmt, `create tag table if not exists simple_tag (name varchar(100) primary key, time datetime basetime, value double)`)
	require.NoError(t, err)
	mach.EngFreeStmt(stmt)

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

	// drop table
	err = mach.EngAllocStmt(conn, &stmt)
	require.NoError(t, err)
	err = mach.EngDirectExecute(stmt, `drop table simple_tag`)
	require.NoError(t, err)
}
