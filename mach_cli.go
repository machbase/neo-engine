package mach

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type SqlType int

const (
	MACHCLI_SQL_TYPE_INT16    SqlType = 0
	MACHCLI_SQL_TYPE_INT32    SqlType = 1
	MACHCLI_SQL_TYPE_INT64    SqlType = 2
	MACHCLI_SQL_TYPE_DATETIME SqlType = 3
	MACHCLI_SQL_TYPE_FLOAT    SqlType = 4
	MACHCLI_SQL_TYPE_DOUBLE   SqlType = 5
	MACHCLI_SQL_TYPE_IPV4     SqlType = 6
	MACHCLI_SQL_TYPE_IPV6     SqlType = 7
	MACHCLI_SQL_TYPE_STRING   SqlType = 8
	MACHCLI_SQL_TYPE_BINARY   SqlType = 9
)

func (st SqlType) String() string {
	switch st {
	case MACHCLI_SQL_TYPE_INT16:
		return "INT16"
	case MACHCLI_SQL_TYPE_INT32:
		return "INT32"
	case MACHCLI_SQL_TYPE_INT64:
		return "INT64"
	case MACHCLI_SQL_TYPE_DATETIME:
		return "DATETIME"
	case MACHCLI_SQL_TYPE_FLOAT:
		return "FLOAT"
	case MACHCLI_SQL_TYPE_DOUBLE:
		return "DOUBLE"
	case MACHCLI_SQL_TYPE_IPV4:
		return "IPV4"
	case MACHCLI_SQL_TYPE_IPV6:
		return "IPV6"
	case MACHCLI_SQL_TYPE_STRING:
		return "STRING"
	case MACHCLI_SQL_TYPE_BINARY:
		return "BINARY"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", st)
	}
}

type CType int

const (
	MACHCLI_C_TYPE_INT16  CType = 101
	MACHCLI_C_TYPE_INT32  CType = 102
	MACHCLI_C_TYPE_INT64  CType = 103
	MACHCLI_C_TYPE_FLOAT  CType = 104
	MACHCLI_C_TYPE_DOUBLE CType = 105
	MACHCLI_C_TYPE_CHAR   CType = 106
)

type CliEnv struct {
	handle     unsafe.Pointer
	tz         *time.Location
	timeformat string
	host       string
	port       int
	user       string
	password   string
}

type CliOption func(*CliEnv)

func WithTimeLocation(loc *time.Location) CliOption {
	return func(c *CliEnv) {
		c.tz = loc
	}
}

// WithTimeformat sets the time format for the time.Time type.
// The default format is "2006-01-02 15:04:05".
func WithTimeformat(fmt string) CliOption {
	return func(c *CliEnv) {
		c.timeformat = fmt
	}
}

func WithHost(host string, port int) CliOption {
	return func(c *CliEnv) {
		c.host = host
		c.port = port
	}
}

func WithUser(user, password string) CliOption {
	return func(c *CliEnv) {
		c.user = user
		c.password = password
	}
}

type HandleType int

const (
	MACHCLI_HANDLE_ENV  HandleType = 1
	MACHCLI_HANDLE_DBC  HandleType = 2
	MACHCLI_HANDLE_STMT HandleType = 3
)

func errorWithCause(obj any, cause error) error {
	var handle unsafe.Pointer
	var handleType HandleType
	switch x := obj.(type) {
	case *CliEnv:
		handle = x.handle
		handleType = MACHCLI_HANDLE_ENV
	case *CliConn:
		handle = x.handle
		handleType = MACHCLI_HANDLE_DBC
	case *CliStmt:
		handle = x.handle
		handleType = MACHCLI_HANDLE_STMT
	default:
		return cause
	}
	var code int
	var msg string
	if reerr := cliError(handle, handleType, &code, &msg); reerr != nil {
		if cause == nil {
			return fmt.Errorf("MACHCLI Fail to get error, %s", reerr.Error())
		} else {
			return fmt.Errorf("MACHCLI Fail to get error code of %s, %s", cause.Error(), reerr.Error())
		}
	} else if msg == "" {
		return cause
	} else {
		if cause == nil {
			if code == 0 {
				return nil // no error
			}
			return fmt.Errorf("MACHCLI ERR-%d, %s", code, msg)
		} else {
			return fmt.Errorf("MACHCLI ERR-%d, %s", code, msg)
		}
	}
}

