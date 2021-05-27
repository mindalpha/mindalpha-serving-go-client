# IndexBatch
[中文文档](README.CN.md) <br>

IndexBatch related operations

## IndexBatch

For a advertise request, it has three level features, the fitst level is request-level/user-level features, the second level is campain-level features, the third level is creative features. <br>
A request may have multiple campains, each campain has multiple creatives. We use IndexBatch to represent this three level features.<br>

We use the following struct to show a ad request <br>
![GitHub](/pictures/request-define.jpg "request definition")

Business code will construct IndexBatch use features from above request <br>

The following picture is a visual show of IndexBatch<br>
![GitHub](/pictures/IndexBatch__of_a_request.png "IndexBatch before optimize")

Above IndexBatch:
1. has 3 level: level 0, level 1, level 2
2. level 0 has 1 feature column, which name is "column 0"
3. level 1 has 1 feature column, which name is "column 1"
4. level 2 has 2 feature columns, which name is "column 2, column 3"
5. has 7 row data, so its batch size is 7.
6. feature column named "column 0" has 1 value, which is "col0_value"
7. feature column named "column 1" has 3 value, which is "col1_val1", "col1_val2" and "col1_val3"
8. feature column named "column 2" has 7 value, which is "col2_val1", "col2_val2", "col2_val3", "col2_val4", "col2_val5", "col2_val6" and "col2_val7"
9. feature column named "column 3" has 7 value, which is "col3_val1", "col3_val2", "col3_val3", "col3_val4", "col3_val5", "col3_val6" and "col3_val7"
10. data of row 0 is "col0_value, col1_val1, col2_val1, co3_val1"
11. data of row 1 is "col0_value, col1_val1, col2_val2, co3_val2"
12. data of row 5 is "col0_value, col1_val3, col2_val6, co3_val6"
13. data of row 6 is "col0_value, col1_val3, col2_val7, co3_val7"

We can conclude from above IndexBatch: 
1. The IndexBatch has 7 rows , so its batch size is 7.
2. feature column "column 0" is a level 0 feature, it has 7 rows, and its value is all the same.
3. feature column "column 1" is a level 1 feature,  it has 7 rows, and: <br>
    a. "column 1" 's row 0/row 1 value is the same <br>
    b. "column 1" 's row 2/row 3 value is the same <br>
    c. "column 1" 's row 4/row 5/row 6 value is the same <br>
4. feature columns "column 2" & "column 3" is  level 2 features: values of each row is different <br>

**In the above IndexBatch, column of level 0 / level 1 has duplicated values, we can compress the duplicated values. The IndexBatch after compression is as follows** <br>
![GitHub](/pictures/IndexBatch_After_Optimize.png "IndexBatch after optimize")
***Note: The above IndexBatch after compression is equivalent to the IndexBatch before compression*** <br>
We can conclude from the above IndexBatch after compression: <br>
1. feature column "column 0" is compressed to 1 row
2. feature column "column 1" is compressed to 3 row
3. we added a field named last_level_index, to record the related upper level's row index.
4. feature column "column 0" has 1 row, its row index is 0(row 0), so all rows last_level_index of feature column "column 1" is 0 (the last_level_index of column 1 is not showed in the above picture)
5. feature column "column 1" 's first value is "col1_val1", its row index is 0 (row 0), so the two feature column of level 2 (column 2 / column 3) 's first 2 row (row 0 / row 1) 's last_level_index is 0.
6. feature column "column 1" 's second value is "col1_val2", its row index is 1 (row 1), so the two feature column of level 2 (column 2 / column 3) 's second 2 row (row 2 / row 3) 's last_level_index is 1.
7. feature column "column 1" 's third value is "col1_val3", its row index is 2 (row 2), so the two feature column of level 2 (column 2 / column 3) 's last 3 row (row 4 / row 5 / row 6) 's last_level_index is 2.


## Code to construct the above IndexBatch
The following code shows how to construct the above compressed IndexBatch

