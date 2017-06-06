package util

import (
	"fmt"
	"reflect"

	"github.com/garyburd/redigo/redis"
)

//数据量大的时候保存，不用每次调用都反射一遍
type RedisOrm struct {
}

//把某个结构保存到redis
func (this *RedisOrm) SaveToKey(t interface{}, key string, predis redis.Conn) error {

	s := reflect.ValueOf(t).Elem()
	typeOfT := s.Type() //把s.Type()返回的Type对象复制给typeofT，typeofT也是一个反射。
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i) //迭代s的各个域，注意每个域仍然是反射。
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface()) //提取了每个域的名字
		//11:Icon int32=0
		predis.Do("HSET", key, typeOfT.Field(i).Name, f.Interface())
	}

	return nil
}

func (this *RedisOrm) TFromRedis(t interface{}) {
	s := reflect.ValueOf(t).Elem()
	typeOfT := s.Type() //把s.Type()返回的Type对象复制给typeofT，typeofT也是一个反射。
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i) //迭代s的各个域，注意每个域仍然是反射。
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface()) //提取了每个域的名字
		//11:Icon int32=0

		//str, _ := redis.String(net.predis.Do("HGET", strKey, "machine"))
		//predis.Do("HGET", key, typeOfT.Field(i).Name, f.Interface())
	}

	return nil
}
