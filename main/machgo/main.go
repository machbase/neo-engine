package main

import (
	"fmt"
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
	fmt.Println(mach.VersionString())

	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	homePath := filepath.Dir(exePath)
	mach.Initialize(homePath)

	mach.DestroyDatabase()
	mach.CreateDatabase()

	db := mach.NewDatabase()
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
	err = db.Exec("create log table log(id int, name varchar(20), pre double)")
	if err != nil {
		panic(err)
	}

	err = db.Exec("insert into log values(?, ?, ?)", 0, "zero", 1.01)
	if err != nil {
		panic(err)
	}
	err = db.Exec("insert into log values(?, ?, ?)", 1, "one", 2.0002)
	if err != nil {
		panic(err)
	}
	err = db.Exec("insert into log select id + 20, name, pre *4 from log")
	if err != nil {
		panic(err)
	}
	fmt.Println("---- insert done")

	appender, err := db.Appender("log")
	if err != nil {
		panic(err)
	}
	defer appender.Close()

	for i := 0; i < 10; i++ {
		err = appender.Append(3+i, "three", float64(3.0003)+float64(i))
		if err != nil {
			panic(err)
		}
	}
	err = appender.Close()
	if err != nil {
		panic(err)
	}

	fmt.Println("---- append done")

	rows, err := db.Query("select id, name, pre from log")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var pre float64

		err := rows.Scan(&id, &name, &pre)
		if err != nil {
			fmt.Printf("error: %s]\n", err.Error())
			break
		}
		fmt.Printf("1st ----> %d %s %v\n", id, name, pre)
	}
	rows.Close()

	rows, err = db.Query("select id, name, pre from log where id = ?", 21)
	for rows.Next() {
		var id int
		var name string
		var pre float64

		err := rows.Scan(&id, &name, &pre)
		if err != nil {
			fmt.Printf("error: %s]\n", err.Error())
			break
		}
		fmt.Printf("2nd ----> %d %s %.5f\n", id, name, pre)
	}
	rows.Close()

	fmt.Println("-------------------------------")
}
