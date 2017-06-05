package main

import (
	"fmt"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

var g_net *NetAPI
var g_redis redis.Conn

func main() {
	connetRedis()
	testGrammar()

	g_net = new(NetAPI)
	g_net.Init(g_redis)

	http.HandleFunc("/hello", Hello)

	err := http.ListenAndServe("127.0.0.1"+":"+"2048", nil)

	if err != nil {
		panic(err)
	}
}

func Hello(w http.ResponseWriter, req *http.Request) {
	g_net.MainHandler(w, req)
}

func connetRedis() {
	var err interface{}
	g_redis, err = redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	//defer c.Close()
}
