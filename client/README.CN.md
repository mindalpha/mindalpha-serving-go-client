# client
[英文文档](README.md)

## 功能
序列化IndexBatch
构造网络请求
反序列化服务端返回的scores数据.


## 测试用例
1. client/client_test.go: <br>
    *  从序列化后的flatbuffer request文件读取Request 直接网络发送给MindAlpha-Serving服务并打印score
    *  从csv文件读取数据构造IndexBatch 并请求MindAlpha-Serving，打印score

## 待完善
1. dumpFBS()
    *  该功能类似一次性使用的功能。在需要dumpFBS时需要修改代码调用一下dumpFBS, dumpFBS会创建一个文件并dump出多个Request/Response. br>
    目前的实现有多个问题. <br>
    问题一: 硬编码dump Request/Response的个数,硬编码文件路径; <br>
    问题二: 需要修改代码调用一次dumpFBS; <br>
    问题三: dumpFBS dump出指定数量的request/response后就再也不会再dump了，除非重启进程.
    问题四: 至于是否需要dump response 用户需要根据自己需要来决定.
