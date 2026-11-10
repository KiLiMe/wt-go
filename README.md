# wt-go

Windows Terminal 风格的终端，用 [mygo](https://github.com/egoist/mygo)（Go 原生 UI，无 webview）写：
多标签、分栏、profile、配色方案、可配置快捷键，配置是一份手写友好的 `settings.json`。
`go build` 产出单文件 exe，没有 Electron。

**完整文档：[docs/使用说明.md](docs/使用说明.md)**

## 快速开始

```pwsh
pwsh .\build.ps1                 # 编译到 build\bin\Release\wtgo.exe
.\build\bin\Release\wtgo.exe
```

发布包用 mygo 自己的构建命令，一次产出可执行文件、原生库与 Linux 包：

```pwsh
pwsh .\build.ps1 -Package        # 等于 go tool mygo build -platform windows/amd64 -o dist
```

## 能力

| 能力 | 状态 |
| --- | --- |
| 多标签、拖拽排序、中键关闭、重命名/置色 | ✅ |
| 分栏（自动/竖/横）、焦点移动、拖动分割线 | ✅ 每标签最多 `maxPanes` 个（默认 4） |
| 查找（`Ctrl+Shift+F`）、命令面板（`Ctrl+Shift+P`） | ✅ |
| profile：自动发现 pwsh / Windows PowerShell / cmd / WSL / POSIX shell | ✅ |
| 配色方案 | ✅ 内置 9 套 + Ghostty 617 套直用 + 自定义 |
| 字体 / 字号 / 行高 / 内边距 / 光标形状 / 窗口透明度 | ✅ |
| 设置页（表单 + 快捷键录制 + 实时预览） | ✅ |
| `settings.json` 热重载、坏配置归档 | ✅ 约一秒内生效 |
| 选中/搜索/链接/剪贴板协议 | ✅ 由终端插件提供 |
| 会话恢复（`startOn`） | ⚠️ 只解析，阶段 5 |
| 每 profile 独立字体、profile 图标 | ⚠️ 字体全局，图标先只做颜色 |
| 标签搜索、交换窗格、分屏广播、背景图/亚克力、Quake 模式 | ❌ |

## 文档

| 文档 | 内容 |
| --- | --- |
| [docs/使用说明.md](docs/使用说明.md) | 运行环境、构建、标签与分栏、快捷键、`settings.json` 全字段、profile、配色、设置页、Linux、与 WT 的差异 |
| [docs/代码导读.md](docs/代码导读.md) | 代码地图：目录、包职责、依赖方向、渲染流程、改动前必读的坑、阅读顺序 |
| [docs/decisions.md](docs/decisions.md) | 设计决策与理由（ADR） |
| [docs/plan.md](docs/plan.md) | 阶段计划、验收标准、不做什么 |
| [AGENTS.md](AGENTS.md) | 工程约定与路径 |

## 目录

```text
cmd/wtgo/          入口：app.Run()
internal/app/      应用状态与动作：标签、窗格、设置读写、热重载、进程生命周期
internal/ui/       只画 State：标签栏（兼标题栏）、窗格、设置页、命令面板、查找栏
internal/action/   动作目录：id、标题、分组、默认键位
internal/pane/     窗格树与标签集合（不做 IO）
internal/layout/   窗格几何计算
internal/profile/  profile 模型与自动发现结果合并
internal/store/    settings.json 的解析、默认值、校验
internal/theme/    配色方案解析
internal/sys/      全部 OS 副作用（路径、原子写、Job Object、shell 探测……）
internal/i18n/     中英文案
docs/              使用说明 / 代码导读 / ADR / 阶段计划
```

依赖方向是硬规则（`app → ui → {pane, store, theme, profile, action} → i18n`），
由 `internal/app/deps_test.go` 机械校验。详见 [docs/代码导读.md](docs/代码导读.md)。

## 运行环境

- Windows 10 1809+ / Windows 11（主要支持与验证目标，ConPTY）
- Linux / macOS：有 pty 与平台路径实现，可交叉编译，属实验性
- Go 1.27+、mygo v0.8.3+
