package logger

import (
	"github.com/mdouchement/shigoto/pkg/logger"
	"github.com/sirupsen/logrus"
)

type wrapper struct {
	logrus *logrus.Entry
}

// WrapLogrus returns Logger based on Logrus backend.
func WrapLogrus(l *logrus.Logger) logger.Logger {
	return &wrapper{
		logrus: logrus.NewEntry(l),
	}
}

func (w *wrapper) WithField(key string, value interface{}) logger.Logger {
	return &wrapper{
		logrus: w.logrus.WithField(key, value),
	}
}

func (w *wrapper) Printf(format string, args ...interface{}) {
	w.logrus.Printf(format, args...)
}

func (w *wrapper) Info(args ...interface{}) {
	w.logrus.Info(args...)
}

func (w *wrapper) Infof(format string, args ...interface{}) {
	w.logrus.Infof(format, args...)
}

func (w *wrapper) Error(args ...interface{}) {
	w.logrus.Error(args...)
}

func (w *wrapper) Errorf(format string, args ...interface{}) {
	w.logrus.Errorf(format, args...)
}
