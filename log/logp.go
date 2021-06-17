package log

import "github.com/coomp/ccs/log/zap"

// L 默认
var L ILog = zap.New()

// SetLogger 设置
func SetLogger(ll ILog) {
	L = ll
}

// Debug 普通日志
func Debug(msg string, args ...interface{}) {
	L.Debug(msg, args...)
}

// Info TODO
func Info(msg string, args ...interface{}) {
	L.Info(msg, args...)
}

// Warn TODO
func Warn(msg string, args ...interface{}) {
	L.Warn(msg, args...)
}

// Error TODO
func Error(msg string, args ...interface{}) {
	L.Error(msg, args...)
}

// Panic TODO
func Panic(msg string, args ...interface{}) {
	L.Panic(msg, args...)
}

// Fatal TODO
func Fatal(msg string, args ...interface{}) {
	L.Fatal(msg, args...)
}

// Debugf TODO
// 其他日志 如：HTTP RPC日志
func Debugf(format string, args ...interface{}) {
	L.Debugf(format, args...)
}

// Infof TODO
func Infof(format string, args ...interface{}) {
	L.Infof(format, args...)
}

// Warnf TODO
func Warnf(format string, args ...interface{}) {
	L.Warnf(format, args...)
}

// Errorf TODO
func Errorf(format string, args ...interface{}) {
	L.Errorf(format, args...)
}

// Panicf TODO
func Panicf(format string, args ...interface{}) {
	L.Panicf(format, args...)
}

// Fatalf TODO
func Fatalf(format string, args ...interface{}) {
	L.Fatalf(format, args...)
}
