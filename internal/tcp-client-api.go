package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "tcp-client/client/list", curd.ApiListHook[TcpClient](getClientsInfo))
	api.Register("POST", "tcp-client/client/search", curd.ApiSearchHook[TcpClient](getClientsInfo))
	api.Register("POST", "tcp-client/client/create", curd.ApiCreateHook[TcpClient](nil, func(m *TcpClient) error {
		_ = FromClient(m)
		return nil
	}))
	api.Register("GET", "tcp-client/client/:id", curd.ApiGetHook[TcpClient](getClientInfo))

	api.Register("POST", "tcp-client/client/:id", curd.ApiUpdateHook[TcpClient](nil, func(m *TcpClient) error {
		_ = FromClient(m)
		return nil
	}, "id", "name", "type", "address", "port", "disabled", "protocol", "protocol_options"))

	api.Register("GET", "tcp-client/client/:id/delete", curd.ApiDeleteHook[TcpClient](nil, func(m *TcpClient) error {
		_ = UnloadClient(m.Id)
		return nil
	}))

	api.Register("GET", "tcp-client/client/:id/enable", curd.ApiDisableHook[TcpClient](false, nil, func(id any) error {
		_ = LoadClient(id.(string))
		return nil
	}))

	api.Register("GET", "tcp-client/client/:id/disable", curd.ApiDisableHook[TcpClient](true, nil, func(id any) error {
		_ = UnloadClient(id.(string))
		return nil
	}))

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
