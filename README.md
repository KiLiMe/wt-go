# wt-go

Windows Terminal 风格的终端：多标签、分栏、profile、配色方案与可录制快捷键，设置是一份 `settings.json`。

## 快速开始

```pwsh
git clone https://github.com/KiLiMe/wt-go.git; cd wt-go   # 1 取源码
pwsh ./Scripts/build.ps1                                  # 2 构建（Linux/macOS：./Scripts/build.sh）
./build/windows-amd64/wt-go.exe                           # 3 运行
```

发布页每个平台一个便携 zip：Windows 含 `wt-go.exe`、`ghostty-vt.dll`、`conpty\` 与 `data\`（便携模式），Linux 含 `wt-go`、`libghostty-vt.so`、`data/` 与 `wt-go.desktop`，均带 `LICENSE`；解压即用，没有安装器或 `.deb`。
构建脚本参数：`-Test` / `--test` 先跑 gofmt 与 vet、测试，`-Clean` / `--clean` 清掉 `build/`，`-Package` / `--package` 只出 `dist/`，`--install` 装到本机（Linux：`~/.local/opt/wt-go` 与菜单项，`data/` 原样保留）。

## 依赖环境

| 项 | 要求 | 出处 |
| --- | --- | --- |
| Go | 1.27.3 | `go.mod` |
| mygo | fork 分支 `wt-go-pin`（见 `go.mod` `replace` 行） | `go.mod` |
| Windows | Windows 10 1809+ / 11（ConPTY） | 主要支持与验证目标 |
| Linux（实验性） | glibc ≥ 2.28；GTK 3、Pango、Cairo（purego 调用，无需 cgo/WebKitGTK）；CJK 字体；`GTK_IM_MODULE` | 见下方安装命令 |
| macOS（实验性） | 可编译，CI 仅跑测试不打包 | `.github/workflows/ci.yml` |

Linux 运行库与字体安装：

```sh
# Debian / Ubuntu
sudo apt install libgtk-3-0 libpango-1.0-0 libpangocairo-1.0-0 fonts-noto-cjk
# Fedora
sudo dnf install gtk3 pango pangocairo google-noto-sans-cjk-fonts
# Arch
sudo pacman -S gtk3 pango noto-fonts-cjk
```

## 文档

- [使用说明](docs/使用说明.md)：功能、交互、设置页与平台差异
- [配置参考](docs/配置参考.md)：`settings.json` 全字段、路径写法与默认键位
- [故障排查](docs/故障排查.md)：`wt-go diag`、故障码 `E1xx`–`E3xx` 与 `wt-go mcp`
- [代码导读](docs/代码导读.md)：包边界、渲染主循环、关键暗坑与发版流程
- [架构决策](docs/decisions.md)：关键架构与设计取舍（ADR）
- 开发约定见 [AGENTS.md](AGENTS.md)

## 许可

wt-go 采用 [MIT 许可证](LICENSE)（Copyright (c) 2026 KiLiMe）。随二进制分发的第三方组件保留原许可：

| 组件 | 用途 | 许可 |
| --- | --- | --- |
| mygo（`KiLiMe/mygo`） | UI、终端插件与打包工具 | MIT |
| libghostty-vt（Ghostty） | 终端内核（`ghostty-vt.dll` / `libghostty-vt.so`） | MIT |
| Windows Terminal ConPTY | Windows 随包控制台（`conpty.dll` / `OpenConsole.exe`） | MIT（`resources/windows-*/conpty/ConPTY-LICENSE.txt`） |
| Remix Icon v4.8.0 | 界面图标（`internal/icons/svg/`） | Apache-2.0（`internal/icons/LICENSE`） |
