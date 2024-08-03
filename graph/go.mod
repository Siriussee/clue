module github.com/anon/execution-property-graph/graph

go 1.19

replace github.com/anon/execution-property-graph/dynamicEVM => ../dynamicEVM

replace github.com/anon/execution-property-graph/go-ethereum-driver => ../go-ethereum-driver

replace github.com/anon/execution-property-graph/trace-emulator => ../trace-emulator

require (
	github.com/anon/execution-property-graph/dynamicEVM v0.0.0
	github.com/anon/execution-property-graph/go-ethereum-driver v0.0.0
	github.com/anon/execution-property-graph/trace-emulator v0.0.0
	github.com/apache/tinkerpop/gremlin-go/v3 v3.6.1
	github.com/deckarep/golang-set/v2 v2.1.0
	github.com/ethereum/go-ethereum v1.10.23
	github.com/pkg/profile v1.7.0
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli/v2 v2.10.2
	go.mongodb.org/mongo-driver v1.11.1
)

require (
	github.com/StackExchange/wmi v0.0.0-20180116203802-5d049714c4a6 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/go-ole/go-ole v1.2.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20230309165930-d61513b1440d // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/holiman/uint256 v1.2.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/nicksnyder/go-i18n/v2 v2.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/tklauser/numcpus v0.2.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)
