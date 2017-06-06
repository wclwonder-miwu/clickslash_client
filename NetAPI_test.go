package main

import (
	. "clickslash/protos"
	"fmt"
	"reflect"
	"testing"

	"github.com/garyburd/redigo/redis"
)

func TestMain(t *testing.T) {

	var some interface{}
	temp := &TUser{}
	temp.Coin = 5
	fmt.Println(temp)

	some = temp

	s := reflect.ValueOf(some).Elem()
	typeOfT := s.Type() //把s.Type()返回的Type对象复制给typeofT，typeofT也是一个反射。
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i) //迭代s的各个域，注意每个域仍然是反射。
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface()) //提取了每个域的名字
		//11:Icon int32=0
	}

}

func testRedis() {
	g_redis := connectRedis()

	id_count, err := redis.String(g_redis.Do("GET", "id_count"))
	if err != nil {
		fmt.Println("get id_count err")
		fmt.Println(err)
		return
	}

	fmt.Println(id_count)
}

func connectRedis() redis.Conn {

	fmt.Println("connectRedis")
	var err interface{}
	g_redis, err = redis.Dial("tcp", REDIS_IP)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return g_redis
}
