# wt-go

Windows Terminal 风格的终端，用 [mygo](https://github.com/egoist/mygo)（Go 原生 UI，无 webview）写：
多标签、分栏、profile、配色方案、可配置快捷键，配置是一份手写友好的 `settings.json`。
`go build` 产出 exe 与同目录的原生库（2 MB 的 `ghostty-vt.dll`），没有 Electron。

**完整文档：[docs/使用说明.md](docs/使用说明.md)**

## 快速开始

```pwsh
pwsh .\build.ps1                 # 编译到 build\windows-amd64\wt-go.exe
.\build\windows-amd64\wt-go.exe
```

发布包用 mygo 自己的构建命令，一次产出可执行文件、原生库与 Linux 包：

```pwsh
pwsh .\build.ps1 -Package        # 等于 go tool mygo build -platform windows/amd64 -skip-installer -o dist
```

发布页只放便携包（Windows 是 `wt-go_<version>_windows-amd64_portable.zip`，里面是
`wt-go.exe` 与 `ghostty-vt.dll`）：解压双击即可运行，不需要安装器。

## 命令行

```pwsh
.\build\windows-amd64\wt-go.exe           # 不开参数就是窗口
.\build\windows-amd64\wt-go.exe diag      # 这台机器上跑不跑得起来：路径、原生库、shell、检查项
.\build\windows-amd64\wt-go.exe diag --json
.\build\windows-amd64\wt-go.exe diag --redact  # 贴到 issue 里：个人路径换成 %USERPROFILE% 之类
.\build\windows-amd64\wt-go.exe mcp       # MCP 服务器（stdin/stdout），给 AI 读现场
```

出问题时先看 [docs/故障排查.md](docs/故障排查.md)。

## 能力

| 能力 | 状态 |
| --- | --- |
| 多标签、拖拽排序、中键关闭、重命名/置色、名字量宽截断 | ✅ |
| 分栏（自动/竖/横）、焦点移动、拖动分割线 | ✅ 每标签最多 `maxPanes` 个（默认 4） |
| 查找（`Ctrl+Shift+F`）、命令面板（`Ctrl+Shift+P`） | ✅ |
| 出故障能自己说清（`wt-go diag`、故障码 `E1xx`/`E2xx`/`E3xx`、崩溃日志） | ✅ 见 [docs/故障排查.md](docs/故障排查.md) |
| 把现场交给 AI（`wt-go mcp`，Model Context Protocol） | ✅ 五个只读工具，不需要窗口开着 |
| exe 图标与版本信息 | ✅ `resources/icon.png`，`mygo build`（含 `build.ps1`）打包时注入 |
| profile：自动发现 pwsh / Windows PowerShell / cmd / WSL / POSIX shell | ✅ |
| 配色方案 | ✅ 内置 9 套 + Ghostty 617 套直用 + 自定义 |
| 字体 / 字号 / 行高 / 内边距 / 光标形状 / 窗口透明度 | ✅ |
| 设置页（表单 + 全部动作可录制键位 + 实时预览） | ✅ |
| `settings.json` 热重载、坏配置归档 | ✅ 约一秒内生效 |
| 选中/搜索/链接/剪贴板协议 | ✅ 由终端插件提供 |
| 会话恢复（`startOn`） | ⚠️ 只解析，阶段 5 |
| 每 profile 独立字体、profile 图标 | ⚠️ 字体全局，图标先只做颜色 |
| 标签搜索、交换窗格、分屏广播、背景图/亚克力、Quake 模式 | ❌ |

## 文档

| 文档 | 内容 |
| --- | --- |
| [docs/使用说明.md](docs/使用说明.md) | 运行环境、构建、标签与分栏、快捷键、`settings.json` 全字段、profile、配色、设置页、Linux、与 WT 的差异 |
| [docs/故障排查.md](docs/故障排查.md) | 每个故障码的意思、怎么确认、改什么；`wt-go diag` 与 MCP 的用法 |
| [docs/代码导读.md](docs/代码导读.md) | 代码地图：目录、包职责、依赖方向、渲染流程、改动前必读的坑、阅读顺序 |
| [docs/代码导读.md](docs/代码导读.md) | 代码地图：目录、包职责、依赖方向、渲染流程、改动前必读的坑、阅读顺序 |
| [docs/decisions.md](docs/decisions.md) | 设计决策与理由（ADR） |
| [docs/plan.md](docs/plan.md) | 阶段计划、验收标准、不做什么 |
| [AGENTS.md](AGENTS.md) | 工程约定与路径 |

## 目录

```text
main.go            入口：app.Run()
internal/app/      应用状态与动作：标签、窗格、设置读写、热重载、进程生命周期
internal/ui/       只画 State：标签栏（兼标题栏）、窗格、设置页、命令面板、查找栏
internal/action/   动作目录：id、标题、分组、默认键位
internal/fault/    故障码与一行含义
internal/cli/      不开窗口的部分：diag 报告、mcp 服务器
internal/mcp/      Model Context Protocol（JSON-RPC 2.0，标准库）
internal/pane/     窗格树与标签集合（不做 IO）
internal/layout/   窗格几何计算
internal/profile/  profile 模型与自动发现结果合并
internal/store/    settings.json 的解析、默认值、校验
internal/theme/    配色方案解析
internal/sys/      全部 OS 副作用（路径、原子写、日志、Job Object、shell 探测……）
internal/i18n/     中英文案
tools/genicon/     画 resources/icon.png（go run ./tools/genicon）
resources/         应用图标（1024×1024 PNG）
docs/              使用说明 / 故障排查 / 代码导读 / ADR / 阶段计划
```

依赖方向是硬规则（`app → ui → {pane, store, theme, profile, action} → i18n`），
由 `internal/app/deps_test.go` 机械校验。详见 [docs/代码导读.md](docs/代码导读.md)。

## 运行环境

- Windows 10 1809+ / Windows 11（主要支持与验证目标，ConPTY）
- Linux / macOS：有 pty 与平台路径实现，可交叉编译，属实验性
- Go 1.27+、mygo v0.8.3+
