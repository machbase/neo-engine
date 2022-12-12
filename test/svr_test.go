package test

import (
	"testing"

	"github.com/machbase/booter"
	_ "github.com/machbase/cemlib/logging"
	_ "github.com/machbase/dbms-mach-go/server"
)

var serverConf = []byte(`
define VARS {
	WORKDIR = "../tmp"
}

module "github.com/machbase/cemlib/logging" {
    config {
        Console                     = false
        Filename                    = "-"
        DefaultPrefixWidth          = 30
        DefaultEnableSourceLocation = true
        DefaultLevel                = "TRACE"
        Levels = [
            { Pattern="machsvr", Level="TRACE" },
        ]
    }
}

module "github.com/machbase/dbms-mach-go/server" {
    name = "machsvr"
    config {
        MachbaseHome     = "${VARS_WORKDIR}/machbase"
        Machbase = {
            HANDLE_LIMIT = 1024
        }
        Grpc = {
            Listeners        = [ 
                "unix://${VARS_WORKDIR}/machsvr.sock", 
                "tcp://127.0.0.1:4056",
            ]
            MaxRecvMsgSize   = 4
            MaxSendMsgSize   = 4
        }
        Http = {
            Listeners        = [ "tcp://127.0.0.1:4088" ]
            Handlers         = [
                { Prefix: "/db",      Handler: "machbase" },
                { Prefix: "/metrics", Handler: "influx" },
            ]
        }
        Mqtt = {
            Listeners        = [ "tcp://127.0.0.1:4083"]
            Handlers         = [
                { Prefix: "db",      Handler: "machbase" },
                { Prefix: "metrics", Handler: "influx" },
            ]
        }
    }
}
`)

func TestMain(m *testing.M) {
	builder := booter.NewBuilder()
	b, err := builder.BuildWithContent(serverConf)
	if err != nil {
		panic(err)
	}
	err = b.Startup()
	if err != nil {
		panic(err)
	}
	m.Run()
	b.Shutdown()
}
