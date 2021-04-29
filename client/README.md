# client
[中文文档](README.CN.md)

## Features
IndexBatch serialization<br>
Construct network message header<br>
Deserialize response scores returned from MindAlpha-Serving service<br>

## Test cases
1. client/client_test.go: <br>
    Construct IndexBatch from csv file and then access MindAlpha-Serving service, print scores

## To be perfected

1. dumpFBS()<br>
    This function dump requests/responses to file. <br>
    There has some problems with this function: <br>
    1. Hard code the number of dumpped Request/Response and dumpped file path <br>
    2. Users need moify code to use dumpFBS <br>
    3. You should restart process if you want to re dump <br>
    4. The format of dumpped file should be user self-defined <br>
