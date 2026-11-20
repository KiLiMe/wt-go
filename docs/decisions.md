# 架构决策记录（ADR）

只保留仍具参考价值的决定、核心理由与关键代价；测试清单、调查过程与实现行号已省略（以代码与 `git log` 为准）。

## 总表

| 编号 | 决定 | 状态 |
| --- | --- | --- |
| ADR-01 | 用 mygo 而不是 WPF/Electron/Tauri | 有效 |
| ADR-02 | 终端内核用 libghostty-vt，不自研 VT | 有效 |
| ADR-03 | 只接管自己创建的进程 | 有效 |
| ADR-04 | 配置格式自定，不兼容 WT 的 settings.json | 有效 |
| ADR-05 | 界面与动作分离 | 有效 |
| ADR-06 | 先不到 sshterm 里改 | 有效 |
| ADR-07 | 窗格树独立为 `internal/pane` | 有效 |
| ADR-08 | OS 副作用集中到 `internal/sys`，纯解析包不做 IO | 有效 |
| ADR-09 | 会话常驻（退出后进程不结束）暂不采用 | 有效 |
| ADR-10 | 配色方案：自定义 > 内置 7 套 > Ghostty 617 套 | 有效 |
| ADR-11 | 标签栏兼作标题栏 | 有效（ADR-28 补齐 Linux） |
| ADR-12 | 窗格几何与焦点移动用矩形判定，不用树序 | 有效 |
| ADR-13 | 键盘焦点每帧用 `Element.Focus()` 保持 | 有效 |
| ADR-14 | 设置页实时预览：改了就应用，保存只负责写盘 | 有效 |
| ADR-15 | profile 编辑用工作副本，保存时才写回 | 有效 |
| ADR-16 | 热重载以文件字节为准，有未保存改动时不覆盖 | 有效 |
| ADR-17 | 多行粘贴确认在 `terminal.Options.OnPaste` 里拦 | 有效 |
| ADR-18 | 启动大小（列×行）用两次测量 | 已被 ADR-46 取代 |
| ADR-19 | 子进程收尾用 Job Object（`BREAKAWAY_OK` 用于重置重启） | 有效 |
| ADR-20 | 崩溃日志写在配置/崩溃目录，写完后重新 panic | 有效 |
| ADR-21 | 失败给编号，日志写文件，`diag` 一次说完 | 有效 |
| ADR-22 | 排查接口用 MCP 走 stdio，不做 HTTP | 有效 |
| ADR-23 | 设置页列出动作目录的每一条，录制以动作 id 为键 | 有效 |
| ADR-24 | 标签宽度用 `Context.MeasureText` 量，不再估算 | 有效 |
| ADR-25 | 关闭确认：按进程树判断（`running` / `always` / `never`） | 有效 |
| ADR-26 | 标签条自己滚动：溢出出箭头 + 视口接管滚轮 | 有效 |
| ADR-27 | `diag --redact`：个人路径脱敏 | 有效 |
| ADR-28 | Linux 也用自绘标题栏，按钮跟随桌面布局 | 有效 |
| ADR-29 | 字体族不是等宽就回落到 `monospace` | 有效 |
| ADR-30 | 设置页控件用现成的，不手搓 | 有效 |
| ADR-31 | 界面主题从配色方案推导 | 有效（ADR-64 拆分） |
| ADR-32 | Linux 用内存 fontconfig 规则保住文字颜色 | 有效 |
| ADR-33 | 界面图标用打进二进制的 SVG，不用字体字形 | 有效 |
| ADR-34 | 自定义配色按「基名 + (custom)」命名，行尾给删除 | 有效 |
| ADR-35 | 启动探测到的 shell 写入配置，仅首次创建文件时执行 | 有效 |
| ADR-36 | 字体列表按需枚举，等宽判定按 `(family, size)` 记忆化 | 有效 |
| ADR-37 | 不做进程内 `malloc_trim`（仅省 3–4 MB 且很快被复用） | 结论：不做 |
| ADR-38 | 设置页一帧取一次快照，配色列表随 `apply` 失效 | 有效 |
| ADR-39 | 等 shell 出现时读内核 children 文件，不再全量扫进程表 | 有效 |
| ADR-40 | 性能基线（2026-11-11 单机实测记录） | 归档 |
| ADR-41 | 文案交给 mygo catalog，`internal/i18n` 只留薄门面 | 有效 |
| ADR-42 | 目录留空回落到 cwd | 已被 ADR-53 取代 |
| ADR-43 | 设置页卡片化，弹菜单画在窗口内，「关于」只读编译期信息 | 有效 |
| ADR-44 | 默认配置文件留空＝跟随设备，不是列表第一项 | 有效 |
| ADR-45 | 标签前是色块图标，不是色点 | 有效 |
| ADR-46 | 窗口恢复上次位置与大小，`startSize` 删除 | 有效 |
| ADR-47 | 暴露 Ghostty 行为：滚轮倍数、最小对比度、内边距色、剪贴板读写 | 有效 |
| ADR-48 | 标签宽度按份额分，强调线删掉 | 有效 |
| ADR-49 | 起始目录认两个保留词：`current` 与 `home` | 有效 |
| ADR-50 | `wt-go -d <目录>` 覆盖本次启动目录，不写文件 | 有效 |
| ADR-51 | Windows 上留空即家目录 | 已被 ADR-53 取代 |
| ADR-52 | 标签先显示启动的 shell 名，再更新为 shell 标题 | 有效 |
| ADR-53 | 起始目录不分平台：留空一律回落家目录 | 有效 |
| ADR-54 | 新建标签只认 `defaultProfile`，不另设开关 | 有效 |
| ADR-55 | 快捷键页：一个动作一个键，改没改看得见 | 有效 |
| ADR-56 | 无可用 profile 时直接跟随设备，不在菜单留空位 | 有效 |
| ADR-57 | ConPTY 载不动的键位在 mygo 编码器里降级为老式字节 | 有效 |
| ADR-58 | fork mygo：拆条写与跳过部分 OSC 回复 | 已被 ADR-72/73 修订 |
| ADR-59 | 程序自身 fork 的剪贴板工具不归 wt-go 管（Linux 装 `xclip`） | 有效 |
| ADR-60 | 设置页整页统一底色，左栏与底栏不铺 `Surface` | 有效 |
| ADR-61 | 「显示隐藏的配置文件」页级视图开关 | 已被 ADR-71 取代 |
| ADR-62 | 「跟随系统」改为「跟随设备」并显式解析本机首个 shell | 有效 |
| ADR-63 | 窗口存在 ≥2 个标签时关闭先确认 | 有效 |
| ADR-64 | 窗口主题与终端配色分离：窗口随 `theme`，仅活动标签随窗格 | 有效 |
| ADR-65 | 浮层同时只开一个，按键分发只问栈顶 `top()` | 有效 |
| ADR-66 | 枚举名写错回落默认值，未知顶层键读入保留 | 有效 |
| ADR-67 | 保存校验不通过时错误横幅留页顶，草稿不丢失 | 有效 |
| ADR-68 | 脱敏按系统路径大小写规则匹配，WSL 探测跳过 Docker 自用发行版 | 有效 |
| ADR-69 | 设置页按用户心智分为 6 页 | 有效 |
| ADR-70 | 「外观与排版」顶部预览吸顶 | 有效 |
| ADR-71 | `hidden` 归属单条 profile，列表始终全量展示并标「已隐藏」 | 有效 |
| ADR-72 | ConPTY 剥掉 OSC 回复头（改写 ADR-58 病因） | 有效（范围见 ADR-73） |
| ADR-73 | ConPTY 不转发回复时整批不发（作废 ADR-58 拆条，扩至全答复） | 有效（条件收窄见 ADR-82） |
| ADR-74 | 按住不放仅连发方向步进类动作（`Action.Repeats`） | 有效 |
| ADR-75 | 内置 profile 支持单条「重置为内置」（保留 `dir`） | 有效 |
| ADR-76 | 标签图标支持预设图形、单字形或按命令自动猜 | 有效 |
| ADR-77 | 配色方案取消下拉，仅保留带色板的可搜索列表 | 有效 |
| ADR-78 | 标题栏空白处右键提供窗口菜单（不做移动/大小模态循环） | 有效 |
| ADR-79 | 字体栏改为文本输入框 + `＋` 菜单 + 实时等宽诊断提示 | 有效 |
| ADR-80 | 字体仅设普通字重，粗体由框架按 `min(max(w+300, 700), 900)` 派生 | 有效 |
| ADR-81 | 系统 conhost 创建的控制台先补 `ENABLE_VIRTUAL_TERMINAL_INPUT` | 有效（系统 conhost 兜底） |
| ADR-82 | 随包 Windows Terminal 的 ConPTY，让程序询问可收到回复 | 有效 |
| ADR-83 | exe 同目录存在 `data` 目录即启用便携模式 | 有效 |
| ADR-84 | 链接开启时悬停即提示、点击打开，路径剥离 `:行:列` 并按 OSC 7 解析 | 有效 |
| ADR-85 | 技术债收尾：校验范围单一来源、空数组输出 `[]`、版本读截断、参数整条传、跳过无效 profile 记日志 | 有效 |
| ADR-86 | Win10 自绘标题栏的顶部 1px 归 mygo：让出客户区第 0 行并用黑刷露出 DWM 的边框线 | 有效 |

