FROM golang:1.21-alpine AS build
WORKDIR /go/src/dlsp

COPY . . 
RUN CGO_ENABLED=0 go build -o /go/bin/dlsp ./cmd/dlsp

# FROM alpine
FROM scratch
COPY --from=build /go/bin/dlsp /bin/dlsp
ENTRYPOINT [ "/bin/dlsp" ]
