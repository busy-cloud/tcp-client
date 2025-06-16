package internal

import (
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

func subscribe() {

	//订阅数据变化
	mqtt.Subscribe("link/tcp-client/+/down", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		conn := clients.Load(ss[2])
		if conn != nil {
			_, _ = conn.Write(payload)
		}
	})

	//关闭连接
	//mqtt.Subscribe("link/tcp-server/+/kill", func(topic string, payload []byte) {
	//	ss := strings.Split(topic, "/")
	//	conn := links.Load(ss[2])
	//	if conn != nil {
	//		_ = conn.Close()
	//	}
	//})
}
