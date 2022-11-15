package main

import (
	"fmt"
	"mach"
	"time"
)

func main() {
	fmt.Println("-------------------------------")

	mach.Initialize("/home/eirny/Developer/sample-machdb/tmp/home")

	mach.DestroyDatabase()
	mach.CreateDatabase()

	mach.Startup(10 * time.Second)

	mach.Execute("alter system set trace_log_level=1023")
	mach.Execute("create log table log(id int, name varchar(20), pre double)")

	mach.ExecuteNewSession("insert into log values(0, 'zero', 1.01)")
	mach.ExecuteNewSession("insert into log values(1, 'one', 2.0002)")
	mach.ExecuteNewSession("insert into log select id + 20, name, pre *4 from log")

	fmt.Println("IsRunning", mach.IsRunning())

	if rows, err := mach.Query("select * from log"); err == nil {
		for rows.Next() {
			var id int
			var name string
			var pre float64

			err := rows.Scan(&id, &name, &pre)
			if err != nil {
				fmt.Printf("error: %s]\n", err.Error())
				break
			}
			fmt.Println("---->", id, name, pre)
		}
	} else {
		fmt.Printf("Error: %s\n", err.Error())
	}

	time.Sleep(5)

	mach.Shutdown()
	fmt.Println("-------------------------------")
}
