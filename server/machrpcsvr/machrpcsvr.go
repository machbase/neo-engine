package machrpcsvr

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	mach "github.com/machbase/dbms-mach-go"
	"github.com/machbase/dbms-mach-go/machrpc"
	"github.com/machbase/dbms-mach-go/pbconv"
	cmap "github.com/orcaman/concurrent-map"
	"google.golang.org/grpc/stats"
)

type Config struct {
}

/**
machrpcSvr, err := machrpcsvr.New(&machrpcsvr.Config{})
// gRPC options
grpcOpt := []grpc.ServerOption{ grpc.StatsHandler(machrpcSvr) }
// crete gRpc server
this.grpcd = grpc.NewServer(grpcOpt...)
// register gRpc server
machrpc.RegisterMachbaseServer(this.grpcd, machrpcSvr)
*/

type Server interface {
	stats.Handler
	machrpc.MachbaseServer // machrpc server interface
}

func New(conf *Config) (Server, error) {
	return &svr{
		conf:     conf,
		ctxMap:   cmap.New(),
		machbase: mach.New(),
	}, nil
}

type svr struct {
	machrpc.MachbaseServer // machrpc server interface

	conf     *Config
	ctxMap   cmap.ConcurrentMap
	machbase *mach.Database
}

func (this *svr) Start() error {
	return nil
}

func (this *svr) Stop() {

}

type sessionCtx struct {
	context.Context
	Id     string
	values map[any]any
}

type stringer interface {
	String() string
}

func contextName(c context.Context) string {
	if s, ok := c.(stringer); ok {
		return s.String()
	}
	return reflect.TypeOf(c).String()
}

func (c *sessionCtx) String() string {
	return contextName(c.Context) + "(" + c.Id + ")"
}

func (c *sessionCtx) Value(key any) any {
	if key == contextCtxKey {
		return c
	}
	if v, ok := c.values[key]; ok {
		return v
	}
	return c.Context.Value(key)
}

type rowsWrap struct {
	id      string
	rows    *mach.Rows
	release func()
}

const contextCtxKey = "machrpc-client-context"

var contextIdSerial int64

//// grpc stat handler

func (this *svr) TagRPC(ctx context.Context, nfo *stats.RPCTagInfo) context.Context {
	return ctx
}

func (this *svr) HandleRPC(ctx context.Context, stat stats.RPCStats) {
}

func (this *svr) TagConn(ctx context.Context, nfo *stats.ConnTagInfo) context.Context {
	id := strconv.FormatInt(atomic.AddInt64(&contextIdSerial, 1), 10)
	ctx = &sessionCtx{Context: ctx, Id: id}
	this.ctxMap.Set(id, ctx)
	return ctx
}

func (this *svr) HandleConn(ctx context.Context, s stats.ConnStats) {
	if sessCtx, ok := ctx.(*sessionCtx); ok {
		switch s.(type) {
		case *stats.ConnBegin:
			fmt.Printf("get connBegin: %v\n", sessCtx.Id)
		case *stats.ConnEnd:
			this.ctxMap.RemoveCb(sessCtx.Id, func(key string, v interface{}, exists bool) bool {
				fmt.Printf("get connEnd: %v\n", sessCtx.Id)
				return true
			})
		}
	}
}

//// machrpc server handler

func (this *svr) Exec(pctx context.Context, req *machrpc.ExecRequest) (*machrpc.ExecResponse, error) {
	rsp := &machrpc.ExecResponse{}
	tick := time.Now()
	defer func() {
		rsp.Elapse = time.Since(tick).String()
	}()

	params := pbconv.ConvertPbToAny(req.Params)
	if err := this.machbase.Exec(req.Sql, params...); err == nil {
		rsp.Success = true
		rsp.Reason = "success"
	} else {
		rsp.Success = false
		rsp.Reason = err.Error()
	}
	return rsp, nil
}

func (this *svr) QueryRow(pctx context.Context, req *machrpc.QueryRowRequest) (*machrpc.QueryRowResponse, error) {
	rsp := &machrpc.QueryRowResponse{}

	tick := time.Now()
	defer func() {
		rsp.Elapse = time.Since(tick).String()
	}()

	// val := pctx.Value(contextCtxKey)
	// ctx, ok := val.(*sessionCtx)
	// if !ok {
	// 	return nil, fmt.Errorf("invlaid session context %T", pctx)
	// }

	params := pbconv.ConvertPbToAny(req.Params)
	row := this.machbase.QueryRow(req.Sql, params...)

	// fmt.Printf("QueryRow : %s  %s   rows: %d\n", ctx.Id, req.Sql, len(row.Values()))

	if row.Err() != nil {
		rsp.Reason = row.Err().Error()
		return rsp, nil
	}

	var err error
	rsp.Success = true
	rsp.Reason = "success"
	rsp.Values, err = pbconv.ConvertAnyToPb(row.Values())
	if err != nil {
		rsp.Success = false
		rsp.Reason = err.Error()
	}

	return rsp, err
}

