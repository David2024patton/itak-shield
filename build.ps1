# iTaK Shield - Cross-Platform Build Script
# Produces binaries for Windows, macOS (Intel + Apple Silicon), and Linux.

$ErrorActionPreference = "Stop"

$version = "0.2.0"
$distDir = "dist"
$module = "github.com/David2024patton/itak-shield"

# Clean dist directory.
if (Test-Path $distDir) {
    Remove-Item -Recurse -Force $distDir
}
New-Item -ItemType Directory -Path $distDir | Out-Null

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "  iTaK Shield v$version Build" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe"; Label = "Windows (64-bit)" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; Ext = ""; Label = "macOS (Intel)" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Ext = ""; Label = "macOS (Apple Silicon)" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Ext = ""; Label = "Linux (64-bit)" }
)

$success = 0
$failed = 0

foreach ($t in $targets) {
    $outFile = "$distDir/itak-shield-$($t.GOOS)-$($t.GOARCH)$($t.Ext)"
    Write-Host "  Building $($t.Label)..." -NoNewline

    $env:GOOS = $t.GOOS
    $env:GOARCH = $t.GOARCH
    $env:CGO_ENABLED = "0"

    try {
        go build -ldflags "-s -w -X main.version=$version" -o $outFile .
        $size = (Get-Item $outFile).Length / 1MB
        Write-Host " OK ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
        $success++
    }
    catch {
        Write-Host " FAILED" -ForegroundColor Red
        Write-Host "    Error: $_" -ForegroundColor Red
        $failed++
    }
}

# Reset env vars.
$env:GOOS = ""
$env:GOARCH = ""
$env:CGO_ENABLED = ""

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "  Results: $success passed, $failed failed" -ForegroundColor Cyan
Write-Host "  Output:  $distDir/" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

if ($failed -gt 0) {
    exit 1
}
