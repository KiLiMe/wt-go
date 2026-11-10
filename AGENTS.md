# AGENTS.md

## 约定

- 始终以干净、模块化的方式写代码，架构合理、职责分离。
- 不要因为是我提的需求就硬塞功能。如果找不到自然的方式实现，且你认为该重构才能让整体更干净，那就提前停下来说明。
- 代码注释用英文（与 mygo 上游一致），文档用中文；注释解释“为什么”，不复述代码。
- 依赖方向是硬规则，由 `internal/app/deps_test.go` 用 `go list` 机械校验。新增包或改 import 时同步它的规则表，以及 `docs/代码导读.md` 的依赖图。
- 纯函数包（`internal/action`、`internal/store`、`internal/profile`、`internal/layout`、`internal/theme`）不 import `os`、`os/exec`、`path/filepath`、注册表；要 OS 能力时由 `internal/app` 读入数据后注入。
- 动作的 id、标题、分组与默认键位只有一份：`internal/action`。菜单、命令面板、设置页、默认 `keybindings` 都从它取名，不另拼字符串。
- 改行为时一起改文档：`docs/使用说明.md`（用户视角）、`docs/代码导读.md`（代码地图）、`docs/decisions.md`（新决策写成 ADR）、`docs/plan.md`（阶段状态）。

## 路径

| 路径 | 内容 |
| --- | --- |
| `cmd/wtgo/main.go` | 入口，一行 `app.Run()` |
| `internal/app/` | 应用状态与动作：标签、窗格、设置读写、热重载、进程生命周期、命令面板与查找栏的状态 |
| `internal/ui/` | 纯绘制：标签栏（兼标题栏）、窗格、设置页、命令面板、查找栏、对话框；动作经 `ui.Actions` 回传 |
| `internal/action/` | 动作目录：id、标题、分组、默认键位（纯函数） |
| `internal/pane/` | 窗格树与标签集合（`Tab`/`Node`/`Set`）：切分、焦点、关闭补位；不做 IO |
| `internal/layout/` | 窗格树的几何计算（矩形与分割线），不认识终端 |
| `internal/profile/` | profile 模型，以及把用户写的与自动发现的结果合并成列表（纯函数） |
| `internal/store/` | `settings.json` 的解析、序列化、默认值与范围校验（纯函数，字节由 app 读写） |
| `internal/theme/` | 配色：设置里的方案 > 内置 > Ghostty 的 617 套（纯函数） |
| `internal/sys/` | 全部 OS 副作用：路径、原子写、坏配置归档、Job Object、shell 探测、打开路径、崩溃日志 |
| `internal/i18n/` | 中英文案，缺条目回落原文 |
| `docs/` | `使用说明.md` / `代码导读.md` / `decisions.md`（ADR）/ `plan.md`（阶段计划） |
| `mygo.json` | `go tool mygo build` 的包名、版本与打包信息 |
| `build.ps1` | 构建脚本：产物固定为 `build\bin\Release\wtgo.exe`，`-Test` / `-Clean` / `-Package` |
