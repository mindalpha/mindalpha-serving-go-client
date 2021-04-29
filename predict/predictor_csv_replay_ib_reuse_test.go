package predict

import (
	"github.com/mindalpha/mindalpha-serving-go-client/fe"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/pool"
	"testing"
	"time"
)

var DefaultTestingRemoteServiceAddr = "cluster.mindalpha-serving.remote.service"

func CommonLoadCsvFile(predictor *Predictor, csv_file, column_name_file string, gen_with_score bool) func(*testing.T) {
	return func(t *testing.T) {
		client_logger.GetMindAlphaServingClientLogger().Errorf("load %s %s with %v", csv_file, column_name_file, gen_with_score)
		nib := fe.NewIndexedColumn(3, 200)
		nib.LoadFromCsvFile(csv_file, column_name_file, gen_with_score)
		client_logger.GetMindAlphaServingClientLogger().Errorf("print debugstring %v", nib.DebugString())
		var base_score_sum float32 = 0.0

		client_logger.GetMindAlphaServingClientLogger().Errorf("fbs pool test start")
		for i := 0; i < 50; i++ {
			predictor.Predict(nib, "demo_model", 100)

			var score_sum float32 = 0.0
			for i := 0; i < predictor.Scores.GetScoreSize(); i++ {
				score_sum += predictor.GetScore(i)
			}

			if 0 == i {
				predictor.DebugScoreString()
				base_score_sum = score_sum
			} else {
				if base_score_sum-score_sum > 0.0001 || score_sum-base_score_sum > 0.0001 {
					client_logger.GetMindAlphaServingClientLogger().Errorf("fbs pool test: i = %v: score_sum not match: ", i)
					predictor.DebugScoreString()
					t.Fatal("fbs pool test: score sum used pool not equal to base score sum not used pool")
				}
			}
			time.Sleep(time.Millisecond * 50)
		}

		nib.Free()

		client_logger.GetMindAlphaServingClientLogger().Errorf("ib/fbs pool test start")
		for i := 0; i < 100; i++ {
			nib = fe.NewIndexedColumn(3, 200)
			nib.LoadFromCsvFile(csv_file, column_name_file, gen_with_score)

			predictor.Predict(nib, "demo_model", 100)

			var score_sum float32 = 0.0
			for i := 0; i < predictor.Scores.GetScoreSize(); i++ {
				score_sum += predictor.GetScore(i)
			}

			if base_score_sum-score_sum > 0.0001 || score_sum-base_score_sum > 0.0001 {
				client_logger.GetMindAlphaServingClientLogger().Errorf("ib/fbs pool test: i = %v: score_sum not match: ", i)
				predictor.DebugScoreString()
				t.Fatal("ib/fbs pool test: score_sum not match")
			}
			time.Sleep(time.Millisecond * 50)

			nib.Free()
		}

		client_logger.GetMindAlphaServingClientLogger().Errorf("multi goroutine ib/fbs pool test start")
		for i := 0; i < 20; i++ {
			go func(i int) {
				for j := 0; j < 300; j++ {
					nib := fe.NewIndexedColumn(3, 200)
					nib.LoadFromCsvFile(csv_file, column_name_file, gen_with_score)

					err := predictor.Predict(nib, "demo_model", 100)

					var score_sum float32 = 0.0
					if nil == predictor.Scores {
						client_logger.GetMindAlphaServingClientLogger().Errorf("multi goroutine ib/fbs pool test: i = %v: score nil, err: %v", i, err)
						t.Fatal("multi goroutine ib/fbs pool test: do not get score: ", err)
						continue
					}
					for i := 0; i < predictor.Scores.GetScoreSize(); i++ {
						score_sum += predictor.GetScore(i)
					}

					if base_score_sum-score_sum > 0.0001 || score_sum-base_score_sum > 0.0001 {
						client_logger.GetMindAlphaServingClientLogger().Errorf("multi goroutine ib/fbs pool test: i = %v: score_sum not match: ", i)
						predictor.DebugScoreString()
						t.Fatal("multi goroutine ib/fbs pool test: score_sum not match")
					}
					time.Sleep(time.Millisecond * 2000)

					nib.Free()
				}
			}(i)
		}
		time.Sleep(time.Second * 150)

	}
}
func TestNewPredictorIndexBatch(t *testing.T) {
	predictor, err := NewPredictor()
	if err != nil {
		panic(err)
	}
	column_name_file := "../data/column_name_criteo.txt"
	fe.LoadColumnNameFile(column_name_file)
	conf := pool.MindAlphaServingClientPoolConfig {
		ConsulAddr:        "127.0.0.1:8500",
		MindAlphaServingService:        DefaultTestingRemoteServiceAddr,
		MaxConnNumPerAddr: 50,
	}
	pool.InitMindAlphaServingClientPool(&conf)
	csv_file := "../data/day_0_0.001_train-ib-format.csv"
	t.Run(csv_file, CommonLoadCsvFile(predictor, csv_file, column_name_file, true))
}