func NewCliEnv(opts ...CliOption) (*CliEnv, error) {
	var h unsafe.Pointer
	if err := cliInitialize(&h); err != nil {
		return nil, err
	}
	ret := &CliEnv{
		handle:     h,
		tz:         time.Local,
		timeformat: "2006-01-02 15:04:05",
		host:       "127.0.0.1",
		port:       5656,
		user:       "sys",
		password:   "manager",
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret, nil
}

func (c *CliEnv) Close() error {
	if err := cliFinalize(c.handle); err == nil {
		return nil
	} else {
		return errorWithCause(c, err)
	}
}

func (c *CliEnv) Error() error {
	return errorWithCause(c, nil)
}

func (c *CliEnv) ConnectionString() string {
	return fmt.Sprintf("SERVER=%s;UID=%s;PWD=%s;CONNTYPE=1;PORT_NO=%d",
		c.host, strings.ToUpper(c.user), strings.ToUpper(c.password), c.port)
}

func (c *CliEnv) SetTimeformat(format string) {
	c.timeformat = format
}

func (c *CliEnv) SetTimeLocation(loc *time.Location) {
	c.tz = loc
}

func (c *CliEnv) Connect(ctx context.Context) (*CliConn, error) {
	var h unsafe.Pointer
	if err := cliConnect(c.handle, c.ConnectionString(), &h); err == nil {
		ret := &CliConn{
			handle: h,
			ctx:    ctx,
			env:    c,
		}
		return ret, nil
	} else {
		return nil, errorWithCause(c, err)
	}
}

type CliConn struct {
	handle unsafe.Pointer
	ctx    context.Context
	env    *CliEnv
}

func (c *CliConn) Close() error {
	if err := cliDisconnect(c.handle); err == nil {
		return nil
	} else {
		return errorWithCause(c, err)
	}
}

func (c *CliConn) Error() error {
	return errorWithCause(c, nil)
}

func (c *CliConn) ExecDirectContext(ctx context.Context, query string) error {
	stmt, err := c.NewStmt()
	if err != nil {
		return err
	}
	defer stmt.Close()
	if err := cliExecDirect(stmt.handle, query); err == nil {
		return nil
	} else {
		return errorWithCause(c, err)
	}
}

func (c *CliConn) ExecContext(ctx context.Context, query string, args ...any) (*CliResult, error) {
	stmt, err := c.NewStmt()
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.prepare(query); err != nil {
		return nil, err
	}
	if err := stmt.bindParams(args...); err != nil {
		return nil, err
	}
	if err := stmt.execute(); err != nil {
		return nil, err
	}

	return &CliResult{}, nil
}

func (c *CliConn) QueryRowContext(ctx context.Context, query string, args ...any) *CliRow {
	ret := &CliRow{env: c.env}

	stmt, err := c.NewStmt()
	if err != nil {
		ret.err = err
		return ret
	}
	defer stmt.Close()

	if err := stmt.prepare(query); err != nil {
		ret.err = err
		return ret
	}
	if err := stmt.bindParams(args...); err != nil {
		ret.err = err
		return ret
	}
	if err := stmt.execute(); err != nil {
		ret.err = err
		return ret
	}
	if values, _, err := stmt.fetch(); err != nil {
		ret.err = err
		return ret
	} else {
		ret.values = values
	}
	return ret
}

func (c *CliConn) QueryContext(ctx context.Context, query string, args ...any) (*CliRows, error) {
	stmt, err := c.NewStmt()
	if err != nil {
		return nil, err
	}
	if err := stmt.prepare(query); err != nil {
		stmt.Close()
		return nil, err
	}
	if err := stmt.bindParams(args...); err != nil {
		stmt.Close()
		return nil, err
	}
	if err := stmt.execute(); err != nil {
		stmt.Close()
		return nil, err
	}
	ret := &CliRows{
		stmt: stmt,
	}
	return ret, nil
}

func (stmt *CliStmt) prepare(query string) error {
	if err := cliPrepare(stmt.handle, query); err != nil {
		return errorWithCause(stmt, err)
	}
	return nil
}

func (stmt *CliStmt) bindParams(args ...any) error {
	numParam, err := cliNumParam(stmt.handle)
	if err != nil {
		return errorWithCause(stmt, err)
	}
	if len(args) != numParam {
		return ErrParamCount(numParam, len(args))
	}

	paramsDesc := make([]CliParamDesc, numParam)
	for i := 0; i < numParam; i++ {
		desc, err := cliDescribeParam(stmt.handle, i)
		if err != nil {
			return errorWithCause(stmt, err)
		}
		paramsDesc[i] = desc
	}

	if len(args) != numParam {
		return ErrParamCount(numParam, len(args))
	}

	for paramNo, pd := range paramsDesc {
		var value unsafe.Pointer
		var valueLen int
		var cType CType
		arg := args[paramNo]
		switch pd.Type {
		default:
			return ErrDatabaseBindUnknownType(paramNo, pd.Type)
		case MACHCLI_SQL_TYPE_INT16:
			switch val := arg.(type) {
			case int16:
				cType = MACHCLI_C_TYPE_INT16
				value = unsafe.Pointer(&val)
				valueLen = 2
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, arg)
			}
		case MACHCLI_SQL_TYPE_INT32:
			switch val := arg.(type) {
			case int32, int:
				cType = MACHCLI_C_TYPE_INT32
				value = unsafe.Pointer(&val)
				valueLen = 4
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_INT64:
			switch val := arg.(type) {
			case int64:
				cType = MACHCLI_C_TYPE_INT64
				value = unsafe.Pointer(&val)
				valueLen = 8
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_DATETIME:
			switch val := arg.(type) {
			case int64:
				cType = MACHCLI_C_TYPE_INT64
				value = unsafe.Pointer(&val)
				valueLen = 8
			case time.Time:
				cType = MACHCLI_C_TYPE_INT64
				v := val.UnixNano()
				value = unsafe.Pointer(&v)
				valueLen = 8
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_FLOAT:
			switch val := arg.(type) {
			case float32:
				cType = MACHCLI_C_TYPE_FLOAT
				value = unsafe.Pointer(&val)
				valueLen = 4
			case float64:
				cType = MACHCLI_C_TYPE_FLOAT
				v := float32(val)
				value = unsafe.Pointer(&v)
				valueLen = 4
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_DOUBLE:
			switch val := arg.(type) {
			case float32:
				cType = MACHCLI_C_TYPE_FLOAT
				value = unsafe.Pointer(&val)
				valueLen = 4
			case float64:
				cType = MACHCLI_C_TYPE_DOUBLE
				value = unsafe.Pointer(&val)
				valueLen = 8
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_IPV4:
			switch val := arg.(type) {
			case net.IP:
				cType = MACHCLI_C_TYPE_CHAR
				v := val.To4()
				value = unsafe.Pointer(&v[0])
				valueLen = 4
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_IPV6:
			switch val := arg.(type) {
			case net.IP:
				cType = MACHCLI_C_TYPE_CHAR
				v := val.To16()
				value = unsafe.Pointer(&v[0])
				valueLen = 16
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_STRING:
			switch val := arg.(type) {
			case string:
				cType = MACHCLI_C_TYPE_CHAR
				bstr := []byte(val)
				value = (unsafe.Pointer)(&bstr[0])
				valueLen = len(bstr)
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		case MACHCLI_SQL_TYPE_BINARY:
			switch val := arg.(type) {
			case []byte:
				cType = MACHCLI_C_TYPE_CHAR
				value = (unsafe.Pointer)(&val[0])
				valueLen = len(val)
			default:
				return ErrDatabaseBindWrongType(paramNo, pd.Type, value)
			}
		}

		if err := cliBindParam(stmt.handle, paramNo, cType, pd.Type, value, valueLen); err != nil {
			return errorWithCause(stmt, err)
		}
	}
	return nil
}

type CliResult struct {
	message string
	// rowsAffected int64
}

func (rs *CliResult) String() string {
	return rs.message
}

func (rs *CliResult) LastInsertId() (int64, error) {
	return 0, ErrNotImplemented("LastInsertId")
}

func (rs *CliResult) RowsAffected() (int64, error) {
	return 0, ErrNotImplemented("RowsAffected")
	// return rs.rowsAffected, nil
}

func (c *CliConn) PrepareContext(ctx context.Context, query string) (*CliStmt, error) {
	ret := &CliStmt{
		conn: c,
	}
	if err := cliAllocStmt(c.handle, &ret.handle); err == nil {
		return ret, nil
	} else {
		return nil, errorWithCause(c, err)
	}
}

func (c *CliConn) NewStmt() (*CliStmt, error) {
	ret := &CliStmt{}
	if err := cliAllocStmt(c.handle, &ret.handle); err == nil {
		ret.conn = c
		return ret, nil
	} else {
		return nil, errorWithCause(c, err)
	}
}

type CliStmt struct {
	handle      unsafe.Pointer
	conn        *CliConn
	columnDescs []CliColumnDesc
}

func (stmt *CliStmt) Close() error {
	if err := cliFreeStmt(stmt.handle); err == nil {
		return nil
	} else {
		return errorWithCause(stmt, err)
	}
}

func (stmt *CliStmt) Error() error {
	return errorWithCause(stmt, nil)
}

func (stmt *CliStmt) execute() error {
	if err := cliExecute(stmt.handle); err != nil {
		return errorWithCause(stmt, err)
	}
	num, err := cliNumResultCol(stmt.handle)
	if err != nil {
		return errorWithCause(stmt, err)
	}
	stmt.columnDescs = make([]CliColumnDesc, num)
	for i := 0; i < num; i++ {
		desc, err := cliDescribeCol(stmt.handle, i)
		if err != nil {
			return errorWithCause(stmt, err)
		}
		stmt.columnDescs[i] = desc
	}
	return nil
}

// fetch fetches the next row from the result set.
// It returns true if it reaches end of the result, otherwise false.
func (stmt *CliStmt) fetch() ([]any, bool, error) {
	end, err := cliFetch(stmt.handle)
	if err != nil {
		return nil, end, err
	}
	if end {
		return nil, true, nil
	}

	row := make([]any, len(stmt.columnDescs))
	for i, desc := range stmt.columnDescs {
		switch desc.Type {
		case MACHCLI_SQL_TYPE_INT16:
			var v = new(int16)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_INT16, unsafe.Pointer(v), 2); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = *v
			}
		case MACHCLI_SQL_TYPE_INT32:
			var v = new(int32)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_INT32, unsafe.Pointer(v), 4); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = *v
			}
		case MACHCLI_SQL_TYPE_INT64:
			var v = new(int64)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_INT64, unsafe.Pointer(v), 8); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = *v
			}
		case MACHCLI_SQL_TYPE_DATETIME:
			var v = new(int64)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_INT64, unsafe.Pointer(v), 8); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = time.Unix(0, *v)
			}
		case MACHCLI_SQL_TYPE_FLOAT:
			var v = new(float32)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_FLOAT, unsafe.Pointer(v), 4); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = *v
			}
		case MACHCLI_SQL_TYPE_DOUBLE:
			var v = new(float64)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_DOUBLE, unsafe.Pointer(v), 8); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = *v
			}
		case MACHCLI_SQL_TYPE_IPV4:
			var v = []byte{0, 0, 0, 0}
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&v[0]), 4); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = net.IP(v)
			}
		case MACHCLI_SQL_TYPE_IPV6:
			var v = make([]byte, 16)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&v[0]), 16); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = net.IP(v)
			}
		case MACHCLI_SQL_TYPE_STRING:
			var v = make([]byte, desc.Size)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&v[0]), desc.Size); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = string(v[0:n])
			}
		case MACHCLI_SQL_TYPE_BINARY:
			var v = make([]byte, desc.Size)
			if n, err := cliGetData(stmt.handle, i, MACHCLI_C_TYPE_CHAR, unsafe.Pointer(&v[0]), desc.Size); err != nil {
				return nil, end, errorWithCause(stmt, err)
			} else if n == -1 {
				row[i] = nil
			} else {
				row[i] = v[0:n]
			}
		}
	}
	return row, end, nil
}

