package configs

import (
	"fmt"
	"sync"

	"github.com/coomp/ccs/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Cfg TODO
var (
	Cfg interface{}
	rw  sync.RWMutex
)

func init() {
	// 把配置文件读取到结构体上
	viper.SetConfigName("config") // 配置文件的文件名，没有扩展名，如 .yaml, .toml 这样的扩展名
	viper.SetConfigType("yaml")   // 设置扩展名。在这里设置文件的扩展名。另外，如果配置文件的名称没有扩展名，则需要配置这个选项
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 获取fsm 后续所有的配置从这里拿
	Cfg = viper.Get("fsms")
	fmt.Println("cfg", Cfg)
	// 获取变动
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		rw.Lock()
		Cfg = viper.Get("fsms")
		rw.Unlock()
		log.L.Debug("config are changing")
	})
}
