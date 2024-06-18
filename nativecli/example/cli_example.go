//go:build !windows

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/machbase/neo-engine/nativecli"
)

func main() {
	ctx := context.TODO()

	// 1. Make Env
	env, err := nativecli.NewEnv(
		nativecli.WithHostPort("127.0.0.1", 5656),
		nativecli.WithUserPassword("sys", "manager"),
		nativecli.WithTimeformat(time.Kitchen),
		nativecli.WithTimeLocation(time.Local),
	)
	if err != nil {
		panic(err)
	}
	defer env.Close()

	// 2. Connect
	conn, err := env.Connect()
	if err != nil {
		panic(err)
	}

	// 3. Append Open
	apd, err := conn.AppendOpen(ctx,
		"example",
		nativecli.WithErrorCheckCount(1),
		nativecli.WithPrependValuesProvider(func() []any {
			return []any{"gocli", time.Now()}
		}),
	)
	if err != nil {
		panic(err)
	}

	// 4. Append
	for i := 0; i < 10; i++ {
		// If you didn't set the .WithPrependValueProvider, you should set the value like this.
		// err = apd.Append("gocli", time.Now(), 1.236)
		// But, in this case, the values will be prepended []{"gocli", time.Now(), 1.236}.
		err = apd.Append(1.236)
		if err != nil {
			panic(err)
		}
	}
	err = apd.Flush()
	if err != nil {
		panic(err)
	}

	// 5. Append Close
	s, f, err := apd.Close()
	fmt.Println("success", s, ", fail", f, ", error", err)
	if err != nil {
		panic(err)
	}
	if f != 0 {
		panic("failed count should be 0")
	}
	conn.Close()

	// 6. New Connection
	conn, err = env.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 7. Query
	sqlText := `select name, time, value from example where name = ? order by time desc limit ?`
	sqlParam := []any{"gocli", 10}
	rows, err := conn.QueryContext(ctx, sqlText, sqlParam...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if cs, err := rows.Columns(); err != nil {
		panic(err)
	} else {
		fmt.Println(cs)
	}
	rownum := 0
	for rows.Next() {
		var time time.Time
		var name string
		var value float64
		err = rows.Scan(&name, &time, &value)
		if err != nil {
			panic(err)
		}

		rownum++
		fmt.Println(rownum, name, time, value)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}
	if rownum != 10 {
		panic("rownum should be 10")
	}

	// 8. QueryRow
	row := conn.QueryRowContext(ctx, "select name, max(time), count(*) from EXAMPLE where name = 'gocli' group by name ")
	if row.Err() != nil {
		panic(row.Err())
	}

	var name string
	var maxTime string
	var count int
	err = row.Scan(&name, &maxTime, &count)
	if err != nil {
		panic(err)
	}
	fmt.Println(name, "maxTime:", maxTime, "count:", count)
}
