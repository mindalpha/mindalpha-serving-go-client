// automatically generated by the FlatBuffers compiler, do not modify

package mindalpha_serving

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Response struct {
	_tab flatbuffers.Table
}

func GetRootAsResponse(buf []byte, offset flatbuffers.UOffsetT) *Response {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Response{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Response) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Response) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Response) Version() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Response) Tensor(obj *Tensor) *Tensor {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Tensor)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *Response) DebugInfo() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func ResponseStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func ResponseAddVersion(builder *flatbuffers.Builder, version flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(version), 0)
}
func ResponseAddTensor(builder *flatbuffers.Builder, tensor flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(tensor), 0)
}
func ResponseAddDebugInfo(builder *flatbuffers.Builder, debugInfo flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(debugInfo), 0)
}
func ResponseEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}