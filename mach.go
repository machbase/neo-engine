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
	"time"
	"unsafe"

	spi "github.com/machbase/neo-spi"
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
		return &database{handle: singleton.handle}, nil
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

type database struct {
	handle unsafe.Pointer
}

var _ spi.Database = &database{}
var _ spi.DatabaseServer = &database{}
var _ spi.DatabaseAuth = &database{}
var _ spi.DatabaseAux = &database{}
var _ spi.Conn = &connection{}
var _ spi.Explainer = &connection{}

// implements spi.DatabaseLife interface
func (db *database) Startup() error {
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
func (db *database) Shutdown() error {
	return shutdown0(db.handle)
}

func (db *database) Error() error {
	return machError0(db.handle)
}

// implements spi.DatabaseAuth interface
func (db *database) UserAuth(username, password string) (bool, error) {
	return machUserAuth(db.handle, username, password)
}

type connection struct {
	ctx         context.Context
	username    string
	password    string
	isTrustUser bool
	handle      unsafe.Pointer
	closeOnce   sync.Once
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

func (db *database) Connect(ctx context.Context, opts ...spi.ConnectOption) (spi.Conn, error) {
	ret := &connection{ctx: ctx}
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
	return ret, nil
}

func (conn *connection) Close() (err error) {
	conn.closeOnce.Do(func() {
		err = machDisconnect(conn.handle)
	})
	return
}

func (conn *connection) Ping() (time.Duration, error) {
	return 0, nil
}

func (conn *connection) Exec(ctx context.Context, sqlText string, params ...any) spi.Result {
	var result = &Result{}
	var stmt unsafe.Pointer
	if err := machAllocStmt(conn.handle, &stmt); err != nil {
		result.err = err
		return result
	}
	defer machFreeStmt(stmt)
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
	rows := &Rows{
		sqlText: sqlText,
	}
	if err := machAllocStmt(conn.handle, &rows.stmt); err != nil {
		return nil, err
	}
	if err := machPrepare(rows.stmt, sqlText); err != nil {
		return nil, err
	}
	if DefaultDetective != nil {
		DefaultDetective.EnlistDetective(rows, sqlText)
	}
	for i, p := range params {
		if err := bind(rows.stmt, i, p); err != nil {
			return nil, err
		}
	}
	if err := machExecute(rows.stmt); err != nil {
		return nil, err
	}
	if stmtType, err := machStmtType(rows.stmt); err != nil {
		return nil, err
	} else {
		rows.stmtType = stmtType
	}
	return rows, nil
}

func (conn *connection) QueryRow(ctx context.Context, sqlText string, params ...any) spi.Row {
	var row = &Row{}
	var stmt unsafe.Pointer
	if row.err = machAllocStmt(conn.handle, &stmt); row.err != nil {
		return row
	}
	defer func() {
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
var BuildVersion spi.Version
var ServicePorts map[string][]*spi.ServicePort

func (db *database) GetServerInfo() (*spi.ServerInfo, error) {
	rsp := &spi.ServerInfo{}

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)

	rsp.Version = spi.Version{
		Engine:         LinkInfo(),
		Major:          BuildVersion.Major,
		Minor:          BuildVersion.Minor,
		Patch:          BuildVersion.Patch,
		GitSHA:         BuildVersion.GitSHA,
		BuildTimestamp: BuildVersion.BuildTimestamp,
		BuildCompiler:  BuildVersion.BuildCompiler,
	}

	rsp.Runtime = spi.Runtime{
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

func (db *database) GetInflights() ([]*spi.Inflight, error) {
	return DefaultDetective.InflightsDetective(), nil
}

func (db *database) GetPostflights() ([]*spi.Postflight, error) {
	return DefaultDetective.PostflightsDetective(), nil
}

func (db *database) GetServicePorts(svc string) ([]*spi.ServicePort, error) {
	ports := []*spi.ServicePort{}
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

var DefaultDetective Detective

type Detective interface {
	EnlistDetective(obj any, sqlTextOrTableName string)
	DelistDetective(any)
	UpdateDetective(any)
	InflightsDetective() []*spi.Inflight
	PostflightsDetective() []*spi.Postflight
}
