package config

import (
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego/logs"
	yaml "gopkg.in/yaml.v2"
)

// ConfigData 配置文件数据
var ConfigData map[interface{}]interface{}

// GetYamlConfig 解析conf配置文件
func GetYamlConfig(path string) map[interface{}]interface{} {
	data, err := ioutil.ReadFile(path)
	m := make(map[interface{}]interface{})
	if err != nil {
		logs.Error(err)
	}

	err = yaml.Unmarshal([]byte(data), &m)
	fmt.Println(m)
	return m
}

func InitConfig(path string) {
	ConfigData = GetYamlConfig(path)
}
