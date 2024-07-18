package main

import (
	"context"
	"fmt"

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

	rows, err := conn.QueryContext(ctx, `select name, time, value from example where name = ? limit 10`, tagName)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var name string
	var ts string // or use time.Time
	var value float64
	for rows.Next() {
		if err := rows.Scan(&name, &ts, &value); err != nil {
			panic(err)
		}
		fmt.Println(">> name", fmt.Sprintf("%q", name), ", time:", ts, ", value:", value)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}
}
