package main

import (
	"fmt"
	"net/http"
)

var g_net *NetAPI

func main() {
	fmt.Println("some")

	g_net = new(NetAPI)
	g_net.Init()

	http.HandleFunc("/hello", Hello)

	err := http.ListenAndServe("127.0.0.1"+":"+"2048", nil)

	if err != nil {
		panic(err)
	}
}

func Hello(w http.ResponseWriter, req *http.Request) {
	g_net.MainHandler(w, req)
}
