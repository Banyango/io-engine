package server

import (
	"encoding/json"
	"fmt"
)

type ServerLogger struct {

}

func (self *ServerLogger) LogInfo(str ...interface{}) {
	fmt.Println(str)
}

func (self *ServerLogger) LogJson(str string, obj interface{}) {
	marshal, _ := json.Marshal(obj)
	self.LogInfo(str, string(marshal))
}