type CliRow struct {
	env    *CliEnv
	err    error
	values []any
}

func (r *CliRow) Err() error {
	return r.err
}

func (r *CliRow) Scan(dest ...any) error {
	if len(dest) > len(r.values) {
		return ErrParamCount(len(r.values), len(dest))
	}
	for i, d := range dest {
		if d == nil {
			continue
		}
		if err := scanConvert(r.values[i], d, r.env); err != nil {
			return err
		}
	}
	return nil
}

type CliRows struct {
	stmt *CliStmt
	err  error
	row  []any
}

func (r *CliRows) Err() error {
	return r.err
}

func (r *CliRows) Close() error {
	return r.stmt.Close()
}

func (r *CliRows) Next() bool {
	row, end, err := r.stmt.fetch()
	if err != nil {
		r.err = err
		return false
	}
	if end {
		return false
	}
	r.row = row
	return true
}

func (r *CliRows) Row() []any {
	return r.row
}

func (r *CliRows) ColumnDescriptions() []CliColumnDesc {
	return r.stmt.columnDescs
}

func (r *CliRows) Scan(dest ...any) error {
	if len(dest) > len(r.row) {
		return ErrParamCount(len(r.row), len(dest))
	}
	for i, d := range dest {
		if d == nil {
			continue
		}
		if err := scanConvert(r.row[i], d, r.stmt.conn.env); err != nil {
			return err
		}
	}
	return nil
}

