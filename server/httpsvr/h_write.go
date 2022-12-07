package httpsvr

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/machbase/dbms-mach-go/server/msg"
)

func (svr *Server) handleWrite(ctx *gin.Context) {
	tick := time.Now()

	tableName := ctx.Param("table")
	req := &msg.WriteRequest{}
	rsp := &msg.WriteResponse{Reason: "not specified"}

	err := ctx.Bind(req)
	if err != nil {
		rsp.Reason = err.Error()
		rsp.Elapse = time.Since(tick).String()
		ctx.JSON(http.StatusBadRequest, rsp)
		return
	}

	msg.Write(svr.db, tableName, req, rsp)
	rsp.Elapse = time.Since(tick).String()

	if rsp.Success {
		ctx.JSON(http.StatusOK, rsp)
	} else {
		ctx.JSON(http.StatusBadRequest, rsp)
	}
}
