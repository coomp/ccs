package UniqueID

import (
	"coomp/log"
	"fmt"
	"github.com/xuyu/goredis"
	"io/ioutil"
	"sync"
)

var luaScript_order string

var (
	once sync.Once
)

func init() {
	once.Do(func() {
		content, err := ioutil.ReadFile("./flowing.lua")
		if err != nil {
			panic(fmt.Sprintf("reading file script failed, err: %v", err))
		}
		luaScript_order = string(content)
	})
}

// Getflowing 获取一个全局的id
func Getflowing() int64 {
	client, err := goredis.Dial(&goredis.DialConfig{Address: "127.0.0.1:6379"})
	if err != nil {
		log.L.Error("conn redis err:%s", err.Error())
		return -1
	}
	evalParams := []interface{}{luaScript_order, 1, RedisKey}
	r, err := client.Eval(luaScript_order, nil, nil)
	if err != nil {

	}
	return -1
}
