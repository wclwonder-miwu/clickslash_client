package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	. "clickslash/protos"

	"github.com/garyburd/redigo/redis"
)

type NetAPI struct {
	x      int
	predis redis.Conn
}

func (this *NetAPI) Init(predis redis.Conn) {
	this.predis = predis

}

func (net *NetAPI) MainHandler(w http.ResponseWriter, req *http.Request) {
	//获取客户端通过GET/POST方式传递的参数
	req.ParseForm()

	fmt.Println("MainHandler COMES")
	param_r, found1 := req.Form["r"]

	if !found1 {
		fmt.Fprint(w, "请勿非法访问")
		return
	}

	var strRet string = ""
	switch param_r[0] {
	case "register/create":
		account, _ := req.Form["account"]
		strRet = net.createAccount(&account[0])
	case "login":
		uid, _ := req.Form["uid"]
		password, _ := req.Form["password"]
		strRet = net.onLogin(&uid[0], &password[0])
	case "user":
		access_token, _ := req.Form["access_token"]
		strRet = net.onUser(&access_token[0])
		//strRet=fmt.Sprintf("%s","{\"ret\":0,\"user\":{\"uid\":2786,\"coin\":0,\"energy\":15,\"max_energy\":15,\"last_recover\":1491965576,\"day_recover\":0,\"last_star\":0,\"cost_star\":0,\"energy_buf\":0,\"gift_bought\":[],\"collect\":[],\"map_gift\":[],\"medal_lv\":0,\"icon\":3,\"stars\":0,\"nick\":\"冒险者2786\",\"level\":0,\"exp\":0,\"comCardIndex\":0,\"bind\":0,\"exchange\":0,\"skin\":\"\"},\"blocks\":[],\"items\":[],\"cur_level\":0}")
	case "item/gift":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"data\":[],\"last_time\":0}")
	case "user/check-energy":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"energy\":15,\"last_recover\":1491965576,\"cur_time\":1491965580,\"day_recover\":0}")
	case "mail":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"total\":\"0\",\"num\":0,\"data\":[]}")
	case "user/auditlist-friend":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"auditList\":[]}")
	case "user/energylist-friend":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"energy_list\":[]}")
	case "user/getenergylist-friend":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"energy_list\":[]}")
	case "user/list-friend":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"friend_list\":[]}")
	case "play/login":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"day\":1,\"reward\":[{\"itemid\":100,\"itemnum\":20,\"NAME\":\"金币\"}]}")
	case "user/change-info":
		strRet = fmt.Sprintf("%s", "{\"ret\":0,\"coin\":20}")
	case "play/begin":
		id, _ := req.Form["id"]
		numid, _ := strconv.Atoi(id[0])
		access_token, _ := req.Form["access_token"]
		strRet = net.onBegin(numid, &access_token[0])
	case "play":
		access_token, _ := req.Form["access_token"]
		strRet = net.onPlay(&access_token[0])
	default:
		w.Write([]byte("error"))
	}

	w.Write([]byte(strRet))
}

func (net *NetAPI) createAccount(machine *string) string {
	//是否已经存在
	var strErr string = fmt.Sprintf("%s", "{\"ret\":1}")
	if net.isUserExist(machine) {
		fmt.Println("UserExist=" + *machine)
		return strErr
	}

	id_count, err := redis.String(net.predis.Do("GET", "id_count"))
	if err != nil {
		net.predis.Do("SET", "id_count", 0)
		id_count = "1"
	}
	net.predis.Do("INCR", "id_count")

	//创建并返回id，密码。创建用户属性表
	strKey := fmt.Sprintf("user%sproperty", id_count)
	net.predis.Do("HSET", strKey, "machine", *machine)
	net.predis.Do("HSET", strKey, "password", "WVlSi01E")
	net.initAccountData(id_count)

	tempMap := make(map[string]interface{})
	tempMap["ret"] = 0
	tempMap["uid"], _ = strconv.Atoi(id_count)
	tempMap["password"] = GetRandomString(10)

	str, err := json.Marshal(tempMap)
	if err != nil {
		fmt.Println(err)
	}

	return string(str)
}

func (this *NetAPI) initAccountData(uid string) {
	temp := &TUser{}
	fmt.Println(temp)
}

