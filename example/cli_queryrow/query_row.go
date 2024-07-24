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

	row := conn.QueryRowContext(ctx, `select name, count(*) as c from example group by name having name = ?`, tagName)
	if row.Err() != nil {
		panic(row.Err())
	}

	var name string
	var count int
	if err := row.Scan(&name, &count); err != nil {
		panic(err)
	}
	fmt.Println(">> name", fmt.Sprintf("%q", name), ", count(*):", count)
}
