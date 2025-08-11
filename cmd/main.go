package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/busy-cloud/boat/apis" //boat的基本接口
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/tcp-client"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("tcp-client")
	//e := viper.SafeWriteConfig()
	////e := viper.WriteConfig()
	//if e != nil {
	//	log.Error(e)
	//}

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

	//web.Engine().Static("app/assets/tcp-server/", "assets")

	err = web.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
