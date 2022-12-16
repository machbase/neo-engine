package shell

import (
	"strings"

	mach "github.com/machbase/dbms-mach-go"
)

func (sess *Session) exec_show(line string) {
	toks := strings.Fields(line)
	if len(toks) == 1 {
		//sess.Println("Usage: SHOW [VERSION | CONFIG]")
		sess.Println("Usage: SHOW [VERSION]")
		return
	}
	if toks[0] != "SHOW" || len(toks) == 1 {
		sess.log.Errorf("invalid show command: %s", line)
		return
	}
	switch toks[1] {
	case "VERSION":
		v := mach.GetVersion()
		sess.Printf("Server v%d.%d.%d #%s", v.Major, v.Minor, v.Patch, v.GitSHA)
		sess.Printf("Engine %s", mach.LibMachLinkInfo)
		// case "CONFIG":
		// 	booter.GetInstance()
		// 	sess.Printf("%+v", sess.server.conf)
	}
}
