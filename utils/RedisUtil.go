/*go**************************************************************************
 File            : RedisUtil.go
 Subsystem       :
 Author          : Frank
 Date&Time       : 2016-01-14
 Description     : 封装redis的一些接口调用
 Revision        :

 History
 -------


 Copyright (c) Shenzhen Team Blemobi.
**************************************************************************go*/
package utils

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

/*
 *@note 验证枚举参数长度是否比结构体变量数目多
		这个函数应该在程序启动的时候调用，如果遇到错误，会直接panic
 *@param obj 结构信息
 *@param len 验证的参数
 *@return 无
*/
func VeifyStructFieldsLen(obj interface{}, len int) {
	values := reflect.ValueOf(obj).Elem()
	if values.NumField() != len {
		panic("struct len wrong")
	}
}

/*
 *@note redisUtil支持的结构信息操作的类型
 *@param kind 结构反射后得到的数据类型
 *@return 无
 */
func supportTypeCheck(kind reflect.Kind) {
	switch kind {
	case reflect.Bool:
		return
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return
	case reflect.Float32, reflect.Float64:
		return
	case reflect.String:
		return
	case reflect.Slice:
		return
	}

	panic(fmt.Sprintf("supportTypeCheck fail, kind=%v", kind))
}

/*
 *@note redisUtil支持的结构信息操作的类型
 *@param v reflect.Value
 *@param reply redis.do 返回的结果
 *@param err redis.do 返回的结果
 *@return 支持返回true，不支持返回false
 */
func setTypeValue(v reflect.Value, reply interface{}, err error) {
	switch v.Type().Kind() {
	case reflect.Bool:
		x, e := redis.Bool(reply, err)
		if e != nil {
			panic(e.Error())
		}
		v.SetBool(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, e := redis.Int64(reply, err)
		if e != nil {
			panic(e.Error())
		}
		v.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, e := redis.Uint64(reply, err)
		if e != nil {
			panic(e.Error())
		}
		v.SetUint(x)
	case reflect.Float32, reflect.Float64:
		x, e := redis.Float64(reply, err)
		if e != nil {
			panic(e.Error())
		}
		v.SetFloat(x)
	case reflect.String:
		x, e := redis.String(reply, err)
		if e != nil {
			panic(e.Error())
		}
		v.SetString(x)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			x, e := redis.Bytes(reply, err)
			if e != nil {
				panic(e.Error())
			}
			v.SetBytes(x)
		}
	default:
		panic("setTypeValue unsupport kind")
	}
}

/*
 *@note 以hash方式存储结构的全部信息到redis
 *@param conn redis连接
 *@param key 存储键值
 *@param obj 结构信息
 *@return error，成功返回nil，失败返回非nil
 */
func RedisSetStruct(conn redis.Conn, key string, obj interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("RedisSetStruct fail, err=%s", r)
			err = errors.New("RedisSetStruct Fail")
		}
	}()

	values := reflect.ValueOf(obj).Elem()
	types := values.Type()

	param := make([]interface{}, 1)
	param[0] = key

	for i := 0; i < values.NumField(); i++ {
		v := values.Field(i)
		kind := v.Type().Kind()

		supportTypeCheck(kind)

		// 空的字符串不做存库处理
		value := v.Interface()
		if kind == reflect.String && len(value.(string)) == 0 {
			continue
		}

		param = append(param, types.Field(i).Name, value)
	}

	_, err = conn.Do("HMSET", param...)
	if err != nil {
		Warning("RedisSetStruct save param:%v fail, err=(%s)", param, err.Error())
		return err
	}

	return nil
}

/*
 *@note 以hash方式存储结构的部分信息到redis
 *@param conn redis连接
 *@param key 存储键值
 *@param obj 结构信息
 *@param fields 结构中需要保存项的idx，具体可参考login/login-common中的结构定义
 *@return error，成功返回nil，失败返回非nil
 */
func RedisHSetStruct(conn redis.Conn, key string, obj interface{}, fields ...int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			Warning("RedisHSetStruct fail, err=%s", r)
			err = errors.New("RedisHSetStruct Fail")
		}
	}()

	values := reflect.ValueOf(obj).Elem()
	types := values.Type()

	param := make([]interface{}, 1)
	param[0] = key

	for _, field := range fields {
		v := values.Field(field)
		t := types.Field(field)
		if !v.IsValid() {
			panic(fmt.Sprintf("RedisHSetStruct !v.IsValid(), field=%d", field))
		}

		kind := v.Type().Kind()
		supportTypeCheck(kind)

		//value := v.Interface()
		//		if kind == reflect.String && len(value.(string)) == 0 {
		//			continue
		//		}

		param = append(param, t.Name, v.Interface())
	}

	_, err = conn.Do("HMSET", param...)
	if err != nil {
		Warning("RedisHSetStruct save param:%v fail, err=(%s)", param, err.Error())
		return err
	}

	return nil
}

/*
 *@note 从redis中获取结构的全部信息
 *@param conn redis连接
 *@param key 存储键值
 *@param obj 结构信息
 *@return error，成功返回nil，失败返回非nil
 */
