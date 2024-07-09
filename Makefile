PROTO_SRC_DIR := $(shell pwd)/proto
PROTO_DST_DIR := $(shell pwd)/internal/pb

RMQ_URL := "amqp://guest:guest@localhost:5672/"
ETCD_ENDPOINT := "localhost:2379"

proto: ${PROTO_SRC_DIR}
	test -d ${PROTO_DST_DIR} || mkdir -p ${PROTO_DST_DIR}
	protoc \
		-I ${PROTO_SRC_DIR} \
		-I ${PROTO_SRC_DIR}/third_party/googleapis \
		--go_opt=paths=source_relative \
		--go_out=${PROTO_DST_DIR} \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_out=${PROTO_DST_DIR} \
		${PROTO_SRC_DIR}/calculator.proto


test:
	go test -v ./... -race

integration:
	RMQ_URL=${RMQ_URL} ETCD_ENDPOINT=${ETCD_ENDPOINT} go test -v ./... -race

.PHONY: proto, test
