package logger

import (
	"github.com/cihub/seelog"
	"testing"
)

type MyLogger struct {
}

func (myLogger *MyLogger) Debugf(format string, v ...interface{}) {
	seelog.Debugf(format, v...)
}
func (myLogger *MyLogger) Infof(format string, v ...interface{}) {
	seelog.Infof(format, v...)
}
func (myLogger *MyLogger) Warnf(format string, v ...interface{}) {
	seelog.Warnf(format, v...)
}
func (myLogger *MyLogger) Errorf(format string, v ...interface{}) {
	seelog.Errorf(format, v...)
}

func TestErrorf(t *testing.T) {
	defer seelog.Flush()
	SetMindAlphaServingClientLogger(&MyLogger{})

	GetMindAlphaServingClientLogger().Debugf("DEBUG")
	GetMindAlphaServingClientLogger().Infof("INFO")
	GetMindAlphaServingClientLogger().Warnf("WARN")
	GetMindAlphaServingClientLogger().Errorf("ERROR")
}
