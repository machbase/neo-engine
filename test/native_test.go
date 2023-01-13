package mach_test

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	mach "github.com/machbase/neo-engine"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var db *mach.Database

func TestMain(m *testing.M) {
	var err error

	fmt.Println("-------------------------------")
	fmt.Println(mach.LinkInfo())

	// exePath, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	// homepath := filepath.Join(filepath.Dir(exePath), "machbase")

	homepath, err := filepath.Abs("./tmp/machbase")
	if err != nil {
		panic(errors.Wrap(err, "abs tmp dir"))
	}
	if err := mkDirIfNotExists("./tmp"); err != nil {
		panic(errors.Wrap(err, "create tmp dir"))
	}
	if err := mkDirIfNotExists(homepath); err != nil {
		panic(errors.Wrap(err, "machbase"))
	}
	if err := mkDirIfNotExists(filepath.Join(homepath, "conf")); err != nil {
		panic(errors.Wrap(err, "machbase conf"))
	}
	if err := mkDirIfNotExists(filepath.Join(homepath, "dbs")); err != nil {
		panic(errors.Wrap(err, "machbase dbs"))
	}
	if err := mkDirIfNotExists(filepath.Join(homepath, "trc")); err != nil {
		panic(errors.Wrap(err, "machbase trc"))
	}

	if len(machbase_conf) == 0 {
		panic("invalid machbase.conf")
	}

	confpath := filepath.Join(homepath, "conf", "machbase.conf")
	if err = os.WriteFile(confpath, machbase_conf, 0644); err != nil {
		panic(errors.Wrap(err, "machbase.conf"))
	}

	mach.Initialize(homepath)

	if mach.ExistsDatabase() {
		if err = mach.DestroyDatabase(); err != nil {
			panic(errors.Wrap(err, "destroy database"))
		}
	}
	if err = mach.CreateDatabase(); err != nil {
		panic(errors.Wrap(err, "create database"))
	}

	db = mach.New()
	if db == nil {
		panic(err)
	}
	err = db.Startup()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("alter system set trace_log_level=1023")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(db.SqlTidy(
		`create log table log(
			short short, ushort ushort, integer integer, uinteger uinteger, long long, ulong ulong, float float, double double, 
			ipv4 ipv4, ipv6 ipv6, varchar varchar(20), text text, json json, binary binary, blob blob, clob clob, 
			datetime datetime, datetime_now datetime
		)`))
	if err != nil {
		panic(err)
	}

	m.Run()

	db.Shutdown()
}

func TestColumns(t *testing.T) {
	rows, err := db.Query("select * from log")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	type ColumnsData struct {
		name string
		typ  string
	}

	data := []ColumnsData{
		{"SHORT", "int16"},
		{"USHORT", "int16"},
		{"INTEGER", "int32"},
		{"UINTEGER", "int32"},
		{"LONG", "int64"},
		{"ULONG", "int64"},
		{"FLOAT", "float32"},
		{"DOUBLE", "float64"},
		{"IPV4", "ipv4"},
		{"IPV6", "ipv6"},
		{"VARCHAR", "string"},
		{"TEXT", "string"},
		{"JSON", "string"},
		{"BINARY", "binary"},
		{"BLOB", "binary"},
		{"CLOB", "binary"},
		{"DATETIME", "datetime"},
		{"DATETIME_NOW", "datetime"},
	}
	for i, cd := range data {
		require.Equal(t, cd.name, cols[i].Name)
	}
}

func TestExec(t *testing.T) {
	var err error
	_, err = db.Exec("insert into log values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		0, 1, 2, 3, 4, 5, 6.6, 7.77,
		net.ParseIP("127.0.0.1"), net.ParseIP("AB:CC:CC:CC:CC:CC:CC:FF"),
		fmt.Sprintf("varchar_1_%s.", randomVarchar()),
		"text_1", "{\"json\":1}", []byte("binary_00"), "blob_01", "clob_01", 1, time.Now())
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("insert into log values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, 3, 4, 5, 6.6, 7.77,
		net.ParseIP("127.0.0.2"), net.ParseIP("AB:CC:CC:CC:CC:CC:CC:DD"),
		fmt.Sprintf("varchar_2_%s.", randomVarchar()),
		"text_2", "{\"json\":1}", []byte("binary_01"), "blob_01", "clob_01", 1, time.Now())
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("insert into log values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		2, 1, 2, 3, 4, 5, 6.6, 7.77,
		net.ParseIP("127.0.0.3"), net.ParseIP("AB:CC:CC:CC:CC:CC:CC:AA"),
		fmt.Sprintf("varchar_3_%s.", randomVarchar()),
		"text_3", "{\"json\":2}", []byte("binary_02"), "blob_01", "clob_01", 1, time.Now())
	if err != nil {
		panic(err)
	}
}

