package httpsvr

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/machbase/dbms-mach-go/server/msg"
)

// Configure telegraf.conf
//
//	[[outputs.http]]
//	url = "http://127.0.0.1:4088/metrics/write"
//	data_format = "influx"
//	content_encoding = "gzip"
func (svr *Server) handleLineProtocol(ctx *gin.Context) {
	oper := ctx.Param("oper")
	method := ctx.Request.Method

	if method == http.MethodPost && oper == "write" {
		svr.handleLineWrite(ctx)
	} else {
		ctx.JSON(
			http.StatusNotImplemented,
			gin.H{"error": fmt.Sprintf("%s %s is not implemented", method, oper)})
	}
}

func (svr *Server) handleLineWrite(c *gin.Context) {
	precision := lineprotocol.Nanosecond
	switch c.Query("precision") {
	case "us":
		precision = lineprotocol.Microsecond
	case "ms":
		precision = lineprotocol.Millisecond
	}

	var body io.Reader
	switch c.Request.Header.Get("Content-Encoding") {
	default:
		body = c.Request.Body
	case "gzip":
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("invalid gzip compression: %s", err.Error())})
			return
		}
		defer gz.Close()
		body = gz
	}

	dec := lineprotocol.NewDecoder(body)
	for dec != nil && dec.Next() {
		m, err := dec.Measurement()
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": fmt.Sprintf("measurement error: %s", err.Error())})
			return
		}
		measurement := string(m)
		tags := make(map[string]string)
		fields := make(map[string]any)

		for {
			key, val, err := dec.NextTag()
			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{"error": fmt.Sprintf("tag error: %s", err.Error())})
				return
			}
			if key == nil {
				break
			}
			tags[string(key)] = string(val)
		}

		for {
			key, val, err := dec.NextField()
			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					gin.H{"error": fmt.Sprintf("field error: %s", err.Error())})
				return
			}
			if key == nil {
				break
			}
			fields[string(key)] = val.Interface()
		}

		ts, err := dec.Time(precision, time.Time{})
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("time error: %s", err.Error())})
			return
		}
		if ts.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no timestamp"})
			return
		}

		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("unsupproted data type tags %s", err.Error())})
			return
		}
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": fmt.Sprintf("unsupproted data type fields %s", err.Error())})
			return
		}

		columns := make([]string, len(fields))
		rows := make([][]any, 1)
		rows[0] = make([]any, len(fields))
		var i = 0
		for k, v := range fields {
			columns[i] = k
			rows[0][i] = v
		}
		writeReq := &msg.WriteRequest{
			Table: measurement,
			Data: &msg.WriteRequestData{
				Columns: columns,
				Rows:    rows,
			},
		}
		writeRsp := &msg.WriteResponse{}
		msg.Write(svr.db, writeReq, writeRsp)
	}
	c.JSON(http.StatusNoContent, "")
}
