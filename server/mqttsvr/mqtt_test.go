package mqttsvr_test

import (
	"sync"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
)

func getOptions(broker string, clientId string, keepAlive time.Duration) *paho.ClientOptions {
	cfg := paho.NewClientOptions()
	cfg.SetKeepAlive(keepAlive)
	cfg.SetCleanSession(true)
	cfg.SetConnectRetry(false)
	cfg.SetAutoReconnect(false)
	cfg.SetProtocolVersion(4)

	cfg.SetClientID(clientId)
	cfg.AddBroker(broker)
	return cfg
}

func TestSmartCsClient(t *testing.T) {
	ops := getOptions("127.0.0.1:4086", "machbase-cli", 3*time.Second)
	ops.SetUsername("user")
	ops.SetPassword("pass")

	client := paho.NewClient(ops)
	require.NotNil(t, client)

	result := client.Connect()
	ok := result.WaitTimeout(time.Second)
	if result.Error() != nil {
		t.Logf("CONNECT: %s", result.Error())
	}
	require.True(t, ok)
	require.Nil(t, result.Error())

	wg := sync.WaitGroup{}

	client.Subscribe("db/reply", 1, func(_ paho.Client, msg paho.Message) {
		buff := msg.Payload()
		t.Logf("RECV: %v", string(buff))
		wg.Done()
	})

	//// ceck table exists
	jsonStr := `{
		"q": "select count(*) from M$SYS_TABLES where name = 'SAMPLE'",
		"limit": 10,
		"cursor": 0
	}`
	wg.Add(1)
	client.Publish("db/query", 1, false, []byte(jsonStr))

	//// create table
	jsonStr = `{
		"q": "create tag table sample (name varchar(200) primary key, time datetime basetime, value double summarized, jsondata json)"
	}`
	wg.Add(1)
	client.Publish("db/query", 1, false, []byte(jsonStr))

	//// insert
	jsonStr = `{
		"data": {
			"columns":["name", "time", "value"],
			"records": [
				[ "sample.tag", 1670380342000000000, 1.0001 ],
				[ "sample.tag", 1670380343000000000, 2.0002 ]
			]
		}
	}`
	wg.Add(1)
	client.Publish("db/write/sample", 1, false, []byte(jsonStr))

	//// select
	wg.Add(1)
	client.Publish("db/query", 1, false, []byte(`{"q":"select * from sample"}`))

	//// wait until receive all replied messages from server
	wg.Wait()
	client.Disconnect(100)
	time.Sleep(time.Second * 1)
}
