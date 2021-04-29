package client

import (
	net_error "github.com/mindalpha/mindalpha-serving-go-client/error"
	"github.com/mindalpha/mindalpha-serving-go-client/fe"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/pool"
	"testing"
	"time"
)

var (
	mindalphaServingService = "cluster.mindalpha-serving.remote.service"
	consulAddr = "127.0.0.1:8500"
)

func TestIndexBatchAndClient(t *testing.T) {
	ib := fe.NewIndexBatch(3, 3)
	ib.LoadFromCsvFile("../data/day_0_0.001_train-ib-format.csv", "../data/column_name_criteo.txt", true)

	_, fbs, err := GenerateMindAlphaServingRequest(ib, "demo_model")
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("GenerateMindAlphaServingRequest() error: %v", err)
	}

	poolConfig := &pool.MindAlphaServingClientPoolConfig{
		MaxConnNumPerAddr: 2,
		MindAlphaServingService:        mindalphaServingService,
		ConsulAddr:        consulAddr,
	}

	p, _ := pool.NewChannelPool(poolConfig)
	v, err := p.Get()
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("pool.Get() error: %v", err)
		t.Errorf("Get error: %s", err)
		return
	}
	if _, ok := v.(*pool.ConnWrap); !ok {
		client_logger.GetMindAlphaServingClientLogger().Errorf("p.Get() return not ConnWrap")
		t.Errorf("p.Get() return not ConnWrap")
	}

	rspChan, err := Request(fbs, v.(*pool.ConnWrap), time.Now().Add(100*time.Millisecond))
	p.Put(v, err)
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("after Request(), err: %v", err)

		if netErr, ok := err.(*net_error.NetError); ok {
			client_logger.GetMindAlphaServingClientLogger().Errorf("Got Net Error: %v", netErr)
		}
	}

	rspData, err := WaitResponse(rspChan, 100*1000*1000)
	if err != nil {
		t.Errorf("WaitResponse() failed: %v", err)
		return
	}
	ts, err := ParseMindAlphaServingResponse(rspData)
	client_logger.GetMindAlphaServingClientLogger().Errorf("after ParseMindAlphaServingResponse(), TensnrScores{}: %v", ts)

	if ts != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("after Request(), score num: %v", ts.GetScoreSize())
		var score_sum float32 = 0.0
		for i := 0; i < ts.GetScoreSize(); i++ {
			score_sum += ts.GetScore(i)
			client_logger.GetMindAlphaServingClientLogger().Errorf("score[%v] = %v", i, ts.GetScore(i))
		}
		client_logger.GetMindAlphaServingClientLogger().Errorf("score sum: %v", score_sum)
	}
}
