package message_header

import (
	"bytes"
	"encoding/binary"
	"errors"
	net_error "github.com/mindalpha/mindalpha-serving-go-client/error"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"net"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"
)

type MessageHeader struct {
	MetaBufferSize   uint64
	DataBufferSize   uint64
	ProcessorClassId uint64
	RequestId        uint64
}

type RequestMeta struct {
	ProcessId uint64
}

var requestId uint64 = 1

const (
	KRequestProcessor_ClassId  uint64 = 0xe683c4f853870912
	KResponseProcessor_ClassId uint64 = 0x8baa59d693539253
	KMessageHeaderLen          uint64 = uint64(unsafe.Sizeof(MessageHeader{0, 0, 0, 0}))
	KRequestMetaLen            uint64 = uint64(unsafe.Sizeof(RequestMeta{0}))
)

var emptyRequestMeta RequestMeta = RequestMeta{KRequestProcessor_ClassId}

func ConstructMessageHeader(dataLen uint64) (*MessageHeader, uint64) {
	reqId := atomic.AddUint64(&requestId, 1)
	mh := MessageHeader{uint64(KRequestMetaLen), dataLen, KRequestProcessor_ClassId, reqId}

	return &mh, reqId
}

func ConstructHeaderBuf(dataLen uint64) ([]byte, uint64) {
	mh, reqId := ConstructMessageHeader(dataLen)
	mhBuf := new(bytes.Buffer)
	_ = binary.Write(mhBuf, binary.LittleEndian, mh)
	_ = binary.Write(mhBuf, binary.LittleEndian, emptyRequestMeta)
	return mhBuf.Bytes(), reqId
}

func WriteExactly(conn net.Conn, buf []byte, deadline time.Time) (int, error) {
	leftData := buf
	leftLen, dataLength := len(buf), len(buf)
	var totalWrite int = 0
	leftTimeForWrite := deadline.Sub(time.Now())
	conn.SetWriteDeadline(deadline)
	for {
		cnt, err := conn.Write(leftData)
		if err != nil {
			client_logger.GetMindAlphaServingClientLogger().Errorf("ERROR: WriteExactly(): conn.Write(%v) error: %v, timeout time: %v", leftLen, err, leftTimeForWrite)
			return totalWrite, &net_error.NetError{Nerr: errors.New("Error: WriteExactly(): conn.Write(" + strconv.Itoa(leftLen) + " error: " + err.Error())}
		}
		if cnt <= 0 {
			client_logger.GetMindAlphaServingClientLogger().Errorf("ERROR: WriteExactly(): do not read enough data")
			return totalWrite, &net_error.NetError{Nerr: errors.New("ERROR: WriteExactly(): do not read enough data")}
		}
		leftLen -= cnt
		totalWrite += cnt
		leftData = buf[totalWrite:]
		if totalWrite >= dataLength {
			return totalWrite, nil
		}
	}
}

func ReadExactly(conn net.Conn, length int) ([]byte, error) {
	rspData := make([]byte, length)
	leftData := rspData
	leftLen := length
	var totalRead int = 0

	for {
		cnt, err := conn.Read(leftData)
		if err != nil {
			return nil, &net_error.NetError{Nerr: errors.New("ERROR: ReadExactly(): conn.Read(" + strconv.Itoa(length) + "), already read " + strconv.Itoa(totalRead) + ", error: " + err.Error())}
		}
		if cnt <= 0 {
			return nil, &net_error.NetError{Nerr: errors.New("ERROR: ReadExactly(): do not read enough data, expect " + strconv.Itoa(length) + ", just read " + strconv.Itoa(cnt) + " bytes")}
		}
		leftLen -= cnt
		totalRead += cnt
		leftData = rspData[totalRead:]
		if totalRead < length {
			//client_logger.GetMindAlphaServingClientLogger().Errorf("response: do not read enough data, read %v bytes, expect %v bytes, try to read left %v bytes", cnt, length, leftLen)
		} else {
			return rspData, nil
		}
	}
}
