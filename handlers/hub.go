package handlers

import (
	"github.com/stinkyfingers/differencebetween/api/game"
	"golang.org/x/net/websocket"
)

type Hub struct {
	ClientMap map[int][]*GameConn
	Broadcast chan *game.Game
}

type GameConn struct {
	Conn      *websocket.Conn
	GameID    int
	WriteChan chan *game.Game
}

func NewHub() *Hub {
	return &Hub{
		ClientMap: make(map[int][]*GameConn),
		Broadcast: make(chan *game.Game),
	}
}

func (h *Hub) Register(gc *GameConn) {
	h.ClientMap[gc.GameID] = append(h.ClientMap[gc.GameID], gc)
}

func (h *Hub) Unregister(gc *GameConn) {
	if clientMap, ok := h.ClientMap[gc.GameID]; ok {
		for i := range h.ClientMap[gc.GameID] {
			if h.ClientMap[gc.GameID][i] == gc {
				clientMap = append(h.ClientMap[gc.GameID][:i], h.ClientMap[gc.GameID][i+1:]...)
			}
		}
		h.ClientMap[gc.GameID] = clientMap
	}
}
