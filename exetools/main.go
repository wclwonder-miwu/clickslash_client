/*
以文件名:ID的hash表导入、保存
*/

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	//"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

var redisConn redis.Conn

const REDIS_IP = "127.0.0.1:6379"

func main() {
	fmt.Println("main start")

	initRedis()
	readCSV()
}

func initRedis() error {
	var err error
	redisConn, err = redis.Dial("tcp", REDIS_IP)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func readCSV() {
	file, err := os.Open("levelConfig.csv")

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return
	}

	title := lines[1]
	var count int = 0

	//遍历行
	for _, line := range lines[2:] {
		var ID string = ""
		params := []interface{}{}

		//读取列，字段
		cols := []interface{}{}
		for j, str := range line {
			cols = append(cols, title[j], str)
			if ID == "" && strings.EqualFold(title[j], "ID") {
				ID = str
			}
		}

		params = append(params, "levelConfig:"+ID)
		params = append(params, cols...)

		//fmt.Println("插入一行")
		//fmt.Println(params)
		redisConn.Do("HMSET", params...)

		count++
		if count >= 2 {
			//break
		}
	}

	//levelConfig:1
}
