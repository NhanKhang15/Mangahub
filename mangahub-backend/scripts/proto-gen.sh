#!/usr/bin/env bash
set -euo pipefail

# Generate Go stubs from proto/*.proto into proto/<name>pb/.
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

for f in proto/*.proto; do
  name=$(basename "$f" .proto)
  out="proto/${name}pb"
  mkdir -p "$out"
  protoc \
    --go_out="$out" --go_opt=paths=source_relative \
    --go-grpc_out="$out" --go-grpc_opt=paths=source_relative \
    -Iproto "$f"
done

echo "proto generation done."
