module github.com/ulfbj/datalake_utils

go 1.24.0

//example on how to use replace to point to fork or local path
//replace github.com/ulfbj/datalake_utils => ./datalake_utils

require (
	github.com/akamensky/argparse v1.3.1
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/gorilla/websocket v1.5.3
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
)

require (
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
)