func RedisGetStruct(conn redis.Conn, key string, obj interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			Warning("RedisGetStruct fail, err=%s", r)
			err = errors.New("RedisHGetStruct Fail")
		}
	}()

	values := reflect.ValueOf(obj).Elem()
	types := values.Type()
	fieldsLen := values.NumField()

	param := make([]interface{}, 1+fieldsLen)
	param[0] = key

	for i := 0; i < fieldsLen; i++ {
		t := types.Field(i)

		supportTypeCheck(t.Type.Kind())

		param[i+1] = t.Name
	}

	reply, err := redis.Values(conn.Do("HMGET", param...))
	if err != nil {
		Warning("RedisGetStruct fail, param=%v, err=%s", param, err.Error())
		return err
	}

	for i := 0; i < fieldsLen; i++ {
		val := reply[i]
		switch val.(type) {
		case nil:
		default:
			setTypeValue(values.Field(i), reply[i], nil)
		}
	}

	return nil
}

/*
 *@note 从redis中获取结构的部分信息
 *@param conn redis连接
 *@param key 存储键值
 *@param obj 结构信息
 *@param fields 结构中需要保存项的idx，具体可参考login/login-common中的结构定义
 *@return error，成功返回nil，失败返回非nil
 */
func RedisHGetStruct(conn redis.Conn, key string, obj interface{}, fields ...int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			Warning("RedisHGetStruct fail, err=%s", r)
			err = errors.New("RedisHGetStruct Fail")
		}
	}()

	values := reflect.ValueOf(obj).Elem()
	types := values.Type()
	fieldsLen := len(fields)

	param := make([]interface{}, 1+fieldsLen)
	param[0] = key

	for i := 0; i < fieldsLen; i++ {
		t := types.Field(fields[i])

		supportTypeCheck(t.Type.Kind())

		param[i+1] = t.Name
	}

	reply, err := redis.Values(conn.Do("HMGET", param...))
	if err != nil {
		Warning("RedisGetStruct fail, param=%v, err=%s", param, err.Error())
		return err
	}

	for i := 0; i < fieldsLen; i++ {
		val := reply[i]
		switch val.(type) {
		case nil:
		default:
			setTypeValue(values.Field(fields[i]), val, nil)
		}
	}

	return nil
}

/*
 *@note 从redis中删除hash的信息
 *@param conn redis连接
 *@param req 预期删除项的总数
 *@param arg hash表中的hash键值
 *@return error，成功返回nil，失败返回非nil
 */
func redisHDel(conn redis.Conn, req int, arg ...interface{}) (err error) {
	del, err := redis.Int(conn.Do("HDEL", arg...))
	if err != nil {
		Warning("redisHDel fail, err=%s", err)
		return err
	}

	if del != req {
		Warning("redisHDel del key num not equal to fields num!")
	}

	return nil
}

/*
 *@note 从redis中删除结构的完整信息
 *@param conn redis连接
 *@param key redis中hash的键值
 *@param obj 结构信息
 *@return error，成功返回nil，失败返回非nil
 */
func RedisDelStruct(conn redis.Conn, key string, obj interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			Warning("RedisDelStruct fail, err=%s", r)
			err = errors.New("RedisDelStruct Fail")
		}
	}()

	arg := []interface{}{key}

	values := reflect.ValueOf(obj).Elem()
	types := values.Type()
	fieldsNum := values.NumField()

	for i := 0; i < fieldsNum; i++ {
		k := types.Field(i).Name

		arg = append(arg, k)
	}

	return redisHDel(conn, fieldsNum, arg...)
}

/*
 *@note 从redis中删除结构的部分信息
 *@param conn redis连接
 *@param key redis中hash的键值
 *@param obj 结构信息
 *@param fields 结构中需要保存项的idx，具体可参考login/login-common中的结构定义
 *@return error，成功返回nil，失败返回非nil
 */
func RedisHDelStruct(conn redis.Conn, key string, obj interface{}, fields ...int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			Warning("RedisHGetStruct fail, err=%s", r)
			err = errors.New("RedisHGetStruct Fail")
		}
	}()

	values := reflect.ValueOf(obj).Elem()
	types := values.Type()

	arg := []interface{}{key}
	for _, field := range fields {
		arg = append(arg, types.Field(field).Name)
	}

	return redisHDel(conn, len(fields), arg...)
}

/*
 *@note 从redis中获取hash值
 *@param conn redis连接
 *@param args HGET的参数
 *@return interface{}，err，interface{}中存储HGET返回的值
 */
func redisHGet(conn redis.Conn, args ...interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			Error("RedisHGet fail, err=%s", r)
		}
	}()

	return conn.Do("HGET", args...)
}

/*
 *@note 从redis中获取hash的字符串值
 *@param conn redis连接
 *@param args HGET的参数
 *@return string，err，如果成功，返回要取的字符串的值，error为nil；否则error不为nil
 */
