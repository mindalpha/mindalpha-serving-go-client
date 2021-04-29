package fe

import (
	"bufio"
	"bytes"
	"fmt"
	client_logger "github.com/mindalpha/mindalpha-serving-go-client/logger"
	"os"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

var Uint64Size uint32 = uint32(unsafe.Sizeof(uint64(0)))
var Uint32Size uint32 = uint32(unsafe.Sizeof(uint32(0)))

var ibPool = sync.Pool{
	New: func() interface{} {
		return NewIndexBatchFromPool()
	},
}

func PutIBToPool(ib *IndexBatch) {
	ibPool.Put(ib)
}

type StringViewHash struct {
	SV       string
	HashCode uint64
}

func BKDRHash(content string) uint64 {
	var seed uint64 = 0
	contentLen := len(content)
	for i := 0; i < contentLen; i++ {
		seed = seed*131 + uint64(int8(content[i]))
	}
	return seed
}

func NewStringViewHash(content string) *StringViewHash {
	var svh StringViewHash
	svh.SV = content
	svh.HashCode = BKDRHash(content)
	return &svh
}

type Cell struct {
	Splits []StringViewHash
}

func splitSVAndFilter(str string, delims string, filter string) []StringViewHash {
	var output []StringViewHash
	fields := strings.Split(str, delims)
	for idx := 0; idx < len(fields); idx++ {
		xstr := fields[idx]
		if len(xstr) <= 0 {
			continue
		}
		if xstr != filter {
			sv := NewStringViewHash(xstr)
			output = append(output, *sv)
		}
	}
	return output
}
func (cell *Cell) Split(delim string) {
	if len(cell.Splits) > 0 {
		cell.Splits = splitSVAndFilter(cell.Splits[0].SV, delim, "none")
	}
}

func (cell *Cell) Access() []StringViewHash {
	return cell.Splits
}
func (cell *Cell) ToString() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i := 0; i < len(cell.Splits); i++ {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(fmt.Sprintf("\"%v\",%v", cell.Splits[i].SV, cell.Splits[i].HashCode))
	}
	out.WriteString("]")
	return out.String()
}

func NewCell(content string) Cell {
	var cell Cell
	if content != "none" && len(content) > 0 {
		sv := NewStringViewHash(content)
		cell.Splits = append(cell.Splits, *sv)
	}
	return cell
}

func NewCellMulti(contents []string) Cell {
	var cell Cell
	xlen := len(contents)
	cell.Splits = make([]StringViewHash, 0, xlen)
	for i := 0; i < xlen; i++ {
		if contents[i] != "none" && len(contents[i]) > 0 {
			sv := NewStringViewHash(contents[i])
			cell.Splits = append(cell.Splits, *sv)
		}
	}

	return cell
}

type Column struct {
	Level      uint64
	CellsIdx   []uint32
	CellsBytes []byte
}

func (column *Column) Append(cell Cell) uint64 {
	column.CellsIdx = append(column.CellsIdx, uint32(len(column.CellsBytes)))

	tmpBuf := make([]byte, 0, 0)
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&tmpBuf)))
	var sv_num uint32 = uint32(len(cell.Splits))

	sliceHeader.Data = (uintptr)(unsafe.Pointer(&sv_num))
	sliceHeader.Len = int(Uint32Size)
	sliceHeader.Cap = int(Uint32Size)
	column.CellsBytes = append(column.CellsBytes, tmpBuf...) //len(cell.Splits)

	for i := 0; i < int(sv_num); i++ {
		hc := cell.Splits[i].HashCode
		//HashCode
		sliceHeader.Data = (uintptr)(unsafe.Pointer(&hc))
		sliceHeader.Len = int(Uint64Size)
		sliceHeader.Cap = int(Uint64Size)
		column.CellsBytes = append(column.CellsBytes, tmpBuf...) //Hashcode
	}

	var sv_len uint32

	for i := 0; i < int(sv_num); i++ {
		sv := cell.Splits[i].SV

		sv_len = uint32(len(sv))
		sliceHeader.Data = (uintptr)(unsafe.Pointer(&sv_len))
		sliceHeader.Len = int(Uint32Size)
		sliceHeader.Cap = int(Uint32Size)
		column.CellsBytes = append(column.CellsBytes, tmpBuf...) //SV Len

		//SV
		x := (*[2]uintptr)(unsafe.Pointer(&sv))
		sliceHeader.Data = x[0]
		sliceHeader.Len = int(x[1])
		sliceHeader.Cap = int(x[1])
		column.CellsBytes = append(column.CellsBytes, tmpBuf...) //SV
	}
	sliceHeader.Len = 0
	sliceHeader.Cap = 0

	return uint64(len(column.CellsIdx))
}

