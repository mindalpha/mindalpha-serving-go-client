# data/
[中文文档](README.CN.md) <br>

Data and files used by MindAlpha-Serving-go-client test code <br>

File list <br>
1. day_0_0.001_train.csv <br>
2. day_0_0.001_train-ib-format.csv <br>
3. column_name_criteo.txt <br>

## File Manual
1. day_0_0.001_train.csv <br>
    Data source: https://www.kaggle.com/c/criteo-display-ad-challenge/data <br>
    This file has 5 row data, for test and demo <br>
2. day_0_0.001_train-ib-format.csv <br>
    This file comes from day_0_0.001_train.csv, but has a different format <br>
    day_0_0.001_train.csv is seperated by \t ,  day_0_0.001_train-ib-format.csv is seperated by \002  <br>
    mindalpha-serving-go-client code needs csv file must be seperated by \002 <br>
    You can translate day_0_0.001_train.csv to day_0_0.001_train-ib-format.csv by command: sed -i $'s/\t/\002/g' day_0_0.001_train.csv
3. column_name_criteo.txt <br>
    Names of Feature columns of day_0_0.001_train-ib-format.csv <br>
    Users can name feature columns by yourself, but the column names must be the same as the MindAlpha-Serving side used column name.
