# logger 
[中文文档](README.CN.md)

log interface

## Usage of log interface
MindAlphaServingClientLogger defines the log interface, and gives the default implementation(print to stdout) and seelog/logrus based implementation. <br>
MindAlphaServingClientLogger defines the following log interface <br>
```
type MindAlphaServingClientLogger interface {
    Debugf(format string, v ...interface{})
    Infof(format string, v ...interface{})
    Warnf(format string, v ...interface{})
    Errorf(format string, v ...interface{})
} 
```
 
The step of using log interface (logger/logger_seelog_test.go 和 logger/logger_logrus_test.go) is as follows: <br>
1. define type mylogger struct {}, implement the above MindAlphaServingClientLogger interface.
2. call logger.SetMindAlphaServingClientLogger(mylogger), let mindalpha-serving-go-client use you log implementation.
3. call logger.GetMindAlphaServingClientLogger().Debugf() logger.GetMindAlphaServingClientLogger().Infof() to log logs.
You can refer to [logger_seelog_test.go](/logger/logger_seelog_test.go) and [logger_logrus_test.go](/logger/logger_logrus_test.go).
