package main

import (
	"net/http"
	"webssh/webssh"

	"github.com/julienschmidt/httprouter"
)

func main() {
	runWebSsh()
}

func runWebSsh() {
	router := httprouter.New()
	router.GET("/wssh", webssh.WsHandler)
	http.ListenAndServe(":59999", router)
}
