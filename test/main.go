package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/busy-cloud/boat-ui"
	_ "github.com/busy-cloud/boat/apis" //boat的基本接口
	"github.com/busy-cloud/boat/apps"
	"github.com/busy-cloud/boat/boot"
	_ "github.com/busy-cloud/boat/broker"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/store"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/modbus" //测试一个协议
	_ "github.com/busy-cloud/tcp-client"
	_ "github.com/busy-cloud/user"
	_ "github.com/god-jason/iot-master"
	"github.com/spf13/viper"
)

func init() {
	manifest, err := os.ReadFile("manifest.json")
	if err != nil {
		log.Fatal(err)
	}

	//注册为内部插件
	var a apps.App
	err = json.Unmarshal(manifest, &a)
	if err != nil {
		log.Fatal(err)
	}

	a.AssetsFS = store.Dir("assets")
	a.PagesFS = store.Dir("pages")
	a.TablesFS = store.Dir("tables")

	apps.Register(&a)
}

func main() {
	viper.SetConfigName("tcp-client")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs

		//关闭web，出发
		_ = web.Shutdown()
	}()

	//安全退出
	defer boot.Shutdown()

	err := boot.Startup()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = web.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
