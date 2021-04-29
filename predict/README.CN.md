# predictor
[English Document](README.md) <br>

在线预测服务客户端接口

## 功能
1. 提供在线预测Predict()接口, 返回score slice
2. 该接口内部利用了该客户端实现的连接池及负载均衡功能

## 用法
下面示例代码描述本接口的用法.

```go
//构造IndexBatch, 构造IndexBatch方法参考 fe/ 或者 predict/目录中的测试文件.


// 生成一个Predictor对象
predictor, err := NewPredictor()

//调用预测在线接口
err = predictor.Predict(ib, "demo_model", 100)

//获得MindAlpha-Serving服务端返回的score
for i := 0; i < predicotor.GetScoreSize(); i++ {
    score := predicotor.GetScore(i)
    //do something with score
}

```

## 测试用例
1. predict/predictor_csv_replay_test.go <br>
    *  从csv文件读取数据构造IndexBatch 并请求MindAlpha-Serving，打印score
    *  该测试文件展示了如何从 csv 文件中读取特征数据并将其构造为IndexBatch, 然后请求MindAlpha-Serving服务在线预测score的基本流程.
    *  该测试文件中用到的csv_file 和 column_name_file 内容及格式说明参考 **[数据文件 data/](data/)**
2. predict/predictor_test.go <br>
    *  调用AddColumn构造一个 3 level，39 column, 1 row 的IndexBatch，
    *  构造IndexBatch所使用的数据,也就是AddColumn()方法所使用的参数, 来自数据文件 **[data/day_0_0.001_train-ib-format.csv](data/day_0_0.001_train-ib-format.csv)** 的第一行和特征列名字文件 **[data/column_name_criteo.txt](/data/column_name_criteo.txt)**.
3. predict/predictor_csv_replay_ib_reuse_test.go <br>
    *  测试IndexBatch内存池. IndexBatch使用完毕后调用Free方法将底层存储放到内存池中。 该测试用例不断的构造IndexBatch(通过fe.NewIndexedColumn()方法，该方法可能会从内存池中获取内存)并调用Free方法, 然后检查每次构造的IndexBatch通过Predict方法在线预测打分是否一致.

用户可以从上面的测试用例中了解该客户端的完整用法.

## 函数说明
1. func NewPredictor() (*Predictor, error) <br>
    *  生成一个Predictor对象, 后续会在该对象上执行在线预测接口Predict(), 并在该对象上操作MindAlpha-Serving返回的score数据.
2. func (t *Predictor) Predict(ib *fe.IndexedColumn, algo_name string, timeout int) error <br>
    *  在线预测接口. 该接口会将ib 用flatbuffer序列化后向MindAlpha-Serving服务端发起在线预测请求. <br>
    *  algo_name 指定模型的名称, 该名称是离线训练导出的模型的名字 <br>
    *  timeout 指定Predict()的超时时间.单位为毫秒. <br>
    *  如果发生错误或者超时，则返回非nil的error. <br>
    *  Predict()函数如果成功了，则会获得服务端返回的scores. 这些score与IndexBatch中的每一行一一对应, scores[0]是IndexBatch 第0行的预测值，scores[1]是IndexBatch 第1行的预测值. <br>

3. func (t *Predictor) GetScoreSize() int <br>
    *  返回score的个数.正常情况下, score 的个数跟 IndexedColumn.GetBatchSize()的返回值是相等的.
4. func (t *Predictor) GetScore(i int) float32 <br>
    *  获取第 i 个score. 该score对应的是IndexBatch的第 i 行.

## 待完善
