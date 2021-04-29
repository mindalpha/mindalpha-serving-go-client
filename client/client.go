package client

import (
	"bufio"
	"encoding/binary"
	"errors"
	flatbuffers "github.com/google/flatbuffers/go"
	net_error "github.com/mindalpha/mindalpha-serving-go-client/error"
	"github.com/mindalpha/mindalpha-serving-go-client/fe"
	"github.com/mindalpha/mindalpha-serving-go-client/gen/mindalpha_serving"
	msg_hdr "github.com/mindalpha/mindalpha-serving-go-client/message_header"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"github.com/mindalpha/mindalpha-serving-go-client/pool"
	"github.com/mindalpha/mindalpha-serving-go-client/tensor"
	"os"
	"sync"
	"time"
)

func uint64ToByte(num uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, num)

	return b
}

var _dump_fbs_file string = "/tmp/drs_fbs.bin"
var _dump_fbs_writer *bufio.Writer = nil

const _dump_fbs_max_count uint64 = 1000

var _dump_fbs_count_mutex = sync.Mutex{}
var _dump_fbs_count uint64 = 0
var _dump_fbs_once sync.Once

//flatbuffer pool.
var fbsBldPool = sync.Pool{
	New: func() interface{} {
		return flatbuffers.NewBuilder(250 * 1024)
	},
}

//TODO
// for debug
// dump req_fbs and rsp_fbs to file.
// req_fbs is flatbuffer encoded data, rsp_fbs is flatbuffer encoded data returned from server side.
// User can read from dump file, get req_fbs and send it to server, then check if the response from server is equal to rsp_fbs.
func DumpFBS(req_fbs []byte, rsp_fbs []byte) error {
	_dump_fbs_once.Do(func() {
		f, err := os.Create(_dump_fbs_file)
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("DumpFBS() os.Create(%v) failed: %v", _dump_fbs_file, err.Error())
		} else {
			_dump_fbs_writer = bufio.NewWriter(f)
		}
	})

	if _dump_fbs_count >= _dump_fbs_max_count {
		return nil
	}

	if _dump_fbs_writer != nil {
		//dump Request:
		var req_fbs_len uint64 = uint64(len(req_fbs))

		//client_logger.GetMindAlphaServingClientLogger().Errorf(" req_fbs len: %v, rsp_fbs len: %v", req_fbs_len, len(rsp_fbs))

		_dump_fbs_count_mutex.Lock()

		_, err := _dump_fbs_writer.Write(uint64ToByte(req_fbs_len))
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("dumpFBS(): Write(req_fbs_len: %v failed: %v", req_fbs_len, err.Error())
			return errors.New("dumpfbs(): Write(req_fbs_len) failed")
		}

		_, err = _dump_fbs_writer.Write(req_fbs)
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("dumpFBS(): Write(req_fbs) failed: %v", err.Error())
			return errors.New("dumpfbs(): Write(req_fbs) failed")
		}

		//dump Response:
		var rsp_fbs_len uint64 = uint64(len(rsp_fbs))
		_, err = _dump_fbs_writer.Write(uint64ToByte(rsp_fbs_len))
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("dumpFBS(): Write(rsp_fbs_len) failed: %v", err.Error())
			return errors.New("dumpfbs(): Write(rsp_fbs_len) failed")
		}
		_, err = _dump_fbs_writer.Write(rsp_fbs)
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("dumpFBS(): Write(rsp_fbs) failed: %v", err.Error())
			return errors.New("dumpfbs(): Write(rsp_fbs) failed")
		}

		//end
		_dump_fbs_count++
		if _dump_fbs_count >= _dump_fbs_max_count {
			//TODO: close file and io writer.
			_dump_fbs_writer.Flush()
		}

		_dump_fbs_count_mutex.Unlock()

		return nil
	}

	return errors.New("dumpFBS failed, before file not ready for open/write")
}

// Request send flatbuffer encoded data to server, with the tcp connection represented by connWrap..
// data: flatbufer encoded data
// connWrap: a wrapper of tcp connection, with some other informations
// deadline: timeout time
func Request(data []byte, connWrap *pool.ConnWrap, deadline time.Time) (<-chan []byte, error) {
	//construct the message header.
	reqHeadBuf, reqId := msg_hdr.ConstructHeaderBuf(uint64(len(data)))
	//send message header.
	_, err := msg_hdr.WriteExactly(*connWrap, reqHeadBuf, deadline)
	if err != nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("Request(): write message header failed, error: %v", err)
		return nil, err
	}

	// Note: putRequest() must be called before WriteExactlry(data), Because when
	// cpu utilization is high, the time window between the two WriteExactlry() function
	// call is big.

	rspChan := connWrap.PutRequest(reqId)

	_, err = msg_hdr.WriteExactly(*connWrap, data, deadline)
	return rspChan, err
}

