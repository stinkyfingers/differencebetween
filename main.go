package main

import (
	"fmt"
	"os"

	"github.com/stinkyfingers/differencebetween/api/handlers"
	"github.com/stinkyfingers/easyrouter"
	"golang.org/x/net/websocket"
)

var (
	port = "7000"
)

var (
	h *handlers.Hub
)

var routes = []easyrouter.Route{
	{
		Path:    "/play/{id}",
		Methods: []string{"GET"},
		WSHandler: websocket.Handler(func(ws *websocket.Conn) {
			handlers.Game(ws, h)
		}),
	},
	{
		Path:    "/game",
		Methods: []string{"POST"},
		Handler: handlers.CreateGame,
	},
	{
		Path:    "/player",
		Methods: []string{"POST"},
		Handler: handlers.AddPlayer,
	},
}

func main() {
	fmt.Println("STARTING API")
	h = handlers.NewHub()

	if portEnv := os.Getenv("PORT"); portEnv != "" {
		port = portEnv
	}
	fmt.Println("PORT: ", port)
	s := easyrouter.Server{
		Port:   port,
		Routes: routes,
		DefaultRoute: easyrouter.Route{
			Path:    "/",
			Methods: []string{"GET"},
			Handler: handlers.Status,
		},
		Middlewares: []easyrouter.Middleware{handlers.Cors},
	}

	s.Run()
}
