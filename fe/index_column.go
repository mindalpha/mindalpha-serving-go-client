package fe

var kFeatureColumnName []string

type IndexedColumn struct {
	Ib *IndexBatch
}

// Load column name file.
// You should refer to fe/README.md.
func LoadColumnNameFile(column_file string) error {
	kFeatureColumnName = LoadColumnNames(column_file)
	return nil
}

func NewIndexedColumn(level, column uint64) *IndexedColumn {
	return &IndexedColumn{
		NewIndexBatch(level, column),
	}
}

func (this *IndexedColumn) Free() {
	this.Ib.Reset()
	PutIBToPool(this.Ib)
}

//TODO
func (this *IndexedColumn) DebugString() string {
	//return this.Ib.ToString()
	return ""
}
func (this *IndexedColumn) GetRowFeatures(row int) string {
	return this.Ib.GetRowFeaturesDelimited(kFeatureColumnName, uint64(row), '\u0002', '\u0001')
}

func (this *IndexedColumn) LoadFromCsvFile(csv_file, column_file string, gen_with_score bool) error {
	return this.Ib.LoadFromCsvFile(csv_file, column_file, gen_with_score)
}

// Add a value to column named key. The level of the column named key is specified by param "level"
// param "pre" indicates the index of the last level features associates to this newly added feature.
// If we do not have the colume, then we will add a new column to IndexBatch named key.
// After this function call, the rows of this column may be increased by 1.
// To learn more about IndexBatch, please refer to fe/README.md.
func (this *IndexedColumn) AddColumn(key, value string, level int, pre int) {
	column := this.Ib.GetColumn(uint64(level), key)
	var cell Cell
	if len(value) == 0 {
		cell = NewCell("none")
	} else {
		cell = NewCell(value)
	}
	this.Ib.AppendCell(column, cell, uint64(pre))
}

// Similar to AddColumn(), but this function adds multiple values to column.
// the column name key is a multi-value feature.
func (this *IndexedColumn) AddColumnArray(key string, values []string, level int, pre int) {
	column := this.Ib.GetColumn(uint64(level), key)
	if len(values) == 0 {
		values = append(values, "none")
	}
	cell := NewCellMulti(values)
	this.Ib.AppendCell(column, cell, uint64(pre))
}

func (this *IndexedColumn) GetBatchSize() int {
	return int(this.Ib.Rows)
}
