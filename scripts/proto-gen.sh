#!/usr/bin/env bash
set -euo pipefail
cd /workspace
if [ ! -f third_party/googleapis/google/api/annotations.proto ]; then
  mkdir -p third_party
  git clone --depth 1 https://github.com/googleapis/googleapis.git third_party/googleapis
fi
if [ ! -f third_party/grpc-gateway/protoc-gen-openapiv2/options/annotations.proto ]; then
  mkdir -p third_party
  git clone --depth 1 https://github.com/grpc-ecosystem/grpc-gateway.git third_party/grpc-gateway
fi
apt-get update -qq
apt-get install -y -qq protobuf-compiler git
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.9
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.26.3
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.26.3
export PATH="${PATH}:$(go env GOPATH)/bin"
protoc \
  -I third_party/googleapis \
  -I third_party/grpc-gateway \
  -I api \
  --go_out=pkg/api \
  --go_opt=paths=source_relative \
  --go-grpc_out=pkg/api \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=pkg/api \
  --grpc-gateway_opt=paths=source_relative \
  --openapiv2_out=api \
  --openapiv2_opt=logtostderr=true \
  --openapiv2_opt=allow_merge=true \
  --openapiv2_opt=merge_file_name=service \
  api/service.proto
echo "proto generation OK"
