.PHONY: proto
proto:
	protoc --go_out=./pb pb/*.proto

.PHONY: run
run:
	go run main.go