package logger

import (
	"errors"
	"github.com/sirupsen/logrus"
	"testing"
)

type MyLogger struct {
}

func (myLogger *MyLogger) Init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	logrus.SetReportCaller(false)
}

func (myLogger *MyLogger) Debugf(format string, v ...interface{}) {
	logrus.Debugf(format, v...)
}
func (myLogger *MyLogger) Infof(format string, v ...interface{}) {
	logrus.Infof(format, v...)
}
func (myLogger *MyLogger) Warnf(format string, v ...interface{}) {
	logrus.Warnf(format, v...)
}
func (myLogger *MyLogger) Errorf(format string, v ...interface{}) {
	logrus.Errorf(format, v...)
}

func TestErrorf(t *testing.T) {

	//use default logger "fmt"
	GetMindAlphaServingClientLogger().Errorf("test default logger. %v", errors.New("test new error"))

	// use logrus
	var myLogger *MyLogger = &MyLogger{}
	myLogger.Init()
	SetMindAlphaServingClientLogger(&MyLogger{})
	SetMindAlphaServingClientLogger(myLogger)
	GetMindAlphaServingClientLogger().Debugf("DEBUG")
	GetMindAlphaServingClientLogger().Infof("INFO")
	GetMindAlphaServingClientLogger().Warnf("WARN")
	GetMindAlphaServingClientLogger().Errorf("ERROR")
}
