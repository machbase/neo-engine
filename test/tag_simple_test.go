package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createSimpleTagTable() {
	_, err := db.Exec(db.SqlTidy(
		`create tag table simple_tag(
			name            varchar(100) primary key, 
			time            datetime basetime, 
			value           double
		)`))
	if err != nil {
		panic(err)
	}
}

func TestAppendSimpleTag(t *testing.T) {
	t.Log("---- append simple_tag")
	appender, err := db.Appender("simple_tag")
	if err != nil {
		panic(err)
	}
	defer appender.Close()

	expectCount := 10000
	for i := 0; i < expectCount; i++ {
		err = appender.Append(
			fmt.Sprintf("name-%d", i%10),
			time.Now(),
			1.001*float64(i+1))
		if err != nil {
			panic(err)
		}
	}
	rows, err := db.Query("select name, time, value from simple_tag order by time")
	if err != nil {
		panic(err)
	}

	for i := 0; rows.Next(); i++ {
		var name string
		var ts time.Time
		var val float64

		err := rows.Scan(&name, &ts, &val)
		if err != nil {
			panic(err)
		}
		require.Equal(t, fmt.Sprintf("name-%d", i%10), name)
	}
	rows.Close()

	r := db.QueryRow("select count(*) from simple_tag")
	if r.Err() != nil {
		panic(r.Err())
	}
	var count int
	err = r.Scan(&count)
	if err != nil {
		panic(err)
	}
	require.Equal(t, expectCount, count)
	t.Log("---- append simple_tag done")
}
