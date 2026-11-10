#Requires -Version 7.0
<#
  重新生成 exe 里的资源：resources\icon.png（图标）与 rsrc_windows_amd64.syso
  （图标 + 版本信息）。版本号取自 mygo.json。

    pwsh .\tools\resync.ps1

  只有改了 tools\genicon 的图形、或 mygo.json 的 version 变了才需要跑。
  产物已入库，日常 go build 与 build.ps1 都不需要 go-winres。
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot

if (-not (Get-Command go-winres -ErrorAction SilentlyContinue)) {
    Write-Host '  [ERR] 找不到 go-winres：go install github.com/tc-hib/go-winres@latest' -ForegroundColor Red
    exit 1
}

$version = (Get-Content -LiteralPath (Join-Path $Root 'mygo.json') -Raw | ConvertFrom-Json).version
if ($version -notmatch '^\d+\.\d+\.\d+$') {
    Write-Host "  [ERR] mygo.json 的 version 应是三段数字：'$version'" -ForegroundColor Red
    exit 1
}
$quad = "$version.0"

Push-Location $Root
try {
    & go run ./tools/genicon
    if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] genicon 失败。' -ForegroundColor Red; exit $LASTEXITCODE }

    # 放在主包目录（仓库根）里，go build 会把它链进 exe；mygo 看到应用自己的 .syso 就不再生成。
    & go-winres simply --icon resources\icon.png --manifest gui --product-name wt-go `
        --file-description 'Terminal multiplexer in one window' `
        --file-version $quad --product-version $quad --out rsrc --arch amd64
    if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] go-winres 失败。' -ForegroundColor Red; exit $LASTEXITCODE }
} finally {
    Pop-Location
}

Write-Host "  [OK] resources\icon.png 与 rsrc_windows_amd64.syso（v$version）" -ForegroundColor Green