---

## 核心架构与技术边界

### 基础选型与包边界（ADR-01..09, 33, 41）
- **GUI 与终端核心（ADR-01/02/41）**：桌面框架选用 `mygo`，终端核心选用 `libghostty-vt`（不自研 VT/解析器）；文案唯一来源为 `locales/*.json`，由 `go tool mygo generate` 生成 `locales/catalog_gen.go`，`internal/i18n` 只作薄门面。
- **进程模型（ADR-03/09）**：不做退出后常驻守护进程，仅管理 wt-go 自己启动的子进程，换取简单生命周期；代价是关窗口即结束未脱管会话。
- **配置策略（ADR-04/66）**：采用扁平自定 `settings.json`（不兼容 WT 两层继承）；枚举拼写错误仅字段回退，JSON/数值越界才拒绝并归档；未知顶层键保留，内部对象未知键不保证。
- **职责隔离（ADR-05/07/08）**：`internal/action` 定义全部动作与默认键；`internal/pane` 与 `internal/layout` 分管树与矩形几何（焦点移动按几何不按树序，ADR-12）；纯解析包不碰 OS，全部副作用收进 `internal/sys` 并由 `internal/app/deps_test.go` 强制检验。
- **内嵌图标（ADR-33）**：界面图标用嵌入 SVG（`internal/icons`）而非字形字体，不依赖系统装字体。

