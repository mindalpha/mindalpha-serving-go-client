# predictor
[中文文档](README.CN.md) <br>

online predict interface for client

## Features
1. provides online predict interface, returns score slice.
2. uses connection pool and load balance to access MindAlpha-Serving service.

## Usage

```go
//construct IndexBatch. The method to constuct IndexBatch please refer to test case files under [fe/](fe/) and [predict/](predict/) directory.

// new a Predictor object
predictor, err := NewPredictor()

// call online predict interface
err = predictor.Predict(ib, "demo_model", 100)

// get the returned score from MindAlpha-Serving service and process it
for i := 0; i < predicotor.GetScoreSize(); i++ {
    score := predicotor.GetScore(i)
    //do something with score
}

```

## Test Cases
1. predict/predictor_csv_replay_test.go <br>
    Construct IndexBatch from csv file, then request MindAlpha-Serving service and print the scores. <br>
    The format of csv_file and column_name_file used by this test file please refer to **[data/](/data/)**
2. predict/predictor_test.go <br>
    Calls AddColumn to construct a 3 level, 39 column, 1 row IndexBatch. <br>
    The data used to construct IndexBatch comes from **[data/day_0_0.001_train-ib-format.csv](/data/day_0_0.001_train-ib-format.csv)** and **[data/column_name_criteo.txt](/data/column_name_criteo.txt)**. <br>
    This test case uses the first row of file **data/day_0_0.001_train-ib-format.csv** as column value, and uses the content of **data/column_name_criteo.txt** as column name.
3. predict/predictor_csv_replay_ib_reuse_test.go <br>
    Test IndexBatch memory pool. When you do not need the IndexBatch, you should put it to memory pool.

Users can learn how to use this mindalpha-serving-go-client from the above test cases.

## Functions
1. func NewPredictor() (*Predictor, error) <br>
    New a Predictor object. we will use this object to do the online predict.
2. func (t *Predictor) Predict(ib *fe.IndexedColumn, algo_name string, timeout int) error <br>
    Online predict interface. This function will serialize ib use flatbuffer protocal, and then request MindAlpha-Serving service to get scores.<br>
    Param algo_name is the name of model your offline trainning exported. <br>
    Param timeout is the timeout time in milliseconds. <br>
    if errors happend or timeout, then will return a non-nil error. <br>
    if succed , Predict() will get the returned scores from MindAlpha-Serving service. These scores relates to the rows of IndexBatch, scores[0] is the score of the IndexBatch's first row, scores[1] is the score of the IndexBatch's second row, etc. <br>

3. func (t *Predictor) GetScoreSize() int <br>
    Get the number of scores returned from MindAlpha-Serving service. the score numbershould be equal to IndexedColumn.GetBatchSize().
4. func (t *Predictor) GetScore(i int) float32 <br>
    Get the i'th score. this score relates to the i'th row of IndexBatch.

