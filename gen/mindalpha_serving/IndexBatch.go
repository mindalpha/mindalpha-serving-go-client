// automatically generated by the FlatBuffers compiler, do not modify

package mindalpha_serving

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type IndexBatch struct {
	_tab flatbuffers.Table
}

func GetRootAsIndexBatch(buf []byte, offset flatbuffers.UOffsetT) *IndexBatch {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &IndexBatch{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *IndexBatch) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *IndexBatch) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *IndexBatch) Rows() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *IndexBatch) MutateRows(n uint64) bool {
	return rcv._tab.MutateUint64Slot(4, n)
}

func (rcv *IndexBatch) Levels() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *IndexBatch) MutateLevels(n uint64) bool {
	return rcv._tab.MutateUint64Slot(6, n)
}

func (rcv *IndexBatch) Names(j int) []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.ByteVector(a + flatbuffers.UOffsetT(j*4))
	}
	return nil
}

func (rcv *IndexBatch) NamesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *IndexBatch) Columns(obj *Column, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *IndexBatch) ColumnsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *IndexBatch) LastLevelIndexTree(obj *LevelIndex, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *IndexBatch) LastLevelIndexTreeLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func IndexBatchStart(builder *flatbuffers.Builder) {
	builder.StartObject(5)
}
func IndexBatchAddRows(builder *flatbuffers.Builder, rows uint64) {
	builder.PrependUint64Slot(0, rows, 0)
}
func IndexBatchAddLevels(builder *flatbuffers.Builder, levels uint64) {
	builder.PrependUint64Slot(1, levels, 0)
}
func IndexBatchAddNames(builder *flatbuffers.Builder, names flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(names), 0)
}
func IndexBatchStartNamesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func IndexBatchAddColumns(builder *flatbuffers.Builder, columns flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(columns), 0)
}
func IndexBatchStartColumnsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func IndexBatchAddLastLevelIndexTree(builder *flatbuffers.Builder, lastLevelIndexTree flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(4, flatbuffers.UOffsetT(lastLevelIndexTree), 0)
}
func IndexBatchStartLastLevelIndexTreeVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func IndexBatchEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
