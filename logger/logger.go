package logger

const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL

	pTrace = "[TRACE]"
	pDebug = "[DEBUG]"
	pInfo  = "[INFO]"
	pWarn  = "[WARN]"
	pError = "[ERROR]"
	pFatal = "[FATAL]"
)

var LOG_LEVEL_PREFIX_MAP map[int]string = map[int]string{
	TRACE: pTrace,
	DEBUG: pDebug,
	INFO:  pInfo,
	WARN:  pWarn,
	ERROR: pError,
	FATAL: pFatal,
}

type Logger interface {
	Trace(records ...interface{})
	Debug(records ...interface{})
	Info(records ...interface{})
	Warn(records ...interface{})
	Error(records ...interface{})
	Fatal(records ...interface{})

	Tracef(format string, records ...interface{})
	Debugf(format string, records ...interface{})
	Infof(format string, records ...interface{})
	Warnf(format string, records ...interface{})
	Errorf(format string, records ...interface{})
	Fatalf(format string, records ...interface{})

	Prefix(prefix string)
	Format(format int)

	// create new logger
	WithPrefix(prefix string) Logger
	WithFormat(format int) Logger
}