func (this *svr) Query(pctx context.Context, req *machrpc.QueryRequest) (*machrpc.QueryResponse, error) {
	rsp := &machrpc.QueryResponse{}

	tick := time.Now()
	defer func() {
		rsp.Elapse = time.Since(tick).String()
	}()

	// val := pctx.Value(contextCtxKey)
	// ctx, ok := val.(*sessionCtx)
	// if !ok {
	// 	return nil, fmt.Errorf("invlaid session context %T", pctx)
	// }
	// fmt.Printf("Query : %s %s\n", ctx.Id, req.Sql)

	params := pbconv.ConvertPbToAny(req.Params)
	realRows, err := this.machbase.Query(req.Sql, params...)
	if err != nil {
		rsp.Reason = err.Error()
		return rsp, nil
	}

	handle := strconv.FormatInt(atomic.AddInt64(&contextIdSerial, 1), 10)
	this.ctxMap.Set(handle, &rowsWrap{
		id:   handle,
		rows: realRows,
		release: func() {
			this.ctxMap.RemoveCb(handle, func(key string, v interface{}, exists bool) bool {
				fmt.Printf("close rows: %v\n", handle)
				realRows.Close()
				return true
			})
		},
	})

	rsp.Success = true
	rsp.Reason = "success"
	rsp.RowsHandle = &machrpc.RowsHandle{
		Handle: handle,
	}

	return rsp, nil
}

func (this *svr) RowsNext(ctx context.Context, rows *machrpc.RowsHandle) (*machrpc.RowsNextResponse, error) {
	rsp := &machrpc.RowsNextResponse{}
	tick := time.Now()
	defer func() {
		rsp.Elapse = time.Since(tick).String()
	}()

	rowsWrapVal, exists := this.ctxMap.Get(rows.Handle)
	if !exists {
		rsp.Reason = fmt.Sprintf("handle '%s' not found", rows.Handle)
		return rsp, nil
	}
	rowsWrap, ok := rowsWrapVal.(*rowsWrap)
	if !ok {
		rsp.Reason = fmt.Sprintf("handle '%s' is not valid", rows.Handle)
		return rsp, nil
	}

	if !rowsWrap.rows.Next() {
		rsp.Success = false
		rsp.Reason = "no rows"
		return rsp, nil
	}

	values := make([]any, rowsWrap.rows.ColumnCount())

	err := rowsWrap.rows.Scan(values...)
	if err != nil {
		rsp.Success = false
		rsp.Reason = err.Error()
		return rsp, nil
	}
	rsp.Values, err = pbconv.ConvertAnyToPb(values)
	if err != nil {
		rsp.Success = false
		rsp.Reason = err.Error()
		return rsp, nil
	}
	rsp.Success = true
	rsp.Reason = "success"
	return rsp, nil
}

func (this *svr) RowsClose(ctx context.Context, rows *machrpc.RowsHandle) (*machrpc.RowsCloseResponse, error) {
	rsp := &machrpc.RowsCloseResponse{}
	tick := time.Now()
	defer func() {
		rsp.Elapse = time.Since(tick).String()
	}()

	rowsWrapVal, exists := this.ctxMap.Get(rows.Handle)
	if !exists {
		rsp.Reason = fmt.Sprintf("handle '%s' not found", rows.Handle)
		return rsp, nil
	}
	rowsWrap, ok := rowsWrapVal.(*rowsWrap)
	if !ok {
		rsp.Reason = fmt.Sprintf("handle '%s' is not valid", rows.Handle)
		return rsp, nil
	}

	rowsWrap.release()
	rsp.Success = true
	rsp.Reason = "success"
	return rsp, nil
}

func (this *svr) Append(stream machrpc.Machbase_AppendServer) error {
	pctx := stream.Context()
	val := pctx.Value(contextCtxKey)
	ctx, ok := val.(*sessionCtx)
	if !ok {
		return fmt.Errorf("invlaid session context %T", pctx)
	}
	fmt.Printf("session id : %s\n", ctx.Id)

	for {
		m, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&machrpc.AppendResponse{})
		} else if err != nil {
			fmt.Printf("Recv %s\n", err.Error())
			return err
		}
		fmt.Printf("==> %+v\n", m)
	}
}
