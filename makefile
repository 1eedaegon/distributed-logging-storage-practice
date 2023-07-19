# Reflect에 의한 동적 컴파일 template 생성

CONFIG_PATH=${HOME}/workspace/golang/distributed-logging-storage-practice/config-path-test

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: pre-gencert
pre-gencert:
	go install github.com/cloudflare/cfssl/cmd/...

.PHONY: gencert
gencert:
	cfssl gencert -initca test/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server test/server-csr.json | cfssljson -bare server
	
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client test/client-csr.json | cfssljson -bare client
	mv *.pem *.csr ${CONFIG_PATH}

.PHONY: test
test:
	go test -race ./... -v

.PHONY: compile
compile:
	protoc api/v1/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.
