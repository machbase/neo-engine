package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	mach "github.com/machbase/dbms-mach-go"
)

func main() {
	defer func() {
		obj := recover()
		if obj != nil {
			fmt.Printf("panic %#v", obj)
		}
	}()

	fmt.Println("-------------------------------")
	fmt.Println(mach.LinkInfo(), mach.VersionString())

	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	homePath := filepath.Dir(exePath)
	mach.Initialize(homePath)

	mach.DestroyDatabase()
	mach.CreateDatabase()

	db := mach.New()
	if db == nil {
		panic(err)
	}
	err = db.Startup(10 * time.Second)
	if err != nil {
		panic(err)
	}
	defer db.Shutdown()

	err = db.Exec("alter system set trace_log_level=1023")
	if err != nil {
		panic(err)
	}
	err = db.Exec(`create log table log(
		short short, ushort ushort, integer integer, uinteger uinteger, long long, ulong ulong, float float, double double, 
		ipv4 ipv4, ipv6 ipv6, varchar varchar(20), text text, json json, binary binary, blob blob, clob clob, 
		datetime datetime, datetime_now datetime)`)
	if err != nil {
		panic(err)
	}

	err = db.Exec("insert into log values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		0, 1, 2, 3, 4, 5, 6.6, 7.77,
		"127.0.0.1", "AB:CC:CC:CC:CC:CC:CC:FF", "varchar_1", "text_1", "{\"json\":1}", "binary_01", "blob_01", "clob_01", 1, -1)
	if err != nil {
		panic(err)
	}

	err = db.Exec("insert into log values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		0, 1, 2, 3, 4, 5, 6.6, 7.77, "127.0.0.2", "AB:CC:CC:CC:CC:CC:CC:DD", "varchar_2", "text_2", "{\"json\":1}", "binary_01", "blob_01", "clob_01", 1, -1)
	if err != nil {
		panic(err)
	}

	err = db.Exec("insert into log values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		0, 1, 2, 3, 4, 5, 6.6, 7.77, "127.0.0.3", "AB:CC:CC:CC:CC:CC:CC:AA", "varchar_3", "text_3", "{\"json\":1}", "binary_01", "blob_01", "clob_01", 1, -1)
	if err != nil {
		panic(err)
	}

	// err = db.Exec("insert into log values(?, ?, ?)", 1, "one", 2.0002)
	// if err != nil {
	// 	panic(err)
	// }
	// err = db.Exec("insert into log select id + 20, name, pre *4 from log")
	// if err != nil {
	// 	panic(err)
	// }
	fmt.Println("---- insert done")

	appender, err := db.Appender("log")
	if err != nil {
		panic(err)
	}
	defer appender.Close()

	for i := 0; i < 10; i++ {
		//err = appender.Append(3+i, "three", float64(3.0003)+float64(i))
		err = appender.Append(i, i*10, i*100, i*1000, i*10000, i*100000, float64(i), float64(i),
			net.IPv4(192, 168, 0, byte(i)).To4(), net.ParseIP("FF:FF:FF:FF:FF:FF:FF:FF"), "varchar_append", "text_append", "{\"json\":999}", "binary_append", "blob_append", "clob_append",
			i*10000000000, time.Now().UnixNano())
		if err != nil {
			panic(err)
		}
	}
	err = appender.Close()
	if err != nil {
		panic(err)
	}

	//rows, err := db.Query("select id, name, pre from log")
	rows, err := db.Query("select short, ushort, integer, uinteger, long, ulong, float, double, varchar, text, json, datetime, datetime_now from log")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
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
		var _varchar string
		var _text string
		var _json string
		var _datetime int64
		var _datetime_now int64

		// var id int
		// var name string
		// var pre float64

		//err := rows.Scan(&id, &name, &pre)
		err := rows.Scan(&_int16, &_uint16, &_int32, &_uint32, &_int64, &_uint64, &_float, &_double, &_varchar, &_text, &_json, &_datetime, &_datetime_now)
		if err != nil {
			fmt.Printf("error: %s]\n", err.Error())
			break
		}
		fmt.Printf("1st ----> %d %d %d %d %d %d %f %f %s %s %s %d %d\n",
			_int16, _uint16, _int32, _uint32, _int64, _uint64, _float, _double,
			_varchar, _text, _json, _datetime, _datetime_now)
		//fmt.Printf("1st ----> %d %s %v\n", id, name, pre)
	}
	rows.Close()

	rows, err = db.Query("select short, ushort, integer, uinteger, long, ulong, float, double, varchar, text, json, datetime, datetime_now from log where short = ? and varchar = ?", 0, "varchar_1")
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

		err := rows.Scan(&_int16, &_uint16, &_int32, &_uint32, &_int64, &_uint64, &_float, &_double, &_varchar, &_text, &_json, &_datetime, &_datetime_now)
		if err != nil {
			fmt.Printf("error: %s]\n", err.Error())
			break
		}
		fmt.Printf("2st ----> %d %d %d %d %d %d %f %f %s %s %s %d %d\n",
			_int16, _uint16, _int32, _uint32, _int64, _uint64, _float, _double,
			_varchar, _text, _json, _datetime, _datetime_now)
	}
	rows.Close()

	// rows, err = db.Query("select id, name, pre from log where id = ? and name = ?", 21, "one")
	// for rows.Next() {
	// 	var id int
	// 	var name string
	// 	var pre float64

	// 	err := rows.Scan(&id, &name, &pre)
	// 	if err != nil {
	// 		fmt.Printf("error: %s]\n", err.Error())
	// 		break
	// 	}
	// 	fmt.Printf("2nd ----> %d %s %.5f\n", id, name, pre)
	// }
	// rows.Close()

	fmt.Println("-------------------------------")
}
