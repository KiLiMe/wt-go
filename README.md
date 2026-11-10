# wt-go

用 [mygo](https://github.com/egoist/mygo)（Go 原生 UI）写的 Windows Terminal 风格终端。
目标：Windows 上开箱即用、外观与操作贴近 Windows Terminal，单文件可执行、无 Electron。

> 当前已完成阶段 0–3：单窗口、多标签、分栏、设置界面、可配置快捷键、配色方案与 settings.json 热重载。搜索、命令面板、会话恢复等列在后续阶段。实现计划见 [docs/plan.md](docs/plan.md)。

## 文档

| 文档 | 内容 |
| --- | --- |
| [docs/plan.md](docs/plan.md) | 分阶段实现计划、验收标准、不做什么 |
| [docs/architecture.md](docs/architecture.md) | 分层、包职责、依赖方向、关键数据结构 |
| [docs/features.md](docs/features.md) | 功能清单与 Windows Terminal 对照表 |
| [docs/settings.md](docs/settings.md) | 配置文件 `settings.json` 的结构与默认值 |
| [docs/keys.md](docs/keys.md) | 快捷键表 |
| [docs/decisions.md](docs/decisions.md) | 设计决策与理由（ADR） |
| [docs/linux.md](docs/linux.md) | Linux 依赖、启动方式、已知缺陷验证清单 |

## 运行环境

- 主要支持并在 Windows 上验证：Windows 10 1809+ / Windows 11（ConPTY）
- Linux/macOS：已提供 pty 与配置目录的平台实现并可交叉编译，但仍属实验性；部分 Windows 专属能力（如 Job Object 进程树收尾）不可用
- Go 1.27+（`go.mod` 里 `go 1.27.3`，与 sshterm 对齐）
- mygo ≥ v0.8.3（`plugins/terminal` 内含 libghostty-vt；发布时建议使用 `mygo build` 打包原生库）

## 构建与运行

```powershell
go build -o wtgo.exe ./cmd/wtgo
.\wtgo.exe
```

普通 `go build` 适合开发验证；发布包应按 mygo 的构建方式打包原生库。

## 与参考项目的关系

- **sshterm**（`D:\Sou\sshterm`）：同栈前辈，验证了 mygo + terminal 插件在 Windows 上
  可用、外观可做到现代程度。wt-go 复用它验证过的做法（tab 样式、四边面板、主题、
  截图冒烟脚本），但不继承它的 SSH/隧道/保管库等功能。
- **Windows Terminal**：交互与配置的参照物，不是代码参照。
