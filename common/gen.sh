./protoc --go_out=./db ./db/*.proto
./protoc --go_out=./proto/rpc ./proto/rpcproto/*.proto
./protoc --go_out=./proto/msg ./proto/msgproto/*.proto

