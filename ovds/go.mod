module github.com/ulfbj/datalake_utils

go 1.13

//example on how to use replace to point to fork or local path
//replace github.com/ulfbj/datalake_utils => ./datalake_utils

require (
	github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
)
