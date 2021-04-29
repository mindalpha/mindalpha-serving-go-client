package logger

import (
	"fmt"
)

// Logger for log
type MindAlphaServingClientLogger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type DefaultMindAlphaServingClientLogger struct {
}

func (myLogger *DefaultMindAlphaServingClientLogger) Debugf(format string, v ...interface{}) {
	err := fmt.Errorf(format, v...)
	fmt.Println(err.Error())
}
func (myLogger *DefaultMindAlphaServingClientLogger) Infof(format string, v ...interface{}) {
	err := fmt.Errorf(format, v...)
	fmt.Println(err.Error())
}
func (myLogger *DefaultMindAlphaServingClientLogger) Warnf(format string, v ...interface{}) {
	err := fmt.Errorf(format, v...)
	fmt.Println(err.Error())
}
func (myLogger *DefaultMindAlphaServingClientLogger) Errorf(format string, v ...interface{}) {
	err := fmt.Errorf(format, v...)
	fmt.Println(err.Error())
}

var clientLogger MindAlphaServingClientLogger

func init() {
	var defaultLogger *DefaultMindAlphaServingClientLogger = &DefaultMindAlphaServingClientLogger{}
	SetMindAlphaServingClientLogger(defaultLogger)

}

func SetMindAlphaServingClientLogger(lg MindAlphaServingClientLogger) {
	clientLogger = lg
}

func GetMindAlphaServingClientLogger() MindAlphaServingClientLogger {
	return clientLogger
}
