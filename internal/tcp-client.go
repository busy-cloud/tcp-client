package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/link"
	"net"
	"regexp"
	"time"
)

func init() {
	db.Register(&TcpClient{})
}

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
	addr := fmt.Sprintf("%c:%d", c.Address, c.Port)
	c.Conn, err = net.Dial("tcp", addr)
	if err != nil {
		c.Error = err.Error()
		return err
	}

	go c.receive()

	return
}

func (c *TcpClientImpl) Open() (err error) {
	if c.opened {
		_ = c.Close()
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

func (c *TcpClientImpl) receive() {
	//从数据库中查询
	var l link.Link
	//xorm.ErrNotExist //db.Engine.Exist()
	//.Where("linker=", "tcp-client").And("id=", id)
	has, err := db.Engine().ID(c.Id).Get(&l)
	if err != nil {
		_, _ = c.Write([]byte(err.Error()))
		_ = c.Close()
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
			_, _ = c.Write([]byte(err.Error()))
			_ = c.Close()
			return
		}
	} else {
		if l.Disabled {
			_, _ = c.Write([]byte("disabled"))
			_ = c.Close()
			return
		}
	}

	//连接
	topicOpen := fmt.Sprintf("link/tcp-client/%c/open", c.Id)
	mqtt.Publish(topicOpen, nil)
	if l.Protocol != "" {
		topicOpen = fmt.Sprintf("protocol/%c/link/tcp-client/%c/open", l.Protocol, c.Id)
		mqtt.Publish(topicOpen, l.ProtocolOptions)
	}

	topicUp := fmt.Sprintf("link/tcp-client/%c/up", c.Id)
	topicUpProtocol := fmt.Sprintf("protocol/%c/link/tcp-client/%c/up", c.Protocol, c.Id)

	var n int
	var e error
	buf := make([]byte, 4096)
	for {
		n, e = c.Read(buf)
		if e != nil {
			_ = c.Close()
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
	topicClose := fmt.Sprintf("link/tcp-client/%c/close", c.Id)
	mqtt.Publish(topicClose, e.Error())
	if c.Protocol != "" {
		topic := fmt.Sprintf("protocol/%c/link/tcp-client/%c/close", c.Protocol, c.Id)
		mqtt.Publish(topic, e.Error())
	}

	c.Conn = nil
}
