package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
)

var clients lib.Map[TcpClientImpl]

func StartClients() {
	//加载连接器
	var servers []*TcpClient
	err := db.Engine().Find(&servers)
	if err != nil {
		log.Error(err)
		return
	}
	for _, server := range servers {
		if server.Disabled {
			log.Info("server %s is disabled", server.Id)
			continue
		}
		err := FromClient(server)
		if err != nil {
			log.Error(err)
		}
	}
}

func StopClients() {
	clients.Range(func(name string, server *TcpClientImpl) bool {
		_ = server.Close()
		return true
	})
}

func FromClient(m *TcpClient) error {
	server := NewTcpClient(m)

	//保存
	val := clients.LoadAndStore(server.Id, server)
	if val != nil {
		err := val.Close()
		if err != nil {
			log.Error(err)
		}
	}

	//启动
	err := server.Open()
	if err != nil {
		return err
	}

	return nil
}

func LoadClient(id string) error {
	var l TcpClient
	has, err := db.Engine().ID(id).Get(&l)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("tcp server %s not found", id)
	}

	return FromClient(&l)
}

func UnloadClient(id string) error {
	val := clients.LoadAndDelete(id)
	if val != nil {
		return val.Close()
	}
	return nil
}
