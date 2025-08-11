package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/link"
	"net"
	"regexp"
	"time"
)

type TcpClient struct {
	Id              string         `json:"id,omitempty" xorm:"pk"`
	Name            string         `json:"name,omitempty"`
	Description     string         `json:"description,omitempty"`
	Address         string         `json:"address,omitempty"`                         //IP或域名
	Port            uint16         `json:"port,omitempty"`                            //端口号
	Protocol        string         `json:"protocol,omitempty"`                        //通讯协议
	ProtocolOptions map[string]any `json:"protocol_options,omitempty" xorm:"json"`    //通讯协议参数
	Disabled        bool           `json:"disabled,omitempty"`                        //禁用
	Created         time.Time      `json:"created,omitempty,omitzero" xorm:"created"` //创建时间

	link.Status `xorm:"-"`
}

type TcpClientImpl struct {
	*TcpClient

	net.Conn

	buf    []byte
	opened bool
}

var idReg = regexp.MustCompile(`^\w{2,128}$`)

func NewTcpClient(l *TcpClient) *TcpClientImpl {
	c := &TcpClientImpl{
		TcpClient: l,
		buf:       make([]byte, 4096),
	}
	return c
}

func (c *TcpClientImpl) connect() (err error) {
	if c.Conn != nil {
		_ = c.Conn.Close()
	}

	//连接
	addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	c.Conn, err = net.Dial("tcp", addr)
	if err != nil {
		c.Error = err.Error()
		return err
	}

	c.Running = true

	go c.receive(c.Conn)

	return
}

func (c *TcpClientImpl) Open() (err error) {
	if c.opened {
		return errors.New("already open")
	}
	c.opened = true

	//保持连接
	go c.keep()

	return c.connect()
}

func (c *TcpClientImpl) keep() {
	for c.opened {
		time.Sleep(time.Minute)

		if c.Conn == nil {
			err := c.connect()
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (c *TcpClientImpl) Close() error {
	c.opened = false

	//停止监听
	if c.Conn != nil {
		err := c.Conn.Close()
		c.Conn = nil
		return err
	}

	return nil
}

func (c *TcpClientImpl) receive(conn net.Conn) {
	//从数据库中查询
	var l link.Link
	//xorm.ErrNotExist //db.Engine.Exist()
	//.Where("linker=", "tcp-client").And("id=", id)
	has, err := db.Engine().ID(c.Id).Get(&l)
	if err != nil {
		_, _ = conn.Write([]byte(err.Error()))
		_ = conn.Close()
		return
	}

	//查不到
	if !has {
		l.Id = c.Id
		l.Linker = "tcp-client"
		l.Protocol = c.Protocol //继承协议
		l.ProtocolOptions = c.ProtocolOptions
		_, err = db.Engine().InsertOne(&l)
		if err != nil {
			_, _ = conn.Write([]byte(err.Error()))
			_ = conn.Close()
			return
		}
	} else {
		if l.Disabled {
			_, _ = conn.Write([]byte("disabled"))
			_ = conn.Close()
			return
		}
	}

	//连接
	topicOpen := fmt.Sprintf("link/tcp-client/%s/open", c.Id)
	mqtt.Publish(topicOpen, nil)
	if c.Protocol != "" {
		topicOpen = fmt.Sprintf("protocol/%s/link/tcp-client/%s/open", c.Protocol, c.Id)
		mqtt.Publish(topicOpen, c.ProtocolOptions)
	}

	topicUp := fmt.Sprintf("link/tcp-client/%s/up", c.Id)
	topicUpProtocol := fmt.Sprintf("protocol/%s/link/tcp-client/%s/up", c.Protocol, c.Id)

	var n int
	var e error
	buf := make([]byte, 4096)
	for {
		n, e = conn.Read(buf)
		if e != nil {
			_ = conn.Close()
			break
		}

		data := buf[:n]
		//转发
		mqtt.Publish(topicUp, data)
		if c.Protocol != "" {
			mqtt.Publish(topicUpProtocol, data)
		}
	}

	//下线
	topicClose := fmt.Sprintf("link/tcp-client/%s/close", c.Id)
	mqtt.Publish(topicClose, e.Error())
	if c.Protocol != "" {
		topic := fmt.Sprintf("protocol/%s/link/tcp-client/%s/close", c.Protocol, c.Id)
		mqtt.Publish(topic, e.Error())
	}

	c.Running = false

	c.Conn = nil
}
