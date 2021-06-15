package configs

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

// ConfPath 默认配置文件地址，可修改
var ConfPath = "../conf/config.yml"

type Config struct {
	Global Global `yaml:"Global"`
	RpcConfig RpcConfig `yaml:"RpcConfig"`
	// 这里补充全部的配置项
}

type Global struct{
	Env int `yaml:"Env"`
}

type RpcConfig struct{
	CodecType int `yaml:"CodecType"`
	RpcTimeout int `yaml:"RpcTimeout"`
}

var (
	once sync.Once
	Conf = &Config{}
)

func init() {
	once.Do(func() {
		config, err := ioutil.ReadFile(ConfPath)
		if err != nil {
			fmt.Print(err)
		}
		yaml.Unmarshal(config,&Conf)
	})
}
