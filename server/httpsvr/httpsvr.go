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

func (this *Server) Start() error {
	return nil
}

func (this *Server) Stop() {
}

func (this *Server) Route(r *gin.Engine) {
	prefix := this.conf.Prefix
	// remove trailing slash
	for strings.HasSuffix(prefix, "/") {
		prefix = prefix[0 : len(prefix)-1]
	}

	r.GET(prefix+"/query", this.handleQuery)
	r.POST(prefix+"/query", this.handleQuery)
	r.POST(prefix+"/write", this.handleWrite)
}
