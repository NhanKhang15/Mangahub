#requires -Version 5.1
# Regenerate gRPC stubs from .proto files (Windows PowerShell variant of proto-gen.sh).

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $repoRoot

$gobin = Join-Path ((& go env GOPATH).Trim()) "bin"

$goPlugin   = Join-Path $gobin "protoc-gen-go.exe"
$grpcPlugin = Join-Path $gobin "protoc-gen-go-grpc.exe"

foreach ($p in @($goPlugin, $grpcPlugin)) {
    if (-not (Test-Path $p)) {
        Write-Error "$p not found. Install with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    }
}

if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    Write-Error "protoc not found on PATH"
}

& protoc `
    "--plugin=protoc-gen-go=$goPlugin" `
    "--plugin=protoc-gen-go-grpc=$grpcPlugin" `
    --proto_path=proto `
    --go_out=. --go_opt=module=mangahub-backend `
    --go-grpc_out=. --go-grpc_opt=module=mangahub-backend `
    proto/catalog.proto `
    proto/artist.proto `
    proto/progress.proto `
    proto/prefs.proto

Write-Host "proto stubs regenerated."
