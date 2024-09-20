package mach

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/sony/sonyflake"
)

type InitOption int

const (
	// machbase-engine takes all control of the signals
	OPT_SIGHANDLER_ON InitOption = 0x0
	// the caller takes all control, machbase-engine can not leave stack dump when the process crashed
	OPT_SIGHANDLER_OFF InitOption = 0x1
	// engine takes all control except SIGINT, so that the caller can take SIGINT control
	OPT_SIGHANDLER_SIGINT_OFF InitOption = 0x2
)

const FactoryName = "machbase-engine"

func Initialize(homeDir string, machPort int) error {
	return InitializeOption(homeDir, machPort, OPT_SIGHANDLER_OFF)
}

func InitializeOption(homeDir string, machPort int, opt InitOption) error {
	homeDir = translateCodePage(homeDir)
	var handle unsafe.Pointer
	err := initialize0(homeDir, machPort, int(opt), &handle)
	if err != nil {
		return err
	}
	singleton.handle = handle
	RegisterFactory(FactoryName, func() (*Database, error) {
		return &Database{
			handle: singleton.handle,
			conns:  cmap.New[*ConnWatcher](),
			idGen:  sonyflake.NewSonyflake(sonyflake.Settings{}),
		}, nil
	})
	return nil
}

func Finalize() {
	finalize0(singleton.handle)
}

func DestroyDatabase() error {
	return destroyDatabase0(singleton.handle)
}

func CreateDatabase() error {
	return createDatabase0(singleton.handle)
}

func ExistsDatabase() bool {
	return existsDatabase0(singleton.handle)
}

func RestoreDatabase(path string) error {
	return restoreDatabase0(singleton.handle, path)
}

type Env struct {
	handle unsafe.Pointer
}

var singleton = Env{}

type Database struct {
	handle unsafe.Pointer
	idGen  *sonyflake.Sonyflake
	conns  cmap.ConcurrentMap[string, *ConnWatcher]
}

func (db *Database) Startup() error {
	// machbase change the current dir to $HOME during startup process.
	// Call chdir() for keeping the working dir of the application.
	cwd, _ := os.Getwd()
	defer func() {
		os.Chdir(cwd)
	}()

	err := startup0(db.handle)
	return err
}

func (db *Database) Shutdown() error {
	return shutdown0(db.handle)
}

func (db *Database) Error() error {
	return machError0(db.handle)
}

func (db *Database) UserAuth(username, password string) (bool, error) {
	return machUserAuth(db.handle, username, password)
}

func (db *Database) RegisterWatcher(key string, conn *Conn) {
	db.SetWatcher(key, &ConnWatcher{
		createdTime: time.Now(),
		conn:        conn,
	})
}

func (db *Database) SetWatcher(key string, cw *ConnWatcher) {
	db.conns.Set(key, cw)
}

func (db *Database) GetWatcher(key string) (*ConnState, bool) {
	w, ok := db.conns.Get(key)
	if ok {
		return w.ConnState(), true
	} else {
		return nil, false
	}
}

func (db *Database) RemoveWatcher(key string) {
	db.conns.Remove(key)
}

func (db *Database) ListWatcher(cb func(*ConnState) bool) {
	if cb == nil {
		return
	}
	var cont = true
	db.conns.IterCb(func(_ string, cw *ConnWatcher) {
		if !cont {
			return
		}
		v := cw.ConnState()
		cont = cb(v)
	})
}

func (db *Database) KillConnection(id string, force bool) error {
	if cw, ok := db.conns.Get(id); ok {
		if cw.conn == nil {
			return ErrDatabaseConnectionInvalid(id)
		}
		if force {
			return cw.conn.Close()
		} else {
			return cw.conn.Cancel()
		}
	} else {
		return ErrDatabaseConnectionNotFound(id)
	}
}

type ConnWatcher struct {
	createdTime time.Time
	conn        *Conn
}

type ConnState struct {
	Id          string
	CreatedTime time.Time
	LatestTime  time.Time
	LatestSql   string
}

func (cw *ConnWatcher) ConnState() *ConnState {
	ret := &ConnState{
		CreatedTime: cw.createdTime,
	}
	if cw.conn != nil {
		ret.Id = cw.conn.id
		ret.LatestTime = cw.conn.latestTime
		ret.LatestSql = cw.conn.latestSql
	}
	return ret
}

type Conn struct {
	ctx         context.Context
	username    string
	password    string
	isTrustUser bool
	handle      unsafe.Pointer
	closeOnce   sync.Once
	closed      bool
	db          *Database

	id            string
	latestTime    time.Time
	latestSql     string
	closeCallback func()
}

func (conn *Conn) SetLatestSql(sql string) {
	conn.latestTime = time.Now()
	conn.latestSql = sql
}

type ConnectOption func(*Conn)

func WithPassword(username string, password string) ConnectOption {
	return func(c *Conn) {
		c.username = username
		c.password = password
	}
}

func WithTrustUser(username string) ConnectOption {
	return func(c *Conn) {
		c.username = username
		c.isTrustUser = true
	}
}

