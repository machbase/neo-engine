package msg

import (
	"errors"
	"fmt"
	"strings"
	"time"

	mach "github.com/machbase/dbms-mach-go"
)

/* Interpreting Influx lineprotocol

   | Machbase            | influxdb                                    |
   | ------------------- | ------------------------------------------- |
   | table name          | db                                          |
   | tag name            | measurement (+ '.' + field named 'name', if exists) |
   | time                | timestamp (if data contains a field named 'time', the field will be ignored) |
   | value               | value of field named 'value', if there is no field of float type named 'value', zero(0.0) will be inserted |
   | additional columnns | other fields than 'name', 'value' and 'time' |
*/

func WriteLineProtocol(db *mach.Database, dbName string, measurement string, fields map[string]any, tags map[string]string, ts time.Time) error {
	columns := make([]string, 0)
	row := make([]any, 0)

	columns = append(columns, "name", "value", "time")
	if v, ok := fields["name"]; ok {
		row = append(row, fmt.Sprintf("%s.%s", measurement, v))
	} else {
		row = append(row, measurement)
	}
	if v, ok := fields["value"]; ok {
		row = append(row, v)
	} else {
		row = append(row, float32(0))
	}
	row = append(row, ts)

	for k, v := range fields {
		switch strings.ToLower(k) {
		case "name", "value", "time":
			continue
		}
		columns = append(columns, k)
		row = append(row, v)
	}

	writeReq := &WriteRequest{
		Table: dbName,
		Data: &WriteRequestData{
			Columns: columns,
			Rows:    [][]any{row},
		},
	}
	writeRsp := &WriteResponse{}
	// fmt.Printf("REQ ==> %s %s %+v\n", writeReq.Table, measurement, writeReq.Data)
	Write(db, writeReq, writeRsp)
	// fmt.Printf("RSP ==> %#v\n", writeRsp)
	if writeRsp.Success {
		return nil
	} else {
		return errors.New(writeRsp.Reason)
	}
}
