# Distributed-logging-storage-practice

> 분산 로깅 스토리지 작성 예제

## Test

1. Run server

`go run cmd/server/main.go`

2. Request produce to server

```bash

❯ curl -X POST localhost:8080 -d \
> '{"record": {"value": "X5gTdUBdJ3sa"}}'
{"offset":0}

❯ curl -X POST localhost:8080 -d \
'{"record": {"value": "X5gTdUBdJ4sa"}}'
{"offset":1}

❯ curl -X POST localhost:8080 -d \
'{"record": {"value": "cr7gTdUBdJ5sa"}}'
illegal base64 data at input byte 12

❯ curl -X POST localhost:8080 -d \
'{"record": {"value": "cr7gTdUBd5sa"}}'
{"offset":2}

```

3. Request consume to server

```bash

❯ curl -X GET localhost:8080 -d \
'{"offset": 1}'
{"record":{"value":"X5gTdUBdJ4sa","offset":1}}

❯ curl -X GET localhost:8080 -d \
'{"offset": 3}'
offset not found

❯ curl -X GET localhost:8080 -d \
'{"offset": 0}'
{"record":{"value":"X5gTdUBdJ3sa","offset":0}}

```

4. Install GRPC presets(Dev)

```bash
# On ubuntu
sudo apt install -y protobuf-compiler
# Install grpc presets
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
# Update PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Implements

- [x] Prototype
- [x] Protocol buffer with struct
  - [x] Schema
  - [x] Domain
- [x] Logging package
  - [x] Storage
  - [x] Indexing
  - [x] Segment
  - [x] Log(Set of segment)
- [x] gRPC request
  - [x] Service define
  - [x] Server testing
  - [x] Client testing
- [x] Security
  - [x] TLS
  - [x] ACL
- [x] Tracing
  - [x] Metrics
  - [x] Tracing
- [x] Distributed service
  - [x] Service discovery
  - [x] Concensus
  - [x] Multiplexer
  - [x] Client load balancing
  - [x] Service discovery with LB
- [x] Deployments
  - [x] Kubernetes
    - [x] FQDN Advertising
  - [x] Helm chart

## Ref

- https://github.com/travisjeffery/proglog/tree/master/WriteALogPackage/
- https://github.com/gorilla/mux
- https://github.com/hashicorp/serf
- https://github.com/golang/protobuf
