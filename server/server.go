package server

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/machbase/booter"
	"github.com/machbase/cemlib/logging"
	mach "github.com/machbase/dbms-mach-go"
	"github.com/machbase/dbms-mach-go/machrpc"
	"github.com/machbase/dbms-mach-go/server/machrpcsvr"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func init() {
	defaultConf := Config{
		MachbaseHome:   ".",
		StartupTimeout: 5 * time.Second,
		Grpc: GrpcConfig{
			Listeners:      []string{"unix://./machsvr.sock"},
			MaxRecvMsgSize: 4,
			MaxSendMsgSize: 4,
		},
	}
	booter.Register(
		"github.com/machbase/dbms-mach-go/server",
		func() *Config {
			conf := defaultConf
			return &conf
		},
		func(conf *Config) (booter.Boot, error) {
			return &svr{
				conf: conf,
			}, nil
		},
	)
}

type Config struct {
	MachbaseHome   string
	StartupTimeout time.Duration
	Grpc           GrpcConfig
}

type GrpcConfig struct {
	Listeners      []string
	MaxRecvMsgSize int
	MaxSendMsgSize int
}

type svr struct {
	conf  *Config
	log   logging.Log
	db    *mach.Database
	grpcd *grpc.Server
}

const TagTableName = "tagdata"

func (this *svr) Start() error {
	this.log = logging.GetLog("machsvr")

	_, err := os.Stat(this.conf.MachbaseHome)
	if err != nil {
		return errors.Wrap(err, "config file not found")
	}
	homepath, err := filepath.Abs(this.conf.MachbaseHome)
	if err != nil {
		return errors.Wrap(err, "config file path")
	}
	if err := mach.Initialize(homepath); err != nil {
		return errors.Wrap(err, "initialize database")
	}
	if !mach.ExistsDatabase() {
		this.log.Info("create database")
		if err := mach.CreateDatabase(); err != nil {
			return errors.Wrap(err, "create database")
		}
	}

	this.db = mach.New()
	if this.db == nil {
		return errors.New("database instance failed")
	}

	if err := this.db.Startup(this.conf.StartupTimeout); err != nil {
		return errors.Wrap(err, "startup database")
	}

	err = this.db.Exec("alter system set trace_log_level=1023")
	if err != nil {
		return errors.Wrap(err, "alter log level")
	}

	if len(this.conf.Grpc.Listeners) > 0 {
		// grpc server
		machrpcSvr, err := machrpcsvr.New(&machrpcsvr.Config{})
		if err != nil {
			return errors.Wrap(err, "grpc handler")
		}
		// ingest gRPC options
		grpcOpt := []grpc.ServerOption{
			grpc.MaxRecvMsgSize(this.conf.Grpc.MaxRecvMsgSize * 1024 * 1024),
			grpc.MaxSendMsgSize(this.conf.Grpc.MaxSendMsgSize * 1024 * 1024),
			grpc.StatsHandler(machrpcSvr),
		}

		// create grpc server
		this.grpcd = grpc.NewServer(grpcOpt...)
		machrpc.RegisterMachbaseServer(this.grpcd, machrpcSvr)

		// listeners
		for _, listen := range this.conf.Grpc.Listeners {
			lsnr, err := makeListener(listen)
			if err != nil {
				return errors.Wrap(err, "cannot start with failed listener")
			}
			this.log.Infof("Listen %s", listen)

			// start go server
			go this.grpcd.Serve(lsnr)
		}
	}
	return nil
}

func (this *svr) Stop() {
	if this.grpcd != nil {
		this.grpcd.Stop()
	}
	this.log.Infof("shutdown.")
}

func makeListener(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, "unix://") {
		pwd, _ := os.Getwd()
		if strings.HasPrefix(addr, "unix://../") {
			addr = fmt.Sprintf("unix:///%s", filepath.Join(filepath.Dir(pwd), addr[len("unix://../"):]))
		} else if strings.HasPrefix(addr, "../") {
			addr = fmt.Sprintf("unix:///%s", filepath.Join(filepath.Dir(pwd), addr[len("../"):]))
		} else if strings.HasPrefix(addr, "unix://./") {
			addr = fmt.Sprintf("unix:///%s", filepath.Join(pwd, addr[len("unix://./"):]))
		} else if strings.HasPrefix(addr, "./") {
			addr = fmt.Sprintf("unix:///%s", filepath.Join(pwd, addr[len("./"):]))
		} else if strings.HasPrefix(addr, "/") {
			addr = fmt.Sprintf("unix://%s", addr)
		}
		path := addr[len("unix://"):]
		// delete existing .sock file
		if _, err := os.Stat(path); err == nil {
			os.Remove(path)
		}
		return net.Listen("unix", path)
	} else if strings.HasPrefix(addr, "tcp://") {
		return net.Listen("tcp", addr[len("tcp://"):])
	} else {
		return nil, fmt.Errorf("unuspported listen scheme %s", addr)
	}
}