func scanConvert(src, dst any, env *CliEnv) error {
	if src == nil {
		dst = nil
		return nil
	}
	switch sv := src.(type) {
	case int16:
		switch dv := dst.(type) {
		case *int16:
			*dv = sv
			return nil
		case *int32:
			*dv = int32(sv)
			return nil
		case *int64:
			*dv = int64(sv)
			return nil
		case *int:
			*dv = int(sv)
			return nil
		}
	case int32:
		switch dv := dst.(type) {
		case *int16:
			*dv = int16(sv)
			return nil
		case *int32:
			*dv = sv
			return nil
		case *int64:
			*dv = int64(sv)
			return nil
		case *int:
			*dv = int(sv)
			return nil
		}
	case int64:
		switch dv := dst.(type) {
		case *int16:
			*dv = int16(sv)
			return nil
		case *int32:
			*dv = int32(sv)
			return nil
		case *int64:
			*dv = sv
			return nil
		case *int:
			*dv = int(sv)
			return nil
		}
	case float64:
		switch dv := dst.(type) {
		case *float32:
			*dv = float32(sv)
			return nil
		case *float64:
			*dv = sv
			return nil
		}
	case float32:
		switch dv := dst.(type) {
		case *float32:
			*dv = sv
			return nil
		case *float64:
			*dv = float64(sv)
			return nil
		}
	case string:
		switch dv := dst.(type) {
		case *string:
			*dv = src.(string)
			return nil
		}
	case time.Time:
		switch dv := dst.(type) {
		case *time.Time:
			*dv = sv
			return nil
		case *int64:
			switch env.timeformat {
			case "us":
				*dv = sv.UnixNano() / 1000
			case "ms":
				*dv = sv.UnixNano() / 1000_000
			case "s":
				*dv = sv.Unix()
			default:
				*dv = sv.UnixNano()
			}
			return nil
		case *string:
			switch env.timeformat {
			case "ns":
				*dv = strconv.FormatInt(sv.UnixNano(), 10)
			case "us":
				*dv = strconv.FormatInt(sv.UnixNano()/1000, 10)
			case "ms":
				*dv = strconv.FormatInt(sv.UnixNano()/1000_000, 10)
			case "s":
				*dv = strconv.FormatInt(sv.Unix(), 10)
			default:
				*dv = sv.In(env.tz).Format(env.timeformat)
			}
			return nil
		}
	case []byte:
		switch dv := dst.(type) {
		case *[]byte:
			*dv = src.([]byte)
			return nil
		}
	}
	return ErrCannotConvertValue(src, dst)
}

