package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/stinkyfingers/differencebetween/api/game"
	"golang.org/x/net/websocket"
)

type GameRequest struct {
	Player string `json:"player"` // name
	Rounds int    `json:"rounds"` // num rounds
}

type PlayerRequest struct {
	Player string `json:"player"` // name
	GameID int    `json:"id"`
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	var gameRequest GameRequest
	err := json.NewDecoder(r.Body).Decode(&gameRequest)
	if err != nil {
		HTTPError(w, err)
		return
	}
	g, err := game.NewGame(game.Player{Name: gameRequest.Player}, gameRequest.Rounds)
	if err != nil {
		HTTPError(w, err)
		return
	}
	j, err := json.Marshal(g)
	if err != nil {
		HTTPError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(j)
}

func AddPlayer(w http.ResponseWriter, r *http.Request) {
	var playerRequest PlayerRequest
	err := json.NewDecoder(r.Body).Decode(&playerRequest)
	if err != nil {
		HTTPError(w, err)
		return
	}
	g, err := game.GetGame(playerRequest.GameID)
	if err != nil {
		HTTPError(w, err)
		return
	}
	err = g.AddPlayer(game.Player{Name: playerRequest.Player})
	if err != nil {
		HTTPError(w, err)
		return
	}
	j, err := json.Marshal(g)
	if err != nil {
		HTTPError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(j)
}

func Game(ws *websocket.Conn, hub *Hub) {
	idStr := ws.Request().URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WSError(ws, err)
		return
	}
	g, err := game.GetGame(id)
	if err != nil {
		WSError(ws, err)
		return
	}

	gameConn := &GameConn{
		GameID:    id,
		Conn:      ws,
		WriteChan: make(chan *game.Game),
	}
	hub.Register(gameConn)
	go gameConn.write(hub)
	hub.Broadcast <- g
	err = gameConn.read(hub)
	if err != nil {
		WSError(ws, err)
		return
	}
}

func (gc *GameConn) write(hub *Hub) error {
	for g := range hub.Broadcast {
		for _, client := range hub.ClientMap[g.ID] {
			err := websocket.JSON.Send(client.Conn, g)
			if err != nil {
				hub.Unregister(client)
				// continue
			}
		}
	}
	return nil
}

func (gc *GameConn) read(hub *Hub) error {
	defer func() {
		hub.Unregister(gc)
		gc.Conn.Close()
	}()
	for {
		var p game.Play
		err := websocket.JSON.Receive(gc.Conn, &p)
		if err != nil {
			return err
		}
		g, err := game.GetGame(gc.GameID)
		if err != nil {
			return err
		}
		if p.Vote != "" && g.CurrentAction == game.VOTE {
			g.Vote(p.Name, p.Vote)
		} else if p.Punchline != "" && g.CurrentAction == game.PLAY {
			g.Play(p.Name, p.Punchline)
		} else {
			log.Print("wrong action") // TODO err
			return errors.New("invalid action")
		}
		hub.Broadcast <- g
	}
}
