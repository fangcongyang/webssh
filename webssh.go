package main

import (
	"fmt"
	"net/http"
	"webssh/webssh"

	"github.com/julienschmidt/httprouter"
)

func main() {
	go runWebSsh()
	fmt.Println("启动webssh成功，监听端口：59999")
}

func runWebSsh()  {
	router := httprouter.New()
	router.GET("/wssh", webssh.WsHandler)
	http.ListenAndServe(":59999", router)
}