func TestAppend(t *testing.T) {

	t.Log("---- insert done")
	appender, err := db.Appender("log")
	if err != nil {
		panic(err)
	}
	defer appender.Close()

	for i := 3; i < 10; i++ {
		err = appender.Append(
			int16(i),         // short
			uint16(i*10),     // ushort
			int(i*100),       // int
			uint(i*1000),     // uint
			int64(i*10000),   // long
			uint64(i*100000), // ulong
			float32(i),       // float
			float64(i),       // double
			net.ParseIP(fmt.Sprintf("192.168.0.%d", i)),              // IPv4
			net.ParseIP(fmt.Sprintf("12:FF:FF:FF:CC:EE:FF:%02X", i)), // IPv6
			fmt.Sprintf("varchar_append-%d", i),
			fmt.Sprintf("text_append-%d-%s.", i, randomVarchar()),
			fmt.Sprintf("{\"json\":%d}", i),
			[]byte(fmt.Sprintf("binary_append_%02d", i)),
			"blob_append",
			"clob_append",
			i*10000000000,
			time.Now())
		if err != nil {
			panic(err)
		}
	}
	err = appender.Close()
	if err != nil {
		panic(err)
	}
	t.Log("---- append done")

	row := db.QueryRow("select count(*) from m$sys_tables  where name = ?", "LOG")
	if row.Err() != nil {
		t.Logf("ERR-query: %s\n", row.Err().Error())
	} else {
		var count int
		err = row.Scan(&count)
		if err != nil {
			t.Logf("ERR-scan: %s\n", err.Error())
		} else {
			t.Logf("============> table 'log' exists=%v\n", count)
		}
	}

	t.Log("---- before select")
	rows, err := db.Query(db.SqlTidy(`
		select
			short, ushort, integer, uinteger, long, ulong, float, double, 
			ipv4, ipv6,
			varchar, text, json, binary, blob, clob, datetime, datetime_now
		from
			log`))
	if err != nil {
		t.Logf("Error: %s\n", err.Error())
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var _int16 int16
		var _uint16 int16
		var _int32 int32
		var _uint32 int32
		var _int64 int64
		var _uint64 int64
		var _float float32
		var _double float64
		var _ipv4 net.IP
		var _ipv6 net.IP
		var _varchar string
		var _text string
		var _json string
		var _bin []byte
		var _blob []byte
		var _clob []byte
		var _datetime int64
		var _datetime_now time.Time

		err := rows.Scan(
			&_int16, &_uint16, &_int32, &_uint32, &_int64, &_uint64, &_float, &_double,
			&_ipv4, &_ipv6,
			&_varchar, &_text, &_json, &_bin, &_blob, &_clob, &_datetime, &_datetime_now)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			panic(err)
		}
		t.Logf("1st ----> %d %d %d %d %d %d %f %f %v %v %s %s %s %v %d %v\n",
			_int16, _uint16, _int32, _uint32, _int64, _uint64, _float, _double,
			_ipv4, _ipv6,
			_varchar, _text, _json, string(_bin),
			_datetime, _datetime_now)
	}
	rows.Close()

	rows, err = db.Query(db.SqlTidy(`
		select 
			short, ushort, integer, uinteger, long, ulong, float, double, varchar, text, json, 
			datetime, datetime_now 
		from 
			log where short = ? and varchar = ?`), 0, "varchar_1")
	if err != nil {
		t.Logf("error:%s\n", err.Error())
	}
	for rows.Next() {
		var _int16 int16
		var _uint16 int16
		var _int32 int32
		var _uint32 int32
		var _int64 int64
		var _uint64 int64
		var _float float32
		var _double float64
		var _varchar string
		var _text string
		var _json string
		var _datetime int64
		var _datetime_now int64

		err := rows.Scan(
			&_int16, &_uint16, &_int32, &_uint32, &_int64, &_uint64, &_float, &_double,
			&_varchar, &_text, &_json,
			&_datetime, &_datetime_now)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			break
		}
		t.Logf("2nd ----> %d %d %d %d %d %d %f %f %s %s %s %d %d\n",
			_int16, _uint16, _int32, _uint32, _int64, _uint64, _float, _double,
			_varchar, _text, _json,
			_datetime, _datetime_now)
	}
	rows.Close()

	// // signal handler
	// fmt.Printf("\npress ^C to quit.\n")
	// quitChan := make(chan os.Signal)
	// signal.Notify(quitChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// // wait signal
	// <-quitChan

	t.Log("-------------------------------")
}

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset)-1)]
	}
	return string(b)
}

