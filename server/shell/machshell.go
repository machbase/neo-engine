package shell

import (
	"fmt"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/machbase/cemlib/logging"
	"github.com/machbase/cemlib/ssh/sshd"
	mach "github.com/machbase/dbms-mach-go"
	"github.com/pkg/errors"
)

type Config struct {
	Listeners     []string
	IdleTimeout   time.Duration
	ServerKeyPath string
}

type Server struct {
	conf  *Config
	log   logging.Log
	sshds []sshd.Server
}

func New(conf *Config) *Server {
	return &Server{
		conf: conf,
	}
}

func (svr *Server) Start() error {
	svr.log = logging.GetLog("machshell")
	svr.sshds = make([]sshd.Server, 0)

	for _, listen := range svr.conf.Listeners {
		listenAddress := strings.TrimPrefix(listen, "tcp://")
		cfg := sshd.Config{
			ListenAddress:      listenAddress,
			ServerKey:          svr.conf.ServerKeyPath,
			IdleTimeout:        svr.conf.IdleTimeout,
			AutoListenAndServe: false,
		}
		s := sshd.New(&cfg)
		err := s.Start()
		if err != nil {
			return errors.Wrap(err, "machsell")
		}
		s.SetHandler(svr.sessionHandler)
		s.SetShellProvider(svr.shellProvider)
		s.SetMotdProvider(svr.motdProvider)
		s.SetPasswordHandler(svr.passwordProvider)
		go func() {
			err := s.ListenAndServe()
			if err != nil {
				svr.log.Warnf("machshell-listen %s", err.Error())
			}
		}()
		svr.log.Infof("SSHD Listen %s", listen)
	}
	return nil
}

func (svr *Server) Stop() {
	for _, s := range svr.sshds {
		s.Stop()
	}
}

func (svr *Server) shellProvider(user string) *sshd.Shell {
	return &sshd.Shell{
		Cmd: "/bin/bash",
	}
}

func (svr *Server) motdProvider(user string) string {
	return fmt.Sprintf(`Greeting, %s
machsvr %v
`, user, mach.VersionString())
}

func (svr *Server) passwordProvider(ctx ssh.Context, password string) bool {
	svr.log.Infof("shell login %s", ctx.User())
	return true
}

func (svr *Server) sessionHandler(ss ssh.Session) {
	svr.log.Tracef("%#v", ss)
	pty, ptyCh, ok := ss.Pty()
	if !ok {
		ss.Write([]byte("ERR unable to get PTY\r\n"))
		return
	}
	ss.Write([]byte(svr.motdProvider(ss.User()) + "\r\n"))

	svr.log.Tracef("PTY %+v WIN %+v", pty.Term, pty.Window)
	for w := range ptyCh {
		svr.log.Tracef("WIN %+v", w)
	}

	if len(ss.Command()) > 0 {
		//svr.commandHandler(ss)
	} else {
		//svr.shellHandler(ss)
	}
}
