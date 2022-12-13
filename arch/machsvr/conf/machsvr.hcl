define VARS {
    GRPC_LISTEN_HOST = flag("--grpc-listen-host", "127.0.0.1")
    GRPC_LISTEN_PORT = flag("--grpc-listen-port", "4056")
    HTTP_LISTEN_HOST = flag("--http-listen-host", "127.0.0.1")
    HTTP_LISTEN_PORT = flag("--http-listen-port", "4088")
    MQTT_LISTEN_HOST = flag("--mqtt-listen-host", "127.0.0.1")
    MQTT_LISTEN_PORT = flag("--mqtt-listen-port", "4083")
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
        }
        Grpc = {
            Listeners        = [ "unix://${execDir()}/mach.sock", "tcp://${VARS_GRPC_LISTEN_HOST}:${VARS_GRPC_LISTEN_PORT}" ]
            MaxRecvMsgSize   = 4
            MaxSendMsgSize   = 4
        }
        Http = {
            Listeners        = [ "tcp://${VARS_HTTP_LISTEN_HOST}:${VARS_HTTP_LISTEN_PORT}" ]
            Prefix           = "/db"
        }
        Mqtt = {
            Listeners        = [ "tcp://${VARS_MQTT_LISTEN_HOST}:${VARS_MQTT_LISTEN_PORT}"]
            Prefix           = "db"
        }
    }
}
