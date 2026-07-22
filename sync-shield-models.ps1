#!/usr/bin/env powershell
# sync-shield-models.ps1
# Queries iTaK Shield's /v1/models endpoint and generates the JSON block
# for opencode.json's "shield" provider models section.
#
# Usage:
#   .\sync-shield-models.ps1                    # prints JSON to stdout
#   .\sync-shield-models.ps1 -OutFile models.json  # writes to file
#   .\sync-shield-models.ps1 | Clip              # copies to clipboard
#
# This lets you add a new model to shield.yaml, restart Shield, run this
# script, and paste the updated model list into opencode.json — no manual
# model ID typing needed.

param(
    [string]$ShieldUrl = "http://127.0.0.1:20979",
    [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"

try {
    $resp = Invoke-RestMethod -Uri "$ShieldUrl/v1/models" -TimeoutSec 10
} catch {
    Write-Host "ERROR: Cannot reach iTaK Shield at $ShieldUrl" -ForegroundColor Red
    Write-Host "Start it with: itak-shield.exe --config shield.yaml --target http://placeholder --port 20979 --no-gui" -ForegroundColor Yellow
    exit 1
}

if (-not $resp.data -or $resp.data.Count -eq 0) {
    Write-Host "ERROR: Shield returned no models. Check shield.yaml gateway.providers." -ForegroundColor Red
    exit 1
}

# Known model capabilities (context/output limits, reasoning, multimodal).
# Unknown models get sensible defaults: 32k context, 8k output, tool_call on.
$known = @{
    "glm-5.2"               = @{ ctx=131072; out=16384; reasoning=$true;  tool=$true;  image=$false }
    "qwen3.8-max-preview"   = @{ ctx=262144; out=32768; reasoning=$true;  tool=$true;  image=$true  }
    "qwen3.7-max"            = @{ ctx=262144; out=32768; reasoning=$true;  tool=$true;  image=$false }
    "qwen3.7-plus"           = @{ ctx=131072; out=16384; reasoning=$true;  tool=$true;  image=$false }
    "qwen3.6-flash"          = @{ ctx=131072; out=16384; reasoning=$false; tool=$true;  image=$false }
    "deepseek-v4-pro"        = @{ ctx=131072; out=16384; reasoning=$true;  tool=$true;  image=$false }
    "deepseek-v4-flash"      = @{ ctx=1000000; out=384000; reasoning=$true; tool=$true; image=$false }
    "gpt-oss:20b-cloud"      = @{ ctx=32768; out=8192; reasoning=$false; tool=$true;  image=$false }
    "gpt-oss:120b-cloud"     = @{ ctx=32768; out=8192; reasoning=$true;  tool=$true;  image=$false }
    "qwen3-coder:480b-cloud" = @{ ctx=32768; out=8192; reasoning=$false; tool=$true;  image=$false }
    "deepseek-r1:671b-cloud" = @{ ctx=32768; out=8192; reasoning=$true;  tool=$true;  image=$false }
    "llama3.3:70b-cloud"     = @{ ctx=32768; out=8192; reasoning=$false; tool=$true;  image=$false }
    "qwen3.6-35b-a3b-mtp"    = @{ ctx=32768; out=8192; reasoning=$true;  tool=$true;  image=$false }
    "qwen3.6-35b-a3b"        = @{ ctx=16384; out=8192; reasoning=$true;  tool=$true;  image=$false }
    "qwen2.5-coder:7b"       = @{ ctx=32768; out=4096; reasoning=$false; tool=$false; image=$false }
    "qwen2.5-coder:3b"       = @{ ctx=32768; out=4096; reasoning=$false; tool=$false; image=$false }
    "qwen2.5-coder:1.5b"     = @{ ctx=32768; out=4096; reasoning=$false; tool=$false; image=$false }
    "qwen2.5-vl:7b"          = @{ ctx=32768; out=4096; reasoning=$false; tool=$false; image=$true  }
    "qwen2.5-vl:3b"          = @{ ctx=32768; out=4096; reasoning=$false; tool=$false; image=$true  }
    "qwen3vl-vla-q4:8k"     = @{ ctx=32768; out=4096; reasoning=$false; tool=$false; image=$true  }
}

# Build a PowerShell object that ConvertTo-Json will serialize correctly.
$modelsObj = [ordered]@{}
foreach ($model in $resp.data) {
    $id = $model.id
    $owner = $model.owned_by

    # Look up known caps or use defaults
    $ctx = 32768; $out = 8192; $reasoning = $false; $tool = $true; $image = $false
    if ($known.ContainsKey($id)) {
        $k = $known[$id]
        $ctx = $k.ctx; $out = $k.out; $reasoning = $k.reasoning; $tool = $k.tool; $image = $k.image
    }

    $inputMods = @("text")
    if ($image) { $inputMods = @("text", "image") }

    # Build a nice display name from the model ID
    $niceName = $id -replace ":", " " -replace "-", " " -replace "_", " "
    $niceName = (Get-Culture).TextInfo.ToTitleCase($niceName.ToLower()) + " (via Shield)"

    $modelsObj[$id] = [ordered]@{
        name        = $niceName
        attachment  = $image
        reasoning   = $reasoning
        temperature = $true
        tool_call   = $tool
        modalities  = [ordered]@{ input = $inputMods; output = @("text") }
        limit       = [ordered]@{ context = $ctx; output = $out }
    }
}

$json = $modelsObj | ConvertTo-Json -Depth 5

# Indent for pasting inside the "shield" provider block
$indented = ($json -split "`n" | ForEach-Object { "      $_" }) -join "`n"

if ($OutFile -ne "") {
    $indented | Set-Content -Path $OutFile -Encoding UTF8
    Write-Host "Written $($resp.data.Count) models to $OutFile" -ForegroundColor Green
} else {
    Write-Host $indented
}