func (db *Database) Connect(ctx context.Context, opts ...ConnectOption) (*Conn, error) {
	ret := &Conn{
		ctx: ctx,
		db:  db,
	}
	for _, o := range opts {
		o(ret)
	}
	var handle unsafe.Pointer
	if ret.isTrustUser {
		if err := machConnectTrust(db.handle, ret.username, &handle); err != nil {
			return nil, err
		}
	} else {
		if err := machConnect(db.handle, ret.username, ret.password, &handle); err != nil {
			return nil, err
		}
	}
	ret.handle = handle

	if id, err := machSessionID(ret.handle); err == nil {
		ret.id = fmt.Sprintf("%d", id)
	} else {
		id, err := db.idGen.NextID()
		if err != nil {
			return nil, ErrDatabaseConnectID(err.Error())
		}
		ret.id = fmt.Sprintf("%X", id)
	}

	statz.AllocConn()
	if statz.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("Conn.Connect() from %s#%d\n", file, no)
		}
	}
	ret.closeCallback = func() {
		ret.SetLatestSql("CLOSE") // 3. set latest sql time
		db.RemoveWatcher(ret.id)
	}
	db.RegisterWatcher(ret.id, ret) // 1. set creTime
	ret.SetLatestSql("CONNECT")     // 2. set latest sql time
	return ret, nil
}

// Close closes connection
func (conn *Conn) Close() (err error) {
	if statz.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("Conn.Close() from %s#%d\n", file, no)
		}
	}
	conn.closeOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in Conn.Close", r)
			}
		}()
		conn.closed = true
		statz.FreeConn()
		err = machDisconnect(conn.handle)
		if conn.closeCallback != nil {
			conn.closeCallback()
		}
	})
	return
}

func (conn *Conn) Cancel() error {
	if err := machCancel(conn.handle); err != nil {
		return err
	}
	return conn.Close()
}

func (conn *Conn) Connected() bool {
	if conn.closed {
		return false
	}
	if len(conn.ctx.Done()) != 0 {
		<-conn.ctx.Done()
		conn.Close()
		return false
	}
	return true
}

func (conn *Conn) Ping() (time.Duration, error) {
	return 0, nil
}

// ExecContext executes SQL statements that does not return result
// like 'ALTER', 'CREATE TABLE', 'DROP TABLE', ...
func (conn *Conn) Exec(ctx context.Context, sqlText string, params ...any) *Result {
	conn.SetLatestSql(sqlText)
	var result = &Result{}
	var stmt unsafe.Pointer
	if err := machAllocStmt(conn.handle, &stmt); err != nil {
		result.err = err
		return result
	}
	statz.AllocStmt()
	defer func() {
		machFreeStmt(stmt)
		statz.FreeStmt()
	}()
	if len(params) == 0 {
		if err := machDirectExecute(stmt, sqlText); err != nil {
			result.err = err
			return result
		}
	} else {
		err := machPrepare(stmt, sqlText)
		if err != nil {
			result.err = err
			return result
		}
		for i, p := range params {
			if err := bind(stmt, i, p); err != nil {
				result.err = err
				return result
			}
		}
		err = machExecute(stmt)
		if err != nil {
			result.err = err
			return result
		}
	}
	affectedRows, err := machEffectRows(stmt)
	if err != nil {
		result.err = err
		return result
	}
	stmtType, err := machStmtType(stmt)
	result.affectedRows = affectedRows
	result.stmtType = stmtType
	result.err = err
	return result
}

// Query executes SQL statements that are expected multiple rows as result.
// Commonly used to execute 'SELECT * FROM <TABLE>'
//
// Rows returned by Query() must be closed to prevent server-side-resource leaks.
//
//	ctx, cancelFunc := context.WithTimeout(5*time.Second)
//	defer cancelFunc()
//
//	rows, err := conn.Query(ctx, "select * from my_table where name = ?", my_name)
//	if err != nil {
//		panic(err)
//	}
//	defer rows.Close()
func (conn *Conn) Query(ctx context.Context, sqlText string, params ...any) (*Rows, error) {
	conn.SetLatestSql(sqlText)
	rows := &Rows{
		sqlText: sqlText,
	}
	if err := machAllocStmt(conn.handle, &rows.stmt); err != nil {
		return nil, err
	}
	if err := machPrepare(rows.stmt, sqlText); err != nil {
		machFreeStmt(rows.stmt)
		return nil, err
	}
	for i, p := range params {
		if err := bind(rows.stmt, i, p); err != nil {
			machFreeStmt(rows.stmt)
			return nil, err
		}
	}
	if err := machExecute(rows.stmt); err != nil {
		machFreeStmt(rows.stmt)
		return nil, err
	}
	if stmtType, err := machStmtType(rows.stmt); err != nil {
		machFreeStmt(rows.stmt)
		return nil, err
	} else {
		rows.stmtType = stmtType
	}
	statz.AllocStmt()
	return rows, nil
}

