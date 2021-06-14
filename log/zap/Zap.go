package zap
import (
	Default "coomp/log/default"
	"coomp/log/fileout"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

type Log struct {
	logger *zap.Logger
}

func (l Log) Debug(s string, i ...interface{}) {
	l.logger.Debug(s)
}

func (l Log) Info(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Warn(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Error(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Panic(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Fatal(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Debugf(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Infof(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Warnf(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Errorf(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Panicf(s string, i ...interface{}) {
	panic("implement me")
}

func (l Log) Fatalf(s string, i ...interface{}) {
	panic("implement me")
}

//var Log *zap.Logger //全局日志

func parseLevel(lvl string) zapcore.Level {
	switch strings.ToLower(lvl) {
	case "panic", "dpanic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.DebugLevel
	}
}

//创建日志
func New(opts ...Default.Option) *Log {
	o := &Default.Options{
		LogPath:     Default.LogPath,
		LogName:     Default.LogName,
		LogLevel:    Default.LogLevel,
		MaxSize:     Default.MaxSize,
		MaxAge:      Default.MaxAge,
		Stacktrace:  Default.Stacktrace,
		IsStdOut:    Default.IsStdOut,
		ProjectName: Default.ProjectName,
	}
	for _, opt := range opts {
		opt(o)
	}
	writers := []zapcore.WriteSyncer{fileout.NewRollingFile(o.LogPath, o.LogName, o.MaxSize, o.MaxAge)}
	if o.IsStdOut == "yes" {
		writers = append(writers, os.Stdout)
	}
	logger := newZapLogger(parseLevel(o.LogLevel), parseLevel(o.Stacktrace), zapcore.NewMultiWriteSyncer(writers...))
	zap.RedirectStdLog(logger)
	logger = logger.With(zap.String("project", o.ProjectName)) //加上项目名称
	return &Log{logger: logger}
}

func newZapLogger(level, stacktrace zapcore.Level, output zapcore.WriteSyncer) *zap.Logger {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "app",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		//EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		//	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		//},
		EncodeTime: zapcore.ISO8601TimeEncoder,
	}

	var encoder zapcore.Encoder
	dyn := zap.NewAtomicLevel()
	//encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	//encoder = zapcore.NewJSONEncoder(encCfg) // zapcore.NewConsoleEncoder(encCfg)
	dyn.SetLevel(level)
	encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder = zapcore.NewJSONEncoder(encCfg)

	return zap.New(zapcore.NewCore(encoder, output, dyn), zap.AddCaller(), zap.AddStacktrace(stacktrace), zap.AddCallerSkip(2))
}
