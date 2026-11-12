# wt-go

Windows Terminal 风格的终端：多标签、分栏、profile、配色方案与可录制快捷键，设置是一份 `settings.json`。

## 快速开始

```pwsh
git clone https://github.com/KiLiMe/wt-go.git; cd wt-go   # 1 取源码
pwsh ./Scripts/build.ps1                                  # 2 构建（Linux/macOS：./Scripts/build.sh）
./build/windows-amd64/wt-go.exe                           # 3 运行
```

发布页每个平台一个便携 zip：Windows 是 `wt-go.exe`、`ghostty-vt.dll` 与 `LICENSE`，Linux 是 `wt-go`、`libghostty-vt.so` 与 `LICENSE`；解压即用，没有安装器、也没有 `.deb`。
构建脚本参数：`-Test` 先跑 vet 与测试，`-Clean` 清掉 `build/`，`-Package` 只出 `dist/`。

## 依赖环境

| 项 | 要求 | 出处 |
| --- | --- | --- |
| Go | 1.27.3 | `go.mod` 的 `go` 行，CI 用 `go-version-file: go.mod` |
| mygo | v0.8.3（提供 `go tool mygo`、UI 与终端插件） | `go.mod` 的 `require` 与 `tool` 行 |
| purego / golang.org/x/sys | v0.11.1 / v0.48.0 | `go.mod` 的 `require` |
| Windows | 10 1809+ / 11，需要 ConPTY | 主要支持与验证目标 |
| Linux | glibc ≥ 2.28；GTK 3、Pango、Cairo 运行库；CJK 字体 | [docs/Linux.md](docs/Linux.md) 第 1 节 |
| macOS | 能编译，CI 只跑测试、不打包 | `.github/workflows/ci.yml` 的构建矩阵 |
| 构建工具 | 无额外依赖；PowerShell 7+（Windows）或 bash（Linux/macOS） | `Scripts/build.ps1` / `Scripts/build.sh` |

## 许可

wt-go 本身用 [MIT 许可证](LICENSE)，版权归 Copyright (c) 2026 KiLiMe：可以自由使用、修改、
再分发（含商用），只需在所有副本里保留版权与许可声明。软件按“原样”提供，不含任何担保。

随二进制分发的第三方组件各自保留原许可：

| 组件 | 用途 | 许可 |
| --- | --- | --- |
| mygo v0.8.3 | UI、终端插件与打包工具 | MIT（`LICENSE`） |
| libghostty-vt（Ghostty） | 终端内核，即 `ghostty-vt.dll` / `libghostty-vt.so` | MIT，见 [Ghostty 上游 LICENSE](https://github.com/ghostty-org/ghostty/blob/main/LICENSE) |
| Remix Icon v4.8.0 | 界面图标，`internal/icons/svg/` | Apache-2.0（全文见 `internal/icons/LICENSE`，出处见 `internal/icons/NOTICE`） |

开发与提交约定见 [AGENTS.md](AGENTS.md)。
