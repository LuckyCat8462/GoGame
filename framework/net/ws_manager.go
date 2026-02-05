package net

import (
	"common/logs"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var (
	websocketUpgrade = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type CheckOriginHandler func(r *http.Request) bool

type Manager struct {
	sync.RWMutex
	websocketUpgrade   *websocket.Upgrader
	ServerId           string
	CheckOriginHandler CheckOriginHandler
	clients            map[string]Connection
	ClientReadChan     chan *MsgPack
	//handlers           map[protocol.PackageType]EventHandler
	//ConnectorHandlers  LogicHandler
	//RemoteReadChan     chan []byte
	//RemoteCli          remote.Client
}

func (m Manager) Run(addr string) {
	go m.clientReadChanHandler()
	http.HandleFunc("/", m.serveWS)
	logs.Fatal("connector listen serve err:%v", http.ListenAndServe(addr, nil))
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	//websocket 基于http
	if m.websocketUpgrade == nil {
		m.websocketUpgrade = &websocketUpgrade
	}
	wsConn, err := m.websocketUpgrade.Upgrade(w, r, nil)
	if err != nil {
		logs.Error("websocketUpgrade.Upgrade err:%v", err)
		return
	}
	//封装连接，方便加入一些我们需要的内容
	client := NewWsConnection(wsConn, m)
	m.addClient(client)
	client.Run()
}

func (m *Manager) addClient(client *WsConnection) {
	m.Lock()
	defer m.Unlock()
	m.clients[client.Cid] = client
}

func (m *Manager) removeClient(wc *WsConnection) {
	for cid, c := range m.clients {
		if cid == wc.Cid {
			c.Close()
			delete(m.clients, cid)
		}
	}
}

func (m *Manager) clientReadChanHandler() {
	for {
		select {
		case body, ok := <-m.ClientReadChan:
			if ok {
				m.decodeClientPack(body)
			}
		}
	}
}

// 解析协议
func (m *Manager) decodeClientPack(body *MsgPack) {
	//解析协议
	logs.Info("receiver message:%v", string(body.Body))
	//packet, err := protocol.Decode(body.Body)
	//if err != nil {
	//	logs.Error("decode message err:%v", err)
	//	return
	//}
	//if err := m.routeEvent(packet, body.Cid); err != nil {
	//	logs.Error("routeEvent err:%v", err)
	//}
}

func (m Manager) Close() {
	for cid, v := range m.clients {
		v.Close()
		delete(m.clients, cid)
	}
}

func NewManager() *Manager {
	return &Manager{
		ClientReadChan: make(chan *MsgPack, 1024),
		clients:        make(map[string]Connection),
		//handlers:       make(map[protocol.PackageType]EventHandler),
		//RemoteReadChan: make(chan []byte, 1024),
	}
}
