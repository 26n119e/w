.PHONY: proto
proto:
	protoc --go_out=./pb pb/*.proto

.PHONY: server
run:
	go run server.go

.PHONY: client
run:
	go run client.go