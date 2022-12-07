package msg

import (
	"fmt"

	mach "github.com/machbase/dbms-mach-go"
)

type QueryRequest struct {
	SqlText    string `json:"q"`
	Cursor     int    `json:"cursor"`
	Limit      int    `json:"limit"`
	Timeformat string `json:"timeformat,omitempty"`
}

type QueryResponse struct {
	Success bool       `json:"success"`
	Reason  string     `json:"reason"`
	Elapse  string     `json:"elapse"`
	Data    *QueryData `json:"data,omitempty"`
}

type QueryData struct {
	Cursor   int      `json:"cursor,omitempty"`
	Columns  []string `json:"colums"`
	Types    []string `json:"types"`
	Recorods [][]any  `json:"records"`
}

func Query(db *mach.Database, req *QueryRequest, rsp *QueryResponse) {
	cursor := req.Cursor
	limit := req.Limit
	timeformat := req.Timeformat

	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	if len(timeformat) == 0 {
		timeformat = "epoch"
	}

	rows, err := db.Query(req.SqlText)
	if err != nil {
		rsp.Reason = err.Error()
		return
	}
	defer rows.Close()
	rows.SetTimeFormat(timeformat)

	if !rows.IsFetchable() {
		rsp.Success = true
		rsp.Reason = "success"
		return
	}
	data := &QueryData{}
	data.Recorods = make([][]any, 0)
	data.Columns, err = rows.ColumnNames()
	if err != nil {
		rsp.Reason = err.Error()
		return
	}
	data.Types, err = rows.ColumnTypes()
	if err != nil {
		rsp.Reason = err.Error()
		return
	}
	rownum := 0
	for {
		rec, next, err := rows.Fetch()
		if err != nil {
			rsp.Reason = err.Error()
			return
		}
		if !next {
			cursor = 0
			break
		}
		rownum++
		if rownum-1 < cursor {
			continue
		}
		// for i, n := range rec {
		// 	if n == nil {
		// 		continue
		// 	}
		// 	switch v := n.(type) {
		// 	case *int64:
		// 		my.log.Tracef("%02d]]%v", i, *v)
		// 	default:
		// 		my.log.Tracef("%02d>>%#v", i, n)
		// 	}
		// }
		data.Recorods = append(data.Recorods, rec)

		if (rownum - cursor) >= limit {
			cursor = req.Cursor + (rownum - cursor)
			break
		}
	}
	data.Cursor = cursor

	rsp.Success = true
	rsp.Reason = fmt.Sprintf("%d records selected", len(data.Recorods))
	rsp.Data = data
}
