/*go**************************************************************************
 File            : RedisUtil.go
 Subsystem       :
 Author          : wcl
 Date&Time       : 2017-06-06
 Description     : 数据类型的转化
 Revision        :

 History
 -------


 Copyright (c) Shenzhen Team Blemobi.
**************************************************************************go*/
package utils

import (
	"fmt"
	"reflect"

	"github.com/garyburd/redigo/redis"
)

//读取redis的map到go的map
func Redis2Map(conn redis.Conn, key string, m map[string]interface{}) error {
	//获取键名
	cachedKeys, err := redis.Strings(conn.Do("HKEYS", key))
	fieldsLen := len(cachedKeys)

	params := make([]interface{}, 3)
	params[0] = key
	for _, v := range cachedKeys {
		params = append(params, v)
	}

	fmt.Println(params)
	//获取所有值
	reply, err := redis.Values(conn.Do("HMGET", params...))
	if err != nil {
		Warning("RedisGetStruct fail, param=%v, err=%s", params, err.Error())
		return err
	}

	for i := 0; i < fieldsLen; i++ {
		val := reply[i]
		switch val.(type) {
		case nil:
		default:
			m[cachedKeys[i]] = reply[i]
		}
	}

	return nil
}

//结构字段放到map
func StructCoverMap(obj interface{}, m map[string]interface{}) {
	values := reflect.ValueOf(obj).Elem()
	types := values.Type()

	for i := 0; i < values.NumField(); i++ {
		m[types.Field(i).Name] = values.Field(i).Interface()
	}
}
