# golang version MindAlpha-Serving client
[中文文档](README.CN.md)

MindAlpha-Serving service is a online prediction service. This code is a go version MindAlpha-Serving client, This client gives you the ability to access MindAlpha-Serving online prediction service.<br>
If you want to use our online prediction service ***MindAlpha-Serving*** please refer to https://github.com/mindalpha/serving-helm-chart <br>
If you want to use our offline training service ***MindAlpha*** please refer to https://github.com/mindalpha/MindAlpha <br>

## Architecture
![GitHub](pictures/architecture.png "architecture")

## Features
1. Construct IndexBatch(by AddColumn() and AddColumnArray())
2. Use flatbuffer protocal to serialize/deserialize network request/response. The proto defination please refer to [proto](proto/)
3. Load balancing based on consul. MindAlpha-Serving service registers consul service, this client then will see it.
4. Connection pool.
5. Weighted Round-robin load balance.
6. Serialize IndexBatch to string representation (by GetRowFeatures() function)
7. Construct IndexBatch from csv file (by LoadFromCsvFile function)
8. Online Predict, return score slice.

## Usage
The complete steps to use this client: <br>
1. Implement log interface, set this client to use it
2. Config consul and connection pool configuration, for service discovery and load banlance depends on it
3. Initialize connection pool with above configuration
4. Construct IndexBatch
5. Online predict
6. Porcess the returned scores from MindAlpha-Serving service
7. Put IndexBatch to memory pool for performence improvement <br>

the following code shows how to use this client 

```go
import (
	"github.com/cihub/seelog"
	"github.com/mindalpha/mindalpha-serving-go-client/fe"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/pool"
	"testing"
)

//implement LOG interface.
type MyLogger struct {
}

func (myLogger *MyLogger) Debugf(format string, v ...interface{}) {
	seelog.Debugf(format, v...)
}
// set log use MyLogger.
client_logger.SetMindAlphaServingClientLogger(&MyLogger{})

// connection pool configuration.
conf := pool.MindAlphaServingClientPoolConfig{
	ConsulAddr: "127.0.0.1:8500", //consul address
	MindAlphaServingService: "cluster.mindalpha-serving.remote.service", // mindalpha-serving service name on consul
	MaxConnNumPerAddr: 10,
}
// initialize connection pool
pool.InitMindAlphaServingClientPool(&conf)

// generate IndexBatch object.
// the first param must be 3.
ib := fe.NewIndexedColumn(3, 200)

// construct IndexBatch, add columns/features to IndexBatch
// The data used to construct IndexBatch comes from data/day_0_0.001_train-ib-format.csv and data/column_name_criteo.txt. 
// This demo code used the first row of file data/day_0_0.001_train-ib-format.csv as column value, and use the content of data/column_name_criteo.txt as column name.
// To learn more about the data we used to construct IndexBatch, please refer to data/
ib.AddColumn("integer_feature_1", "", 2, 0)
ib.AddColumn("integer_feature_2", "478", 2, 0)
ib.AddColumn("integer_feature_3", "1", 2, 0)
ib.AddColumn("integer_feature_4", "2", 2, 0)
ib.AddColumn("integer_feature_5", "9", 2, 0)
ib.AddColumn("integer_feature_6", "6", 2, 0)
ib.AddColumn("integer_feature_7", "0", 2, 0)
ib.AddColumn("integer_feature_8", "36", 2, 0)
ib.AddColumn("integer_feature_9", "3", 2, 0)
ib.AddColumn("integer_feature_10", "1", 2, 0)
ib.AddColumn("integer_feature_11", "5", 2, 0)
ib.AddColumn("integer_feature_12", "721", 2, 0)
ib.AddColumn("integer_feature_13", "1", 2, 0)

ib.AddColumn("categorical_feature_1", "265366bf", 2, 0)
ib.AddColumn("categorical_feature_2", "b1feb7c7", 2, 0)
ib.AddColumn("categorical_feature_3", "fddc0f59", 2, 0)
ib.AddColumn("categorical_feature_4", "67ecc871", 2, 0)
ib.AddColumn("categorical_feature_5", "4dc31926", 2, 0)
ib.AddColumn("categorical_feature_6", "6fcd6dcb", 2, 0)
ib.AddColumn("categorical_feature_7", "ee3c4dac", 2, 0)
ib.AddColumn("categorical_feature_8", "ab96c6b2", 2, 0)
ib.AddColumn("categorical_feature_9", "25dd8f9a", 2, 0)
ib.AddColumn("categorical_feature_10", "e63d98b4", 2, 0)
ib.AddColumn("categorical_feature_11", "c939136f", 2, 0)
ib.AddColumn("categorical_feature_12", "8490a3ea", 2, 0)
ib.AddColumn("categorical_feature_13", "a77a4a56", 2, 0)
ib.AddColumn("categorical_feature_14", "", 2, 0)
ib.AddColumn("categorical_feature_15", "5cbc7f6a", 2, 0)
ib.AddColumn("categorical_feature_16", "", 2, 0)
ib.AddColumn("categorical_feature_17", "", 2, 0)
ib.AddColumn("categorical_feature_18", "a1eb1511", 2, 0)
ib.AddColumn("categorical_feature_19", "108a0699", 2, 0)
ib.AddColumn("categorical_feature_20", "47849e55", 2, 0)
ib.AddColumn("categorical_feature_21", "73b3f46d", 2, 0)
ib.AddColumn("categorical_feature_22", "d994ba60", 2, 0)
ib.AddColumn("categorical_feature_23", "", 2, 0)
ib.AddColumn("categorical_feature_24", "4dc8c296", 2, 0)
ib.AddColumn("categorical_feature_25", "321935cd", 2, 0)
ib.AddColumn("categorical_feature_26", "2ba8d787", 2, 0)


// generate Predictor
predictor, _ := NewPredictor()

// online predict. "demo_model" is exported model name,
// third param is timeout in milliseconds
predictor.Predict(ib, "demo_model", 100)

// do something with  mindalpha-serving returned scores
score_num := predictor.GetScoreSize()

// do something with scores
// the first score related to the first row of ib
// the second score related to the second row of ib, etc.
for i := 0; i < score_num; i++ {
	score := predictor.GetScore(i)
	client_logger.GetMindAlphaServingClientLogger().Debugf("score[%v] = %v", i, score)
}

// put IndexBatch to sync.Pool when you do not need ib.
ib.Free()

```

**The detail usage please refer to: [demo code](predict/predictor_test.go) and [predict/](predict/)** <br>
**The data used by demo code comes from [data/](data)**

## Data
Data used by this client code. <br>
The above demo code uses data from which to construct IndexBatch. <br>
To learn more about the data information, please refer to **[data/](data/)**

## Submodules
### Log interface
Defination and document of log interface please refer to [logger/](logger/)

### Service discovery / Connection pool
Please refer to [pool](pool/), usage of it please refer to [Predic() method](predict/predictor.go) and test cases under path [predict/](predict). <br>
**Users just need to config the configuration of consul and connection pool when you use this client**

### IndexBatch
The defination and usage of IndexBatch please refer to [fe/](fe/)

### Online predict
The usage of online predict method please refer to [Predic() method](predict/predictor.go) and [demo code](predict/predictor_test.go)

### Proto
Protocal defination: [proto/](proto/) <br>
Generated code: [gen/](gen/) <br>
**Do Not Modify Protocal, it will not compatible with MindAlpha-Serving service**
