package mach

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/machbase/neo-server/spi"
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
	spi.RegisterFactory(FactoryName, func() (spi.Database, error) {
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

type Env struct {
	handle unsafe.Pointer
}

var singleton = Env{}

type Database struct {
	handle unsafe.Pointer
	idGen  *sonyflake.Sonyflake
	conns  cmap.ConcurrentMap[string, *ConnWatcher]
}

var _ spi.Database = &Database{}
var _ spi.Conn = &connection{}

// implements spi.DatabaseLife interface
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

// implements spi.DatabaseLife interface
func (db *Database) Shutdown() error {
	return shutdown0(db.handle)
}

func (db *Database) Error() error {
	return machError0(db.handle)
}

// implements spi.DatabaseAuth interface
func (db *Database) UserAuth(username, password string) (bool, error) {
	return machUserAuth(db.handle, username, password)
}

func (db *Database) RegisterWatcher(key string, conn *connection) {
	db.SetWatcher(key, &ConnWatcher{
		Key:     key,
		Created: time.Now(),
		conn:    conn,
	})
}

func (db *Database) SetWatcher(key string, cw *ConnWatcher) {
	db.conns.Set(key, cw)
}

func (db *Database) GetWatcher(key string) (*ConnWatcher, bool) {
	return db.conns.Get(key)
}

func (db *Database) RemoveWatcher(key string) {
	db.conns.Remove(key)
}

func (db *Database) ListWatcher(cb func(*ConnWatcher) bool) {
	if cb == nil {
		return
	}
	var cont = true
	db.conns.IterCb(func(_ string, v *ConnWatcher) {
		if !cont {
			return
		}
		cont = cb(v)
	})
}

type ConnWatcher struct {
	Key     string
	Created time.Time
	conn    *connection
}

func (cw *ConnWatcher) LatestSQL() string {
	return cw.conn.latestSQL
}

type connection struct {
	ctx         context.Context
	username    string
	password    string
	isTrustUser bool
	handle      unsafe.Pointer
	closeOnce   sync.Once
	closed      bool
	db          *Database

	latestSQL string
}

func WithPassword(username string, password string) spi.ConnectOption {
	return func(c spi.Conn) {
		c.(*connection).username = username
		c.(*connection).password = password
	}
}

func WithTrustUser(username string) spi.ConnectOption {
	return func(c spi.Conn) {
		c.(*connection).username = username
		c.(*connection).isTrustUser = true
	}
}

func (db *Database) Connect(ctx context.Context, opts ...spi.ConnectOption) (spi.Conn, error) {
	id, err := db.idGen.NextID()
	if err != nil {
		return nil, fmt.Errorf("connection id fail, %s", err.Error())
	}
	strId := fmt.Sprintf("%X", id)
	ret := &connection{
		ctx:       ctx,
		db:        db,
		latestSQL: "CONNECT",
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
	statz.AllocConn()
	if statz.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("Conn.Connect() from %s#%d\n", file, no)
		}
	}

	db.RegisterWatcher(strId, ret)
	return ret, nil
}

func (conn *connection) Close() (err error) {
	conn.latestSQL = "CLOSE"
	if statz.Debug {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("Conn.Close() from %s#%d\n", file, no)
		}
	}
	conn.closeOnce.Do(func() {
		conn.closed = true
		statz.FreeConn()
		err = machDisconnect(conn.handle)
	})
	return
}

func (conn *connection) Connected() bool {
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

func (conn *connection) Ping() (time.Duration, error) {
	return 0, nil
}

func (conn *connection) Exec(ctx context.Context, sqlText string, params ...any) spi.Result {
	conn.latestSQL = sqlText
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

func (conn *connection) Query(ctx context.Context, sqlText string, params ...any) (spi.Rows, error) {
	conn.latestSQL = sqlText
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

func (conn *connection) QueryRow(ctx context.Context, sqlText string, params ...any) spi.Row {
	conn.latestSQL = sqlText
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
			row.err = fmt.Errorf("QueryRow unsupported type %d", typ)
		}
	}
	row.err = scan(stmt, row.values...)
	if row.err == nil {
		row.ok = true
	}
	return row
}

func (conn *connection) Explain(ctx context.Context, sqlText string, full bool) (string, error) {
	conn.latestSQL = "EXPLAIN " + sqlText
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

var startupTime = time.Now()
var BuildVersion Version
var ServicePorts map[string][]*ServicePort

func (db *Database) GetServerInfo() (*ServerInfo, error) {
	rsp := &ServerInfo{}

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)

	rsp.Version = Version{
		Engine:         LinkInfo(),
		Major:          BuildVersion.Major,
		Minor:          BuildVersion.Minor,
		Patch:          BuildVersion.Patch,
		GitSHA:         BuildVersion.GitSHA,
		BuildTimestamp: BuildVersion.BuildTimestamp,
		BuildCompiler:  BuildVersion.BuildCompiler,
	}

	rsp.Runtime = Runtime{
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		Pid:            int32(os.Getpid()),
		UptimeInSecond: int64(time.Since(startupTime).Seconds()),
		Processes:      int32(runtime.GOMAXPROCS(-1)),
		Goroutines:     int32(runtime.NumGoroutine()),
		MemSys:         mem.Sys,
		MemHeapSys:     mem.HeapSys,
		MemHeapAlloc:   mem.HeapAlloc,
		MemHeapInUse:   mem.HeapInuse,
		MemStackSys:    mem.StackSys,
		MemStackInUse:  mem.StackInuse,
	}

	return rsp, nil
}

type ServerInfo struct {
	Version Version
	Runtime Runtime
}

type Version struct {
	Major          int32
	Minor          int32
	Patch          int32
	GitSHA         string
	BuildTimestamp string
	BuildCompiler  string
	Engine         string
}

type Runtime struct {
	OS             string
	Arch           string
	Pid            int32
	UptimeInSecond int64
	Processes      int32
	Goroutines     int32
	MemSys         uint64
	MemHeapSys     uint64
	MemHeapAlloc   uint64
	MemHeapInUse   uint64
	MemStackSys    uint64
	MemStackInUse  uint64
}

func (db *Database) GetServicePorts(svc string) ([]*ServicePort, error) {
	ports := []*ServicePort{}
	for k, s := range ServicePorts {
		if len(svc) > 0 {
			if strings.ToLower(svc) != k {
				continue
			}
		}
		ports = append(ports, s...)
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Service == ports[j].Service {
			return ports[i].Address < ports[j].Address
		}
		return ports[i].Service < ports[j].Service
	})
	return ports, nil
}

type ServicePort struct {
	Service string
	Address string
}

type Statz struct {
	Conns          int64
	Stmts          int64
	Appenders      int64
	ConnsInUse     int32
	StmtsInUse     int32
	AppendersInUse int32
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

func StatzSnapshot() map[string]any {
	ret := map[string]any{
		"conns":          statz.ConnsInUse,
		"conns_used":     statz.Conns,
		"stmts":          statz.StmtsInUse,
		"stmts_used":     statz.Stmts,
		"appenders":      statz.AppendersInUse,
		"appenders_used": statz.Appenders,
	}
	if singleton.handle != nil {
		ret["conns_raw"] = machConnectionCount(singleton.handle)
	}
	return ret
}
