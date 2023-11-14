FROM quay.io/ibmdpdev/golang:1.19.3-bullseye AS builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Copy the go source
COPY cmd/ cmd/
COPY auth/ auth/

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

#build all the executables at once, they will be copied out into individual images 
RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/token_proxy cmd/http_proxy_token_auth/proxy.go

FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV BIN_DIR=/usr/local/bin
RUN mkdir -p ${BIN_DIR}/

COPY --from=builder /workspace/bin/token_proxy ${BIN_DIR}/token_proxy

USER 1001
CMD ["/usr/local/bin/token_proxy"]
