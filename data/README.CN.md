# data/ 数据文件说明
[English Document](README.md)<br>

mindalpha-serving-go-client 测试代码使用的数据文件.<br>

文件列表如下<br>
1. day_0_0.001_train.csv <br>
2. day_0_0.001_train-ib-format.csv <br>
3. column_name_criteo.txt <br>

## 各个文件说明
1. day_0_0.001_train.csv <br>
    数据来源及格式请参考: https://www.kaggle.com/c/criteo-display-ad-challenge/data <br>
    该文件中只包含5行数据，用来测试及演示.
2. day_0_0.001_train-ib-format.csv <br>
    该文件内容来源自day_0_0.001_train.csv,只不过格式略有不同. <br>
    day_0_0.001_train.csv 文件以 \t 作为分隔符, 而 day_0_0.001_train-ib-format.csv 文件以 \002 作为分隔符. <br>
    mindalpha-serving-go-client代码要求csv文件必须以 \002 作为分隔符. <br>
    将day_0_0.001_train.csv 转换为 day_0_0.001_train-ib-format.csv 的命令: sed -i $'s/\t/\002/g' day_0_0.001_train.csv
3. column_name_criteo.txt <br>
    day_0_0.001_train-ib-format.csv文件的特征列名字. 特征列的名字是我们自己定义的.<br>
    用户可以根据自己需要自行定义特征列名字. 本客户端使用的特征列名字必须和MindAlpha-Serving端用的特征列名字完全一致.
