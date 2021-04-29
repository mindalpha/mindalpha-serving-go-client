package predict

import (
	"github.com/cihub/seelog"
	"github.com/mindalpha/mindalpha-serving-go-client/fe"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/pool"
	"testing"
)

var DefaultTestingRemoteServiceAddr = "cluster.mindalpha-serving.remote.service"

//implement LOG interface.
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

func AddIndexBatch(ib *fe.IndexedColumn) {
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
}

func TestPredictorAddColumn(t *testing.T) {
	defer seelog.Flush()
	// set log use MyLogger.
	client_logger.SetMindAlphaServingClientLogger(&MyLogger{})

	// connection pool configuration.
	conf := pool.MindAlphaServingClientPoolConfig{
		ConsulAddr: "127.0.0.1:8500", //consul address
		MindAlphaServingService: DefaultTestingRemoteServiceAddr, // mindalpha-serving service name on consul
		MaxConnNumPerAddr: 10,
	}
	// initialize connection pool
	pool.InitMindAlphaServingClientPool(&conf)

	// generate IndexBatch object.
	ib := fe.NewIndexedColumn(3, 200)
	// construct IndexBatch, add columns/features to IndexBatch
	AddIndexBatch(ib)

	// generate Predictor
	predictor, err := NewPredictor()
	if err != nil {
		panic(err)
	}
	// online predict. "demo_model" is exported model name,
	// third param is timeout in milliseconds
	err = predictor.Predict(ib, "demo_model", 100)
	if nil != err {
		t.Fatal("Predict() failed: ", err)
	}
	// do something with  mindalpha-serving returned scores
	batch_size := ib.GetBatchSize()
	score_num := predictor.GetScoreSize()

	if batch_size != score_num {
		t.Fatal("batch size should equal to score num")
	}

	// do something with scores
	// the first score related to the first row of ib
	// the second score related to the second row of ib, etc.
	for i := 0; i < score_num; i++ {
		score := predictor.GetScore(i)
		client_logger.GetMindAlphaServingClientLogger().Infof("score[%v] = %v", i, score)
	}

	// put IndexBatch to sync.Pool when you do not need ib.
	ib.Free()
}
