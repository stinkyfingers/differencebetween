package handlers

import (
	"fmt"
	"net/http"
)

func Status(w http.ResponseWriter, r *http.Request) {
	fmt.Println("status called - path: ", r.URL.Path, " method: ", r.Method)
	w.Write([]byte("OK"))
}
