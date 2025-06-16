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
	var clients []*TcpClient
	err := db.Engine().Find(&clients)
	if err != nil {
		log.Error(err)
		return
	}
	for _, client := range clients {
		if client.Disabled {
			log.Info("client %s is disabled", client.Id)
			continue
		}
		err := FromClient(client)
		if err != nil {
			log.Error(err)
		}
	}
}

func StopClients() {
	clients.Range(func(name string, client *TcpClientImpl) bool {
		_ = client.Close()
		return true
	})
}

func FromClient(m *TcpClient) error {
	client := NewTcpClient(m)

	//保存
	val := clients.LoadAndStore(client.Id, client)
	if val != nil {
		err := val.Close()
		if err != nil {
			log.Error(err)
		}
	}

	//启动
	err := client.Open()
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
		return fmt.Errorf("tcp client %s not found", id)
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
