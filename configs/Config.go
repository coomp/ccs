package configs

import (
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

// ConfPath 默认配置文件地址，可修改
var ConfPath = "../conf/config.yml"

// Config TODO
type Config struct {
	Global    Global    `yaml:"Global"`
	RpcConfig RpcConfig `yaml:"RpcConfig"`
	// 这里补充全部的配置项
}

// Global TODO
type Global struct {
	Env int `yaml:"Env"`
}

// RpcConfig TODO
type RpcConfig struct {
	CodecType  int    `yaml:"CodecType"`
	RpcTimeout int    `yaml:"RpcTimeout"`
	NetType    string `yaml:"NetType"`
	Key        string `yaml:"Key"`
	Address    string `yaml:"address"`
}

var (
	once sync.Once
	// Conf TODO
	Conf = &Config{}
)

func init() {
	once.Do(func() {
		config, err := ioutil.ReadFile(ConfPath)
		if err != nil {
			fmt.Print(err)
		}
		yaml.Unmarshal(config, &Conf)
	})
}
