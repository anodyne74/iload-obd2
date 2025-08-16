module iload-obd2

go 1.21

require (
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/influxdata/influxdb-client-go/v2 v2.12.3
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/rzetterberg/elmobd v0.0.0-20240426091703-01e7bbc11e6c
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)

require (
	github.com/brutella/can v0.0.2 // indirect
	github.com/deepmap/oapi-codegen v1.8.2 // indirect
	github.com/go-daq/canbus v0.2.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)
// This module replaces the local packages with their respective directories
// to facilitate local development and testing. 
replace (
	github.com/anodyne74/iload-obd2/capture => ./capture
	github.com/anodyne74/iload-obd2/analyze => ./analyze
	github.com/anodyne74/iload-obd2/query => ./query
	github.com/anodyne74/iload-obd2/replay => ./replay
)