func randomVarchar() string {
	rangeStart := 0
	rangeEnd := 10
	offset := rangeEnd - rangeStart
	randLength := seededRand.Intn(offset) + rangeStart

	charSet := "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ"
	randString := StringWithCharset(randLength, charSet)
	return randString
}

func mkDirIfNotExists(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

var machbase_conf = []byte(`
#################################################################################
# Copyright of this product 2013-2023
# MACHBASE Corporation (or Inc.) or its subsidiaries.
# All Rights reserved.
#################################################################################

#################################################################################
# Server
#################################################################################
PORT_NO = 5656

#################################################################################
# DB Path   ? => $MACHBASE_HOME
#################################################################################
DBS_PATH=?/dbs

#################################################################################
# Trace Log
#################################################################################
TRACE_LOGFILE_SIZE  = 10485760
TRACE_LOGFILE_COUNT = 1000
TRACE_LOGFILE_PATH  = ?/trc

#################################################################################
# MIN value: 0
# MAX value: 2
# Def Value: 0
#
# More detailed logs are written in the higher level.
# More levels you set, more details are written.
#################################################################################
TRACE_LOG_LEVEL = 277

#################################################################################
# Error Call Stack & Core Generation
# GEN_CALLSTACK_FOR_ABORT_ERROR specifies whether to record call stacks
# after an abnormal server shutdown.
#
# GEN_CORE_FILE determines whether to record core files
# after an abnormal server shutdown.
#################################################################################
GEN_CALLSTACK_FOR_ABORT_ERROR = 0
GEN_CORE_FILE = 1

#################################################################################
# Default DURATION (Second)
# If DURATION_GAP is set to 0, it means that you can search all data.
# If set a non-zero value as duration, it can search the latest data
# for the specified duration.
#################################################################################
DURATION_GAP = 0

#################################################################################
# Parallel Factor
#################################################################################
CPU_PARALLEL = 1

#################################################################################
# MIN value: 1   (SEC)
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 120 (SEC)
#
# The meta of a table is written to a disk at a specified time interval.
#################################################################################
DISK_COLUMNAR_TABLE_CHECKPOINT_INTERVAL_SEC = 120

#################################################################################
# MIN value: 1   (SEC)
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 120 (SEC)
#
# The meta of an index is written to a disk at a specified time interval.
#################################################################################
DISK_COLUMNAR_INDEX_CHECKPOINT_INTERVAL_SEC = 120

#################################################################################
# MIN value: 0
# MAX value: 1
# Def Value: 0
#
# If this is set to 1, machbased tries to write column data when column partitions are full
# in order to minimize I/O frequency. If the value is set to 0, machbased writes column
# partitions as soon as possible to minimize data loss caused by power failure or
# process failure.
#################################################################################
DISK_COLUMNAR_TABLE_COLUMN_PART_FLUSH_MODE = 0

#################################################################################
# MIN value: 0
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 3
#
# This indicates the I/O interval for writing data to the same column partition.
# If the gap between the last and current-write time is less than the specified value,
# it doesnt write. But if the partition is full, machbased writes it.
#################################################################################
DISK_COLUMNAR_TABLE_COLUMN_PART_IO_INTERVAL_MIN_SEC = 3

#################################################################################
# MIN value: 1
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 3
#
# The number of disk I/O thread which writes appended data to disks.
#################################################################################
DISK_IO_THREAD_COUNT = 3

#################################################################################
# MIN value: 1048576 (1MB)
# MAX value: unsigned long MAX (2^64 - 1)
#
# When machbased starts up, it reserves the amount of memory with this
# size in advance in order to reduce system call of the memory
# allocation from OS.
#################################################################################
DISK_COLUMNAR_TABLESPACE_MEMORY_MIN_SIZE = 104857600

#################################################################################
# MIN value: 268435456 (256MB)
# MAX value: unsigned long MAX (2^64 - 1)
# If it is not specified, 8589934592(8GB) will be used.
# cf. 1073741824(1GB), 2147483648(2GB), 4294967296(4GB)
#     3221225472(3GB), 6442450944(6GB)
# Log tables that you created cannot use more than the specified size.
# It is recommended to set this value as 50%-80% of the physical memory size on your system.
#################################################################################
DISK_COLUMNAR_TABLESPACE_MEMORY_MAX_SIZE = 268435456

#################################################################################
# MIN value: 1048576 (1MB)
# MAX value: unsigned long MAX (2^64 - 1)
#
# This value specifies the block-size of memory allocation for the column partition.
#################################################################################
DISK_COLUMNAR_TABLESPACE_MEMORY_EXT_SIZE = 2097152

#################################################################################
# MIN value: 0
# MAX value: 100
# Def Value: 80
#
# If the current memory consumption of the disk column tablespace exceeds
# the specified memory limit, which is DISK_COLUMNAR_TABLESPACE_MEMORY_MAX_SIZE *
# (DISK_COLUMNAR_TABLESPACE_MEMORY_SLOWDOWN_HIGH_LIMIT_PCT / 100), the appending
# operation will be suspended.
#################################################################################
DISK_COLUMNAR_TABLESPACE_MEMORY_SLOWDOWN_HIGH_LIMIT_PCT = 80

#################################################################################
# MIN value: 0
# MAX value: unsigned int max (2^32 - 1)
# Def Value: 1
#
# If the current memory consumption of the disk column tablespace exceeds
# a threshold, the appending operation waits for the specified time (in milliseconds).
#################################################################################
DISK_COLUMNAR_TABLESPACE_MEMORY_SLOWDOWN_MSEC = 1

#################################################################################
# MIN value: 1048576 (1MB)
# MAX value: unsigned int max (2^32 - 1)
# Def Value: 2097152 (2MB)
#
# The double write file for data consistency and recovery is created with this
# size when creating a database.
#################################################################################
DISK_COLUMNAR_TABLESPACE_DWFILE_INT_SIZE = 2097152

#################################################################################
# MIN value: 1048576 (1MB)
# MAX value: unsigned int max (2^32 - 1)
# Def Value: 1048576 (1MB)
#
# If there is not enough space in a double write file, dwfile (double write file)
# will be extended by this value.
#################################################################################
DISK_COLUMNAR_TABLESPACE_DWFILE_EXT_SIZE = 1048576

#################################################################################
# MIN value: 1073741824 (1GB)
# MAX value: unsigned long MAX (2^64 - 1)
# Def Value: 8589934592 (8GB)
#
# This property is the maximum size of the memory of machbased.
# If machbased's memory consumption exceeds this value, machbased tries to minimize
# memory usage by do the followings.
#  1. Suspend or abort appending rows to a table.
#  2. Delay building indexes until there is memory available.
#################################################################################
PROCESS_MAX_SIZE = 4294967296

#################################################################################
# MIN value: 0
# MAX value: 1
# Def Value: 0 (Not recorded)
#
# If this valus was set '1' and the appending program that used SQLAppend API
# had some errors, it will write error logs in $MACHBASE_HOME/trc/machbase.trc file.
# Caution: It makes the appending performance slow.
#          Refer to 'SQLAppendSetErrorCallback' API from CLI chapter.
#################################################################################
DUMP_APPEND_ERROR = 0

#################################################################################
# MIN value: 0 (Disabled direct I/O)
# MAX value: 1
# Def Value: 1
#
# If the value is 1, the direct I/O is used to store data and
# indexes to disks, otherwise buffered I/O is used.
#################################################################################
DISK_TABLESPACE_DIRECT_IO_WRITE = 1

#################################################################################
# MIN value: 0 (Disabled direct I/O)
# MAX value: 1
# Def Value: 0
#
# If the value is 1, direct I/O is used to read data and
# indexes from the disks, otherwise buffered I/O is used.
#################################################################################
DISK_TABLESPACE_DIRECT_IO_READ = 0

#################################################################################
# MIN value: 0
# MAX value: 3
# Def Value: 1
#
# 0 - OFF.   No synchronize at all
# 1 - NORMAL Synchronize on DW file write, and backup
# 2 - FULL   Synchronize on disk file close, adjusting end RID, including 1
# 3 - EXTRA  Synchronize on every write, including 2
#################################################################################
DISK_TABLESPACE_SYNCHRONOUS = 1

#################################################################################
# MIN value: 0
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 3
#
# This value specifies the number of threads that build indexes.
# If the value is 0, the index building is disabled.
#################################################################################
INDEX_BUILD_THREAD_COUNT = 3

#################################################################################
# MIN value: 1
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 3
#
# Specify maximum index partition in the memory for an index.
#################################################################################
INDEX_FLUSH_MAX_REQUEST_COUNT_PER_INDEX = 3

#################################################################################
# MIN value: 1
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 100000
#
# Each thread of building index appends rows to a table by the unit of this value.
# After that, it switches to the next index.
#################################################################################
INDEX_BUILD_MAX_ROW_COUNT_PER_THREAD = 100000

#################################################################################
# MIN value: 0 (False)
# MAX value: 1 (True)
# Def Value: 0 (False)
#
# When machbased shuts down and this value is set to '1', machbased waits for all indexes
# to be built for all key values in its table. If not, machbased does not wait.
#################################################################################
DISK_COLUMNAR_INDEX_SHUTDOWN_BUILD_FINISH = 0

#################################################################################
# MIN value: 0
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 0
#
# This value defines the first ID of CPU that machbased uses in the system.
#################################################################################
CPU_AFFINITY_BEGIN_ID = 0

#################################################################################
# MIN value: 0
# MAX value: unsigned int MAX (2^32 - 1)
# Def Value: 0
#
# This value defines the number of CPUs that machbased uses in the system.
# If the value is 0, machbased will use all CPUs in the system.
#################################################################################
CPU_AFFINITY_COUNT = 0

#################################################################################
# MIN value: 1
# MAX value: 1024
# Def Value: 3
#
# This value defines the number of threads building LSM partitions if
# the level of LSM index is greater than 0.
#################################################################################
INDEX_LEVEL_PARTITION_BUILD_THREAD_COUNT = 3

#################################################################################
# MIN value: 1
# MAX value: 1024
# Def Value: 1
#
# This value defines the number of threads deleting the partitions in the LSM index.
#################################################################################
INDEX_LEVEL_PARTITION_AGER_THREAD_COUNT = 1

#################################################################################
# MIN value: 0
# MAX value: unsigned long MAX (2^64 - 1)
# Def Value: 2147483648 (2GB)
#
# This value defines the buffer size for the page cache.
#################################################################################
DISK_COLUMNAR_PAGE_CACHE_MAX_SIZE = 2147483648

#################################################################################
# MIN value: 0
# MAX value: 100
# Def Value: 70
#
# Set the maximum portion of LSM Index build memory.
# The portion is expressed as a percentage of the total memory that machbased uses.
# When it exceeds the limit, LSM partition merge thread will be blocked.
#################################################################################
INDEX_LEVEL_PARTITION_BUILD_MEMORY_HIGH_LIMIT_PCT = 70


#################################################################################
# MIN value: 0
# MAX value: unsigned long MAX (2^64 - 1)
# Def Value: 2147483648 (2GB)
#
# If the memory size consumed by volatile tables exceeds the value of the property,
# no more insertion to volatile table is allowed.
#################################################################################
VOLATILE_TABLESPACE_MEMORY_MAX_SIZE = 536870912


#################################################################################
# MIN value: 0 (False)
# MAX value: 1 (True)
# Def Value: 1 (True)
#
# Result-Cache mode ON/OFF.
#################################################################################
RS_CACHE_ENABLE = 0

#################################################################################
# MIN value: 0
# MAX value: unsigned long MAX (2^64 - 1)
# Def Value: 1000 (msec)
#
# If the execution time of a query is slower than specified value, the results will
# be stored in the cache.
# If set to 0, all query results will be stored.
#################################################################################
RS_CACHE_TIME_BOUND_MSEC = 1000

#################################################################################
# MIN value: 32 * 1024
# MAX value: unsigned long MAX (2^64 - 1)
# Def Value: 536870912 (512MB)
#
#  The maximum size of memory that cache can use.
#################################################################################
RS_CACHE_MAX_MEMORY_SIZE = 33554432

#################################################################################
# MIN value: 1
# MAX value: unsigned long MAX (2^64 - 1)
# Def Value: 10000
#
# The maximum number of a query result that is stored in the cache.
#################################################################################
RS_CACHE_MAX_RECORD_PER_QUERY = 50000

#################################################################################
# MIN value: 1024
# MAX value: unsigned long MAX (2^64 - 1)
# Def Value: 16777216 (16MB)
#
# Memory size of a query result that is stored in the cache.
# It if exceeds the memory size, cache does not save the results.
#################################################################################
RS_CACHE_MAX_MEMORY_PER_QUERY = 4194304

#################################################################################
# MIN value: 0 (False)
# MAX value: 1 (True)
# Def Value: 0 (False)
#
# Result-Cache Approximate result mode ON/OFF.
# When cache is enabled, cached results can be used for getting approximate values.
# To get the exact values, set it to 0.
#################################################################################
RS_CACHE_APPROXIMATE_RESULT_ENABLE = 0

#################################################################################
# MIN value: 0 (False)
# MAX value: 1 (True)
# Def Value: 1 (True)
#
# Remote access allow/disallow flag.
# If it is 0, remote access will not be granted.
#################################################################################
GRANT_REMOTE_ACCESS = 1

#################################################################################
# MIN value: 0 (False)
# MAX value: 1 (True)
# Def Value: 1 (True)
#
# This value indicates that it is allowed to append a new value whose arrival time is
# less than the latest value's in a table. If it is set to 0, it is not allowed.
#################################################################################
DISK_COLUMNAR_TABLE_TIME_INVERSION_MODE = 1

#################################################################################
# It indicates the default value of LSM MAX_LEVEL for BITMAP/KEYWORD indexes
# min: 0, max: 3, default: 2
#################################################################################
DEFAULT_LSM_MAX_LEVEL = 0

#################################################################################
# MIN value: 1048576    (1024 * 1024)
# MAX value: unsigned long MAX (2^64 - 1)
# Def value: 104857600  (100 * 1024 * 1024)
#
# Limits total memory consumption during each query execution.
#################################################################################
MAX_QPX_MEM = 268435456

#################################################################################
# Min value: 0
# Max value: 2
# Def Value: 2
#
# [TAG TABLE]
# It determines whether it insert TAG META columns automatically
# when the corresponding meta columns are not found during INSERT or APPEND.
#
# 0: DO NOT insert tag name and metadata columns
# 1: Insert tag name (excluding other metadata columns)
# 2: Insert tag name and other metadata columns (default)
#################################################################################
TAGDATA_AUTO_META_INSERT = 2

#################################################################################
# MIN value: 0 (OFF)
# MAX value: unsigned long MAX (2^64 - 1)
# DEF value: 0 (OFF)
#
# Set timeout of communicating between server and client.
#################################################################################
SESSION_IDLE_TIMEOUT_SEC = 0

#################################################################################
# Rest-API port
#################################################################################
HTTP_PORT_NO = 5657

#################################################################################
# Maximum memory per web session.
# Default Value: 536870912 (512MB)
#################################################################################
HTTP_MAX_MEM = 536870912

#################################################################################
# Min Value:     0
# Max Value:     1
# Default Value: 0
#
# Enable REST-API service.
#################################################################################
HTTP_ENABLE = 0

#################################################################################
# Min Value:     0
# Max Value:     1
# Default Value: 0
#
# Enable Basic Authentication for Rest-API service
#################################################################################
HTTP_AUTH = 0

#################################################################################
# This property defines the behavior of a LOOKUP table
# when a duplicate key tries to update a row with the same pre-existing key
# while appending on the table.
#
# 0: Returns error.
# 1: Updates the corresponding row.
#################################################################################
LOOKUP_APPEND_UPDATE_ON_DUPKEY = 0

#################################################################################
# MIN value: 0 (unlimited)
# MAX value: unsigned int MAX (2^32 - 1)
# DEF value: 3000000
#
# Limit the number of rows to fetch in a rollup thread.
#################################################################################
ROLLUP_FETCH_COUNT_LIMIT = 10000

HANDLE_LIMIT = 1024

TAG_CACHE_MAX_MEMORY_SIZE = 33554432
DISK_TAG_INDEX_BLOCKS = 128
STREAM_THREAD_COUNT = 0
TAG_TABLE_META_MAX_SIZE = 1048576
DISK_BUFFER_COUNT = 1
TAG_CACHE_ENABLE = 3
`)