func RedisHGetString(conn redis.Conn, args ...interface{}) (string, error) {
	return redis.String(redisHGet(conn, args...))
}

/*
 *@note 从redis中获取hash的整形值
 *@param conn redis连接
 *@param args HGET的参数
 *@return int，err，如果成功，返回要取的整形的值，error为nil；否则error不为nil
 */
func RedisHGetInt(conn redis.Conn, args ...interface{}) (int, error) {
	return redis.Int(redisHGet(conn, args...))
}

func RedisHGetInt64(conn redis.Conn, args ...interface{}) (int64, error) {
	return redis.Int64(redisHGet(conn, args...))
}

type CRedisPool struct {
	rp              *redis.Pool // redigo pool连接池实体
	mu              sync.Mutex  // 保护errDialCount的读写正确
	addr            string      // 拨号地址
	errDialCount    int         // redigo pool连接失败次数
	maxErrDialCount int         // 程序设置的允许连接失败次数，如果为0，则永远不panic
}

/*
 *@note 在redigo pool的dial后根据err进行判断逻辑
		当maxErrDialCount为0的时候，不做任何处理
		当累计连接失败次数（errDialCount）大于上限（maxErrDialCount）的时候，panic报错
		errDialCount在连接成功之后，次数会清空
 *@return error，成功返回nil，失败返回非nil
*/
func (this *CRedisPool) OnDial(err error) {
	if this.maxErrDialCount == 0 {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	if nil != err {
		this.errDialCount = this.errDialCount + 1
		if this.errDialCount > this.maxErrDialCount {
			Error("redis pool dial fail...")
			panic("redis pool dial fail...")
		}
	} else {
		if this.errDialCount > 0 {
			Warning("redis pool redial success")
			this.errDialCount = 0
		}
	}
}

/*
 *@note 获取一个redis的连接
 *@return redis.Conn
 */
func (this *CRedisPool) Get() redis.Conn {
	return this.rp.Get()
}

/*
 *@note 默认的redispool拨号函数
 *@param rp redis连接池
 *@param addr 拨号地址
 *@param maxIdle 连接池最大空闲连接数
 *@param maxActive 连接池最大连接数
 *@param opt 拨号选项
 *@return redis.Pool
 */
func DefRedisPoolDialFun(rp *CRedisPool, addr string, maxIdle, maxActive int, wait bool, idelTimeout time.Duration, opt ...redis.DialOption) *redis.Pool {
	return &redis.Pool{
		Wait:        wait,
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idelTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr, opt...)
			rp.OnDial(err)
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

/*
 *@note 单元测试时redis拨号选项
 *@param c redigomock连接对象
 *@return 无
 */
func EmptyTestRedisDialOpt(c *redigomock.Conn) {
}

// 设置redis单元测试拨号选项
var TestRedisDialOpt = EmptyTestRedisDialOpt

/*
 *@note 单元测试的redispool拨号函数
 *@param rp redis连接池
 *@param addr 拨号地址
 *@param maxIdle 连接池最大空闲连接数
 *@param maxActive 连接池最大连接数
 *@param opt 拨号选项
 *@return redis.Pool
 */
func TestRedisPoolDialFun(rp *CRedisPool, addr string, maxIdle, maxActive int, wait bool, opt ...redis.DialOption) *redis.Pool {
	return &redis.Pool{
		Wait:      wait,
		MaxIdle:   0,
		MaxActive: 1,
		Dial: func() (redis.Conn, error) {
			c := redigomock.NewConn()
			c.Command("PING").Expect("PONG")
			TestRedisDialOpt(c)
			return c, nil
		},
	}
}

// 设置redis拨号函数
var RedisPoolDialFun = DefRedisPoolDialFun

/*
 *@note 创建一个redispool
 *@param addr 拨号地址
 *@param opt 拨号选项
 *@return *CRedisPool，
 */
func NewRedisPool(addr string, opt ...redis.DialOption) *CRedisPool {
	return NewRedisPoolEx(addr, 2000, 0, 0, true, time.Second*30, opt...)
}

/*
 *@note 创建一个redispool
 *@param addr 拨号地址
 *@param maxIdle 连接池最大空闲连接数
 *@param maxActive 连接池最大连接数
 *@param maxErrDialCount 连接池最大连接失败次数
 *@param opt 拨号选项
 *@return *CRedisPool，
 */
func NewRedisPoolEx(addr string, maxIdle, maxActive, maxErrDialCount int, wait bool, idelTimeout time.Duration, opt ...redis.DialOption) *CRedisPool {
	this := &CRedisPool{
		addr:            addr,
		maxErrDialCount: maxErrDialCount,
	}

	this.rp = RedisPoolDialFun(
		this, addr, maxIdle, maxActive, wait, idelTimeout, opt...)

	conn := this.Get()
	defer conn.Close()

	ret, err := redis.String(conn.Do("PING"))
	if err != nil {
		panic("NewRedisPoolEx, Do(PING) fail, err=" + err.Error())
	}

	if ret != "PONG" {
		panic("NewRedisPoolEx, get ret is not equal PONG")
	}

	return this
}
