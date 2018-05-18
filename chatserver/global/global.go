package global

import (
	"chatserver/mylog"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

// AppLog 日志对象指针
var AppLog = mylog.NewLog(10*1024*1024, "./log")

//AppConfig 配置对象
var AppConfig Config

//LoadConfig 加载配置文件
func LoadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(data, &AppConfig)
	if err != nil {
		return err
	}
	return verifyConfig(&AppConfig)
}

func verifyConfig(c *Config) error {
	if len(c.HTTPListenIP) == 0 {
		return fmt.Errorf("httpListenIP is empty")
	}
	if len(c.ListenIP) == 0 {
		return fmt.Errorf("listenIP is empty")
	}
	if len(c.WsListenIP) == 0 {
		return fmt.Errorf("wsListenIP is empty")
	}
	if len(c.RedisConfig.RedisSplice) == 0 {
		return fmt.Errorf("not found redis config")
	}
	if len(c.MysqlConfig.MysqlSplice) == 0 {
		return fmt.Errorf("not found mysql config")
	}
	return nil
}

//Config 配置
type Config struct {
	ListenIP       string       `xml:"listenIP"`
	ListenPort     uint16       `xml:"listenPort"`
	HTTPListenIP   string       `xml:"httpListenIP"`
	HTTPListenPort uint16       `xml:"httpListenPort"`
	WsListenIP     string       `xml:"wsListenIP"`
	WsListenPort   uint16       `xml:"wsListenPort"`
	RedisConfig    RedisServers `xml:"redisServers"`
	MysqlConfig    MysqlServers `xml:"mysqlServers"`
}

//Redis redis配置
type Redis struct {
	Name       string `xml:"name"`
	IP         string `xml:"ip"`
	Port       uint16 `xml:"port"`
	DB         string `xml:"db"`
	MaxConnect uint16 `xml:"maxconnect"`
	Passwd     string `xml:"passwd"`
}

//RedisServers 配置
type RedisServers struct {
	RedisSplice []Redis `xml:"redis"`
}

//MysqlServers 配置
type MysqlServers struct {
	MysqlSplice []Mysql `xml:"mysql"`
}

//Mysql mysql配置
type Mysql struct {
	Name       string `xml:"name"`
	Dsn        string `xml:"dsn"`
	MaxConnect uint16 `xml:"maxconnect"`
}