### Windows / ConPTY 与 Fork 边界（ADR-57, 72, 73, 81, 82, 84）
- `github.com/KiLiMe/mygo` 的 `my-go-custom` 分支维护 wt-go 所需的定制改动。
  1. **键位降级（ADR-57）**：`kittyKeyboard` 未开启或仅部分标志位启用时，将 ConPTY 无法透传的按键降为传统字节序列。
  2. **ConPTY 随包与静默策略（ADR-72/73/81/82）**：Windows 系统自带 `conhost` 会剥掉终端发回给程序的 OSC/CSI 回复且延后开启 `ENABLE_VIRTUAL_TERMINAL_INPUT`。wt-go 随包分发 MIT 许可的 `Microsoft.Windows.Console.ConPTY`（`resources/windows-<goarch>/conpty/`，哈希钉在 `Scripts/conpty.json`）并按 `MYGO_CONPTY → conpty/ → exe 同目录 → 系统 conhost` 选用；仅在未启用转发的系统 conhost 路径下开启 `setVTInput()`（ADR-81）并由 `writeReply` 整批静默回复（ADR-73/82）。随包 ConPTY 的 `ClearPseudoConsole` 不主动清屏，清屏后依赖 `wt-go` VT 层重绘。
  3. **链接控制与路径（ADR-84）**：fork 提供 `SetLinks`、`SetLinkHoverHint`（免 `Ctrl` 悬停提示与点击打开、拖动则选区）及 `file:///C:/...` 盘符修正；wt-go 在 `OpenLink` 钩子中通过 `sys.OpenURL`/`sys.OpenPath` 自行打开并捕获 `ShellExecuteW` 错误（`E209`），相对路径依次按 OSC 7 `Terminal.Dir()` 与窗格启动目录解析（暂不嗅探 PowerShell 的 OSC 9;9）。

