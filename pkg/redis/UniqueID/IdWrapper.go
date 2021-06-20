package UniqueID

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/coomp/ccs/log"
	"github.com/xuyu/goredis"
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

// Getflowing 获取一个全局的id ,请求参数是给与一个随机id,业务自己保证下唯一性(雪花？xid？都可以),作为操作的幂等保证,这边会给一个针对全局的流水号
func Getflowing(flowingId string) int64 {
	client, err := goredis.Dial(&goredis.DialConfig{Address: "127.0.0.1:6379"})
	if err != nil {
		log.L.Error("conn redis err:%s", err.Error())
		return -1
	}
	r, err := client.Eval(luaScript_order, []string{flowingId}, nil)
	if err != nil {
		log.L.Error("Eval redis err:%s", err.Error())
		return -1
	}
	if final, rerr := r.IntegerValue(); rerr != nil {
		log.L.Error("Eval redis IntegerValue err:%s", err.Error())
		return -1
	} else {
		return final
	}
}