// QueryRow executes a SQL statement that expects a single row result.
//
//	ctx, cancelFunc := context.WithTimeout(5*time.Second)
//	defer cancelFunc()
//
//	var cnt int
//	row := conn.QueryRow(ctx, "select count(*) from my_table where name = ?", "my_name")
//	row.Scan(&cnt)
func (conn *Conn) QueryRow(ctx context.Context, sqlText string, params ...any) *Row {
	conn.SetLatestSql(sqlText)
	var row = &Row{}
	var stmt unsafe.Pointer
	statz.AllocStmt()
	if row.err = machAllocStmt(conn.handle, &stmt); row.err != nil {
		return row
	}
	defer func() {
		statz.FreeStmt()
		err := machFreeStmt(stmt)
		if err != nil && row.err == nil {
			row.err = err
		}
	}()

	if row.err = machPrepare(stmt, sqlText); row.err != nil {
		return row
	}
	for i, p := range params {
		if row.err = bind(stmt, i, p); row.err != nil {
			return row
		}
	}
	if row.err = machExecute(stmt); row.err != nil {
		return row
	}

	if typ, err := machStmtType(stmt); err != nil {
		row.err = err
		return row
	} else {
		row.stmtType = typ
	}

	// Do not proceed if the statement is not a SELECT
	if !row.stmtType.IsSelect() {
		affectedRows, err := machEffectRows(stmt)
		if err != nil {
			row.err = err
			return row
		}
		row.affectedRows = affectedRows
		row.ok = true
		return row
	}

	var fetched bool
	if fetched, row.err = machFetch(stmt); row.err != nil {
		// fetch error
		return row
	}

	// nothing fetched
	if !fetched {
		row.err = sql.ErrNoRows
		return row
	}

	var count int
	count, row.err = machColumnCount(stmt)
	if row.err != nil {
		return row
	}
	if count == 0 {
		row.err = sql.ErrNoRows
		return row
	}
	row.values = make([]any, count)
	for i := 0; i < count; i++ {
		typ, siz, err := machColumnType(stmt, i)
		if err != nil {
			row.err = err
			return row
		}
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			row.values[i] = new(int16)
		case 1: // MACH_DATA_TYPE_INT32
			row.values[i] = new(int32)
		case 2: // MACH_DATA_TYPE_INT64
			row.values[i] = new(int64)
		case 3: // MACH_DATA_TYPE_DATETIME
			row.values[i] = new(time.Time)
		case 4: // MACH_DATA_TYPE_FLOAT
			row.values[i] = new(float32)
		case 5: // MACH_DATA_TYPE_DOUBLE
			row.values[i] = new(float64)
		case 6: // MACH_DATA_TYPE_IPV4
			row.values[i] = new(net.IP)
		case 7: // MACH_DATA_TYPE_IPV6
			row.values[i] = new(net.IP)
		case 8: // MACH_DATA_TYPE_STRING
			row.values[i] = new(string)
		case 9: // MACH_DATA_TYPE_BINARY
			row.values[i] = make([]byte, siz)
		default:
			row.err = ErrDatabaseUnsupportedType("QueryRow", int(typ))
		}
	}
	row.err = scan(stmt, row.values...)
	if row.err == nil {
		row.ok = true
	}
	return row
}

func (conn *Conn) Explain(ctx context.Context, sqlText string, full bool) (string, error) {
	conn.SetLatestSql("EXPLAIN " + sqlText)
	var stmt unsafe.Pointer
	if err := machAllocStmt(conn.handle, &stmt); err != nil {
		return "", err
	}
	defer machFreeStmt(stmt)

	if full {
		if err := machDirectExecute(stmt, sqlText); err != nil {
			return "", err
		}
	} else {
		if err := machPrepare(stmt, sqlText); err != nil {
			return "", err
		}
	}
	return machExplain(stmt, full)
}

type Statz struct {
	Conns          int64
	Stmts          int64
	Appenders      int64
	ConnsInUse     int32
	StmtsInUse     int32
	AppendersInUse int32
	RawConns       int32
	Debug          bool
}

var statz Statz

func (s *Statz) AllocConn() {
	atomic.AddInt32(&s.ConnsInUse, 1)
	atomic.AddInt64(&s.Conns, 1)
}

func (s *Statz) FreeConn() {
	atomic.AddInt32(&s.ConnsInUse, -1)
}

func (s *Statz) AllocStmt() {
	atomic.AddInt32(&s.StmtsInUse, 1)
	atomic.AddInt64(&s.Stmts, 1)
}

func (s *Statz) FreeStmt() {
	atomic.AddInt32(&s.StmtsInUse, -1)
}

func (s *Statz) AllocAppender() {
	atomic.AddInt32(&s.AppendersInUse, 1)
	atomic.AddInt64(&s.Appenders, 1)
}

func (s *Statz) FreeAppender() {
	atomic.AddInt32(&s.AppendersInUse, -1)
}

func StatzDebug(flag bool) {
	statz.Debug = flag
}

func StatzSnapshot() *Statz {
	ret := statz
	if singleton.handle != nil {
		ret.RawConns = int32(machConnectionCount(singleton.handle))
	}
	return &ret
}
