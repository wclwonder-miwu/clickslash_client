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

func (this *Redisbase) CreateMapUser(uid *string) map[string]interface{} {
	retMap := map[string]interface{}{}

	//先获取到struct，redis可能缺一些数据
	tUser := &TUser{}
	RedisGetStruct(this.redisConn, "user:"+*uid+":property", tUser)
	StructCoverMap(tUser, retMap)

	delete(retMap, "password")
	//其他数组字段
	//user:id:gift_bought已购买礼包
	//user:id:collect已收集勋章
	//user:id:map_gift礼包

	propName := []string{"gift_bought", "collect", "map_gift"}

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

//拼接道具信息数组
func (this *Redisbase) GetPropsData(uid *string) []interface{} {
	//可能多个字段，可能没有
	ret := []interface{}{}
	strKey := fmt.Sprintf("user:%s:props", *uid)
	keys, err := redis.Values(this.redisConn.Do("HKEYS", strKey))
	if err != nil {
		return ret
	}

	if len(keys) > 0 {
		for _, v := range keys {
			retMap := map[string]interface{}{}
			key, _ := redis.String(v, err)

			itemid, _ := strconv.Atoi(key)
			retMap["id"] = itemid
			value, _ := redis.Int64(this.redisConn.Do("HGET", strKey, key))
			retMap["num"] = value
			ret = append(ret, retMap)
		}
	}
	return ret
}

func (this *Redisbase) GetGiftData() []interface{} {
	ret := []interface{}{}
	return ret
}

func (this *Redisbase) GetBlocksData(uid *string, level1, level2 int) []interface{} {
	ret := []interface{}{}

	if level1 == -1 {
		level1 = 1
		level222, err := redis.Int(this.redisConn.Do("GET", "user:"+*uid+"cur_level"))
		if err != nil {
			level222 = 1
		}
		level2 = level222
	}

	for i := level1; i <= level2; i++ {
		levelSave := &TLevelSave{}
		strKey := fmt.Sprintf("user:%s:blocks:%d", *uid, i)
		RedisGetStruct(this.redisConn, strKey, levelSave)
		levelMap := map[string]interface{}{}
		StructCoverMap(levelSave, levelMap)

		ret = append(ret, levelMap)
	}
	return ret
}

func (this *Redisbase) GetLevelConfig(level string) *TLevelConfig {
	levelCfg := &TLevelConfig{}
	RedisGetStruct(this.redisConn, "levelConfig:"+level, levelCfg)
	return levelCfg
}

func (this *Redisbase) AddUserData(ret map[string]interface{}) {
	//额外的一些数据
	ret["reward_id"] = 0
	ret["gift_back_time"] = 0
	ret["pick_reward_id"] = 0
	ret["reward_double"] = 0
	ret["pig_card"] = 0
}

//如果打了新的一关，更新当前关，返回当前关
func (this *Redisbase) UpateCurLevel(uid *string, level int) int {
	strKey := fmt.Sprintf("user:%s:cur_level", *uid)
	cur_level, err := redis.Int(this.redisConn.Do("GET", strKey))
	if err != nil {
		cur_level = 1
	}
	if level >= cur_level {
		cur_level = level
		this.redisConn.Do("SET", strKey, strconv.Itoa(level))
		fmt.Println("update curlevel=%d")
	} else {
		fmt.Println("play old level")
	}
	return cur_level
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
