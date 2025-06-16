package main

import (
	_ "github.com/busy-cloud/boat/apis" //boat的基本接口
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/tcp-client/internal"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	viper.SetConfigName("tcp-server")
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
