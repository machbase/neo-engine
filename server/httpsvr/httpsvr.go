package httpsvr

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/machbase/cemlib/logging"
	mach "github.com/machbase/dbms-mach-go"
)

func New(conf *Config) (*Server, error) {
	return &Server{
		conf: conf,
		log:  logging.GetLog("httpsvr"),
		db:   mach.New(),
	}, nil
}

type Config struct {
	Prefix string
}

type Server struct {
	conf *Config
	log  logging.Log
	db   *mach.Database
}

func (my *Server) Start() error {
	return nil
}

func (my *Server) Stop() {
}

func (my *Server) Route(r *gin.Engine) {
	prefix := my.conf.Prefix
	// remove trailing slash
	for strings.HasSuffix(prefix, "/") {
		prefix = prefix[0 : len(prefix)-1]
	}

	r.GET(prefix+"/query", my.handleQuery)
	r.POST(prefix+"/query", my.handleQuery)
	r.POST(prefix+"/write", my.handleWrite)
}
