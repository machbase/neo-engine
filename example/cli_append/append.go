package main

import (
	"context"
	"fmt"
	"time"

	mach "github.com/machbase/neo-engine"
)

const (
	machPort = 5656
	machHost = "127.0.0.1"

	tableName = "example"
	tagName   = "helloworld"
)

func main() {
	env, err := mach.NewCliEnv(
		mach.WithHost("127.0.0.1", machPort),
	)
	if err != nil {
		panic(err)
	}
	defer env.Close()

	ctx := context.TODO()

	conn, err := env.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	appender, err := conn.Appender(tableName, mach.WithErrorCheckCount(10))
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10000; i++ {
		if i%1000 == 0 {
			if err := appender.Flush(); err != nil {
				panic(err)
			}
		}
		if err := appender.Append(tagName, time.Now(), i); err != nil {
			panic(err)
		}
	}
	appender.Flush()
	s, f, err := appender.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(">> success:", s, ", fail:", f)
}
