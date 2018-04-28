package global

import (
	"chatServer/mylog"
	"encoding/xml"
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
	return nil
}

//Config 配置
type Config struct {
	ListenIP   string `xml:"listenIP"`
	ListenPort uint16 `xml:"listenPort"`
}
