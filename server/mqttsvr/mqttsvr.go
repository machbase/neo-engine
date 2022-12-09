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
	cmap "github.com/orcaman/concurrent-map"
	"github.com/tidwall/gjson"
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
	log   logging.Log

	appenders cmap.ConcurrentMap
}

func (svr *Server) Start() error {
	svr.log = logging.GetLog("mqttsvr")
	svr.appenders = cmap.New()

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
	svr.appenders.RemoveCb(evt.PeerId, func(key string, v interface{}, exists bool) bool {
		if !exists {
			return false
		}
		appenders := v.([]*mach.Appender)
		for _, ap := range appenders {
			ap.Close()
		}
		return true
	})
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
		////////////////////////
		// query
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
	} else if strings.HasPrefix(topic, "write") {
		////////////////////////
		// write
		req := &msg.WriteRequest{}
		rsp := &msg.WriteResponse{Reason: "not specified"}
		err := json.Unmarshal(evt.Raw, req)
		if err != nil {
			rsp.Reason = err.Error()
			rsp.Elapse = time.Since(tick).String()
			reply(rsp)
			return nil
		}
		if len(req.Table) == 0 {
			req.Table = strings.TrimPrefix(topic, "write/")
		}

		if len(req.Table) == 0 {
			rsp.Reason = "table is not specified"
			rsp.Elapse = time.Since(tick).String()
			reply(rsp)
			return nil
		}
		msg.Write(svr.db, req, rsp)
		rsp.Elapse = time.Since(tick).String()
		reply(rsp)
	} else if strings.HasPrefix(topic, "append/") {
		////////////////////////
		// append
		table := strings.ToUpper(strings.TrimPrefix(topic, "append/"))
		if len(table) == 0 {
			return nil
		}

		var err error
		var appenderSet []*mach.Appender
		var appender *mach.Appender

		val, exists := svr.appenders.Get(evt.PeerId)
		if exists {
			appenderSet = val.([]*mach.Appender)
			for _, a := range appenderSet {
				if a.Table() == table {
					appender = a
					break
				}
			}
		}
		if appender == nil {
			appender, err = svr.db.Appender(table)
			if err != nil {
				svr.log.Error("fail to create appender, %s", err.Error())
				return nil
			}
			if len(appenderSet) == 0 {
				appenderSet = []*mach.Appender{}
			}
			appenderSet = append(appenderSet, appender)
			svr.appenders.Set(evt.PeerId, appenderSet)
		}

		result := gjson.ParseBytes(evt.Raw)

		head := result.Get("0")
		if head.IsArray() {
			// if payload contains multiple tuples
			result.ForEach(func(key, value gjson.Result) bool {
				vals := []any{}
				value.ForEach(func(key, value gjson.Result) bool {
					vals = append(vals, value.Value())
					return true
				})
				err = appender.Append(vals...)
				if err != nil {
					svr.log.Warnf("append fail %s", err.Error())
					return false
				}
				return true
			})
		} else {
			// a single tuple
			vals := []any{}
			result.ForEach(func(key, value gjson.Result) bool {
				vals = append(vals, value.Value())
				return true
			})
			err = appender.Append(vals...)
			if err != nil {
				svr.log.Warnf("append fail %s", err.Error())
			}
		}
	}
	return nil
}
