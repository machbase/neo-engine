package mqttsvr

import (
	"strings"

	"github.com/machbase/cemlib/allowance"
	"github.com/machbase/cemlib/mqtt"
)

func New(conf *Config) *Server {
	svr := &Server{
		conf: conf,
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
	Listeners   []string
	TopicPrefix string
	Passwords   map[string]string
	Prefix      string
}

type Server struct {
	conf  *Config
	mqttd mqtt.Server
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
	return mqtt.AuthSuccess, nil, nil
}

func (svr *Server) OnDisconnect(evt *mqtt.EvtDisconnect) {

}

func (svr *Server) OnMessage(evt *mqtt.EvtMessage) error {
	return nil
}
