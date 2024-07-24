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

	err = conn.ExecDirectContext(ctx, fmt.Sprintf(`
		create tag table if not exists %s (
			name     varchar(200) primary key,
			time     datetime basetime,
			value    double summarized
	)`, tableName))
	if err != nil {
		panic(err)
	}
}
