package mach

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

func LinkInfo() string {
	return LibMachLinkInfo
}

func Edition() string {
	if strings.Contains(LibMachLinkInfo, "edge") {
		return "edge"
	} else if strings.Contains(LibMachLinkInfo, "fog") {
		return "fog"
	} else {
		return "none"
	}
}

type InitOption int

const (
	OPT_NONE               InitOption = 0x0
	OPT_SIGHANDLER_DISABLE InitOption = 0x1
)

func Initialize(homeDir string) error {
	return InitializeOption(homeDir, OPT_SIGHANDLER_DISABLE)
}

func InitializeOption(homeDir string, opt InitOption) error {
	var handle unsafe.Pointer
	err := initialize0(homeDir, int(opt), &handle)
	if err == nil {
		singleton.handle = handle
	}
	return err
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
}

func New() *Database {
	return &Database{
		handle: singleton.handle,
	}
}

func (db *Database) Startup(timeout time.Duration) error {
	// machbase startup 과정에서 현재 디렉터리를 HOME으로 변경하는데,
	// application의 Working directory를 유지하기 위해 chdir()을 호출한다.
	cwd, _ := os.Getwd()
	defer func() {
		os.Chdir(cwd)
	}()

	err := startup0(db.handle, timeout)
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

func (db *Database) SqlTidy(sqlText string) string {
	lines := strings.Split(sqlText, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimSpace(ln)
	}
	return strings.TrimSpace(strings.Join(lines, " "))
}

func (db *Database) Exec(sqlText string, params ...any) (int64, error) {
	var stmt unsafe.Pointer
	if err := machAllocStmt(db.handle, &stmt); err != nil {
		return 0, err
	}
	defer machFreeStmt(db.handle, stmt)
	if len(params) == 0 {
		if err := machDirectExecute(stmt, sqlText); err != nil {
			return 0, err
		}
	} else {
		err := machPrepare(stmt, sqlText)
		if err != nil {
			return 0, err
		}
		for i, p := range params {
			if err := bind(stmt, i, p); err != nil {
				return 0, err
			}
		}
		err = machExecute(stmt)
		if err != nil {
			return 0, err
		}
	}
	return machEffectRows(stmt)
}

func (db *Database) Query(sqlText string, params ...any) (*Rows, error) {
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

func (db *Database) QueryRow(sqlText string, params ...any) *Row {
	var row = &Row{}

	var stmt unsafe.Pointer
	if row.err = machAllocStmt(db.handle, &stmt); row.err != nil {
		return row
	}
	defer func() {
		row.err = machFreeStmt(db.handle, stmt)
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

	typ, err := machStmtType(stmt)
	if err != nil {
		row.err = err
		return row
	}

	row.ok = true

	// select 가 아니면 fetch를 진행하지 않는다.
	if !isFetchableStmtType(typ) {
		return row
	}

	if _, row.err = machFetch(stmt); row.err != nil {
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
	return row
}

func (db *Database) Appender(tableName string) (*Appender, error) {
	appender := &Appender{}
	appender.handle = db.handle
	appender.tableName = strings.ToUpper(tableName)
	if err := machAllocStmt(db.handle, &appender.stmt); err != nil {
		return nil, err
	}
	if err := machAppendOpen(appender.stmt, tableName); err != nil {
		return nil, err
	}

	row := db.QueryRow("select type from M$SYS_TABLES where name = ?", appender.tableName)
	var typ int32 = -1
	if err := row.Scan(&typ); err != nil {
		return nil, err
	}
	appender.tableType = TableType(typ)

	var err error
	appender.colCount, err = machColumnCount(appender.stmt)
	if err != nil {
		return nil, err
	}
	appender.colTypes = make([]ColumnType, appender.colCount)
	appender.colSizes = make([]ColumnSize, appender.colCount)
	for i := 0; i < appender.colCount; i++ {
		typ, siz, err := machColumnType(appender.stmt, i)
		if err != nil {
			return nil, err
		}
		appender.colTypes[i] = typ
		appender.colSizes[i] = siz
	}
	return appender, nil
}

type Appender struct {
	handle       unsafe.Pointer
	stmt         unsafe.Pointer
	tableName    string
	tableType    TableType
	colCount     int
	colTypes     []ColumnType
	colSizes     []ColumnSize
	SuccessCount uint64
	FailureCount uint64
}

func (ap *Appender) Table() string {
	return ap.tableName
}

func (ap *Appender) Close() error {
	if ap.stmt == nil {
		return nil
	}
	s, f, err := machAppendClose(ap.stmt)
	if err != nil {
		return err
	}
	ap.SuccessCount = s
	ap.FailureCount = f

	if err := machFreeStmt(ap.handle, ap.stmt); err != nil {
		return err
	}
	ap.stmt = nil
	return nil
}

func (ap *Appender) matchColumnTypes(cols []any) ([]any, error) {
	// TODO JSON encoding으로 전달받은 데이터를 append에 적용하려면
	// float, int 형이 모두 number로 표현되는 json의 한계로 인해
	// 반드시 table schema를 먼저 확인해서 적절한 type으로 변경하는 과정이
	// 필요하다.
	if ap.tableType == TagTableType {
		var ok = false
		var err error
		// tag name
		if cols[0], ok = cols[0].(string); !ok {
			return cols, errors.New("first value of tuple should be tag name")
		}
		// time
		switch v := cols[1].(type) {
		case float32:
			cols[1] = int64(v)
		case float64:
			cols[1] = int64(v)
		case string:
			cols[1], err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return cols, errors.Wrap(err, "time conversion")
			}
		}
	}
	return cols, nil
}

func (ap *Appender) Append(cols ...any) error {
	cols, err := ap.matchColumnTypes(cols)
	if err != nil {
		return err
	}
	if ap.tableType == TagTableType {
		return ap.appendTagTable(cols)
	} else {
		return ap.appendLogTable(time.Time{}, cols)
	}
}

// supports only Log Table
func (ap *Appender) AppendWithTimestamp(ts time.Time, cols ...any) error {
	if ap.tableType == 0 {
		return ap.appendLogTable(ts, cols)
	} else {
		return fmt.Errorf("%s is not a log table, use Append() instead", ap.tableName)
	}
}

func (ap *Appender) appendLogTable(ts time.Time, cols []any) error {
	if ap.colCount-1 != len(cols) {
		return fmt.Errorf("value count %d, table '%s' (type %s) has %d columns", len(cols), ap.tableName, ap.tableType, ap.colCount-1)
	}
	vals := make([]*machAppendDataNullValue, ap.colCount)
	if ts.IsZero() {
		vals[0] = bindValue(nil)
	} else {
		vals[0] = bindValue(ts)
	}
	for i, c := range cols {
		vals[i+1] = bindValue(c)
	}
	defer func() {
		for _, v := range vals {
			v.Free()
		}
	}()
	if err := machAppendData(ap.stmt, vals); err != nil {
		return err
	}
	return nil
}

func (ap *Appender) appendTagTable(cols []any) error {
	if ap.colCount == 0 {
		return fmt.Errorf("table '%s' has no columns", ap.tableName)
	}
	if ap.colCount != len(cols) {
		return fmt.Errorf("value count %d, table '%s' has %d columns", len(cols), ap.tableName, ap.colCount)
	}
	vals := make([]*machAppendDataNullValue, ap.colCount)
	for i, c := range cols {
		vals[i] = bindValue(c)
	}
	defer func() {
		for _, v := range vals {
			v.Free()
		}
	}()
	if err := machAppendData(ap.stmt, vals); err != nil {
		return err
	}
	return nil
}
