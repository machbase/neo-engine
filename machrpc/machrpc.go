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

	closeTimeout time.Duration
	queryTimeout time.Duration
}

func NewClient(options ...ClientOption) *Client {
	client := &Client{
		closeTimeout: 3 * time.Second,
		queryTimeout: 0,
	}
	for _, opt := range options {
		switch o := opt.(type) {
		case *queryTimeoutOption:
			client.queryTimeout = o.timeout
		case *closeTimeoutOption:
			client.closeTimeout = o.timeout
		}
	}
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

func (this *Client) queryContext() (context.Context, context.CancelFunc) {
	if this.queryTimeout > 0 {
		return context.WithTimeout(context.Background(), this.queryTimeout)
	} else {
		ctx := context.Background()
		return ctx, func() {}
	}
}

func (this *Client) Exec(sqlText string, params ...any) error {
	ctx, cancelFunc := this.queryContext()
	defer cancelFunc()
	return this.ExecContext(ctx, sqlText, params...)
}

func (this *Client) ExecContext(ctx context.Context, sqlText string, params ...any) error {
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

func (this *Client) Query(sqlText string, params ...any) (*Rows, error) {
	ctx, cancelFunc := this.queryContext()
	defer cancelFunc()
	return this.QueryContext(ctx, sqlText, params...)
}

func (this *Client) QueryContext(ctx context.Context, sqlText string, params ...any) (*Rows, error) {
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
	err    error
}

func (rows *Rows) Close() error {
	var ctx context.Context
	if rows.client.closeTimeout > 0 {
		ctx0, cancelFunc := context.WithTimeout(context.Background(), rows.client.closeTimeout)
		defer cancelFunc()
		ctx = ctx0
	} else {
		ctx = context.Background()
	}
	_, err := rows.client.cli.RowsClose(ctx, rows.handle)
	return err
}

func (rows *Rows) Next() bool {
	if rows.err != nil {
		return false
	}
	ctx, cancelFunc := rows.client.queryContext()
	defer cancelFunc()
	rsp, err := rows.client.cli.RowsFetch(ctx, rows.handle)
	if err != nil {
		rows.err = err
		return false
	}
	if rsp.Success {
		if rsp.HasNoRows {
			return false
		}
		rows.values = pbconv.ConvertPbToAny(rsp.Values)
	} else {
		if len(rsp.Reason) > 0 {
			rows.err = errors.New(rsp.Reason)
		}
		rows.values = nil
	}
	return !rsp.HasNoRows
}

func (rows *Rows) Scan(cols ...any) error {
	if rows.err != nil {
		return rows.err
	}
	if rows.values == nil {
		return sql.ErrNoRows
	}
	return scan(rows.values, cols)
}

func (this *Client) QueryRow(sqlText string, params ...any) *Row {
	ctx, cancelFunc := this.queryContext()
	defer cancelFunc()
	return this.QueryRowContext(ctx, sqlText, params...)
}

func (this *Client) QueryRowContext(ctx context.Context, sqlText string, params ...any) *Row {
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
	err := scan(row.values, cols)
	return err
}

func scan(src []any, dst []any) error {
	var err error
	for i := range dst {
		if i >= len(src) {
			return fmt.Errorf("column %d is out of range %d", i, len(src))
		}
		switch v := src[i].(type) {
		default:
			return fmt.Errorf("column %d is %T, not compatible with %T", i, v, dst[i])
		case *int:
			err = valconv.Int32ToAny(int32(*v), dst[i])
		case *int16:
			err = valconv.Int16ToAny(*v, dst[i])
		case *int32:
			err = valconv.Int32ToAny(*v, dst[i])
		case *int64:
			err = valconv.Int64ToAny(*v, dst[i])
		case *time.Time:
			err = valconv.DateTimeToAny(*v, dst[i])
		case *float32:
			err = valconv.Float32ToAny(*v, dst[i])
		case *float64:
			err = valconv.Float64ToAny(*v, dst[i])
		case *net.IP:
			err = valconv.IPToAny(*v, dst[i])
		case *string:
			err = valconv.StringToAny(*v, dst[i])
		case []byte:
			err = valconv.BytesToAny(v, dst[i])
		case int:
			err = valconv.Int32ToAny(int32(v), dst[i])
		case int16:
			err = valconv.Int16ToAny(v, dst[i])
		case int32:
			err = valconv.Int32ToAny(v, dst[i])
		case int64:
			err = valconv.Int64ToAny(v, dst[i])
		case time.Time:
			err = valconv.DateTimeToAny(v, dst[i])
		case float32:
			err = valconv.Float32ToAny(v, dst[i])
		case float64:
			err = valconv.Float64ToAny(v, dst[i])
		case net.IP:
			err = valconv.IPToAny(v, dst[i])
		case string:
			err = valconv.StringToAny(v, dst[i])
		}
		if err != nil {
			return err
		}
	}
	return nil
}