### 平台、生命周期与可观测性（ADR-19..22, 27, 28, 32, 39, 59, 68, 83）
- **子进程与关闭确认（ADR-19/25/39/59）**：Windows 将自身加入带 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` 与 `BREAKAWAY_OK` 的 Job Object（重置重启时脱离）；关闭确认通过读取内核子进程信息判断（不再全量扫进程表），但程序自己 fork 出的外部进程（如 Linux 下需要 `xclip` 保活选区）不归 wt-go 管理。
- **便携模式（ADR-83）**：exe 同目录存在 `data` 目录即便携版，配置/日志/崩溃日志均收进 `data/{config,logs,crashes}`，并将 `mygo` 的 `PathUserData` 钉进 `data/config`。
- **诊断与日志（ADR-20..22, 27, 68）**：panic 写 `crash-*.log` 后重抛；已知故障统一走 `internal/fault` 编号；`wt-go diag --redact` 按各平台路径大小写规则对个人目录脱敏；`wt-go mcp` 仅走 stdio（JSON-RPC 2.0）。
- **Linux 标题栏与字体渲染（ADR-28/32）**：Linux 同样使用自绘标题栏并读取 `gtk-decoration-layout`；启动时向内存 fontconfig 注入 `rgba=none` 避免 cairo 将次像素抗锯齿误染为彩色。

### 设置、外观与交互设计（ADR-10..17, 23..26, 29..31, 34..38, 43..56, 60..71, 74..80, 85）
- **状态与重载（ADR-14..16, 38, 67）**：设置页草稿实时预览但仅「保存」时校验落盘；校验失败保留红条与草稿；后台每秒按字节比对热重载（有未保存草稿时不覆盖）。
- **外观与字体（ADR-10, 29, 34, 36, 47, 64, 70, 77, 79, 80）**：`theme` 管窗口、`colorScheme` 管终端，仅活动标签随窗格底色；配色仅保留带色板预览的可搜索列表（选色板颜色按内置/Ghostty 名缓存）；字体栏为逗号列表输入框 + 等宽诊断菜单（按 `(family, size)` 记忆化测量，非等宽回落 `monospace`），粗体字重由框架从单一 `font.weight` 派生。
- **标签、窗口与交互（ADR-11, 13, 24, 26, 45, 46, 48, 52, 63, 65, 74, 76, 78, 86）**：窗口几何由 `mygo` 存入 `window-state.json`；浮层栈只保留一个活跃浮层并将键盘交 `top()`；按住按键仅移动焦点/切标签动作连发；标签宽度用 `MeasureText` 与份额均分并支持横向滚动；标签图标支持预设 12 组自带配色图形、单字形或按命令猜测。
- **窗口边框归属（ADR-86）**：Win10 只在非客户区画那圈 1px 边框线，自绘标题栏的窗口要自己让出位置。这件事收在 `mygo`：`TitleBarHidden` 且非最大化/全屏时，客户区第 0 行不交给 surface（`topLine()`），surface 与各填充矩形下移一行，`WM_PAINT` 里用黑刷填那一行，让 DWM 的边框线透出。`wt-go` 只传 `TitleBarHidden`，不读系统主题色也不调 `DwmSetWindowAttribute(DWMWA_BORDER_COLOR)`（该属性在 Win10 无效，Win11 由 DWM 自己画）；颜色跟 `DWMWA_USE_IMMERSIVE_DARK_MODE` 所代表的外观走。代价：`go.mod` 只能 pin 到带这段修复的 mygo 提交，回退版本会让 Win10 顶部缺线。
- **Profile 与路径解析（ADR-35, 44, 49, 50, 53, 54, 56, 62, 71, 75, 85）**：仅首次写配置时探测 shell；`defaultProfile` 留空时按固定优先级显式解析本机首个可用 shell（跳过隐藏项与 Docker WSL 发行版）；`hidden` 为单条属性，列表全量展示以便取消隐藏；内建项支持单条「重置为内置」；起始目录统一回落家目录，`current`/`home` 为保留词，`-d` 覆盖单次启动目录。
