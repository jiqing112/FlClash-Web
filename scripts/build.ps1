# Build flclash-web for all desktop targets (PowerShell).
# Requires: Go >= 1.27 (GOTOOLCHAIN=auto fetches it), Node.js >= 20.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $root
Set-Location $root

Write-Host "== 1/3 building frontend =="
Push-Location frontend
npm install --no-audit --no-fund
npm run build
Pop-Location

Write-Host "== 2/3 embedding frontend =="
Remove-Item -Recurse -Force backend/internal/api/dist -ErrorAction SilentlyContinue
Copy-Item -Recurse frontend/dist backend/internal/api/dist

Write-Host "== 3/3 building binaries =="
New-Item -ItemType Directory -Force -Path dist | Out-Null
Push-Location backend
$env:GOTOOLCHAIN = "auto"
$targets = @(
    @{ os = "windows"; arch = "amd64"; ext = ".exe" },
    @{ os = "linux"; arch = "amd64"; ext = "" },
    @{ os = "linux"; arch = "arm64"; ext = "" },
    @{ os = "darwin"; arch = "amd64"; ext = "" },
    @{ os = "darwin"; arch = "arm64"; ext = "" }
)
foreach ($t in $targets) {
    $env:GOOS = $t.os
    $env:GOARCH = $t.arch
    $env:CGO_ENABLED = "0"
    $out = "../dist/flclash-web-$($t.os)-$($t.arch)$($t.ext)"
    Write-Host "building $out"
    go build -trimpath -ldflags "-s -w" -o $out .
}
Pop-Location
$env:GOOS = ""
$env:GOARCH = ""
Write-Host "done -> dist/"
