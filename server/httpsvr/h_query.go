package httpsvr

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type QueryRequest struct {
	SqlText string `json:"q"`
	Cursor  int    `json:"cursor"`
	Limit   int    `json:"limit"`
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

func (my *Server) handleQuery(ctx *gin.Context) {
	req := &QueryRequest{}
	rsp := &QueryResponse{Success: false, Reason: "not specified"}
	tick := time.Now()

	var err error
	var strCursor string
	var strLimit string
	var timeformat string
	if ctx.Request.Method == http.MethodPost {
		req.SqlText = ctx.PostForm("q")
		strCursor = ctx.PostForm("cursor")
		strLimit = ctx.PostForm("limit")
		timeformat = ctx.PostForm("timeformat")
	} else if ctx.Request.Method == http.MethodGet {
		req.SqlText = ctx.Query("q")
		strCursor = ctx.Query("cursor")
		strLimit = ctx.Query("limit")
		timeformat = ctx.PostForm("timeformat")
	}
	if len(req.SqlText) == 0 {
		rsp.Reason = "empty sql"
		rsp.Elapse = time.Since(tick).String()
		ctx.JSON(http.StatusBadRequest, rsp)
		return
	}
	if len(strCursor) == 0 {
		req.Cursor = 0
	} else {
		req.Cursor, err = strconv.Atoi(strCursor)
		if err != nil {
			rsp.Reason = "invalid cursor"
			rsp.Elapse = time.Since(tick).String()
			ctx.JSON(http.StatusBadRequest, rsp)
			return
		}
	}
	if len(strLimit) == 0 {
		req.Limit = 10
	} else {
		req.Limit, err = strconv.Atoi(strLimit)
		if err != nil {
			rsp.Reason = "invalid limit"
			rsp.Elapse = time.Since(tick).String()
			ctx.JSON(http.StatusBadRequest, rsp)
			return
		}
	}

	if len(timeformat) == 0 {
		timeformat = "epoch"
	}

	cursor := req.Cursor
	limit := req.Limit

	rows, err := my.db.Query(req.SqlText)
	if err != nil {
		rsp.Reason = err.Error()
		rsp.Elapse = time.Since(tick).String()
		ctx.JSON(http.StatusInternalServerError, rsp)
		return
	}
	defer rows.Close()
	rows.SetTimeFormat(timeformat)

	data := &QueryData{}
	data.Recorods = make([][]any, 0)
	data.Columns, err = rows.ColumnNames()
	if err != nil {
		rsp.Reason = err.Error()
		rsp.Elapse = time.Since(tick).String()
		ctx.JSON(http.StatusInternalServerError, rsp)
		return
	}
	data.Types, err = rows.ColumnTypes()
	if err != nil {
		rsp.Reason = err.Error()
		rsp.Elapse = time.Since(tick).String()
		ctx.JSON(http.StatusInternalServerError, rsp)
		return
	}
	rownum := 0
	for {
		rec, next, err := rows.Fetch()
		if err != nil {
			rsp.Reason = err.Error()
			rsp.Elapse = time.Since(tick).String()
			ctx.JSON(http.StatusInternalServerError, rsp)
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
	rsp.Elapse = time.Since(tick).String()
	rsp.Data = data
	ctx.JSON(http.StatusOK, rsp)
}
