# logger 
[English Document](README.md)

客户端日志接口.

## 日志接口的使用
MindAlphaServingClientLogger提供了日志接口，并提供了默认的实现(直接打印到标准输出)和基于seelog、logrus的实现示例代码。
日志定义了如下的接口
```
type MindAlphaServingClientLogger interface {
    Debugf(format string, v ...interface{})
    Infof(format string, v ...interface{})
    Warnf(format string, v ...interface{})
    Errorf(format string, v ...interface{})
} 
```
 
日志使用步骤(可以参考logger/logger_seelog_test.go 和 logger/logger_logrus_test.go)：
1. 定义type mylogger struct {} 实现上述的MindAlphaServingClientLogger接口
2. 调用logger.SetMindAlphaServingClientLogger(mylogger)， 设置日志库使用我们在步骤1中实现的接口
3. 使用上述接口logger.GetMindAlphaServingClientLogger().Debugf() 打日志
