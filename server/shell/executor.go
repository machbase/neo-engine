package shell

import (
	"fmt"
	"strings"
	"time"
)

func (sess *Session) executor(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if line == "exit" {
		sess.Close()
		return
	}

	sess.log.Debugf("SQL: %s", line)

	rows, err := sess.db.Query(line)
	if err != nil {
		sess.WriteStr(err.Error() + "\r\n")
		return
	}
	defer rows.Close()

	colNames, err := rows.ColumnNames()
	if err != nil {
		sess.WriteStr(err.Error() + "\r\n")
		return
	}
	sess.WriteStr(strings.Join(colNames, " | ") + "\r\n")
	for {
		rec, next, err := rows.Fetch()
		if err != nil {
			sess.WriteStr(err.Error() + "\r\n")
			return
		}
		if !next {
			return
		}
		cols := make([]string, len(rec))
		for i, r := range rec {
			if r == nil {
				cols[i] = fmt.Sprintf("%-10s", "NULL")
				continue
			}
			switch v := r.(type) {
			case *string:
				cols[i] = fmt.Sprintf("%-10s", *v)
			case *time.Time:
				cols[i] = fmt.Sprintf("%-26s", v.Format("2006-01-02 15:04:05.000000"))
			case *float64:
				cols[i] = fmt.Sprintf("%.5f", *v)
			case *int64:
				cols[i] = fmt.Sprintf("%10d", *v)
			default:
				cols[i] = fmt.Sprintf("%-10T", r)
			}
		}
		sess.WriteStr(strings.Join(cols, " ") + "\r\n")
	}
}
