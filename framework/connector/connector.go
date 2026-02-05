package connector

import (
	"common/logs"
	"fmt"
	"framework/game"
	"framework/net"
)

type Connector struct {
	isRunning bool
	wsManager *net.Manager
	//handlers  net.LogicHandler
	//remoteCli remote.Client
}

func Default() *Connector {
	return &Connector{
		//handlers: make(net.LogicHandler),
	}
}

func (c *Connector) Run(serverId string) {

	if !c.isRunning {
		//启动websocket和nats
		c.wsManager = net.NewManager()
		//c.wsManager.ConnectorHandlers = c.handlers
		////启动nats nats server不会存储消息
		//c.remoteCli = remote.NewNatsClient(serverId, c.wsManager.RemoteReadChan)
		//c.remoteCli.Run()
		//c.wsManager.RemoteCli = c.remoteCli
		c.Serve(serverId)
	}
}

func (c *Connector) Close() {
	if c.isRunning {
		//关闭websocket和nats
		c.wsManager.Close()
	}
}

func (c *Connector) Serve(serverId string) {
	logs.Info("run connector:%v", serverId)
	//需要一个地址，读取配置文件获取
	//在游戏中可能会有多个游戏，可能会有很多信息（配置），如果都写到yml中可能太复杂，所以一般写到json格式的配置文件中
	c.wsManager.ServerId = serverId
	connectorConfig := game.Conf.GetConnector(serverId)
	if connectorConfig == nil {
		logs.Fatal("no connector config found")
	}
	addr := fmt.Sprintf("%s:%d", connectorConfig.Host, connectorConfig.ClientPort)
	c.isRunning = true
	c.wsManager.Run(addr)
}

//func (c *Connector) RegisterHandler(handlers net.LogicHandler) {
//	c.handlers = handlers
//}
