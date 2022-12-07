package mqttsvr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/machbase/cemlib/allowance"
	"github.com/machbase/cemlib/logging"
	"github.com/machbase/cemlib/mqtt"
	mach "github.com/machbase/dbms-mach-go"
	"github.com/machbase/dbms-mach-go/server/msg"
)

func New(conf *Config) *Server {
	svr := &Server{
		conf: conf,
		db:   mach.New(),
	}
	mqttdConf := &mqtt.MqttConfig{
		Name:             "machbase",
		TcpListeners:     []mqtt.TcpListenerConfig{},
		UnixSocketConfig: mqtt.UnixSocketListenerConfig{},
		Allowance: allowance.AllowanceConfig{
			Policy: "NONE",
		},
	}
	for _, c := range conf.Listeners {
		if strings.HasPrefix(c, "tcp://") {
			mqttdConf.TcpListeners = append(mqttdConf.TcpListeners, mqtt.TcpListenerConfig{
				ListenAddress: strings.TrimPrefix(c, "tcp://"),
				SoLinger:      0,
				KeepAlive:     10,
				NoDelay:       true,
			})
		} else if strings.HasPrefix(c, "unix://") {
			mqttdConf.UnixSocketConfig = mqtt.UnixSocketListenerConfig{
				Path:       strings.TrimPrefix(c, "unix://"),
				Permission: 0644,
			}
		}
	}
	if len(conf.Prefix) > 0 {
		conf.Prefix = strings.TrimSuffix(conf.Prefix, "/")
	}
	svr.mqttd = mqtt.NewServer(mqttdConf, svr)
	return svr
}

type Config struct {
	Listeners []string
	Prefix    string
	Passwords map[string]string
}

type Server struct {
	conf  *Config
	mqttd mqtt.Server
	db    *mach.Database
}

func (svr *Server) Start() error {
	err := svr.mqttd.Start()
	if err != nil {
		return err
	}

	return nil
}

func (svr *Server) Stop() {
	if svr.mqttd != nil {
		svr.mqttd.Stop()
	}
}

func (svr *Server) OnConnect(evt *mqtt.EvtConnect) (mqtt.AuthCode, *mqtt.ConnectResult, error) {
	peer, ok := svr.mqttd.GetPeer(evt.PeerId)
	if ok {
		peer.SetLogLevel(logging.LevelDebug)
	}
	result := &mqtt.ConnectResult{
		AllowedPublishTopicPatterns:   []string{fmt.Sprintf("%s/*", svr.conf.Prefix)},
		AllowedSubscribeTopicPatterns: []string{"*"},
	}
	return mqtt.AuthSuccess, result, nil
}

func (svr *Server) OnDisconnect(evt *mqtt.EvtDisconnect) {

}

func (svr *Server) OnMessage(evt *mqtt.EvtMessage) error {
	topic := evt.Topic
	topic = strings.TrimPrefix(topic, svr.conf.Prefix+"/")
	tick := time.Now()

	reply := func(msg any) {
		peer, ok := svr.mqttd.GetPeer(evt.PeerId)
		if ok {
			buff, err := json.Marshal(msg)
			if err != nil {
				return
			}
			peer.Publish(svr.conf.Prefix+"/reply", 1, buff)
		}
	}
	if topic == "query" {
		req := &msg.QueryRequest{}
		rsp := &msg.QueryResponse{Reason: "not specified"}
		err := json.Unmarshal(evt.Raw, req)
		if err != nil {
			rsp.Reason = err.Error()
			rsp.Elapse = time.Since(tick).String()
			reply(rsp)
			return nil
		}
		msg.Query(svr.db, req, rsp)
		rsp.Elapse = time.Since(tick).String()
		reply(rsp)
	} else if strings.HasPrefix(topic, "write/") {
		tableName := strings.TrimPrefix(topic, "write/")
		req := &msg.WriteRequest{}
		rsp := &msg.WriteResponse{Reason: "not specified"}
		err := json.Unmarshal(evt.Raw, req)
		if err != nil {
			rsp.Reason = err.Error()
			rsp.Elapse = time.Since(tick).String()
			reply(rsp)
			return nil
		}
		msg.Write(svr.db, tableName, req, rsp)
		rsp.Elapse = time.Since(tick).String()
		reply(rsp)
	}
	return nil
}

func (svr *Server) handleQuery(req *msg.QueryRequest, rsp *msg.QueryResponse, reply func(any)) {
	reply("OK")
}
