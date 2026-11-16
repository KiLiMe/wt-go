# AGENTS.md

## 约定

- 始终以干净、模块化的方式写代码，架构合理、职责分离。
- 不要因为提了需求就硬塞功能。如果找不到自然的方式实现，且认为该重构才能让整体更干净，提前停下来说明。
- 代码注释用英文（与 mygo 上游一致），文档用中文；注释解释“为什么”，不复述代码。
- 依赖方向是硬规则，由 `internal/app/deps_test.go` 用 `go list` 机械校验。新增包或改 import 时同步它的规则表与 `docs/代码导读.md` 的依赖图。
- 纯函数包（`internal/action`、`internal/cli`、`internal/fault`、`internal/i18n`、`internal/icons`、`internal/layout`、`internal/mcp`、`internal/pane`、`internal/profile`、`internal/store`、`internal/theme`、`internal/ui`、`locales`）不 import `os`、`os/exec`、`path/filepath`、注册表（测试文件也不例外，仅放过 `io/fs` 读夹具）；要 OS 能力时由 `internal/sys` 提供、`internal/app` 读入数据后注入。
- 文案只有一份：`locales/en.json`（原文，`mygo.json` 的 `i18n.source`）与 `locales/zh-Hans.json`，`go tool mygo generate` 写出 `locales/catalog_gen.go`（typed 访问器，生成物提交）。静态调用点用 `locales.CloseTab()`；只有 `internal/action` 的 `TitleKey`/`GroupKey` 与动态 key 走 `i18n.T`/`i18n.F`。改完 json 跑 `go tool mygo generate` 与 `go tool mygo i18n check`；语言切换一律经 `internal/app` 的 `locale()` 调 `mygo.App.SetLocale`，纯函数包不 import 根包 `mygo`。
- 动作的 id、标题、分组与默认键位只有一份：`internal/action`。菜单、命令面板、设置页、默认 `keybindings` 都从它取名，不另拼字符串；用户向键位说明只写在 `docs/配置参考.md`。
- 文档按职责单一维护，不跨文件复制：改用户可见行为按需同步 `docs/使用说明.md` 或 `docs/配置参考.md`；改包结构/架构取舍再同步 `docs/代码导读.md` 或 `docs/decisions.md`。
- 新的失败路径要编号：`internal/fault` 加码、`internal/cli` 加检查、`docs/故障排查.md` 写确认与修法；日志用 `fault.Line(code, detail)` 打印，不写自由文本。
- 版本号与发版流程唯一来源：`mygo.json` 与 `docs/代码导读.md` 的「构建与发布」。

## 路径

| 路径 | 内容 |
| --- | --- |
| `main.go` | 入口，一行 `os.Exit(app.Run())` |
| `internal/app/` | 应用状态与动作：标签、窗格、设置读写、热重载、浮层管理、进程生命周期、命令面板与查找栏 |
| `internal/ui/` | 纯绘制：标签栏（兼标题栏）、窗格、设置页、命令面板、查找栏、对话框；动作经 `ui.Actions` 回传 |
| `internal/action/` | 动作目录：id、标题、分组、默认键位、连发标志（纯函数） |
| `internal/pane/` | 窗格树与标签集合（`Tab`/`Node`/`Set`）：切分、焦点、关闭补位；不做 IO |
| `internal/layout/` | 窗格树的几何计算（矩形与分割线），不认识终端 |
| `internal/profile/` | profile 模型、自动发现合并、默认/设备解析、标签标记（纯函数） |
| `internal/store/` | `settings.json` 的解析、序列化、默认值、未知键保留与范围校验（纯函数） |
| `internal/theme/` | 配色：自定义 > 内置 7 套 > Ghostty 617 套（纯函数） |
| `internal/sys/` | 全部 OS 副作用：路径（便携版 `data` 目录，ADR-83）、原子写、坏配置归档、Job Object、shell 探测、进程树判定、打开路径/网址、崩溃日志 |
| `internal/i18n/` | 按 key 说一句话（`T`/`F`/`Has`），语言跟随 `mygo.App` 当前语言（纯函数） |
| `locales/` | 消息目录：`en.json` / `zh-Hans.json` 与 `mygo generate` 出的 `catalog_gen.go` |
| `internal/icons/` | Remix Icon v4.8.0 SVG（`//go:embed`），由 `ui.Icon` 绘制（纯函数） |
| `internal/fault/` | 故障码与一行含义：日志、`wt-go diag`、文档共用一套编号（纯函数） |
| `internal/cli/` | 不开窗口部分：`wt-go diag` 报告、`wt-go mcp` 工具（只调 `internal/sys`） |
| `internal/mcp/` | 精简 MCP 服务器：JSON-RPC 2.0 over stdio + tools/resources（只用标准库） |
| `tools/genicon/` | 画 `resources/icon.png`：`go run ./tools/genicon` |
| `resources/icon.png` | 应用图标（1024×1024 PNG），`mygo build` 打包嵌进 exe |
| `resources/windows-<goarch>/conpty/` | 随包 ConPTY（`conpty.dll` / `OpenConsole.exe` / `ConPTY-LICENSE.txt`，ADR-82）；哈希钉在 `Scripts/conpty.json`，`conpty_test.go` 校验 |
| `resources/data/README.txt` | 装出来是 exe 旁的 `data/README.txt`：同目录存在 `data` 目录即走便携模式（ADR-83） |
| `Scripts/build.ps1` / `build.sh` | 日常构建与打包（`-Test`/`--test`、`-Clean`/`--clean`、`-Package`/`--package`、`-Version`/`--expect <版本>`） |
| `docs/` | `使用说明.md` / `配置参考.md` / `故障排查.md` / `代码导读.md` / `decisions.md` |
| `mygo.json` | `go tool mygo build` 的包名、版本与打包信息 |
| `refs/` | 只读参考项目（`.gitignore`），清单见 `refs/AGENTS.md` |
