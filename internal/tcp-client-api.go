package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func init() {

	api.Register("GET", "tcp-client/client/:id/open", clientOpen)
	api.Register("GET", "tcp-client/client/:id/close", clientClose)

	api.Register("GET", "tcp-client/client/:id/status", clientStatus)
}

func getClientsInfo(ds []*TcpClient) error {
	for _, d := range ds {
		_ = getClientInfo(d)
	}
	return nil
}

func getClientInfo(d *TcpClient) error {
	l := clients.Load(d.Id)
	if l != nil {
		d.Status = l.Status
	}
	return nil
}

func clientClose(ctx *gin.Context) {
	l := clients.Load(ctx.Param("id"))
	if l == nil {
		api.Fail(ctx, "找不到服务器")
		return
	}

	err := l.Close()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}

func clientOpen(ctx *gin.Context) {
	l := clients.Load(ctx.Param("id"))
	if l != nil {
		err := l.Open()
		if err != nil {
			api.Error(ctx, err)
			return
		}
		api.OK(ctx, nil)
		return
	}

	err := LoadClient(ctx.Param("id"))
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}

func clientStatus(ctx *gin.Context) {
	l := clients.Load(ctx.Param("id"))
	if l == nil {
		api.Fail(ctx, "找不到服务器")
		return
	}

	api.OK(ctx, l.Status)
}
