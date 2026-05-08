#!/usr/bin/env bash
# Regenerate gRPC stubs from .proto files into proto/<svc>pb/.
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc.
# On Windows, the plugin binaries end in `.exe`, which protoc cannot resolve
# implicitly, so we always pass them via --plugin=<name>=<path>.

set -euo pipefail

cd "$(dirname "$0")/.."

PROTOC_BIN="${PROTOC:-protoc}"

if ! command -v "${PROTOC_BIN}" >/dev/null 2>&1; then
  echo "protoc not found on PATH (override with PROTOC=...)" >&2
  exit 1
fi

GOBIN="$(go env GOPATH)/bin"

# Pick `.exe` variants on Windows, plain names elsewhere.
GO_PLUGIN="${GOBIN}/protoc-gen-go"
GRPC_PLUGIN="${GOBIN}/protoc-gen-go-grpc"
if [ ! -x "${GO_PLUGIN}" ] && [ -x "${GO_PLUGIN}.exe" ]; then
  GO_PLUGIN="${GO_PLUGIN}.exe"
fi
if [ ! -x "${GRPC_PLUGIN}" ] && [ -x "${GRPC_PLUGIN}.exe" ]; then
  GRPC_PLUGIN="${GRPC_PLUGIN}.exe"
fi

for tool in "${GO_PLUGIN}" "${GRPC_PLUGIN}"; do
  if [ ! -x "${tool}" ]; then
    echo "${tool} not found. Install with:" >&2
    echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" >&2
    echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" >&2
    exit 1
  fi
done

"${PROTOC_BIN}" \
  --plugin=protoc-gen-go="${GO_PLUGIN}" \
  --plugin=protoc-gen-go-grpc="${GRPC_PLUGIN}" \
  --proto_path=proto \
  --go_out=. --go_opt=module=mangahub-backend \
  --go-grpc_out=. --go-grpc_opt=module=mangahub-backend \
  proto/catalog.proto \
  proto/artist.proto \
  proto/progress.proto \
  proto/prefs.proto

echo "proto stubs regenerated."
