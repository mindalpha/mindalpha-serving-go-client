# IndexBatch
[English Document](README.md) <br>

构造IndexBatch及相关操作

## IndexBatch说明

对于一个广告的请求，该广告请求有三层特征，分别是请求级/设备级信息，广告的单子级信息，以及单子的素材信息。<br>
这一条请求可能包含多个单子级信息，每个单子又包含多个素材信息。我们将这三级特征用IndexBatch来表示，IndexBatch的level <br>
表示特征的层级，比如level 0 对应该请求的请求/设备信息， level 1 对应该请求的单子级信息, level 2对应该请求的素材级信息。<br>

我们用如下的代码大致展示一下一个请求的定义: <br>
![GitHub](/pictures/request-define.jpg "request definition")

业务代码会根据上面提到的广告请求，提取、构造一些特征， 然后用这些提取构造的特征去构造IndexBatch <br>

下图是IndexBatch的一个形象的展示。

![GitHub](/pictures/IndexBatch__of_a_request.png "IndexBatch before optimize")


上面的IndexBatch中：
1. 有三级level，分别是level 0, level 1, level 2
2. level 0 有1列特征，这1列特征的名字为column 0
3. level 1 有1列特征，这1列特征的名字为column 1
4. level 2 有2列特征，这2列特征的名字分别为column 2, column 3
5. 该IndexBatch 共填充了7行数据, 即 batch size 为 7
6. column 0 这1列有1个值 col0_value
7. column 1 这1列有3个值, 分别为 col1_val1, col1_val2, col1_val3
8. column 2 这1列有7个值, 分别为 col2_val1, col2_val2, col2_val3, col2_val4, col2_val5, col2_val6, col2_val7
9. column 3 这1列有7个值, 分别为 col3_val1, col3_val2, col3_val3, col3_val4, col3_val5, col3_val6, col3_val7
10. row 0 的数据为 col0_value, col1_val1, col2_val1, co3_val1
11. row 1 的数据为 col0_value, col1_val1, col2_val2, co3_val2
12. row 5 的数据为 col0_value, col1_val3, col2_val6, co3_val6
12. row 6 的数据为 col0_value, col1_val3, col2_val7, co3_val7

从上面的IndexBatch我们可以看出:
1. 该IndexBatch总共有7行, batch size 为 7
2. level 0 : column 0 这一列特征，总共有7行, 而这7行在 column 0 这一列的特征值是完全一样的
3. level 1: column 1  这一列特征，总共有7行，而
    a. row 0/row 1 这两行 column 1 这一列的特征值是完全一样的
    b. row 2/row 3 这两行 column 1 这一列的特征值是完全一样的
    c. row 4/row 5/row 6 column 1 这一列这三行的特征值是完全一样的
4. level 2: column 2 & column 3 特征列, 每一行的值是完全不一样的

**由于上述的IndexBatch中, 不同行的 level 0 / level 1 的特征列的值存在重复的, 我们可以将重复的特征值合并压缩, 这样可以节省空间. 压缩后的IndexBatch如下:** <br>

![GitHub](/pictures/IndexBatch_After_Optimize.png "IndexBatch after optimize")

**上面的压缩后的IndexBatch 和 压缩前的IndexBatch是等价的**.<br>
从上面的压缩后的IndexBatch我们可以看出:
1. level 0 的列的值压缩到1行
2. level 1 的列的值压缩到3行
3. 添加了一个辅助信息记录last_level_index, 用来记录下级特征对应的上级特征的row index.
4. column 0 这一列有1行，其row index 为 0(即 row 0)。 所以 column 1 这一列所有行对应的column 0 列的last_level_index 为 0(图中没有画出column 1 这一列的每行的last_level_index)
5. column 1 这一列 col1_val1 的row index 为 0 (即 row 0), 所以level 2 两列(column 2, column 3) 的row 0 / row 1 行对应的level 1 的列(column 1) 的last_level_index 为 0
6. column 1 这一列 col1_val2 的row index 为 1 (即 row 1), 所以level 2 两列(column 2, column 3) 的row 2 / row 3 行对应的level 1 的列(column 1) 的last_level_index 为 1
7. column 1 这一列 col1_val3 的row index 为 2 (即 row 2), 所以level 2 两列(column 2, column 3) 的row 4 / row 5 / row 6 行对应的level 1 列(column 1) 的的last_level_index 为 2



## 构造上述IndexBatch
下面的示例代码演示了如何构造上图中展示的压缩后的IndexBatch

```go

// 加载特征列列名文件. 文件格式参考 data/column_name_criteo.txt
// 非必须.如果不调用GetRowFeatures(row) 的话则不需要加载列名文件.
fe.LoadColumnNameFile("column_name_file")

// 初始化一个IndexedColumn
// 3表示IndexBatch的level, 必须为3.
// 4 表示有多少列特征,可以不准确; 该值用来指示IndexBatch内部预分配的空间.
ib := NewIndexedColumn(3, 4)

// 添加特征列.
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
###  函数说明
1. func NewIndexedColumn(level, column uint64) *IndexedColumn <br>
    初始化一个新的IndexBatch，后续会在该IndexBatch上添加特征. <br>
    参数level表示该IndexBatch的特征的层级。该值目前必须为3. <br>
    参数column表示该IndexBatch总共有列特征。该值可以不准确，因为该值用来指示我们要预分配多少内存。 <br>
2. func (this *IndexedColumn) Free() <br>
    释放IndexBatch. 为了提高性能, 该函数调用会将IndexBatch的底层存储放到内存池中. <br>
3. func (this *IndexedColumn) AddColumn(key, value string, level int, pre int) <br>
    参数key表示特征列的名字，参数value表示要添加的特征的value，参数level表示该列特征所在的level.<br>
    参数pre表示该新添加的特征对应的上一级特征的index. <br>
    如果key所指定的column不存在于IndexBatch中的话，则会向IndexBatch添加一个特征列，该列的名字即为key。 <br>
    该函数会向key所指定的column添加一个value，所以该函数调用会使该特征列的行数增加一行。 <br>
4. func (this *IndexedColumn) AddColumnArray(key string, values []string, level int, pre int) <br>
    同AddColumn类似，只不过该函数添加的是一个多值特征. <br>
5. func (this *IndexedColumn) GetRowFeatures(row int) string <br>
    获取IndexBatch中的某一行特征. 该函数会将该行特征转换为string表示, 方便用户用于事后的分析、特征回流等等. 可结合LoadFromCsvFile()使用 <br>
6. func (this *IndexedColumn) LoadFromCsvFile(csv_file, column_name_file string, gen_with_score bool) <br>
    从csv_file 文件中加载数据，并将其转换为IndexBatch. <br>
    csv_file文件中的数据即为GetRowFeatures()返回的数据(用户也可以在将GetRowFeatures()返回的数据落文件前加上该行预测得到的score信息). <br>
    column_name_file中保存有各个特征列的名字, 其格式参考文件 [column_name_file](/data/column_name_criteo.txt) <br>
    gen_with_score： 如果用户将GetRowFeatures()返回的数据直接落文件，则将该值设置为false；如果用户在GetRowFeatures()返回的数据前添加了该行对应的score值，则该值设置为true.
 
### 
