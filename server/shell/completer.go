package shell

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

func (sess *Session) completer(d prompt.Document) []prompt.Suggest {
	prev := strings.ToLower(d.Text)
	if strings.HasSuffix(prev, " from ") {
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
		return rt
	}

	prefix := d.GetWordBeforeCursor()
	if len(prefix) == 0 {
		return nil
	}
	suggests := []prompt.Suggest{
		{Text: "SELECT", Description: "select query"},
		{Text: "exit", Description: "exit shell"},
	}
	return prompt.FilterHasPrefix(suggests, prefix, true)
}
