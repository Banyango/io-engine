package web

type WebLogger struct {

}

func (self *WebLogger) LogInfo(str ...interface{}) {
	//js.Global().Get("console").Call("log", str...)
}

func (self *WebLogger) LogJson(str string, obj interface{}) {
	//marshal, _ := json.Marshal(obj)
	//self.LogInfo(str, string(marshal))
}


