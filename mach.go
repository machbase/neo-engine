package mach

import (
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"
	"unsafe"
)

func LinkInfo() string {
	return LibMachLinkInfo
}

func Initialize(homeDir string) error {
	return initialize0(homeDir)
}

func DestroyDatabase() error {
	return destroyDatabase0()
}

func CreateDatabase() error {
	return createDatabase0()
}

func ExistsDatabase() bool {
	return existsDatabase0()
}

type Database struct {
	handle unsafe.Pointer
}

func New() *Database {
	return &Database{}
}

func (this *Database) Startup(timeout time.Duration) error {
	return startup0(&this.handle, timeout)
}

func (this *Database) Shutdown() error {
	return shutdown0(this.handle)
}

func (this *Database) Error() error {
	return machError0(this.handle)
}

func (this *Database) SqlTidy(sqlText string) string {
	lines := strings.Split(sqlText, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimSpace(ln)
	}
	return strings.TrimSpace(strings.Join(lines, " "))
}

func (this *Database) Exec(sqlText string, params ...any) error {
	var stmt unsafe.Pointer
	if err := machAllocStmt(this.handle, &stmt); err != nil {
		return err
	}
	defer machFreeStmt(stmt)
	if len(params) == 0 {
		if err := machDirectExecute(stmt, sqlText); err != nil {
			return err
		}
	} else {
		err := machPrepare(stmt, sqlText)
		if err != nil {
			return err
		}
		for i, p := range params {
			if err := bind(stmt, i, p); err != nil {
				return err
			}
		}
		err = machExecute(stmt)
	}
	return nil
}

func (this *Database) Query(sqlText string, params ...any) (*Rows, error) {
	rows := &Rows{
		sqlText: sqlText,
	}
	if err := machAllocStmt(this.handle, &rows.stmt); err != nil {
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
	return rows, nil
}

func (this *Database) QueryRow(sqlText string, params ...any) *Row {
	var row = &Row{}

	var stmt unsafe.Pointer
	if row.err = machAllocStmt(this.handle, &stmt); row.err != nil {
		return row
	}
	defer func() {
		row.err = machFreeStmt(stmt)
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

	if _, row.err = machFetch(stmt); row.err != nil {
		return row
	}
	row.ok = true

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

func (this *Database) Appender(tableName string) (*Appender, error) {
	appender := &Appender{}
	appender.tableName = strings.ToUpper(tableName)
	if err := machAllocStmt(this.handle, &appender.stmt); err != nil {
		return nil, err
	}
	if err := machAppendOpen(appender.stmt, tableName); err != nil {
		return nil, err
	}

	row := this.QueryRow("select type from M$SYS_TABLES where name = ?", appender.tableName)
	if err := row.Scan(&appender.tableType); err != nil {
		return nil, err
	}

	var err error
	appender.colCount, err = machColumnCount(appender.stmt)
	if err != nil {
		return nil, err
	}
	for i := 0; i < appender.colCount; i++ {
		/*typ, siz*/ _, _, err := machColumnType(appender.stmt, i)
		if err != nil {
			return nil, err
		}
	}
	return appender, nil
}

type Appender struct {
	stmt         unsafe.Pointer
	tableName    string
	tableType    int // 0: Log Table, 1: Fixed Table, 3: Volatile Table, 4: Lookup Table, 5: KeyValue Table, 6: Tag Table
	colCount     int
	SuccessCount uint64
	FailureCount uint64
}

func (this *Appender) Close() error {
	if this.stmt == nil {
		return nil
	}
	s, f, err := machAppendClose(this.stmt)
	if err != nil {
		return err
	}
	this.SuccessCount = s
	this.FailureCount = f

	if err := machFreeStmt(this.stmt); err != nil {
		return err
	}
	this.stmt = nil
	return nil
}

func (this *Appender) Append(cols ...any) error {
	if this.tableType == 0 {
		return this.appendLogTable(time.Time{}, cols)
	} else {
		return this.appendTagTable(cols)
	}
}

// supports only Log Table
func (this *Appender) AppendWithTimestamp(ts time.Time, cols ...any) error {
	if this.tableType == 0 {
		return this.appendLogTable(ts, cols)
	} else {
		return fmt.Errorf("%s is not a log table, use Append() instead", this.tableName)
	}
}

func (this *Appender) appendLogTable(ts time.Time, cols []any) error {
	if this.colCount-1 != len(cols) {
		return fmt.Errorf("value count %d, table '%s' has %d columns", len(cols), this.tableName, this.colCount-1)
	}
	vals := make([]*machAppendDataNullValue, this.colCount)
	if ts.IsZero() {
		vals[0] = bindValue(nil)
	} else {
		vals[0] = bindValue(ts)
	}
	for i, c := range cols {
		vals[i+1] = bindValue(c)
	}
	if err := machAppendData(this.stmt, vals); err != nil {
		return err
	}
	return nil
}

func (this *Appender) appendTagTable(cols []any) error {
	if this.colCount != len(cols) {
		return fmt.Errorf("value count %d, table '%s' has %d columns", len(cols), this.tableName, this.colCount)
	}
	vals := make([]*machAppendDataNullValue, this.colCount)
	for i, c := range cols {
		vals[i] = bindValue(c)
	}
	if err := machAppendData(this.stmt, vals); err != nil {
		return err
	}
	return nil
}
