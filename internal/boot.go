package internal

import (
	"github.com/busy-cloud/boat/boot"
	"time"
)

func init() {
	boot.Register("tcp-client", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"log", "mqtt", "database"},
	})
}

func Startup() error {

	//订阅通知
	subscribe()

	//5秒后再启动，先让其他准备好
	time.AfterFunc(time.Second*5, StartClients)

	return nil
}

func Shutdown() error {

	StopClients()

	return nil
}
