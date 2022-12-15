package shell

import (
	"regexp"
	"strings"

	"github.com/c-bata/go-prompt"
)

func (sess *Session) completer(d prompt.Document) []prompt.Suggest {
	line := strings.ToUpper(d.Text)
	match, _ := regexp.MatchString(`\s+FROM\s+.*$`, line)
	if match {
		rows, err := sess.db.Query("select NAME from M$SYS_TABLES order by NAME")
		if err != nil {
			sess.log.Errorf("select m$sys_tables fail; %s", err.Error())
			return nil
		}
		defer rows.Close()
		rt := []prompt.Suggest{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			rt = append(rt, prompt.Suggest{Text: name, Description: ""})
		}
		tableNamePrefix := d.GetWordBeforeCursor()
		if len(tableNamePrefix) == 0 {
			return rt
		}
		return prompt.FilterHasPrefix(rt, tableNamePrefix, true)
	}

	// prefix := d.GetWordBeforeCursor()
	// if len(prefix) == 0 {
	// 	return nil
	// }
	// suggests := []prompt.Suggest{
	// 	{Text: "SELECT", Description: ""},
	// 	{Text: "exit", Description: ""},
	// }
	// return prompt.FilterHasPrefix(suggests, prefix, true)
	return nil
}
