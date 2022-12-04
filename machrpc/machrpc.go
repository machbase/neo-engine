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
	conn, err := MakeGrpcConn(serverAddr)
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
	if !rsp.Success {
		return fmt.Errorf(rsp.Reason)
	}
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
		return &Rows{client: this, handle: rsp.RowsHandle}, nil
	} else {
		if len(rsp.Reason) > 0 {
			return nil, errors.New(rsp.Reason)
		}
		return nil, errors.New("unknown error")
	}
}

type Rows struct {
	client *Client
	handle *RowsHandle
	values []any
}

func (rows *Rows) Next() bool {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()
	rsp, err := rows.client.cli.RowsNext(ctx, rows.handle)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rsp.Success {
		rows.values = pbconv.ConvertPbToAny(rsp.Values)
	} else {
		rows.values = nil
	}
	return rsp.Success
}

func (rows *Rows) Scan(cols ...any) error {
	if rows.values != nil {
		return sql.ErrNoRows
	}
	return scan(rows.values, cols)
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

func scan(src []any, dst []any) error {
	for i := range dst {
		if i >= len(src) {
			return fmt.Errorf("column %d is out of range %d", i, len(src))
		}
		switch v := src[i].(type) {
		default:
			return fmt.Errorf("column %d is %T, not compatible with %T", i, v, dst[i])
		case *int:
			valconv.Int32ToAny(int32(*v), dst[i])
		case *int16:
			valconv.Int16ToAny(*v, dst[i])
		case *int32:
			valconv.Int32ToAny(*v, dst[i])
		case *int64:
			valconv.Int64ToAny(*v, dst[i])
		case *time.Time:
			valconv.DateTimeToAny(*v, dst[i])
		case *float32:
			valconv.Float32ToAny(*v, dst[i])
		case *float64:
			valconv.Float64ToAny(*v, dst[i])
		case *net.IP:
			valconv.IPToAny(*v, dst[i])
		case *string:
			valconv.StringToAny(*v, dst[i])
		case []byte:
			valconv.BytesToAny(v, dst[i])
		case int:
			valconv.Int32ToAny(int32(v), dst[i])
		case int16:
			valconv.Int16ToAny(v, dst[i])
		case int32:
			valconv.Int32ToAny(v, dst[i])
		case int64:
			valconv.Int64ToAny(v, dst[i])
		case time.Time:
			valconv.DateTimeToAny(v, dst[i])
		case float32:
			valconv.Float32ToAny(v, dst[i])
		case float64:
			valconv.Float64ToAny(v, dst[i])
		case net.IP:
			valconv.IPToAny(v, dst[i])
		case string:
			valconv.StringToAny(v, dst[i])
		}
	}
	return nil
}
