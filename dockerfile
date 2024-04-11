FROM golang:1.21-alpine AS build
WORKDIR /go/src/dlsp

COPY . . 
RUN CGO_ENABLED=0 go build -o /go/bin/dlsp ./cmd/dlsp
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \
    wget -q0 /go/bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/\
$(GRPC_HEALTH_PROBE_VERSION)/grpc_health_probe-linux-amd64 && \
    chmod +x /go/bin/grpc_health_probe
# FROM alpine
FROM scratch
COPY --from=build /go/bin/dlsp /bin/dlsp
COPY --from=build /go/bin/grpc_health_probe /bin/grpc_health_probe
ENTRYPOINT [ "/bin/dlsp" ]
