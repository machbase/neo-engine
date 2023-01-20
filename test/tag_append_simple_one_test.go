package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createSimpleOneTagTable() {
	_, err := db.Exec(db.SqlTidy(
		`create tag table simple_one_tag(
			name            varchar(100) primary key, 
			time            datetime basetime, 
			value           double,
			svalue          varchar(80)
		)`))
	if err != nil {
		panic(err)
	}
}

func TestAppendTagSimpleOne(t *testing.T) {
	pr := db.QueryRow("select count(*) from simple_one_tag")
	if pr.Err() != nil {
		panic(pr.Err())
	}
	var existingCount int
	err := pr.Scan(&existingCount)
	if err != nil {
		panic(err)
	}

	t.Log("---- append simple_one_tag")
	appender, err := db.Appender("simple_one_tag")
	if err != nil {
		panic(err)
	}

	expectCount := 10000
	ts := time.Now()
	for i := 0; i < expectCount; i++ {
		err = appender.AppendWithTimestamp(
			ts.Add(time.Duration(i)),
			fmt.Sprintf("name-%d", i%10),
			1.001*float64(i+1),
			fmt.Sprintf("strvalue-%d", i),
		)
		if err != nil {
			panic(err)
		}
	}
	sc, fc, err := appender.Close()
	if err != nil {
		panic(err)
	}
	require.Equal(t, uint64(expectCount), sc)
	require.Equal(t, uint64(0), fc)

	rows, err := db.Query(`
		select
			name, time, value, svalue
		from
			simple_one_tag
		where
			time >= ?
		order by time`, ts)

	if err != nil {
		panic(err)
	}

	for i := 0; rows.Next(); i++ {
		var name string
		var ts time.Time
		var val float64
		var sval string

		err := rows.Scan(&name, &ts, &val, &sval)
		if err != nil {
			panic(err)
		}
		require.Equal(t, fmt.Sprintf("name-%d", i%10), name)
		require.Equal(t, fmt.Sprintf("strvalue-%d", i), sval)
	}
	rows.Close()

	r := db.QueryRow("select count(*) from simple_one_tag where time >= ?", ts)
	if r.Err() != nil {
		panic(r.Err())
	}
	var count int
	err = r.Scan(&count)
	if err != nil {
		panic(err)
	}
	require.Equal(t, expectCount, count)
	t.Log("---- append simple_one_tag done")
}
