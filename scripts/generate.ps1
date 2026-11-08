param(
    [ValidateSet('all', 'rbac', 'blog', 'product', 'file')]
    [string]$Service = 'all'
)
$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$services = if ($Service -eq 'all') { @('rbac', 'blog', 'product', 'file') } else { @($Service) }
Push-Location $projectRoot
try {
    foreach ($domain in $services) {
        & goctl api go -api "service/$domain/api/$domain.api" -dir "service/$domain/api" --style go_zero
        if ($LASTEXITCODE -ne 0) { throw "API generation failed: $domain" }
        Push-Location "service/$domain/rpc"
        try {
            & goctl rpc protoc "$domain.proto" --go_out=. --go-grpc_out=. --zrpc_out=. --style go_zero
            if ($LASTEXITCODE -ne 0) { throw "RPC generation failed: $domain" }
        } finally { Pop-Location }
    }
    & go mod tidy
    if ($LASTEXITCODE -ne 0) { throw 'go mod tidy failed' }
    & go build ./...
    if ($LASTEXITCODE -ne 0) { throw 'generated services do not build' }
} finally { Pop-Location }
