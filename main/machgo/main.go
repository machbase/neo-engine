package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	mach "github.com/machbase/dbms-mach-go"
)

func main() {
	fmt.Println("-------------------------------")
	fmt.Println(mach.VersionString())

	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	homePath := filepath.Dir(exePath)
	mach.Initialize(homePath)

	mach.DestroyDatabase()
	mach.CreateDatabase()

	db := mach.NewDatabase()
	if db == nil {
		fmt.Printf("Error: %s\n", db.Error())
	}
	err = db.Startup(10 * time.Second)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	defer db.Shutdown()

	err = db.Exec("alter system set trace_log_level=1023")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	err = db.Exec("create log table log(id int, name varchar(20), pre double)")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	err = db.Exec("insert into log values(?, ?, ?)", 0, "zero", 1.01)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	err = db.Exec("insert into log values(1, 'one', 2.0002)")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	err = db.Exec("insert into log select id + 20, name, pre *4 from log")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

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