func (column *Column) Access(row uint64) []StringViewHash {
	cellStartIdx := column.CellsIdx[uint32(row)]

	pStart := uintptr(unsafe.Pointer(&(column.CellsBytes[0]))) + uintptr(cellStartIdx)
	sv_num := *(*uint32)(unsafe.Pointer(pStart))
	pStart = pStart + uintptr(Uint32Size)

	svRet := make([]StringViewHash, sv_num)
	//HashCode
	for i := 0; i < int(sv_num); i++ {
		svRet[i].HashCode = *(*uint64)(unsafe.Pointer(pStart))
		pStart = pStart + uintptr(Uint64Size)
	}

	//SV
	var sv_len uint32
	for i := 0; i < int(sv_num); i++ {
		sv_len = *(*uint32)(unsafe.Pointer(pStart))
		pStart = pStart + uintptr(Uint32Size)

		strHeader := reflect.StringHeader{pStart, int(sv_len)}
		svRet[i].SV = *(*string)(unsafe.Pointer(&strHeader))
		pStart = pStart + uintptr(sv_len)
	}

	return svRet
}

func (column *Column) AccessHashCodesUnsafe(row uint64) []uint64 {
	cellStartIdx := column.CellsIdx[uint32(row)]

	pStart := uintptr(unsafe.Pointer(&(column.CellsBytes[0]))) + uintptr(cellStartIdx)
	sv_num := *(*uint32)(unsafe.Pointer(pStart))
	pStart = pStart + uintptr(Uint32Size)

	tmpBuf := make([]uint64, 0, 0)
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&tmpBuf)))
	sliceHeader.Data = (uintptr)(unsafe.Pointer(pStart))
	sliceHeader.Len = int(sv_num)
	sliceHeader.Cap = int(sv_num)

	return tmpBuf
}

//TODO
/*
func (column *Column) ToString() string {
	var buffer bytes.Buffer
	xlen := len(column.Cells)
	buffer.WriteString(fmt.Sprintf("{\"level\": %v, \"cells\":%v [", column.Level, xlen))
	for idx := 0; idx < xlen; idx++ {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(column.Cells[idx].ToString())
	}
	buffer.WriteString("]}")
	return buffer.String()
}
*/

func (column *Column) Reset() {
	column.CellsIdx = column.CellsIdx[:0]
	column.CellsBytes = column.CellsBytes[:0]
}

func NewColumnFromPool() *Column {
	var column Column
	column.CellsIdx = make([]uint32, 0, 4)
	column.CellsBytes = make([]byte, 0, 32)
	return &column
}

func (ib *IndexBatch) IbNewColumn(level uint64) *Column {
	var column *Column = ib.columnPool.Get().(*Column)
	column.Level = level
	return column
}

type IndexBatch struct {
	Rows                  uint64
	Levels                uint64
	Columns               []*Column
	Names                 []string
	Name_col_map          map[string]uint64
	Last_level_index_tree [][]uint64
	columnPool            *sync.Pool //pool local to IndexBatch. If not, this pool will lead to performence case.
}

func (ib *IndexBatch) Reset() {
	for _, col := range ib.Columns {
		col.Reset()
		ib.columnPool.Put(col)
	}
	ib.Rows = 0
	ib.Levels = 0
	ib.Columns = ib.Columns[:0]
	ib.Names = ib.Names[:0]
	ib.Name_col_map = make(map[string]uint64)
	for i, _ := range ib.Last_level_index_tree {
		ib.Last_level_index_tree[i] = ib.Last_level_index_tree[i][:0]
	}
}
func NewIndexBatchFromPool() *IndexBatch {
	var ib IndexBatch
	ib.columnPool = &sync.Pool{
		New: func() interface{} {
			return NewColumnFromPool()
		},
	}

	return &ib
}

func NewIndexBatch(level, column uint64) *IndexBatch {
	var ib *IndexBatch = ibPool.Get().(*IndexBatch)
	ib.Levels = level
	ib.Rows = 0

	if cap(ib.Names) == 0 {
		ib.Names = make([]string, 0, column)
	}
	if cap(ib.Columns) == 0 {
		ib.Columns = make([]*Column, 0, column)
	}

	ib.Name_col_map = make(map[string]uint64)
	if len(ib.Last_level_index_tree) != int(level) {
		ib.Last_level_index_tree = make([][]uint64, level)
	}
	return ib
}

type ProcessLine func(string)

