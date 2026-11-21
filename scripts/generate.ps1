param(
    [ValidateSet('all', 'rbac', 'blog', 'product')]
    [string]$Service = 'all'
)
$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$services = if ($Service -eq 'all') { @('rbac', 'blog', 'product') } else { @($Service) }
Push-Location $projectRoot
try {
    foreach ($domain in $services) {
        & python scripts/generate-api.py $domain
        if ($LASTEXITCODE -ne 0) { throw "API generation failed: $domain" }
        if ($domain -ne 'rbac') { continue }
        Push-Location "service/$domain/rpc"
        try {
            & goctl rpc protoc "$domain.proto" --go_out=. --go-grpc_out=. --zrpc_out=. --style go_zero
            if ($LASTEXITCODE -ne 0) { throw "RPC generation failed: $domain" }
        } finally { Pop-Location }
    }
    & go build ./...
    if ($LASTEXITCODE -ne 0) { throw 'generated services do not build' }
} finally { Pop-Location }
