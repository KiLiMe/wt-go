#Requires -Version 7.0
<#
  wt-go 构建脚本
    pwsh .\Scripts\build.ps1            -> mygo 打包到 build\windows-amd64\wt-go.exe（真机测试就用它）
    pwsh .\Scripts\build.ps1 -Test      -> 编译前先跑 go vet 与 go test
    pwsh .\Scripts\build.ps1 -Clean     -> 先删掉 build 目录再编译
    pwsh .\Scripts\build.ps1 -Package   -> 只出分发包：go tool mygo build -o dist（不造安装器）

  图标、版本信息、manifest 与 native 库由 mygo 在打包时注入（go build 带不了资源），
  所以本地构建的产物与发布版一致。产物目录由 mygo.json 的 out 决定（build\），
  目录名带平台，升级清单 .mygo-files 才能与产物对上：build\ 已在 .gitignore 里。
#>
[CmdletBinding()]
param(
    [switch]$Clean,
    [switch]$Test,
    [switch]$Package
)

$ErrorActionPreference = 'Stop'
# The script sits in Scripts\; the repository is the directory above it, which
# go and mygo both run from.
$Root = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $Root 'build'
$Exe = Join-Path $BuildDir 'windows-amd64\wt-go.exe'

if (-not (Get-Command go.exe -ErrorAction SilentlyContinue)) {
    Write-Host '  [ERR] 找不到 go.exe，请先安装 Go 1.27+。' -ForegroundColor Red
    exit 1
}

if ($Clean -and (Test-Path -LiteralPath $BuildDir)) {
    Remove-Item -LiteralPath $BuildDir -Recurse -Force
}

Push-Location $Root
try {
    if ($Test) {
        & go vet ./...
        if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] go vet 失败。' -ForegroundColor Red; exit $LASTEXITCODE }
        & go test ./internal/...
        if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] 测试失败。' -ForegroundColor Red; exit $LASTEXITCODE }
    }

    if ($Package) {
        & go tool mygo build -platform windows/amd64 -skip-installer -skip-dmg -o dist
        if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] 打包失败。' -ForegroundColor Red; exit $LASTEXITCODE }
        Write-Host "  [OK] $(Join-Path $Root 'dist')" -ForegroundColor Green
        return
    }

    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Exe) | Out-Null

    # 正在跑的测试实例占着输出文件，Windows 会拒绝覆盖：先关掉再编译。
    $busy = @(Get-Process -Name 'wt-go' -ErrorAction SilentlyContinue | Where-Object {
        try { $_.Path -eq $Exe } catch { $false }
    })
    if ($busy.Count -gt 0) {
        $ids = ($busy | ForEach-Object { $_.Id }) -join ', '
        Write-Host "  [ERR] 正在运行的 wt-go.exe（PID $ids）占着输出文件：先关掉它的窗口再编译。" -ForegroundColor Red
        exit 1
    }

    # 图标、版本信息、manifest 与 libghostty-vt 都靠 mygo 打包时注入，
    # go build 只会出一个没图标没版本号的裸 exe。mygo 自己按 out 写 build\<os>-<arch>\，
    # 不另设中转目录，产物才与 .mygo-files 的清单一致。
    & go tool mygo build -platform windows/amd64 -skip-installer -skip-dmg
    if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] 编译失败。' -ForegroundColor Red; exit $LASTEXITCODE }
} finally {
    Pop-Location
}

$size = [math]::Round((Get-Item -LiteralPath $Exe).Length / 1MB, 2)
Write-Host "  [OK] $Exe ($size MB)" -ForegroundColor Green
