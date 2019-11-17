package config

import (
	"encoding/json"
	"log"
	"os"
)

type conf struct {
	TSLCertFile string // tsl cert file
	TSLKeyFile string // tsl key file
	WssAllowOrigin []string // websocket允许的来源域名
	HttpAddr string // http接口的Addr:[domain:port]
	TokenTimeOut int64 // token的过期时间（秒）
}

// 公共配置变量
var Config conf

// 读取配置文件
func ReadConfig(configPath string) {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal("open config file err:", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("close config file err:", err)
		}
	}()

	Config = conf{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		log.Fatal("decode config file err:", err)
	}
}