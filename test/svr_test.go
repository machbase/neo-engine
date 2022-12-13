package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/machbase/booter"
	_ "github.com/machbase/cemlib/logging"
	mach "github.com/machbase/dbms-mach-go"
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

var benchmarkTableName = strings.ToUpper("samplebench")

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

	/// preparing benchmark table
	db := mach.New()
	var count int

	checkTableSql := fmt.Sprintf("select count(*) from M$SYS_TABLES where name = '%s'", benchmarkTableName)
	row := db.QueryRow(checkTableSql)
	err = row.Scan(&count)
	if err != nil {
		panic(err)
	}

	if count == 1 {
		dropTableSql := fmt.Sprintf("drop table %s", benchmarkTableName)
		_, err = db.Exec(dropTableSql)
		if err != nil {
			panic(err)
		}
	}

	creTableSql := fmt.Sprintf(db.SqlTidy(`
            create tag table %s (
                name     varchar(200) primary key,
                time     datetime basetime,
                value    double summarized,
                id       varchar(80),
                jsondata json
        )`), benchmarkTableName)
	_, err = db.Exec(creTableSql)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE INDEX %s_id_idx ON %s (id)", benchmarkTableName, benchmarkTableName))
	if err != nil {
		panic(err)
	}

	row = db.QueryRow("select count(*) from " + benchmarkTableName)
	err = row.Scan(&count)
	if err != nil {
		panic(err)
	}
	/// end of preparing benchmark table

	m.Run()
	b.Shutdown()
}