// WaitResponse wait response data from server side.
// rspChan: return value of connWrap.PutRequest(reqId).
func WaitResponse(rspChan <-chan []byte, rspTimeout time.Duration) ([]byte, error) {
	select {
	case rspData, ok := <-rspChan:
		if ok {
			return rspData, nil
		} else { // chan closed by pool because of some error.
			return nil, errors.New("WaitResponse: no related chan")
		}
	case <-time.After(rspTimeout):
		return nil, &net_error.TimeoutError{TOErr: errors.New("wait for response timeout " + rspTimeout.String())}
	}

}

func PutFBSBuilder(builder *flatbuffers.Builder) {
	builder.Reset()
	fbsBldPool.Put(builder)
}

//encode IndexBatch to flatbufer.
func GenerateMindAlphaServingRequest(batch *fe.IndexBatch, algoName string) (*flatbuffers.Builder, []byte, error) {
	builder := fbsBldPool.Get().(*flatbuffers.Builder)
	ib_num := 1
	ib_list := make([]flatbuffers.UOffsetT, ib_num)
	for i := 0; i < ib_num; i += 1 {
		// columns
		xlen := len(batch.Columns)
		columns := make([]flatbuffers.UOffsetT, xlen)
		for j := 0; j < xlen; j += 1 {
			column := batch.Columns[j]
			cells := make([]flatbuffers.UOffsetT, len(column.CellsIdx))
			for k := 0; k < int(len(column.CellsIdx)); k++ {

				hcs := column.AccessHashCodesUnsafe(uint64(k))
				mindalpha_serving.CellStartHashCodesVector(builder, len(hcs))
				for idx := len(hcs); idx > 0; idx-- {
					builder.PrependUint64(hcs[idx-1])
				}
				hashlist := builder.EndVector(len(hcs))
				mindalpha_serving.CellStart(builder)
				mindalpha_serving.CellAddHashCodes(builder, hashlist)
				cell := mindalpha_serving.CellEnd(builder)
				cells[k] = cell
			}
			mindalpha_serving.ColumnStartCellsVector(builder, len(column.CellsIdx))
			for idx := len(cells); idx > 0; idx-- {
				builder.PrependUOffsetT(cells[idx-1])
			}
			offset_cells := builder.EndVector(len(cells))

			mindalpha_serving.ColumnStart(builder)
			mindalpha_serving.ColumnAddCells(builder, offset_cells)
			mindalpha_serving.ColumnAddLevel(builder, column.Level)
			col := mindalpha_serving.ColumnEnd(builder)
			columns[j] = col
		}
		mindalpha_serving.IndexBatchStartColumnsVector(builder, xlen)
		for idx := len(columns); idx > 0; idx-- {
			builder.PrependUOffsetT(columns[idx-1])
		}
		offset_columns := builder.EndVector(xlen)
		// names
		xlen = len(batch.Names)
		names_list := make([]flatbuffers.UOffsetT, xlen)
		for j := 0; j < xlen; j += 1 {
			str := builder.CreateString(batch.Names[j])
			names_list[j] = str
		}
		mindalpha_serving.IndexBatchStartNamesVector(builder, xlen)
		for idx := len(names_list); idx > 0; idx-- {
			builder.PrependUOffsetT(names_list[idx-1])
		}
		names := builder.EndVector(xlen)
		//level_tree
		xlen = len(batch.Last_level_index_tree)
		level_tree_list := make([]flatbuffers.UOffsetT, xlen)
		for j := 0; j < xlen; j += 1 {
			level_index_tree := batch.Last_level_index_tree[j]
			mindalpha_serving.LevelIndexStartIndexsVector(builder, len(level_index_tree))
			for idx := len(level_index_tree); idx > 0; idx-- {
				builder.PrependUint64(level_index_tree[idx-1])
			}
			level_index := builder.EndVector(len(level_index_tree))
			mindalpha_serving.LevelIndexStart(builder)
			mindalpha_serving.LevelIndexAddIndexs(builder, level_index)
			level_tree := mindalpha_serving.LevelIndexEnd(builder)
			level_tree_list[j] = level_tree
		}
		mindalpha_serving.IndexBatchStartLastLevelIndexTreeVector(builder, xlen)
		for idx := len(level_tree_list); idx > 0; idx-- {
			builder.PrependUOffsetT(level_tree_list[idx-1])
		}
		level_tree := builder.EndVector(xlen)

		mindalpha_serving.IndexBatchStart(builder)
		mindalpha_serving.IndexBatchAddRows(builder, (uint64)(batch.Rows))
		mindalpha_serving.IndexBatchAddLevels(builder, (uint64)(batch.Levels))
		mindalpha_serving.IndexBatchAddNames(builder, names)
		mindalpha_serving.IndexBatchAddColumns(builder, offset_columns)
		mindalpha_serving.IndexBatchAddLastLevelIndexTree(builder, level_tree)
		// finish
		ib := mindalpha_serving.IndexBatchEnd(builder)
		ib_list[i] = ib
	}
	mindalpha_serving.RequestStartIndexBatchsVector(builder, ib_num)
	for idx := len(ib_list); idx > 0; idx-- {
		builder.PrependUOffsetT(ib_list[idx-1])
	}
	ibs := builder.EndVector(ib_num)
	algo_name := builder.CreateString(algoName)
	version := builder.CreateString("v2")
	mindalpha_serving.RequestStart(builder)
	mindalpha_serving.RequestAddIndexBatchs(builder, ibs)
	mindalpha_serving.RequestAddAlgoName(builder, algo_name)
	mindalpha_serving.RequestAddVersion(builder, version)
	rqst := mindalpha_serving.RequestEnd(builder)
	builder.Finish(rqst)
	return builder, builder.FinishedBytes(), nil
}

