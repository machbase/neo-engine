package httpsvr

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type WriteRequest struct {
	Data *WriteRequestData `json:"data"`
}

type WriteRequestData struct {
	Columns []string `json:"columns"`
	Records [][]any  `json:"records"`
}

type WriteResponse struct {
	Success bool               `json:"success"`
	Reason  string             `json:"reason"`
	Elapse  string             `json:"elapse"`
	Data    *WriteResponseData `json:"data,omitempty"`
}

type WriteResponseData struct {
	AffectedRows uint64 `json:"affectedRows"`
}

func (svr *Server) handleWrite(ctx *gin.Context) {
	tick := time.Now()

	tableName := ctx.Param("table")
	req := &WriteRequest{}
	rsp := &WriteResponse{Reason: "not specified"}

	err := ctx.Bind(req)
	if err != nil {
		rsp.Reason = err.Error()
		rsp.Elapse = time.Since(tick).String()
		ctx.JSON(http.StatusBadRequest, rsp)
		return
	}
	vf := make([]string, len(req.Data.Columns))
	for i := range vf {
		vf[i] = "?"
	}
	valuesFormat := strings.Join(vf, ",")
	columns := strings.Join(req.Data.Columns, ",")

	sqlText := fmt.Sprintf("insert into %s (%s) values(%s)", tableName, columns, valuesFormat)
	var nrows uint64
	for i, rec := range req.Data.Records {
		_, err := svr.db.Exec(sqlText, rec...)
		if err != nil {
			rsp.Reason = fmt.Sprintf("record[%d] %s", i, err.Error())
			rsp.Data = &WriteResponseData{
				AffectedRows: nrows,
			}
			rsp.Elapse = time.Since(tick).String()
			ctx.JSON(http.StatusBadRequest, rsp)
			return
		}
		nrows++
	}

	rsp.Success = true
	rsp.Reason = fmt.Sprintf("%d rows inserted", nrows)
	rsp.Data = &WriteResponseData{
		AffectedRows: nrows,
	}
	rsp.Elapse = time.Since(tick).String()
	ctx.JSON(http.StatusOK, rsp)
}