func (c *CliConn) Appender(tableName string, opts ...CliAppenderOption) (*CliAppender, error) {
	ret := &CliAppender{tableName: strings.ToUpper(tableName)}
	for _, opt := range opts {
		opt(ret)
	}

	stmt, err := c.NewStmt()
	if err != nil {
		return nil, err
	}
	ret.stmt = stmt

	if err := cliAppendOpen(stmt.handle, ret.tableName, ret.errCheckCount); err != nil {
		return nil, errorWithCause(stmt, err)
	}

	num, err := cliNumResultCol(stmt.handle)
	if err != nil {
		stmt.Close()
		return nil, err
	}

	// XXX
	// TODO temporary solution to skip the first column
	stmt.columnDescs = make([]CliColumnDesc, num-1)
	for i := 0; i < num; i++ {
		if i == 0 {
			continue
		}
		desc, err := cliDescribeCol(stmt.handle, i)
		if err != nil {
			err = errorWithCause(stmt, err)
			stmt.Close()
			return nil, err
		}
		stmt.columnDescs[i-1] = desc
	}

	return ret, nil
}

type CliAppender struct {
	stmt          *CliStmt
	tableName     string
	errCheckCount int
}

type CliAppenderOption func(*CliAppender)

func WithErrorCheckCount(count int) CliAppenderOption {
	return func(c *CliAppender) {
		c.errCheckCount = count
	}
}

// Close returns the number of success and fail rows.
func (a *CliAppender) Close() (int64, int64, error) {
	if success, fail, err := cliAppendClose(a.stmt.handle); err == nil {
		return success, fail, nil
	} else {
		c := a.stmt.conn
		if err := a.stmt.Close(); err != nil {
			return success, fail, errorWithCause(c, err)
		}
		return success, fail, errorWithCause(a.stmt, err)
	}
}

func (a *CliAppender) Flush() error {
	if err := cliAppendFlush(a.stmt.handle); err == nil {
		return nil
	} else {
		return errorWithCause(a.stmt, err)
	}
}

func (a *CliAppender) Append(args ...any) error {
	if err := cliAppendData(a.stmt.handle, a.stmt.columnDescs, args); err != nil {
		return errorWithCause(a.stmt, err)
	}
	return nil
}
