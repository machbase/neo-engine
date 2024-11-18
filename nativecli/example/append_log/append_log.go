//go:build !windows

package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/machbase/neo-engine/v8/nativecli"
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
	conn, err := env.Connect(ctx)
	if err != nil {
		panic(err)
	}

	// 3. Create Table
	err = conn.ExecDirectContext(ctx,
		`CREATE TABLE IF NOT EXISTS CLI_SAMPLE(
				seq short,
				score integer,
				total long,
				percentage float,
				ratio double,
				id varchar(10),
				srcip ipv4,
				dstip ipv6,
				reg_date datetime,
				textlog text,
				image binary
			)`,
	)
	if err != nil {
		panic(err)
	}

	// 4. Append Open
	apd, err := conn.AppendOpen(ctx,
		"CLI_SAMPLE",
		nativecli.WithErrorCheckCount(1),
	)
	if err != nil {
		panic(err)
	}

	// 5. Append
	for i := 1; i <= 10; i++ {
		sSeq := i
		sScore := i + i
		sTotal := (sSeq + sScore) * 100
		sPercentage := float32(sScore) / float32(sTotal)
		sRatio := float64(sSeq) / float64(sTotal)
		sId := fmt.Sprintf("id-%d", i)
		sSrcIP := net.ParseIP(fmt.Sprintf("192.168.0.%d", i))
		if sSrcIP == nil {
			panic("invalid ipv4")
		}
		sDstIP := net.ParseIP(fmt.Sprintf("2001:0DB8:0000:0000:0000:0000:1428:%04d", i))
		if sDstIP == nil {
			panic("invalid ipv6")
		}
		sRegDate := time.Unix(int64(i), 0)
		sLog := fmt.Sprintf("text log-%d", i)
		sImage := []byte(fmt.Sprintf("binary image-%d", i))
		data := []any{
			sSeq,        // seq short
			sScore,      // score integer
			sTotal,      // total long
			sPercentage, // percentage float
			sRatio,      // ratio double
			sId,         // id varchar(10)
			sSrcIP,      // srcip ipv4
			sDstIP,      // dstip ipv6
			sRegDate,    // reg_date datetime
			sLog,        // textlog text
			sImage,      // image binary
		}

		err = apd.Append(data...)
		if err != nil {
			panic(err)
		}
	}
	err = apd.Flush()
	if err != nil {
		panic(err)
	}

	// 6. Append Close
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
	conn, err = env.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 7. Query
	sqlText := `select * from cli_sample`
	rows, err := conn.QueryContext(ctx, sqlText)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var seq int16
		var score int32
		var total int64
		var percentage float32
		var ratio float64
		var id string
		var srcip string
		var dstip string
		var regDate time.Time
		var textlog string
		var image []byte
		err = rows.Scan(&seq, &score, &total, &percentage, &ratio, &id, &srcip, &dstip, &regDate, &textlog, &image)
		if err != nil {
			panic(err)
		}
		fmt.Println(seq, score, total, percentage, ratio, id, srcip, dstip, regDate, textlog, string(image))
		if srcip != "192.168.0."+fmt.Sprint(seq) {
			panic("invalid srcip")
		}
		if dstip != "2001:db8::1428:"+fmt.Sprintf("%d", seq) {
			panic("invalid dstip")
		}
	}
	/* ignore err */
	conn.ExecDirectContext(ctx, `DROP TABLE CLI_SAMPLE`)
}