//for debug. deserialize flatbuffered byte slice to IndexBatch, and print Indexbatch
func ParseMindAlphaServingRequest(rqstBuf []byte) {
	rqst := mindalpha_serving.GetRootAsRequest(rqstBuf, 0)
	client_logger.GetMindAlphaServingClientLogger().Errorf("algoname %v Version %v", string(rqst.AlgoName()), string(rqst.Version()))
	var ib mindalpha_serving.IndexBatch
	rqst.IndexBatchs(&ib, 0)
	client_logger.GetMindAlphaServingClientLogger().Errorf("Rows=%v\tColumns=%v\t Levels=%v", ib.Rows(), ib.ColumnsLength(), ib.LastLevelIndexTreeLength())
	for i := 0; i < ib.NamesLength(); i++ {
		client_logger.GetMindAlphaServingClientLogger().Errorf("names[%v]=%v", i, string(ib.Names(i)))
	}
	for i := 0; i < ib.ColumnsLength(); i++ {
		var column mindalpha_serving.Column
		ib.Columns(&column, i)
		for j := 0; j < column.CellsLength(); j++ {
			var cell mindalpha_serving.Cell
			column.Cells(&cell, j)
		}
		client_logger.GetMindAlphaServingClientLogger().Errorf("columns[%v]= {level=%v, cell_size=%v}\n", i, column.Level(), column.CellsLength())
	}
	for i := 0; i < ib.LastLevelIndexTreeLength(); i++ {
		var level_tree mindalpha_serving.LevelIndex
		ib.LastLevelIndexTree(&level_tree, i)
		for j := 0; j < level_tree.IndexsLength(); j++ {
			client_logger.GetMindAlphaServingClientLogger().Errorf("i=%v j=%v value=%v\n", i, j, level_tree.Indexs(j))
		}
	}
}

func ParseMindAlphaServingResponse(rspBuf []byte) (*tensor.TensorScores, error) {

	// flatbuffer deserialize
	psResponse := mindalpha_serving.GetRootAsResponse(rspBuf, 0)
	if psResponse == nil {
		client_logger.GetMindAlphaServingClientLogger().Errorf("parse Response failed")
		return nil, errors.New("ParseMindAlphaServingResposne: respose is nil")
	}
	version := psResponse.Version()
	debugInfo := psResponse.DebugInfo()
	if len(debugInfo) > 0 {
		client_logger.GetMindAlphaServingClientLogger().Errorf("version: %v, debug_info: %v", string(version), string(debugInfo))
	}
	ts := tensor.NewTensorScores()
	ts.ParseTensorScores(rspBuf)
	if ts.GetScoreSize() == 0 {
		if len(debugInfo) > 0 {
			return nil, errors.New("ParseMindAlphaServingResposne(): ERROR: version: " + string(version) + ", debug_info: " + string(debugInfo))
		} else {
			return nil, errors.New("ParseMindAlphaServingResposne(): ERROR: version: " + string(version) + ", tensor score size = 0")
		}
	}

	return ts, nil
}
