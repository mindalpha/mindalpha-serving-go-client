package tensor

import "C"
import (
	"github.com/mindalpha/mindalpha-serving-go-client/gen/mindalpha_serving"
)

type TensorScores struct {
	scores_size  int
	scores_slice []float32
	version      string
}

func NewTensorScores() *TensorScores {
	var ts TensorScores
	ts.scores_size = 0
	return &ts
}

func (ts *TensorScores) ParseTensorScores(buf []byte) {
	resp := mindalpha_serving.GetRootAsResponse(buf, 0)
	ts.version = string(resp.Version())
	tensor := new(mindalpha_serving.Tensor)
	resp.Tensor(tensor)
	ts.scores_size = tensor.ScoresLength()
	ts.scores_slice = make([]float32, ts.scores_size)
	for i := 0; i < ts.scores_size; i++ {
		ts.scores_slice[i] = tensor.Scores(i)
	}
}
func (ts *TensorScores) GetVersion() string {
	return ts.version
}
func (ts *TensorScores) GetScoreSize() int {
	return ts.scores_size
}
func (ts *TensorScores) GetScoreUnCheck(idx int) float32 {
	return ts.scores_slice[idx]
}

func (ts *TensorScores) GetScore(idx int) float32 {
	if idx >= ts.scores_size {
		panic("GetScore: index exceed total score nums")
	}
	return ts.GetScoreUnCheck(idx)
}
