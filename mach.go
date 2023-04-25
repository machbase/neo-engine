package mach

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
	"unsafe"

	spi "github.com/machbase/neo-spi"
)

type InitOption int

const (
	OPT_NONE               InitOption = 0x0
	OPT_SIGHANDLER_DISABLE InitOption = 0x1
)

const FactoryName = "machbase-engine"

func Initialize(homeDir string) error {
	return InitializeOption(homeDir, OPT_SIGHANDLER_DISABLE)
}

func InitializeOption(homeDir string, opt InitOption) error {
	var handle unsafe.Pointer
	err := initialize0(homeDir, int(opt), &handle)
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

// implements spi.DatabaseLife interface
func (db *database) Startup() error {
	// machbase startup 과정에서 현재 디렉터리를 HOME으로 변경하는데,
	// application의 Working directory를 유지하기 위해 chdir()을 호출한다.
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

func (db *database) Explain(sqlText string, full bool) (string, error) {
	var stmt unsafe.Pointer
	if err := machAllocStmt(db.handle, &stmt); err != nil {
		return "", err
	}
	defer machFreeStmt(db.handle, stmt)
	if err := machPrepare(stmt, sqlText); err != nil {
		return "", err
	}
	return machExplain(stmt, full)
}

func (db *database) ExecContext(ctx context.Context, sqlText string, params ...any) spi.Result {
	// TODO apply context
	return db.Exec(sqlText, params...)
}

func (db *database) Exec(sqlText string, params ...any) spi.Result {
	var result = &Result{}

	var stmt unsafe.Pointer
	if err := machAllocStmt(db.handle, &stmt); err != nil {
		result.err = err
		return result
	}
	defer machFreeStmt(db.handle, stmt)
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

func (db *database) QueryContext(ctx context.Context, sqlText string, params ...any) (spi.Rows, error) {
	return db.Query(sqlText, params...)
}

func (db *database) Query(sqlText string, params ...any) (spi.Rows, error) {
	rows := &Rows{
		handle:  db.handle,
		sqlText: sqlText,
	}
	if err := machAllocStmt(db.handle, &rows.stmt); err != nil {
		return nil, err
	}
	if err := machPrepare(rows.stmt, sqlText); err != nil {
		return nil, err
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

func (db *database) QueryRowContext(ctx context.Context, sqlText string, params ...any) spi.Row {
	return db.QueryRow(sqlText, params...)
}

func (db *database) QueryRow(sqlText string, params ...any) spi.Row {
	var row = &Row{}

	var stmt unsafe.Pointer
	if row.err = machAllocStmt(db.handle, &stmt); row.err != nil {
		return row
	}
	defer func() {
		err := machFreeStmt(db.handle, stmt)
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

var startupTime = time.Now()
var BuildVersion spi.Version

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
