package logger

// A Logger is the interface used in this package for logging,
// so that any backend can be plugged in.
type Logger interface {
	WithField(key string, value interface{}) Logger
	Printf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}
