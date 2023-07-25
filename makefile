
CONFIG_PATH=${HOME}/workspace/golang/distributed-logging-storage-practice/config-path-test

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: pre-gencert
pre-gencert:
	go install github.com/cloudflare/cfssl/cmd

.PHONY: gencert
gencert: init
	cfssl gencert -initca test/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client test/client-csr.json | cfssljson -bare client

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server test/server-csr.json | cfssljson -bare server
	
	# Client 1
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=client \
		-cn="nobody" \
		test/client-csr.json | cfssljson -bare nobody-client
	# Client 2
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-cn="root" \
		-profile=client test/client-csr.json | cfssljson -bare root-client
				
	mv *.pem *.csr ${CONFIG_PATH}

$(CONFIG_PATH)/model.conf:
	cp test/model.conf $(CONFIG_PATH)/model.conf
	
$(CONFIG_PATH)/policy.csv:
	cp test/policy.csv $(CONFIG_PATH)/policy.csv
	
.PHONY: test
test: gencert $(CONFIG_PATH)/model.conf $(CONFIG_PATH)/policy.csv
	go test -race ./... -v

# Reflect에 의한 동적 컴파일 template 생성
.PHONY: compile
compile:
	protoc api/v1/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.
