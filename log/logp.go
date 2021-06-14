package log

import "github.com/coomp/ccs/log/zap"

//默认
var L ILog = zap.New()

//设置
func SetLogger(ll ILog) {
	L = ll
}

//普通日志
func Debug(msg string, args ...interface{}) {
	L.Debug(msg, args...)
}
func Info(msg string, args ...interface{}) {
	L.Info(msg, args...)
}
func Warn(msg string, args ...interface{}) {
	L.Warn(msg, args...)
}
func Error(msg string, args ...interface{}) {
	L.Error(msg, args...)
}
func Panic(msg string, args ...interface{}) {
	L.Panic(msg, args...)
}
func Fatal(msg string, args ...interface{}) {
	L.Fatal(msg, args...)
}

//其他日志 如：HTTP RPC日志
func Debugf(format string, args ...interface{}) {
	L.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	L.Infof(format, args...)
}
func Warnf(format string, args ...interface{}) {
	L.Warnf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	L.Errorf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	L.Panicf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	L.Fatalf(format, args...)
}
