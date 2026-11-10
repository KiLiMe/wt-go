#Requires -Version 7.0
<#
  wt-go 构建脚本
    pwsh .\build.ps1            -> 编译到 build\bin\Release\wtgo.exe（真机测试就用它）
    pwsh .\build.ps1 -Test      -> 编译前先跑 go vet 与 go test
    pwsh .\build.ps1 -Clean     -> 先删掉 build 目录再编译
    pwsh .\build.ps1 -Package   -> 换用 mygo 打包：go tool mygo build -o dist

  exe 要放在固定位置才好反复测试：build\ 已在 .gitignore 里，产物不会被提交。
#>
[CmdletBinding()]
param(
    [switch]$Clean,
    [switch]$Test,
    [switch]$Package
)

$ErrorActionPreference = 'Stop'
$Root = $PSScriptRoot
$BuildDir = Join-Path $Root 'build'
$Exe = Join-Path $BuildDir 'bin\Release\wtgo.exe'

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
        & go tool mygo build -platform windows/amd64 -o dist
        if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] 打包失败。' -ForegroundColor Red; exit $LASTEXITCODE }
        Write-Host "  [OK] $(Join-Path $Root 'dist')" -ForegroundColor Green
        return
    }

    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Exe) | Out-Null

    # 正在跑的测试实例占着输出文件，Windows 会拒绝覆盖：先关掉再编译。
    $busy = @(Get-Process -Name wtgo -ErrorAction SilentlyContinue | Where-Object {
        try { $_.Path -eq $Exe } catch { $false }
    })
    if ($busy.Count -gt 0) {
        $ids = ($busy | ForEach-Object { $_.Id }) -join ', '
        Write-Host "  [ERR] 正在运行的 wtgo.exe（PID $ids）占着输出文件：先关掉它的窗口再编译。" -ForegroundColor Red
        exit 1
    }

    $ldflags = '-s -w'
    if ($IsWindows) {
        # 终端窗口就是界面，不需要控制台。
        $ldflags += ' -H=windowsgui'
    }
    & go build -trimpath -ldflags $ldflags -o $Exe .
    if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] 编译失败。' -ForegroundColor Red; exit $LASTEXITCODE }
} finally {
    Pop-Location
}

$size = [math]::Round((Get-Item -LiteralPath $Exe).Length / 1MB, 2)
Write-Host "  [OK] $Exe ($size MB)" -ForegroundColor Green