```go

// Load feature column names from file. The format of column name file please refer to data/column_name_criteo.txt.
// This function is not necessary if you do not call GetRowFeatures(row).
fe.LoadColumnNameFile("column_name_file")

// Init a IndexedColumn
// First param is the level of IndexBatch, it must be 3.
// Second param tells how many columns you may have, it can be not precise.
ib := NewIndexedColumn(3, 4)

// Add featuren column and value to IndexBatch

// level0, row 0
ib.AddColumn("column 0", "col0_value", 0, 0)



// level1, row 0, last_level = level0, last_level_index = 0
ib.AddColumn("column 1", "col1_val1", 1, 0)

// level2 row 0, last_level = level1, last_level_index = 0
ib.AddColumn("column 2", "col2_val1", 2, 0)
ib.AddColumn("column 3", "col3_val1", 2, 0)

// level2 row 1, last_level = level1, last_level_index = 0
ib.AddColumn("column 2", "col2_val2", 2, 0)
ib.AddColumn("column 3", "col3_val2", 2, 0)


// level1, row 1, last_level = level0, last_level_index = 0
ib.AddColumn("column 1", "col1_val2", 1, 0)

// level2 row 2, last_level = level1, last_level_index = 1
ib.AddColumn("column 2", "col2_val3", 2, 1)
ib.AddColumn("column 3", "col3_val3", 2, 1)

// level2 row 3, last_level = level1, last_level_index = 1
ib.AddColumn("column 2", "col2_val4", 2, 1)
ib.AddColumn("column 3", "col3_val4", 2, 1)



// level1, row 2, last_level = level0, last_level_index = 0
ib.AddColumn("column 1", "col1_val3", 1, 0)

// level2 row 4, last_level = level1, last_level_index = 2
ib.AddColumn("column 2", "col2_val5", 2, 2)
ib.AddColumn("column 3", "col3_val5", 2, 2)

// level2 row 5, last_level = level1, last_level_index = 2
ib.AddColumn("column 2", "col2_val6", 2, 2)
ib.AddColumn("column 3", "col3_val6", 2, 2)

// level2 row 6, last_level = level1, last_level_index = 2
ib.AddColumn("column 2", "col2_val7", 2, 2)
ib.AddColumn("column 3", "col3_val7", 2, 2)



```
### Functions
1. func NewIndexedColumn(level, column uint64) *IndexedColumn <br>
    Init a IndexBatch, then we will add feature to this IndexBatch <br>
    Param level show how many levels this IndexBatch have. ***Its value must be 3*** <br>
    Param column shows how many columns this IndexBatch may have. Its value can be not precise, its a hint for IndexBatch to decide how many memory to allocate.<br>
3. func (this *IndexedColumn) AddColumn(key, value string, level int, pre int) <br>
    Add a value to column named key. <br>
    Param key is the name of feature column <br>
    Param value is the newly added value <br>
    Param pre is the last_level_index <br>
    If the feature column named key not exists, then will add a new column, which name is key <br>
    This function will add a new value to column named key, so the rows of this column will increase by 1 <br>
4. func (this *IndexedColumn) AddColumnArray(key string, values []string, level int, pre int) <br>
    Same as AddColumn, but this function as a multi-value feature to column named key. <br>
5. func (this *IndexedColumn) GetRowFeatures(row int) string <br>
    Get a row from IndexBatch. This row is formated to string. <br>
    Users can store this returned string to file, and then use LoadFromCsvFile() load it <br>
6. func (this *IndexedColumn) LoadFromCsvFile(csv_file, column_name_file string, gen_with_score bool) <br>
    Load data from csv_file and use it constuct IndexBatch. <br>
    Data in csv_file is the returned string of GetRowFeatures(). (User may add a label to the returned string of GetRowFeatures()) <br>
    column_name_file stores the column names, its format please refer to [column_name_file](/data/column_name_criteo.txt) <br>
    gen_with_score: If users save the returned string of GetRowFeatures() to file directly, the value of gen_with_score is false; If users add a label before the returned string of GetRowFeatures(), then this value should be true. <br>
