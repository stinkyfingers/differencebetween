package handlers

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/websocket"
)

type Error struct {
	Message string `json:"message"`
}

func HTTPError(w http.ResponseWriter, err error) {
	message := "unspecified error"
	if err != nil {
		message = err.Error()
	}
	e := &Error{
		Message: message,
	}
	j, err := json.Marshal(e)
	if err != nil {
		w.Write([]byte("error encoding error"))
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(j)
}

func WSError(ws *websocket.Conn, err error) {
	message := "unspecified error"
	if err != nil {
		message = err.Error()
	}
	websocket.JSON.Send(ws, Error{Message: message})
}
