
define DEF {
    LISTEN_HOST       = flag("--host", "127.0.0.1")
    SHELL_PORT        = flag("--shell-port", "4055")
    GRPC_PORT         = flag("--grpc-port", "4056")
    MQTT_PORT         = flag("--mqtt-port", "4057")
    HTTP_PORT         = flag("--http-port", "4058")
}

define VARS {
    SHELL_LISTEN_HOST = flag("--shell-listen-host", DEF_LISTEN_HOST)
    SHELL_LISTEN_PORT = flag("--shell-listen-port", DEF_SHELL_PORT)
    GRPC_LISTEN_HOST = flag("--grpc-listen-host", DEF_LISTEN_HOST)
    GRPC_LISTEN_PORT = flag("--grpc-listen-port", DEF_GRPC_PORT)
    HTTP_LISTEN_HOST = flag("--http-listen-host", DEF_LISTEN_HOST)
    HTTP_LISTEN_PORT = flag("--http-listen-port", DEF_HTTP_PORT)
    MQTT_LISTEN_HOST = flag("--mqtt-listen-host", DEF_LISTEN_HOST)
    MQTT_LISTEN_PORT = flag("--mqtt-listen-port", DEF_MQTT_PORT)
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

module "github.com/machbase/cemlib/banner" {
    config {
        Label = pname()
        Info  = version()
    }
}

module "github.com/machbase/dbms-mach-go/server" {
    name = "machsvr"
    config {
        MachbaseHome     = "${execDir()}/machbase"
        Machbase         = {
            HANDLE_LIMIT     = 2048
        }
        Shell = {
            Listeners        = [ "tcp://${VARS_SHELL_LISTEN_HOST}:${VARS_SHELL_LISTEN_PORT}" ]
        }
        Grpc = {
            Listeners        = [ 
                "unix://${execDir()}/mach.sock",
                "tcp://${VARS_GRPC_LISTEN_HOST}:${VARS_GRPC_LISTEN_PORT}",
            ]
            MaxRecvMsgSize   = 4
            MaxSendMsgSize   = 4
        }
        Http = {
            Listeners        = [ "tcp://${VARS_HTTP_LISTEN_HOST}:${VARS_HTTP_LISTEN_PORT}" ]
            Handlers         = [
                { Prefix: "/db",      Handler: "machbase" },
                { Prefix: "/metrics", Handler: "influx" },
            ]
        }
        Mqtt = {
            Listeners        = [ "tcp://${VARS_MQTT_LISTEN_HOST}:${VARS_MQTT_LISTEN_PORT}"]
            Handlers         = [
                { Prefix: "db",       Handler: "machbase" },
                { Prefix: "metrics",  Handler: "influx" },
            ]
        }
    }
}
