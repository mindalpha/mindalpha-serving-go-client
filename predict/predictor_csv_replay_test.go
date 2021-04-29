package predict

import (
	"github.com/mindalpha/mindalpha-serving-go-client/fe"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/pool"
	"testing"
)

var DefaultTestingRemoteServiceAddr = "cluster.mindalpha-serving.remote.service"

func CommonLoadCsvFile(predictor *Predictor, csv_file, column_name_file string, gen_with_score bool) func(*testing.T) {
	return func(t *testing.T) {
		// load column name file
		fe.LoadColumnNameFile(column_name_file)
		client_logger.GetMindAlphaServingClientLogger().Errorf("load %s %s with %v", csv_file, column_name_file, gen_with_score)

		// generate IndexBatch object.
		nib := fe.NewIndexedColumn(3, 200)
		// use features/values in csv file to construct IndexBatch, add columns/features to IndexBatch
		nib.LoadFromCsvFile(csv_file, column_name_file, gen_with_score)

		batch_size := nib.GetBatchSize()
		for i := 0; i < batch_size; i++ {
			row_features := nib.GetRowFeatures(i)
			client_logger.GetMindAlphaServingClientLogger().Errorf("rows[%v]: %v", i, row_features)
		}
		client_logger.GetMindAlphaServingClientLogger().Errorf("print debugstring %v", nib.DebugString())

		// online predict. "demo_model" is exported model name,
		// third param is timeout in milliseconds
		err := predictor.Predict(nib, "demo_model", 50)
		if nil != err {
			t.Fatal("Predict() failed: ", err)
		}
		score_num := predictor.GetScoreSize()
		if score_num != batch_size {
			t.Fatal("score num should be equal to batch size")
		}

		// do someting with scores
		// the first score related to the first row of ib
		// the second score related to the second row of ib, etc.
		for i := 0; i < score_num; i++ {
			score := predictor.GetScore(i)
			client_logger.GetMindAlphaServingClientLogger().Infof("score[%v] = %v", i, score)
		}

		// put IndexBatch to sync.Pool when you do not need ib.
		nib.Free()
	}
}
func TestPredictorIndexBatch(t *testing.T) {
	predictor, err := NewPredictor()
	if err != nil {
		t.Fatal("NewPredictor() failed: ", err)
	}

	// connection pool configuration.
	conf := pool.MindAlphaServingClientPoolConfig{
		ConsulAddr: "127.0.0.1:8500",
		//ConsulAddr:        "vg-consul-aws.rayjump.com:8500",
		MindAlphaServingService:        DefaultTestingRemoteServiceAddr,
		MaxConnNumPerAddr: 5,
	}
	// initialize connection pool
	pool.InitMindAlphaServingClientPool(&conf)

	column_name_file := "../data/column_name_criteo.txt"
	csv_file := "../data/day_0_0.001_train-ib-format.csv"
	t.Run(csv_file, CommonLoadCsvFile(predictor, csv_file, column_name_file, true))
}
