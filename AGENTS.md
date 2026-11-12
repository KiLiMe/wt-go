# AGENTS.md

## 约定

- 始终以干净、模块化的方式写代码，架构合理、职责分离。
- 不要因为是我提的需求就硬塞功能。如果找不到自然的方式实现，且你认为该重构才能让整体更干净，那就提前停下来说明。
- 代码注释用英文（与 mygo 上游一致），文档用中文；注释解释“为什么”，不复述代码。
- 依赖方向是硬规则，由 `internal/app/deps_test.go` 用 `go list` 机械校验。新增包或改 import 时同步它的规则表，以及 `docs/代码导读.md` 的依赖图。
- 纯函数包（`internal/action`、`internal/cli`、`internal/fault`、`internal/i18n`、`internal/mcp`、`internal/store`、`internal/profile`、`internal/layout`、`internal/theme`、`locales`）不 import `os`、`os/exec`、`path/filepath`、注册表；要 OS 能力时由 `internal/sys` 提供、`internal/app` 读入数据后注入。
- 文案只有一份：`locales/en.json`（原文，`mygo.json` 的 `i18n.source`）与 `locales/zh-Hans.json`，`go tool mygo generate` 写出 `locales/catalog_gen.go`（typed 访问器，生成物提交）。静态调用点用 `locales.CloseTab()`；只有 `internal/action` 的 `TitleKey`/`GroupKey` 与动态 key 走 `i18n.T`/`i18n.F`。改完 json 跑 `go tool mygo generate`，再跑 `go tool mygo i18n check`；语言切换一律经 `internal/app` 的 `locale()` 调 `mygo.App.SetLocale`，纯函数包不 import 根包 `mygo`。
- 动作的 id、标题、分组与默认键位只有一份：`internal/action`。菜单、命令面板、设置页、默认 `keybindings` 都从它取名，不另拼字符串；用户能看到的键位说明也只在 `docs/键位.md` 写一遍。
- 改行为时一起改文档：`docs/使用说明.md`（用户视角）、`docs/配置参考.md`（`settings.json` 字段）、`docs/代码导读.md`（代码地图）、`docs/decisions.md`（新决策写成 ADR）、`docs/plan.md`（阶段状态）。
- 新的失败路径要编号：`internal/fault` 里加一个码，`docs/故障排查.md` 里写它的确认与修法；日志里用 `fault.Line(code, detail)` 打印，别写自由文本。
- 版本号与发版流程不在本文件：见 `docs/代码导读.md` 的「版本与发布」与 `docs/使用说明.md` 第 2 节。

## 路径

| 路径 | 内容 |
| --- | --- |
| `main.go` | 入口，一行 `app.Run()` |
| `internal/app/` | 应用状态与动作：标签、窗格、设置读写、热重载、进程生命周期、命令面板与查找栏的状态 |
| `internal/ui/` | 纯绘制：标签栏（兼标题栏）、窗格、设置页、命令面板、查找栏、对话框；动作经 `ui.Actions` 回传 |
| `internal/action/` | 动作目录：id、标题、分组、默认键位（纯函数） |
| `internal/pane/` | 窗格树与标签集合（`Tab`/`Node`/`Set`）：切分、焦点、关闭补位；不做 IO |
| `internal/layout/` | 窗格树的几何计算（矩形与分割线），不认识终端 |
| `internal/profile/` | profile 模型，以及把用户写的与自动发现的结果合并成列表（纯函数） |
| `internal/store/` | `settings.json` 的解析、序列化、默认值与范围校验（纯函数，字节由 app 读写） |
| `internal/theme/` | 配色：设置里的方案 > 内置 > Ghostty 的 617 套（纯函数） |
| `internal/sys/` | 全部 OS 副作用：路径、原子写、坏配置归档、Job Object、shell 探测、打开路径、崩溃日志 |
| `internal/i18n/` | 按 key 说一句话（`T`/`F`/`Has`），语言跟随 `mygo.App` 的当前语言；不 import 根包 `mygo` |
| `locales/` | 消息目录：`en.json` / `zh-Hans.json` 与 `mygo generate` 写出的 `catalog_gen.go`（typed 访问器） |
| `internal/icons/` | 界面图标：Remix Icon v4.8.0（Apache-2.0）的 SVG，`//go:embed` 打进二进制，由 `ui.Icon` 画（纯函数） |
| `internal/fault/` | 故障码与一行含义：日志、`wt-go diag`、文档共用一套编号（纯函数） |
| `internal/cli/` | 不开窗口的那部分：`wt-go diag` 的报告、`wt-go mcp` 的工具（不碰 OS，全走 `internal/sys`） |
| `internal/mcp/` | 精简的 Model Context Protocol 服务器：JSON-RPC 2.0 + tools/resources（只用标准库） |
| `tools/genicon/` | 画 `resources/icon.png`：`go run ./tools/genicon` |
| `resources/icon.png` | 应用图标（1024×1024 PNG），`mygo build` 打包时嵌进 exe |
| `Scripts/build.ps1` | 日常构建（Windows）：`go tool mygo build` 出 `build\windows-amd64\wt-go.exe`（带图标、版本与 native 库） |
| `Scripts/build.sh` | 同上（Linux / macOS）：出 `build/<os>-<arch>/wt-go`，平台按 `go env` |
| `docs/` | `使用说明.md` / `配置参考.md`（`settings.json` 全字段）/ `键位.md`（默认键位与键位写法）/ `故障排查.md`（故障码）/ `MCP.md`（`wt-go mcp` 的工具、资源与协议）/ `Linux.md` / `代码导读.md` / `decisions.md`（ADR）/ `plan.md`（阶段计划） |
| `mygo.json` | `go tool mygo build` 的包名、版本与打包信息 |
| `Scripts/build.ps1` / `Scripts/build.sh` | 参数：`-Test`/`--test` 先跑 vet 与测试、`-Clean`/`--clean` 清掉 `build/`、`-Package`/`--package` 只出 `dist/`、`-Version`/`--expect <版本>` 断言 `mygo.json` 的版本就是这一串（发版对 tag 用） |
