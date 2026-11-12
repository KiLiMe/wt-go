#Requires -Version 7.0
<#
  wt-go 构建脚本
    pwsh .\Scripts\build.ps1            -> mygo 打包到 build\windows-amd64\wt-go.exe（真机测试就用它）
    pwsh .\Scripts\build.ps1 -Test      -> 编译前先跑 go vet 与 go test
    pwsh .\Scripts\build.ps1 -Clean     -> 先删掉 build 目录再编译
    pwsh .\Scripts\build.ps1 -Package   -> 只出分发包：go tool mygo build -o dist（不造安装器）
    pwsh .\Scripts\build.ps1 -Version 0.3.0
                                        -> 断言模式：mygo.json 里的版本必须是这一串（发版时对 tag 用）

  图标、版本信息、manifest 与 native 库由 mygo 在打包时注入（go build 带不了资源），
  所以本地构建的产物与发布版一致。产物目录由 mygo.json 的 out 决定（build\），
  目录名带平台，升级清单 .mygo-files 才能与产物对上：build\ 已在 .gitignore 里。

  版本号只有一个来源：mygo.json 的 version。mygo build 把它写进 exe 的版本信息，
  打包完这里再从 exe 属性里读回来对账，对不上就报错——免得产物自称的版本与
  mygo.json 里写的不是一个数。
#>
[CmdletBinding()]
param(
    [switch]$Clean,
    [switch]$Test,
    [switch]$Package,
    [string]$Version = ''
)

$ErrorActionPreference = 'Stop'
# The script sits in Scripts\; the repository is the directory above it, which
# go and mygo both run from.
$Root = Split-Path -Parent $PSScriptRoot
$Config = Join-Path $Root 'mygo.json'
$BuildDir = Join-Path $Root 'build'
$Exe = Join-Path $BuildDir 'windows-amd64\wt-go.exe'

if (-not (Get-Command go.exe -ErrorAction SilentlyContinue)) {
    Write-Host '  [ERR] 找不到 go.exe，请先安装 Go 1.27+。' -ForegroundColor Red
    exit 1
}
if (-not (Test-Path -LiteralPath $Config)) {
    Write-Host "  [ERR] 找不到 $Config：版本、图标与产物名都在它里面。" -ForegroundColor Red
    exit 1
}

# 版本：mygo.json 是唯一来源，-Version 只是断言。
$target = [string](Get-Content -LiteralPath $Config -Raw | ConvertFrom-Json).version
if ($target -notmatch '^\d+\.\d+\.\d+$') {
    Write-Host "  [ERR] mygo.json 的 version 要是不带 v 的三段数字（tag 去掉 v）：'$target'。" -ForegroundColor Red
    exit 1
}
if ($Version -and $Version -ne $target) {
    Write-Host "  [ERR] 版本对不上：-Version 是 $Version，mygo.json 里是 '$target'。" -ForegroundColor Red
    exit 1
}

# 让 wt-go version 也能报出发布版本：build info 说的是“哪个构建回答的”，
# 这个值是 mygo.json 宣布的版本。mygo 把自己的 -ldflags 与这个值拼在一起，
# 所以这里不带 -ldflags= 前缀。
$env:MYGO_LDFLAGS = "-X wtgo/internal/cli.packagedVersion=$target"

if ($Clean -and (Test-Path -LiteralPath $BuildDir)) {
    Remove-Item -LiteralPath $BuildDir -Recurse -Force
}

Push-Location $Root
try {
    if ($Test) {
        & go vet ./...
        if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] go vet 失败。' -ForegroundColor Red; exit $LASTEXITCODE }
        # ./... 而不是 ./internal/...：locales 的覆盖测试也在 CI 里跑，本地不该少这一份。
        & go test ./...
        if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] 测试失败。' -ForegroundColor Red; exit $LASTEXITCODE }
    }

    if ($Package) {
        & go tool mygo build -platform windows/amd64 -skip-installer -skip-dmg -o dist
        if ($LASTEXITCODE -ne 0) { Write-Host '  [ERR] 打包失败。' -ForegroundColor Red; exit $LASTEXITCODE }
        $dist = Join-Path $Root 'dist\windows-amd64\wt-go.exe'
        $fileVersion = (Get-Item -LiteralPath $dist).VersionInfo.FileVersion
        if ($fileVersion -ne $target) {
            Write-Host "  [ERR] dist 里 exe 属性中的版本是 '$fileVersion'，mygo.json 是 '$target'。" -ForegroundColor Red
            exit 1
        }
        Write-Host "  [OK] $(Join-Path $Root 'dist') v$target" -ForegroundColor Green
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
# exe 属性里的版本来自 mygo.json；读回来对账，免得产物自称的版本与文件里写的不是一回事。
$fileVersion = (Get-Item -LiteralPath $Exe).VersionInfo.FileVersion
if ($fileVersion -ne $target) {
    Write-Host "  [ERR] exe 属性里的版本是 '$fileVersion'，mygo.json 是 '$target'：查有没有别的 .syso 占了资源，或者重新打包。" -ForegroundColor Red
    exit 1
}
Write-Host "  [OK] $Exe v$target ($size MB)" -ForegroundColor Green
