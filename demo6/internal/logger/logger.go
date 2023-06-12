package logger

import "envoy-demo4/pkg/logger"

var log logger.Logger

func BootLogger(debug bool) {
	log = newLogrusLogger(debug)
}

func Instance() logger.Logger {
	return log
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

func WithFields(keyValues logger.Fields) logger.Logger {
	return log.WithFields(keyValues)
}
