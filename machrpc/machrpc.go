package machrpc

import (
	context "context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/machbase/dbms-mach-go/pbconv"
	"github.com/machbase/dbms-mach-go/valconv"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Client struct {
	conn grpc.ClientConnInterface
	cli  MachbaseClient
}

func NewClient() *Client {
	client := &Client{}
	return client
}

func (this *Client) Connect(serverAddr string) error {
	conn, err := MakeGrpcConn("unix://../../tmp/tagd.sock")
	if err != nil {
		return errors.Wrap(err, "NewClient")
	}
	this.conn = conn
	this.cli = NewMachbaseClient(conn)
	return nil
}

func (this *Client) Disconnect() {
	this.conn = nil
	this.cli = nil
}

func (this *Client) Exec(ctx context.Context, sqlText string, params ...any) error {
	pbparams, err := pbconv.ConvertAnyToPb(params)
	if err != nil {
		return err
	}
	req := &ExecRequest{
		Sql:    sqlText,
		Params: pbparams,
	}
	rsp, err := this.cli.Exec(ctx, req)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", rsp)
	return nil
}

func (this *Client) Query(ctx context.Context, sqlText string, params ...any) (*Rows, error) {
	pbparams, err := pbconv.ConvertAnyToPb(params)
	if err != nil {
		return nil, err
	}

	req := &QueryRequest{Sql: sqlText, Params: pbparams}
	rsp, err := this.cli.Query(ctx, req)
	if err != nil {
		return nil, err
	}

	if rsp.Success {
		return rsp.Rows, nil
	} else {
		if len(rsp.Reason) > 0 {
			return nil, errors.New(rsp.Reason)
		}
		return nil, errors.New("unknown error")
	}
}

func (rows *Rows) Next() bool {
	return false
}

func (this *Client) QueryRow(ctx context.Context, sqlText string, params ...any) *Row {
	pbparams, err := pbconv.ConvertAnyToPb(params)
	if err != nil {
		return &Row{success: false, err: err}
	}

	req := &QueryRowRequest{Sql: sqlText, Params: pbparams}
	rsp, err := this.cli.QueryRow(ctx, req)
	if err != nil {
		return &Row{success: false, err: err}
	}

	var row = &Row{}
	row.success = rsp.Success
	row.err = nil
	if !rsp.Success && len(rsp.Reason) > 0 {
		row.err = errors.New(rsp.Reason)
	}
	row.values = pbconv.ConvertPbToAny(rsp.Values)
	return row
}

type Row struct {
	success bool
	err     error
	values  []any
}

func (row *Row) Err() error {
	return row.err
}

func (row *Row) Scan(cols ...any) error {
	if row.err != nil {
		return row.err
	}
	if !row.success {
		return sql.ErrNoRows
	}
	for i := range cols {
		if i >= len(row.values) {
			return fmt.Errorf("column %d is out of range %d", i, len(row.values))
		}
		switch v := row.values[i].(type) {
		default:
			return fmt.Errorf("column %d is %T, not compatible with %T", i, v, cols[i])
		case *int:
			valconv.Int32ToAny(int32(*v), cols[i])
		case *int16:
			valconv.Int16ToAny(*v, cols[i])
		case *int32:
			valconv.Int32ToAny(*v, cols[i])
		case *int64:
			valconv.Int64ToAny(*v, cols[i])
		case *time.Time:
			valconv.DateTimeToAny(*v, cols[i])
		case *float32:
			valconv.Float32ToAny(*v, cols[i])
		case *float64:
			valconv.Float64ToAny(*v, cols[i])
		case *net.IP:
			valconv.IPToAny(*v, cols[i])
		case *string:
			valconv.StringToAny(*v, cols[i])
		case []byte:
			valconv.BytesToAny(v, cols[i])
		case int:
			valconv.Int32ToAny(int32(v), cols[i])
		case int16:
			valconv.Int16ToAny(v, cols[i])
		case int32:
			valconv.Int32ToAny(v, cols[i])
		case int64:
			valconv.Int64ToAny(v, cols[i])
		case time.Time:
			valconv.DateTimeToAny(v, cols[i])
		case float32:
			valconv.Float32ToAny(v, cols[i])
		case float64:
			valconv.Float64ToAny(v, cols[i])
		case net.IP:
			valconv.IPToAny(v, cols[i])
		case string:
			valconv.StringToAny(v, cols[i])
		}
	}
	return nil
}
