package master

import (
	"encoding/json"
	"io/ioutil"
)

// 配置
type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeout  int      `json:"apiReadTimeout"`
	ApiWriteTimeout int      `json:"apiWriteTimeout"`
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	WebRoot         string   `json:"webroot"`
}

var (
	Conf *Config
)

func InitConfig(filename string) error {
	var (
		content []byte
		conf    Config
		err     error
	)

	// 读取配置文件
	if content, err = ioutil.ReadFile(filename); err != nil {
		return err
	}

	// 反序列化
	if err = json.Unmarshal(content, &conf); err != nil {
		return err
	}

	Conf = &conf

	return nil
}
