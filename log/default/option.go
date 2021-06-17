package Default

// Option TODO
type Option func(*Options)

// Options TODO
type Options struct {
	LogPath     string //日志保存路径
	LogName     string //日志保存的名称，不些随机生成
	LogLevel    string //日志记录级别
	MaxSize     int    //日志分割的尺寸 MB
	MaxAge      int    //分割日志保存的时间 day
	Stacktrace  string //记录堆栈的级别
	IsStdOut    string //是否标准输出console输出
	ProjectName string //项目名称
}

// WithLogPath TODO
func WithLogPath(logpath string) Option {
	return func(o *Options) {
		o.LogPath = logpath
	}
}

// WithLogName TODO
func WithLogName(logname string) Option {
	return func(o *Options) {
		o.LogName = logname
	}
}

// WithLogLevel TODO
func WithLogLevel(loglevel string) Option {
	return func(o *Options) {
		o.LogLevel = loglevel
	}
}

// WithMaxSize TODO
func WithMaxSize(maxsize int) Option {
	return func(o *Options) {
		o.MaxAge = maxsize
	}
}

// WithMaxAge TODO
func WithMaxAge(maxage int) Option {
	return func(o *Options) {
		o.MaxAge = maxage
	}
}

// WithStacktrace TODO
func WithStacktrace(stacktrace string) Option {
	return func(o *Options) {
		o.Stacktrace = stacktrace
	}
}

// WithIsStdOut TODO
func WithIsStdOut(isstdout string) Option {
	return func(o *Options) {
		o.IsStdOut = isstdout
	}
}

// WithProjectName TODO
func WithProjectName(projectname string) Option {
	return func(o *Options) {
		o.ProjectName = projectname
	}
}
