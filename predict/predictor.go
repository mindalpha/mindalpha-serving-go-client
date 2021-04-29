package predict

import (
	"errors"
	"github.com/mindalpha/mindalpha-serving-go-client/client"
	net_error "github.com/mindalpha/mindalpha-serving-go-client/error"
	"github.com/mindalpha/mindalpha-serving-go-client/fe"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/pool"
	"github.com/mindalpha/mindalpha-serving-go-client/tensor"
	"time"
)

type Predictor struct {
	Scores *tensor.TensorScores
}

func NewPredictor() (*Predictor, error) {
	var predictor Predictor
	return &predictor, nil
}

// online predict.
// This method will serialize the ib and then send it to mindalpha-serving server, and wait
// response, deserialize response to score slice.
// param ib: IndexBatch you constructed.
// param algo_name: model name
// param timeout: timeout time in milliseconds.
func (t *Predictor) Predict(ib *fe.IndexedColumn, algo_name string, timeout int) error {
	start1 := time.Now()
	endTimeForAll := start1.Add(time.Duration(timeout) * time.Millisecond)
	builder, fbs, err := client.GenerateMindAlphaServingRequest(ib.Ib, algo_name)
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("GenerateMindAlphaServingRequest() error: %v", err)
		return err
	}
	defer client.PutFBSBuilder(builder)

	p, err := pool.GetConnPoolInstance()
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("GetConnPoolInstance() error: %v", err)
		return errors.New("Predict(): GetConnPoolInstance() error: " + err.Error())
	}
	v, err := p.Get()
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Predict(): get connection from pool failed: %v", err)
		return err
	}
	if nil == v {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Predict(): get nil connection: %v", err)
		return err
	}

	defer func() {
		p.Put(v, err) //Note: we use "err" as param of p.Put, so follow code should set error message to variable "err".
	}()

	if _, ok := v.(*pool.ConnWrap); !ok {
		client_logger.GetMindAlphaServingClientLogger().Errorf("predictor.go: p.Get() return not ConnWrap")
		err = errors.New("Predict(): p.Get() return not ConnWrap")
		return err
	}

	if endTimeForAll.Before(time.Now()) {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Predict(): endtime before now")
		err = &net_error.TimeoutError{TOErr: errors.New("Predict() end time before current time")}
		return err
	}

	rspChan, err := client.Request(fbs, v.(*pool.ConnWrap), endTimeForAll)
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Predict(): call request failed: %v", err)
		return err
	}

	timeForRsp := time.Now()
	if endTimeForAll.Before(timeForRsp) {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Predict(): no time left to wait for response")
		err = &net_error.TimeoutError{TOErr: errors.New("Predict(): no time left to wait for response")}
		return err
	}

	leftTimeForRsp := endTimeForAll.Sub(timeForRsp)
	rspData, err := client.WaitResponse(rspChan, leftTimeForRsp)
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Predict(): WaitResponse error: %v", err)
		return err
	}
	//client.DumpFBS(fbs, rspData)
	t.Scores, err = client.ParseMindAlphaServingResponse(rspData)
	return err
}
func (t *Predictor) GetScore(i int) float32 {
	return t.Scores.GetScore(i)
}
func (t *Predictor) GetScoreSize() int {
	return t.Scores.GetScoreSize()
}
func (t *Predictor) DebugScoreString() {
	client_logger.GetMindAlphaServingClientLogger().Errorf("predictor.scores: %v", t.Scores)
	if t.Scores != nil {
		var score_sum float32 = 0.0
		for i := 0; i < t.Scores.GetScoreSize(); i++ {
			score_sum += t.GetScore(i)
			client_logger.GetMindAlphaServingClientLogger().Errorf("score[%v] = %v", i, t.GetScore(i))
		}
		client_logger.GetMindAlphaServingClientLogger().Errorf("score sum: %v", score_sum)
	}
}
