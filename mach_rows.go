package mach

import (
	"database/sql"
	"fmt"
	"net"
	"time"
	"unsafe"

	spi "github.com/machbase/neo-spi"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Result struct {
	err          error
	affectedRows int64
	stmtType     StmtType
}

func (r *Result) RowsAffected() int64 {
	return r.affectedRows
}

func (r *Result) Err() error {
	return r.err
}

func (r *Result) Message() string {
	if r.err != nil {
		return r.err.Error()
	}

	rows := "no row"
	if r.affectedRows == 1 {
		rows = "a row"
	} else if r.affectedRows > 1 {
		p := message.NewPrinter(language.English)
		rows = p.Sprintf("%d rows", r.affectedRows)
	}
	if r.stmtType.IsSelect() {
		return rows + " selected."
	} else if r.stmtType.IsInsert() {
		return rows + " inserted."
	} else if r.stmtType.IsUpdate() {
		return rows + " updated."
	} else if r.stmtType.IsDelete() {
		return rows + " deleted."
	} else if r.stmtType.IsAlterSystem() {
		return "system altered."
	} else if r.stmtType.IsDDL() {
		return "ok."
	}
	return fmt.Sprintf("ok.(%d)", r.stmtType)
}

type Row struct {
	ok     bool
	err    error
	values []any

	affectedRows int64
	stmtType     StmtType
}

func (row *Row) Success() bool {
	return row.ok
}

func (row *Row) Err() error {
	return row.err
}

func (row *Row) Values() []any {
	return row.values
}

func (row *Row) RowsAffected() int64 {
	return row.affectedRows
}

func (r *Row) Message() string {
	if r.err != nil {
		return r.err.Error()
	}

	rows := "no row"
	if r.affectedRows == 1 {
		rows = "a row"
	} else if r.affectedRows > 1 {
		p := message.NewPrinter(language.English)
		rows = p.Sprintf("%d rows", r.affectedRows)
	}
	if r.stmtType.IsSelect() {
		return rows + " selected."
	} else if r.stmtType.IsInsert() {
		return rows + " inserted."
	} else if r.stmtType.IsUpdate() {
		return rows + " updated."
	} else if r.stmtType.IsDelete() {
		return rows + " deleted."
	} else if r.stmtType.IsAlterSystem() {
		return "system altered."
	} else if r.stmtType.IsDDL() {
		return "ok."
	}
	return fmt.Sprintf("ok.(%d)", r.stmtType)
}

func (row *Row) Scan(cols ...any) error {
	if row.err != nil {
		return row.err
	}
	if !row.ok {
		return sql.ErrNoRows
	}
	for i := range cols {
		if i >= len(row.values) {
			return fmt.Errorf("column %d is out of range %d", i, len(row.values))
		}
		var isNull bool
		switch v := row.values[i].(type) {
		case *int16:
			row.err = ScanInt16(*v, cols[i])
		case *int32:
			row.err = ScanInt32(*v, cols[i])
		case *int64:
			row.err = ScanInt64(*v, cols[i])
		case *time.Time:
			row.err = ScanDateTime(*v, cols[i])
		case *float32:
			row.err = ScanFloat32(*v, cols[i])
		case *float64:
			row.err = ScanFloat64(*v, cols[i])
		case *net.IP:
			row.err = ScanIP(*v, cols[i])
		case *string:
			row.err = ScanString(*v, cols[i])
		case []byte:
			row.err = ScanBytes(v, cols[i])
		default:
			return fmt.Errorf("column %d can not assign to %T", i, v)
		}
		if row.err != nil {
			return row.err
		}
		if isNull {
			cols[i] = nil
		}
	}
	return nil
}

type Rows struct {
	handle     unsafe.Pointer
	stmt       unsafe.Pointer
	stmtType   StmtType
	sqlText    string
	timeFormat string
}

func (rows *Rows) Close() error {
	if rows.stmt != nil {
		machFreeStmt(rows.handle, rows.stmt)
		rows.stmt = nil
	}
	rows.sqlText = ""
	return nil
}

func (rows *Rows) IsFetchable() bool {
	return rows.stmtType.IsSelect()
}

func (rows *Rows) StatementType() StmtType {
	return rows.stmtType
}

func (rows *Rows) RowsAffected() int64 {
	if rows.IsFetchable() {
		return 0
	}
	nrow, err := machEffectRows(rows.stmt)
	if err != nil {
		return 0
	}
	return nrow
}

func (rows *Rows) SetTimeFormat(format string) {
	rows.timeFormat = format
}

func (rows *Rows) Columns() (spi.Columns, error) {
	count, err := machColumnCount(rows.stmt)
	if err != nil {
		return nil, err
	}
	cols := make([]*Column, count)
	for i := 0; i < count; i++ {
		col, err := machColumnInfo(rows.stmt, i)
		if err != nil {
			return nil, errors.Wrap(err, "ColumnTypes")
		}
		cols[i] = col
	}
	result := make([]*spi.Column, len(cols))
	for i := range cols {
		result[i] = &spi.Column{
			Name:   cols[i].Name,
			Type:   cols[i].Type,
			Size:   cols[i].Size,
			Length: cols[i].Len,
		}
	}
	return result, nil
}

func (rows *Rows) Message() string {
	nrows := rows.RowsAffected()
	stmtType := rows.stmtType
	var verb = ""

	if stmtType >= 1 && stmtType <= 255 {
		return "executed."
	} else if stmtType >= 256 && stmtType <= 511 {
		// "ALTER SYSTEM"
		return "system altered."
	} else if stmtType == 512 {
		verb = "selected."
	} else if stmtType == 513 {
		verb = "inserted."
	} else if stmtType == 514 || stmtType == 515 {
		verb = "deleted."
	} else if stmtType == 516 {
		verb = "select and inserted."
	} else if stmtType == 517 {
		verb = "updated."
	} else {
		return fmt.Sprintf("executed (%d).", stmtType)
	}
	if nrows == 0 {
		return fmt.Sprintf("no row %s", verb)
	} else if nrows == 1 {
		return fmt.Sprintf("a row %s", verb)
	} else {
		p := message.NewPrinter(language.English)
		return p.Sprintf("%d rows %s", nrows, verb)
	}
}

// internal use only from machrpcserver
func (rows *Rows) Fetch() ([]any, bool, error) {
	// select 가 아니면 fetch를 진행하지 않는다.
	if !rows.IsFetchable() {
		return nil, false, sql.ErrNoRows
	}

	next, err := machFetch(rows.stmt)
	if err != nil {
		return nil, next, errors.Wrap(err, "Fetch")
	}
	if !next {
		return nil, false, nil
	}

	colCount, err := machColumnCount(rows.stmt)
	if err != nil {
		return nil, next, err
	}

	values := make([]any, colCount)
	for i := range values {
		typ, _, err := machColumnType(rows.stmt, i)
		if err != nil {
			return nil, next, errors.Wrap(err, "Fetch")
		}
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			values[i] = new(int)
		case 1: // MACH_DATA_TYPE_INT32
			values[i] = new(int32)
		case 2: // MACH_DATA_TYPE_INT64
			values[i] = new(int64)
		case 3: // MACH_DATA_TYPE_DATETIME
			switch rows.timeFormat {
			case "epoch":
				values[i] = new(int64)
			case "":
				values[i] = new(time.Time)
			default:
				values[i] = new(string)
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			values[i] = new(float32)
		case 5: // MACH_DATA_TYPE_DOUBLE
			values[i] = new(float64)
		case 6: // MACH_DATA_TYPE_IPV4
			values[i] = new(net.IP)
		case 7: // MACH_DATA_TYPE_IPV6
			values[i] = new(net.IP)
		case 8: // MACH_DATA_TYPE_STRING
			values[i] = new(string)
		case 9: // MACH_DATA_TYPE_BINARY
			values[i] = []byte{}
		default:
			return nil, next, fmt.Errorf("Fetch unsupported type %T", typ)
		}
	}
	err = scan(rows.stmt, values...)
	if err != nil {
		return nil, next, errors.Wrap(err, "Fetch")
	}
	return values, next, nil
}

func (rows *Rows) Next() bool {
	// select 가 아니면
	if !rows.IsFetchable() {
		return false
	}

	next, err := machFetch(rows.stmt)
	if err != nil {
		return false
	}
	return next
}

func (rows *Rows) Scan(cols ...any) error {
	return scan(rows.stmt, cols...)
}

func scan(stmt unsafe.Pointer, cols ...any) error {
	for i, c := range cols {
		typ, _ /*size*/, err := machColumnType(stmt, i)
		if err != nil {
			return errors.Wrap(err, "Scan")
		}
		switch typ {
		case 0: // MACH_DATA_TYPE_INT16
			if v, nonNull, err := machColumnDataInt16(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else if nonNull {
				if err = ScanInt16(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 1: // MACH_DATA_TYPE_INT32
			if v, nonNull, err := machColumnDataInt32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else if nonNull {
				if err = ScanInt32(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 2: // MACH_DATA_TYPE_INT64
			if v, nonNull, err := machColumnDataInt64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan int16")
			} else if nonNull {
				if err = ScanInt64(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 3: // MACH_DATA_TYPE_DATETIME
			if v, nonNull, err := machColumnDataDateTime(stmt, i); err != nil {
				return errors.Wrap(err, "Scan datetime")
			} else if nonNull {
				if err = ScanDateTime(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 4: // MACH_DATA_TYPE_FLOAT
			if v, nonNull, err := machColumnDataFloat32(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else if nonNull {
				if err = ScanFloat32(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 5: // MACH_DATA_TYPE_DOUBLE
			if v, nonNull, err := machColumnDataFloat64(stmt, i); err != nil {
				return errors.Wrap(err, "Scan float32")
			} else if nonNull {
				if err = ScanFloat64(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 6: // MACH_DATA_TYPE_IPV4
			if v, nonNull, err := machColumnDataIPv4(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else if nonNull {
				if err = ScanIP(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 7: // MACH_DATA_TYPE_IPV6
			if v, nonNull, err := machColumnDataIPv6(stmt, i); err != nil {
				return errors.Wrap(err, "scal IPv4")
			} else if nonNull {
				if err = ScanIP(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 8: // MACH_DATA_TYPE_STRING
			if v, nonNull, err := machColumnDataString(stmt, i); err != nil {
				return errors.Wrap(err, "Scan string")
			} else if nonNull {
				if err = ScanString(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		case 9: // MACH_DATA_TYPE_BINARY
			if v, nonNull, err := machColumnDataBinary(stmt, i); err != nil {
				return errors.Wrap(err, "Scan binary")
			} else if nonNull {
				if err = ScanBytes(v, c); err != nil {
					return err
				}
			} else {
				cols[i] = nil
			}
		default:
			return fmt.Errorf("MachGetColumnData unsupported type %T", c)
		}
	}
	return nil
}