//用户是否存在
func (net *NetAPI) isUserExist(machine *string) bool {
	values, err := redis.Values(net.predis.Do("lrange", "id_list", "0", "-1"))
	fmt.Println("isUserExist1")

	if err != nil {
		fmt.Println(err)
		return true
	}

	fmt.Println("isUserExist2")
	//机器码是否存在
	for _, v := range values {
		//检查每个用户的ID

		strKey := fmt.Sprintf("user%sproperty", string(v.([]uint8)[0]))
		str, _ := redis.String(net.predis.Do("HGET", strKey, "machine"))

		if strings.EqualFold(str, *machine) {
			fmt.Println("user already exist")
			return true
		}
	}

	return false
}

func (net *NetAPI) onLogin(uid *string, password *string) string {

	strKey := fmt.Sprintf("user%sproperty", *uid)
	fmt.Println(strKey)

	password_server, _ := redis.String(net.predis.Do("HGET", strKey, "password"))

	md5Ctx := md5.New()
	md5Ctx.Write([]byte(password_server))
	cipherStr := md5Ctx.Sum(nil)
	password_server = hex.EncodeToString(cipherStr)

	if !strings.EqualFold(*password, password_server) {
		var strErr string = fmt.Sprintf("%s", "{\"ret\":1}")
		return strErr
	}

	tempMap := make(map[string]interface{})
	tempMap["ret"] = 0
	tempMap["access_token"] = *uid + "#" + GetRandomString(20)

	str, err := json.Marshal(tempMap)
	if err != nil {
		fmt.Println(err)
	}

	return string(str)

	//return fmt.Sprintf("%s","{\"ret\":0,\"access_token\":\""+GetRandomString(37)+"2786#Njb2jhRCzXJlKQqSRbQy3rWVY23QqvzH\"}")
}

func (this *NetAPI) onUser(access_token *string) string {
	str := strings.Split(*access_token, "#")
	uid := str[0]

	strKey := fmt.Sprintf("user%sproperty", uid)

	tempMap := make(map[string]interface{})
	tempMap["ret"] = 0
	tempMap["blocks"] = []int{}
	tempMap["items"] = []int{}
	tempMap["cur_level"] = 0

	user := createMapUser()

	values, _ := redis.StringMap(this.predis.Do("HGETALL", strKey))
	for k, v := range values {
		user[k] = v
	}
	tempMap["user"] = user

	str1, err := json.Marshal(tempMap)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("showlog=")
	fmt.Println(string(str1))

	return string(str1)
}

//关卡开始，只要扣体力
func (this *NetAPI) onBegin(level int, access_token *string) string {
	str := strings.Split(*access_token, "#")
	uid := str[0]
	strKey := fmt.Sprintf("user%sproperty", uid)

	//剩余体力
	temp, _ := redis.String(this.predis.Do("HGET", strKey, "energy"))
	energy, _ := strconv.Atoi(temp)

	tempMap := make(map[string]interface{})
	tempMap["id"] = level

	if energy >= 3 {
		tempMap["ret"] = 0
	} else {
		tempMap["ret"] = 0
	}

	user := createMapUser()
	//user info
	values, _ := redis.StringMap(this.predis.Do("HGETALL", strKey))
	for k, v := range values {
		user[k] = v
	}
	tempMap["user"] = user

	//返回
	str1, err := json.Marshal(tempMap)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("showlog=")
	fmt.Println(string(str1))

	return string(str1)
}

//关卡结算
func (this *NetAPI) onPlay(access_token *string) string {
	str := strings.Split(*access_token, "#")
	uid := str[0]
	strKey := fmt.Sprintf("user%sproperty", uid)

	tempMap := make(map[string]interface{})
	tempMap["ret"] = 0
	tempMap["pass"] = 1

	user := createMapUser()
	//user info
	values, _ := redis.StringMap(this.predis.Do("HGETALL", strKey))
	for k, v := range values {
		user[k] = v
	}
	tempMap["user"] = user

	//关卡数据
	blocks := []map[string]interface{}{
		map[string]interface{}{
			"pig_save": 2, "draw_count": 0, "ex": 3, "id": 1, "score": 17509,
		}}
	tempMap["blocks"] = blocks

	addUserData(tempMap)

	//返回
	str1, err := json.Marshal(tempMap)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("showlog=")
	fmt.Println(string(str1))

	return string(str1)
}