func LoadFileCommon(filename string, callback ProcessLine) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		callback(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func LoadCsvFileRows(filename string, delim string) [][]string {
	var rows [][]string
	rows = make([][]string, 0, 100)
	LoadFileCommon(filename, func(line string) {
		row := strings.Split(line, delim)
		rows = append(rows, row)
	})
	return rows
}
func LoadColumnNames(filename string) (column_names []string) {
	const kColumnInitSize = 160
	column_names = make([]string, 0, kColumnInitSize)
	var line_count = 0
	LoadFileCommon(filename, func(line string) {
		field_list := strings.Split(line, " ")
		column_names = append(column_names, field_list[1])
		line_count += 1
	})
	return
}

func (ib *IndexBatch) LoadFromCsvFile(csv_file, column_file string, gen_with_score bool) error {
	var rows [][]string
	rows = LoadCsvFileRows(csv_file, string('\002'))
	column_names := LoadColumnNames(column_file)
	var start_idx = 0
	var rows_size = len(rows)
	if gen_with_score {
		start_idx++
	}
	for j, column_name := range column_names {
		column := ib.GetColumn(2, column_name)
		for i := 0; i < rows_size; i += 1 {
			item := rows[i][j+start_idx]
			cell := NewCell(item)
			cell.Split("\001")
			ib.AppendCell(column, cell, 0)
		}
	}
	return nil
}

func (ib *IndexBatch) GetColumn(level uint64, name string) *Column {
	if col, ok := ib.Name_col_map[name]; !ok {
		ib.Name_col_map[name] = uint64(len(ib.Columns))
		ib.Names = append(ib.Names, name)
		column := ib.IbNewColumn(level)
		ib.Columns = append(ib.Columns, column)
		return column
	} else {
		return ib.Columns[col]
	}
}

func (ib *IndexBatch) AppendCell(column *Column, cell Cell, last_level_index uint64) {
	levelSize := column.Append(cell)
	if levelSize > ib.Rows {
		ib.Rows = levelSize
	}
	vec := ib.Last_level_index_tree[column.Level]
	if uint64(len(vec)) < levelSize {
		ib.Last_level_index_tree[column.Level] = append(ib.Last_level_index_tree[column.Level], last_level_index)
	}
}

func (ib *IndexBatch) GetLastLevelIndex(row, level uint64) uint64 {
	if len(ib.Last_level_index_tree[level]) == 0 {
		return row
	}
	return ib.Last_level_index_tree[level][row]
}

func (ib *IndexBatch) GetCellAt(row, col uint64) []StringViewHash {
	column := ib.Columns[col]
	if ib.Levels-1 == column.Level {
		return column.Access(row)
	} else if 0 == column.Level {
		return column.Access(0)
	} else {
		for j := ib.Levels - 1; j > column.Level; j-- {
			row = ib.GetLastLevelIndex(row, j)
		}
		return column.Access(row)
	}
}

func (ib *IndexBatch) SafeGetCell(row uint64, column_name string, col uint64) []StringViewHash {
	if col > uint64(len(ib.Columns)) {
		client_logger.GetMindAlphaServingClientLogger().Errorf("SafeGetCell(%v, %v, %v), column_index exceed IndexBatch.Columns: %v", row, column_name, col, len(ib.Columns))
		return nil
	}
	return ib.GetCellAt(row, col)
}

func (ib *IndexBatch) GetRowFeaturesDelimited(column_array []string, row uint64, feature_delim, value_delim byte) string {
	var buffer bytes.Buffer
	column_name_map := ib.Name_col_map
	var j = 0
	for _, column := range column_array {
		if j > 0 {
			buffer.WriteByte(feature_delim)
		}
		if idx, ok := column_name_map[column]; ok {
			cell_splits := ib.SafeGetCell(row, column, idx)
			if nil == cell_splits {
				continue
			}
			for i, val := range cell_splits {
				if i > 0 {
					buffer.WriteByte(value_delim)
				}
				buffer.WriteString(val.SV)
			}
		} else {
			// add more debug
		}
		j++
	}
	return buffer.String()
}

//TODO
/*
func (ib *IndexBatch) ToString() string {
	var buffer bytes.Buffer
	column_num := len(ib.Columns)
	buffer.WriteString(fmt.Sprintf("{\"rows\": %v, \"levels\": %v, \"column_num\": %v, ", ib.Rows, ib.Levels, column_num))
	buffer.WriteString("\"columns\": [")
	for idx := 0; idx < column_num; idx++ {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(ib.Columns[idx].ToString())
	}
	buffer.WriteString("], \"names\": [")
	buffer.WriteString(strings.Join(ib.Names, ","))
	buffer.WriteString("]\"name_col_map\": {")
	//for _, _:= range ib.Name_col_map {
	//}
	buffer.WriteString("},\"last_level_index_tree\": [")
	for idx := range ib.Last_level_index_tree {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(strings.Replace(fmt.Sprint(ib.Last_level_index_tree[idx]), " ", ",", -1))
	}
	buffer.WriteString("]}")
	return buffer.String()
}
*/
