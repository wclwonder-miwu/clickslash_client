package model

import (
	. "clickslash/protos"
	. "clickslash/utils"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Redisbase struct {
	redisConn redis.Conn
}

var g_redisbase *Redisbase

var (
	redisBase *Redisbase
	redisConn redis.Conn
	once      sync.Once
)

func NewRedisbase() (*Redisbase, redis.Conn) {
	once.Do(func() {
		redisBase = new(Redisbase)
		redisConn = redisBase.init()
		redisBase.redisConn = redisConn
	})

	return redisBase, redisConn
}

func (this *Redisbase) init() redis.Conn {
	var err interface{}
	redisConn, err = redis.Dial("tcp", REDIS_IP)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return redisConn
	//defer c.Close()
}

//获取用户数据
func GetUser(key string) {
	fmt.Println("getuserdata")
}

func (this *Redisbase) CreateMapUser(uid *string) map[string]interface{} {
	retMap := map[string]interface{}{}

	//先获取到struct，redis可能缺一些数据
	tUser := &TUser{}
	RedisGetStruct(this.redisConn, "user:"+*uid+":property", tUser)
	StructCoverMap(tUser, retMap)
	//其他数组字段
	//user:id:gift_bought已购买礼包
	//user:id:collect已收集勋章
	//user:id:map_gift礼包

	propName := []string{"gift_bought", "collectStr", "map_giftStr"}

	for _, v := range propName {
		retStr, err := redis.String(this.redisConn.Do("GET", "user:"+*uid+":"+v))
		if err != nil {
			retMap[v] = []int{}
		} else {
			retMap[v] = strings.Split(retStr, ",")
		}
	}

	return retMap
}

func (this *Redisbase) GetLevelConfig(level string) *TLevelConfig {
	levelCfg := &TLevelConfig{}
	RedisHGetStruct(this.redisConn, "levelConfig:"+level, levelCfg)
	return levelCfg
}

//更新token
func (this *Redisbase) UpdateToken(uid *string, token *string) {

	//更新token的时间
	now := time.Now().Unix()
	this.redisConn.Do("HSET", "user:"+*uid+":token", "time", now)
	this.redisConn.Do("HSET", "user:"+*uid+":token", "token", *token)
}

//检查token
func (this *Redisbase) CheckToken(uid *string, token *string) bool {
	//检查token是否一致和超时
	timeStr, err := redis.String(this.redisConn.Do("HGET", "user:"+*uid+":token", "time"))
	if err != nil {
		return false //没有值
	}

	//时间
	timeLast, _ := strconv.ParseInt(timeStr, 10, 64)
	if time.Now().Unix()-timeLast > int64(time.Hour*TOKEN_TIMEOUT) {
		return false
	}

	//是否一致
	tokenStr, _ := redis.String(this.redisConn.Do("HGET", "user:"+*uid+":token", "token"))
	if !strings.EqualFold(*token, tokenStr) {
		return false
	}

	this.UpdateToken(uid, token)
	return true
}
