package ecs

type Logger interface {
	LogInfo(str... interface{})
	LogJson(str string, obj interface{})
}
