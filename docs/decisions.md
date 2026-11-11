# 设计决策

按时间记录。后加的决定追加在末尾，不覆盖旧的。

---

## ADR-001 用 mygo 而不是 WPF/Electron/Tauri

**决定**：Go + mygo。

**理由**：用户要求用 mygo；单文件分发、库体积小、启动快；与 sshterm 同栈，
验证过的做法能直接沿用。

**代价**：mygo 生态较新，遇到框架缺陷需要绕或提交上游；无 cgo 意味着某些
Win32 API 要手写 syscall。

---

## ADR-002 终端内核用 libghostty-vt，不自研 VT

**决定**：`github.com/egoist/mygo/plugins/terminal`。

**理由**：终端模拟器正确性极难（宽字符、reflow、Kitty 键盘协议、OSC 52、
bracketed paste、鼠标上报……）。libghostty-vt 是 Ghostty 的成熟实现，且已解决
Windows ConPTY 接入。

**代价**：受限于它的 Options 暴露面；需要的功能（如某些自定义键位）要通过
`Keybinds` / `OnKey` 绕。

---

## ADR-003 只接管自己创建的进程

**决定**：不做"附加到已存在的终端窗口/ConPTY 会话"。

**理由**：Windows 没有官方的"终端复用"接口，ConPTY 管道的句柄只对创建者有效；
第三方接管要靠注入，不稳定且触发安全软件。Windows Terminal 本身也只管理自己
创建的子进程。

**推论**：想"在 wt-go 里看到别的终端"只能新开一个进程，而不是搬过来。

---

## ADR-004 配置格式自定，不兼容 WT 的 settings.json

**决定**：wt-go 用自己的扁平结构。

**理由**：WT 的配置是 `defaults.json` + `settings.json` 两层合并、`profiles.list`
数组、`actions` 命令对象，实现与解释成本都对不上当前规模。扁平结构更好读、好写、
好热重载。

**代价**：从 WT 迁移要手抄一次。

---

## ADR-005 界面与动作分离

**决定**：`internal/ui` 只实现 `Actions` 接口的调用方，不直接改状态里的进程/文件。

**理由**：沿用 sshterm 的经验——界面单测可以用假 Actions（`idle{}`），不必起真进程；
异步回调统一 `Do(fn)` 回主线程，避免竞态。

---

## ADR-006 先不到 sshterm 里改

**曾被尝试**：在 sshterm 里给本地 shell 加"管道模式"跑 pi。

**放弃原因**：偏离用户目标（要的是一个终端软件，不是 pi 的容器）；且管道模式
引入大量非标准代码（隐藏控制台、光标探针记历史），收益不抵风险。代码保留在
sshterm 分支未提交，不影响 wt-go。

---

---

## ADR-007 窗格树独立为 `internal/pane`

**决定**：`Tab` / `Node` / `Pane` 与切分、焦点、关闭补位逻辑放在 `internal/pane`，不放进 `internal/ui`，也不放进 `internal/app`。

**理由**：`ui.State.Tabs` 需要 `*pane.Set`，而 `ui` 不能 import `app`；`layout` 只算矩形、不认识终端。`pane` 持有 `*terminal.Terminal` 句柄，因此单独成包，依赖 `layout` 与 mygo `terminal`。

**约定**：`pane` 不做 IO，不关闭进程。移除窗格时返回被移除的句柄，由 `app` 关闭。

**参考**：`egoist/gorex` 把窗格树放在 `package main`（`state.go`），与 UI 混在一起。wt-go 为了可单测，拆出独立包。

---

## ADR-008 OS 副作用集中到 `internal/sys`，纯解析包不做 IO

**背景**：原文档“只有 `internal/app` 能碰文件系统和进程”与 `store`（读写 `settings.json`、热重载）、`profile`（阶段 1 读注册表发现 WSL）的职责冲突。

**决定**：
- `internal/sys` 封装文件读写与监听、注册表读取、Job Object，只被 `internal/app` import。
- `store`、`profile` 只做纯函数的解析与序列化，字节或数据由 `app` 读入后注入。
- 其他包需要 OS 能力时，在使用方定义接口，由 `app` 注入实现。

**参考**：`egoist/gorex` 的 `internal/rex`（会话服务、客户端、ConPTY、进程查询）只依赖标准库与 `golang.org/x/sys/windows`，不依赖 mygo，也不依赖 UI。UI 通过 `rex.Stream` 把字节流接给 mygo `terminal`。这证明了“协议与进程层”可以与 UI 层干净分离。

---

## ADR-009 会话常驻（应用退出后进程不结束）暂不采用

**背景**：`egoist/gorex` 用独立的会话服务进程（经 Unix socket 通信，JSON 行协议）持有 ConPTY，应用退出后进程继续运行，重开窗口时恢复标签与窗格。

**决定**：阶段 0–4 不做，保持 ADR-003 的“自己创建、自己管理生命周期”模型，应用退出 = 关闭所有标签。

**理由**：会话服务带来多进程、版本匹配（gorex 为此做了 `staleServer` 检测）、Windows 上分离进程等复杂度，超出阶段 0–4 的范围。ADR-003 不冲突（进程仍是自己创建的），但会改变生命周期模型。

**约束**：`internal/sys` 与 `internal/pane` 的接口不得假设“进程随窗口结束”，以便阶段 5 的“启动恢复”可以再评估。

---

## ADR-010 配色方案：自定义 > 内置 > Ghostty 主题

**背景**：`plugins/terminal` 自带 Ghostty 的 617 套主题（`GhosttyTheme(name)`），但 WT 的
Campbell / Solarized / Tango / Vintage 都不在其中。

**决定**：`internal/theme` 按 `settings.schemes` → 内置（手写 9 套 WT 取值）→ Ghostty →
`theme` 对应的默认配色（`terminal.DarkTheme()` / `LightTheme()`）的顺序解析；自定义方案只写
部分字段时，未写的字段沿用同名内置/Ghostty 方案的值。

**理由**：不自己维护调色板库（Ghostty 已覆盖绝大部分需求），同时保证 WT 用户的
`colorScheme` 名字能直接用；局部覆盖让“只改背景色”这类需求不用抄一整套 16 色。

**代价**：名字写错静默回退到 `theme` 的默认配色（不报错、不崩溃），排错要看 `docs/使用说明.md` 的顺序说明。

---

## ADR-011 标签栏兼作标题栏

**决定**：Windows/macOS 上窗口用 `mygo.TitleBarHidden` + `TitleBarHeight = ui.TitleBarHeight`（36dp），
标签栏自己就是拖窗区域（空白处 `DragWindow`），不再单独画 40dp 标题栏。
Linux 上改用 `mygo.TitleBarDefault`（`internal/app/titlebar_linux.go`）。

**理由**：与 Windows Terminal 一致（用户提供的参考图即如此），省掉一行高度；
`mygo` 的 `DragWindow` 放在容器上不影响子元素点击（`ui/input.go:481`）。
Linux 除外，是因为 mygo 的 hidden 在 Wayland 上靠合成器补标题栏，在 X11 上只做
`gtk_window_set_decorated(false)`，既没有服务端装饰也没有 CSD，窗口于是无边框、
无按钮、拖不动；交给平台画标题栏后，X11 由窗口管理器装饰，Wayland 与之前一样由合成器装饰。

**代价**：标签栏右侧必须留空白给拖窗，标签放不下时没有横向滚动（已知退化，待做文本测量）。
Linux 上多一行合成器/窗口管理器的标题栏，标签栏不再顶到窗口上缘。

**后续**：Linux 的例外在 ADR-029 取消。

---

## ADR-012 窗格几何与焦点移动用矩形判定，不用树序

**背景**：切分出来的窗格树里，树的顺序不等于视觉相邻：`A | (B1 / B2)` 中 B2 的“左边”
在视觉上是 A，在树里却先遇到 B1。

**决定**：`internal/layout` 只算矩形（`Split` 给出两格与分割线，`Share` 把比例夹在两端之内）；
`pane.Tab.Move` 按几何选目标：同方向的候选中取跨度重叠最多者，其次取最近者。
窗格上限（`maxPanes`）是策略而非几何，因此 `Tab.Split` 不收上限参数，由 `app` 拦。

**理由**：树序解法要在每个 split 上做方向归一，代码量大且易错；矩形判定只用现成的
几何，行为与用户看到的一致。

**代价**：每次移动焦点都重算一遍布局（窗格数 ≤ `maxPanes`，成本可忽略）。

---

## ADR-013 键盘焦点每帧用 `Element.Focus()` 保持

**背景**：`mygo` 的 `AutoFocus()` 只在元素“出现的那一帧”生效；窗格切分后旧窗格已经在树上，
只给它 `AutoFocus` 并不会把焦点从别的窗格拿回来（实测：`Alt+←` 后输入仍进原窗格）。

**决定**：焦点窗格每帧调 `Element.Focus()`（文档：“called in every frame, it keeps it there”）；
对话框（重命名）打开时不调，让 `Modal` 里的输入框拿到焦点。

**理由**：状态里只有一个“焦点窗格”，界面按它决定谁拿键盘，不需要额外的焦点请求队列。

**代价**：每帧都发一次焦点请求；如果将来有多个可输入控件（如设置页）同时存在，
这条规则要重新考虑。（阶段 3 落地后仍然成立：设置标签页没有窗格，两者不会同时出现；
对话框打开时 `st.Rename`/`st.Confirm` 非空，窗格那一路本来就不画。）

---

## ADR-014 设置页实时预览：改了就应用，保存只负责写盘

**背景**：WT 的设置页是“改完保存才生效”，但它能预览是因为它的设置模型是“重启式”的；
而字体、配色、透明度这些项的预览价值很高，等保存才看到效果会很别扭。

**决定**：表单直接改当前生效的设置（`a.settings`），每帧与 `a.applied` 比较，不同就调用一次幂等的
`apply()`（语言、窗口明暗、键位、profile、每个终端的字体/配色/选中即复制、窗口透明度/底色/置顶）。
「保存」= 写文件并把 `a.saved` 更新为当前值；「放弃」= 把 `a.settings` 换回 `a.saved` 再 `apply()`。

**理由**：只有一条“当前设置”通路，不需要“预览层 + 生效层”两套状态，也不需要为每个控件写预览逻辑。

**代价**：不点保存就关掉程序，改动会丢（和 WT 一致）；`apply()` 必须幂等且便宜，
新增设置项时要记得把它加进 `apply()`，否则只会写进文件而不生效。

---

## ADR-015 profile 编辑用工作副本，保存时才写回 `profiles`

**背景**：设置页要能看/改自动发现的 profile，但自动发现的结果只在内存里（不落盘）。
若直接绑到 `a.settings.Profiles[i]`，翻一遍页面就会把发现项写成用户配置。

**决定**：`internal/app` 持有 `a.editing *store.Profile` + `a.editingArgs string`（多行文本框的原始文本），
打开某页时从现有 profile 复制；`SaveSettings` 时才 upsert 进 `a.settings.Profiles`。

**理由**：读写分离清楚；`args` 用文本保存，避免每帧 `join`→`split` 把回车产生的空行吃掉。

**代价**：多一份状态（`editing`/`editingFrom`/`editingArgs`），dirty 判定要看三处。

---

## ADR-016 热重载以文件字节为准，有未保存改动时不覆盖

**背景**：设置页自己也会写文件，若把“文件变了”当成外部修改，会和自己的保存打架。

**决定**：`internal/app` 每秒读一次 `settings.json`，与上次读到的字节（`a.savedData`）比较；
不同且当前没有未保存改动 → 解析并应用；有未保存改动 → 不动，在设置页底部提示“放弃当前修改才能应用”；
解析失败 → 保留旧值并把错误写进提示。

**理由**：比“监听文件事件”简单（不依赖平台 API），一秒的延迟对配置文件完全够用；
比较字节而不是时间戳，避免程序自己写入后又被自己触发。

**代价**：每秒一次读文件（很小）；外部改动在有一秒内的延迟。

---

## ADR-017 多行粘贴确认在 `terminal.Options.OnPaste` 里拦

**背景**：要问“确定粘贴 N 行？”，需要拿到粘贴内容并在写入前取消它。

**决定**：`OnPaste func(transfer.Data) bool`（主线程、粘贴前调用，返回 true 表示已接管）：
按行数判断，需要问就弹确认框，确认时调用 `term.PasteData(d)`，并返回 true。

**理由**：插件已经提供了这个钩子，且在主线程上调用，可以直接改应用状态（不需要 `Do`）。

**代价**：粘贴内容要读两次（一次数行、一次写入），大段粘贴会有少量重复拷贝。

---

## ADR-018 启动大小（列×行）用两次测量，不估算单元格尺寸

**背景**：终端只报告“能放下多少格”，不报告一格多宽；mygo 也没有字体度量接口。
第一版用“内容区宽 ÷ 列数”估算，实测 100×30 得到 97×29（还漏算了 36dp 标题栏）。

**决定**：`internal/app/size.go` 的状态机：先把窗口撑大 120px 并读两次稳定后的格数，
用差值算出 `cellW/cellH`，再按 `startSize` 反推窗口尺寸（窗口尺寸 = 单元格区 + 窗口与客户区的差值，
其中客户区高度 = 内容区 + 标题栏）。终端尺寸比窗口晚一帧，所以每次改窗口后都要求“连续两次读数相同”才算稳定。

**理由**：只用终端自己报的数据，不依赖字体文件解析，换字体/DPI 都自然成立。

**代价**：启动时窗口会先大一点再收回来（几帧，肉眼可见一瞬）；设置 `startSize` 时才有这段逻辑。

---

## ADR-019 子进程收尾用“把本进程加进 Job Object”

**背景**：窗格的 shell 是本进程的子进程。窗口被强杀（任务管理器、崩溃、`Stop-Process -Force`）时，
`Terminal.Close()` 不会跑，`pwsh.exe` 与它起的进程会留在系统里。

**决定**：`sys.HoldChildren()` 在 `Run` 开头建一个带 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` 的作业，
把**当前进程**加进去，句柄不关。子进程自动继承作业成员资格，本进程一结束（无论怎么结束）
Windows 关句柄 → 作业里的进程全部终止。

**理由**：比“每开一个窗格就枚举子进程再 AssignProcessToJobObject”简单得多，
也不会漏掉 shell 后来自己起的子进程；一个调用覆盖所有窗格、所有切分。

**代价**：本进程自己也在作业里，所以“脱离父作业”这类能力（`BREAKAWAY_OK`）用不上；
如果以后要把某个窗格做成“关窗后继续跑”（会话常驻，ADR-009），这条规则要重新考虑。
实测：强杀 wt-go 后 `pwsh.exe` 与两个 `conhost.exe` 全部退出。

---

## ADR-020 崩溃日志写在配置目录，写完后重新 panic

**背景**：上一阶段的空指针崩溃只能靠 `Start-Process -RedirectStandardError` 才看到，
用户那边就是一个窗口消失了。

**决定**：`Run` 与 `watch` 协程各带一个 `recover`，调 `sys.WriteCrashLog`（时间、原因、栈）
写到 `%LOCALAPPDATA%\wt-go\crash-<时间>.log`；`Run` 写完后 `panic(r)` 继续，保留非零退出码与原有栈。

**理由**：写文件不依赖 stderr 被重定向；日志与 `settings.json` 并排，用户找得到。

**代价**：`watch` 协程的 panic 只记录不退出（该协程停掉，热重载失效），下次启动看日志才知道。

---

## ADR-021 失败给编号，日志写文件，`diagnose` 一次说完

**背景**：用户看到的症状是“窗口不开”“某个快捷键没反应”；打包的 exe 是 GUI 子系统，
没有控制台，`log.Printf` 就这么消失了。设备上的事实（哪个 dll、哪个 shell、哪行键位）
只有机器自己知道。

**决定**：
1. `internal/fault` 定义八个码（`E101` 无可用 profile、`E102` 找不到 libghostty-vt、`E103` 配置目录不可用、
   `E201` 设置不是合法 JSON、`E202` 字段越界、`E203` 键位字符串读不懂、`E301` shell 起不来、`E302` panic），
   每码一行英文含义；日志里用 `fault.Line(code, detail)` 打印，不再写自由文本。
2. `app.Run` 一开始把日志接到 `%LOCALAPPDATA%\wt-go\wt-go.log`（超过 1 MiB 下次启动轮换成 `.log.1`）。
3. `wt-go diag` 打印同一套事实：版本/提交、exe、设置文件、原生库、每个 profile 能不能启动、
   日志与崩溃日志位置、检查项；`--json` 供机器读。

**理由**：重复的排查对话全部变成“先跑 `diag`”。码与文档一对一，用户报 `E203` 时不需要
解释现象，也不需要在聊天记录里找文件；测试也能断言“这个配置应得到 E201 而不是 E202”。

**代价**：多了一层间接（码 → 含义 → 文档）；新增失败路径必须三处同步（`fault`、`cli.checks`、
`docs/故障排查.md`），漏一处就会“有码没解释”。

---

## ADR-022 排查接口用 MCP 走 stdio，不做 HTTP

**背景**：“把问题交给 AI”需要机器可读的现场。候选是本地 HTTP 端口与 MCP（Model Context Protocol）。

**决定**：实现 `wt-go mcp`：JSON-RPC 2.0 走 stdin/stdout，五个只读工具
（`wtgo_diag`/`wtgo_status`/`wtgo_settings`/`wtgo_profiles`/`wtgo_log`）与两个资源，
实际协议实现自写（`internal/mcp`，只用标准库）。

**理由**：编辑器与 agent 已经支持 MCP，用户只需在配置里加一段 `command`+`args`；
没有端口就没有监听、防火墙弹窗、端口冲突与鉴权问题；stdio 进程自然随调用方退出。
上层（`internal/cli`）复用 `diag` 的数据，两个入口不会各说一套。

**代价**：没有“一直开着的接口”，也不能从另一台机器访问（不需要）；协议版本需跟进。

---

## ADR-023 设置页列出动作目录的每一条，录制以动作 id 为键

**背景**：旧实现只渲染 `settings.json` 里已有的键位，于是没有默认键的动作（例如「关闭标签」）
在设置页里根本不存在，永远无法绑定；录制状态还是以“行号”记的，而行的意义来自文件数组。

**决定**：「操作」分组遍历 `action.All()`，每个动作一行（`entry{action,row}`，`row == -1` 表示
文件里没绑定，显示 `无`）；录制/删除都按动作 id；`SetKeybinding` 找不到就 `append` 一条。

**理由**：动作的“全集”本来就只有 `action` 一个来源，界面不列它就会出现“看得到的才是存在的”这种错误；
以 id 为键让状态在文件变化（新增/删除键位、恢复默认）后仍然对得上，而行号会错位。
删除按动作过滤，也顺手修了“一个动作有多个键时只删掉一条”的问题。

**代价**：设置页的行数不再等于文件的键位数，新增动作会让这一页变长（目前 35 行，可接受）。

---

## ADR-024 标签宽度用 `Context.MeasureText` 量，不再估算

**背景**：标签条按 `float32(220)` 换算宽度，中文名、长短不一的名字都会留下过大或过小的空隙。

**决定**：`measuredName` 用 `Context.MeasureText` 取文本宽度，夹在 `minTabName=48` 与
`maxTabName=240` 之间，超出由 `Ellipsis("…")` 截断；等宽模式仍是定宽的 `tabWidth - 60`。

**理由**：mygo 已经提供文本度量，估算等于把排版交给常量；夹上下限防止单字标签太窄、长名字挤掉其它标签。

**代价**：每帧多一次文本测量（每标签一次，可忽略）；宽度现在依赖字体，换字体后布局会跟着变——
这正是想要的。

---

## ADR-025 关闭确认按进程树判断，`confirmClose` 改三态

**背景**：`pane.Tab.Running()` 只看终端本身是否结束，停在提示符的 shell 也不算结束，
于是关任何一个活标签都弹「这个标签里的程序还在运行。」——用户一眼就看穿："实际上并没有程序"。
插件不给 PID（`terminal.Terminal` 没有进程句柄，`internal/pty` 是 internal 且 Windows 上返回 0），
所以「程序在跑」只能自己看进程树。

**决定**：

- `internal/sys/process.go` 枚举全机进程（Windows 走 Toolhelp32，其它读 `/proc/<pid>/stat`）：
  `Children`、`Descendants`（BFS，防环）、`Spawned`（两次快照里新增的、父进程是本进程的子进程）。
  创建终端时 `before`/`after` 各取一次快照，记下 `App.shells[term]`；关标签时该 PID 有子孙才算「还在跑」。
- `Spawned` 跳过 pty 的后台宿主（`conhost.exe` / `OpenConsole.exe`）：ConPTY 会把它和 shell 一起
  作为本进程的子进程拉起，pid 更大，不排除就会认错。
- 插件在 goroutine 里启动程序，`terminal.New` 返回时进程还没起来，所以等待放到后台
  （50 × 10ms 轮询），结果经 `App.Do` 回到 UI 线程写入 `shells`。
- `confirmClose` 从 bool 变三态字符串：`running`（默认）/ `always` / `never`；
  `UnmarshalJSON` 仍接受旧的 `true`（= `running`）与 `false`（= `never`）。

**理由**：只有进程树能区分"提示符"和"正在跑东西"；三态让不想被问的人一次关掉，也让喜欢确认的人保留旧行为，
而旧的 bool 无法表达"有程序才问"这个默认意图。

**代价**：每次关标签枚举一次全机进程（几百条，微秒级）；快照拿不到 PID 时退回"总是问"（宁可多问，不可误杀）；
Windows 上靠进程名排除 conhost 是启发式，改名后就会退化成"总是问"。

---

## ADR-026 关闭提示改用 `Activity` 三态，认不出就不假装没程序

**背景**：ADR-025 的判断是"有子孙进程就算在跑"，但落地后有三处说不通：
macOS 没有 `/proc`，`Processes()` 必然失败，退回 `tab.Running()` 于是空闲 shell 也提示；
两个标签同时启动时 `Spawned` 取"最大 Pid"会把两个标签认成同一个 shell；
"读不到进程列表"和"确实没程序"共用一句话，用户没法区分。

**决定**：

- `internal/sys` 把判断抽成纯函数：`Proc{Pid, Parent, Name, State}` + `ShellProc{Pid, Name}` + `ShellProc.Activity(all []Proc) Activity`，
  三态 `Idle` / `Busy` / `Unknown`。shell 自己的 Pid 找不到（已退出）→ `Idle`；
  名字对不上（PID 被复用）→ `Idle`；有非僵尸、非 pty 宿主的子孙 → `Busy`；
  拿不到 shell 的 Pid → `Unknown`。`Tab.Running()` 随之删掉。
- macOS 走 `unix.SysctlKinfoProcSlice("kern.proc.all")`（`P_stat` 映射成 Linux 风格的 `State` 字母），
  与 Linux 的 `/proc/<pid>/stat` 解析同构，`Activity` 不再知道平台。
- `Spawned(before, after, parent, taken)` 返回 `Proc` 并新增 `taken`：多个标签同时启动时，
  先看的那次先认领，后看的跳过已被认领的 Pid；选择规则从"最大 Pid"改成"最小 Pid"（更早启动的更可能是这次的那个）。
- 认领表 `App.taken`（`spawnMu` 保护）随 `forget` 释放；等待窗口从"50×10ms"改成"3 秒自适应轮询"（10ms 起每次翻倍，上限 100ms）。
- `confirmClose` 仍是三态，但 `running` 的语义变成 `Activity != Idle`；`Unknown` 时说另一句话
  （"wt-go 无法判断……"），`always` 仍用"关闭即结束"那句。读不到进程列表时打 `E303`。
- 提示会点名：`ShellProc.Programs(all)` 列出让标签忙的程序名（去重、去掉 `.exe`、跳过僵尸与 pty 宿主），
  `Activity` 就是 `len(Programs) > 0`，两处共用同一套过滤。`app` 用 `tabActivity{act, programs}` 把两者一起传给对话框，
  文案变成「还在运行：PING」。理由：用户问的是"为什么 pwsh 会提示"，点名程序比任何解释都短，
  而程序名恰好就是 `ping`/`vim` 这类用户自己起的。
- 认领时跟着"同名重执行链"往下走一层：`sys.Reexec(all, p)` 找出 p 唯一一个同名子进程，
  `spawned` 找到候选后最多再看 250ms（`shellSettle`），有就用 `reclaim` 把认领的 Pid 换到子进程上。
  理由：scoop 这类包管理器在 `PATH` 上放的是 **shim**（`…\shims\pwsh.exe`），它再启动真正的 `…\apps\pwsh\current\pwsh.exe`；
  于是标签的 shell 底下永远有一个同名子进程，空闲标签也被判成 `Busy`，提示「还在运行：pwsh」。
  实测：从资源管理器双击启动（`PATH` 命中 shim）必现，从带真 exe 的 `PATH` 启动则复现不出来。
  `taken` 随之从 `map[int]bool` 变成 `map[int]*terminal.Terminal`，`forget` 按终端清掉它的全部 Pid，
  否则 shim 的 Pid 会永久占位，将来被系统复用后新标签就认不出自己的 shell。

**理由**：产品要回答的是"关了会不会杀掉用户还在用的东西"，这个问题有三个答案，不是一个。
把 `Unknown` 单独拿出来，既不会把误报推给用户（假阳性），也不会在真的不知道时说"没事"（假阴性）。
纯函数化之后，macOS 的采集错误不可能再伪装成"有程序在跑"。

**代价**：shell 内置的长命令（`Start-Sleep`）看不到，这是进程树的固有盲区（插件内部的 OSC 133 提示符状态没有导出）；
只在文档里说明，不为此引入新的采集层。启动器若只在一个短暂窗口内启动子进程（超过 `shellSettle`），仍会被当成 shell 本身。

---

## ADR-027 标签条自己滚动：溢出了才出现的箭头 + 视口接管滚轮

**背景**：标签多到放不下时，原来是把标签压窄（名字截断、`×` 被挤掉），再往后连标题栏的拖动区都没了。
计划要求：统一的最小宽度、独立的可视范围、`<`/`>` 指示、滚轮横向浏览，且不能改 `internal/ui` 的纯绘制职责。

**决定**：

- 标签条拆成四段：滚动视口（`ScrollHorizontal` + `TrackScroll` + `Grow(1)`）、`❮`/`❯`、`▾`、`+`、以及一段固定 56dp 的拖动区。
  标签显式 `.Shrink(0)`，宽度 = 60（有圆点）或 44（无圆点）+ 名字宽，最小 108。
- 箭头只在标签放不下时出现（`MaxX > 0`），出现后无处可滚的一侧把颜色压到 45% 透明度。
  理由：箭头本身要占 48dp，不溢出时也占着，等于从标签嘴里抢地方；而用 `Disabled` 会连点击一起关掉，
  用户点到"灰但该能用"的箭头会以为坏了。
  箭头占地方使视口变窄，所以"出现"不能每帧重新算：mygo 的布局输出（`MaxX`）在构建期只读得到上一帧的值，
  一旦显示，多出的 48dp 正好让内容真的溢出，状态自洽；反过来从显示到隐藏只需一帧——
  关标签后 `MaxX` 归零的那一帧读到的还是旧值，下一帧才收起来。
  边界（内容只多出不到 48dp）会取决于历史：从无箭头变成有箭头后不再收回，这是有意的——
  此时视口确实比内容窄，箭头确实能用。
- 滚轮由视口自己的 `HandleInput` 把 `DY` 当横向位移吃掉（一格 ≈ 99dp）。理由：mygo 的 `scrollBy` 只让纵向滚轮动纵向容器，
  纯横向容器永远收不到滚轮。
- "滚到当前标签"由 `State.RevealTab` 传意图，`ui` 侧对该标签调 `ScrollIntoView`。
  意图持续两帧（`App.revealTabs = revealFrames`）：箭头出现的下一帧视口才变窄，只问一帧的话
  活动标签会在箭头挤进来之后又被截掉半个；问多了则会把用户手动滚走的视口拉回来。
  几何仍全在 `ui` 里算，`app` 只提供 `ScrollTabs(by)` 与滚动状态。
- 标签名 `.Unselectable()`：名字是标签，不是可选文本，否则按住名字拖拽会变成选字。

**理由**：最小宽度 + 滚动才是终端标签栏的通行做法（WT 也是），压缩会让每个标签都变得不可点；
把"滚多少"交给 `ui`、把"滚到哪"留给 `app` 的一帧意图，`Actions` 只多一个 `ScrollTabs(by float32)`。

**代价**：箭头出现时占 48dp（标签因此少显示一点），以及 `RevealTab` 需要 app 侧记两帧。

---

## ADR-028 `diag --redact`：报告要能贴出去，但不说明这台机器属于谁

**背景**：`wt-go diag` 的全部价值在于把现场交给别人看（issue、聊天、AI），而报告里的每一行都带着
绝对路径：`C:\Users\<名字>\AppData\Local\wt-go\settings.json`。贴出去等于连账号名一起贴。

**决定**：

- 新增 `wt-go diag --redact`（可和 `--json` 同时用）：把 `%USERPROFILE%`、`%LOCALAPPDATA%`、`%APPDATA%`、
  `%TEMP%`（Linux/macOS 上是 `~`、`$XDG_CONFIG_HOME`）这些前缀换成占位符。
- 占位符由 `sys.Redactions()` 给（它才知道家目录在哪），替换本身是 `cli` 里的纯字符串处理，
  保持 `cli` 不碰 OS 的约定；长前缀优先，嵌套目录不会被外层吃掉。
- 只改报告，不改磁盘；默认不开，因为本地自己看时完整路径更有用。

**理由**：把"该不该脱敏"交给用户决定，但把能力内置：让用户自己去手改几十行路径，实际结果是没人贴报告。
不用"默认脱敏"是因为报告要回答"是哪个文件坏了"，路径本身是信息。

**代价**：多一个开关要记；`--redact` 之后 `exe:` 这类不在家目录下的路径仍然是原样（那是安装位置，不是身份）。

---

## ADR-029 Linux 也用自绘标题栏，按钮跟随桌面布局

**背景**：ADR-011 让 Linux 把标题栏交给平台，因为 mygo 的 `TitleBarHidden` 在 X11 上只做
`gtk_window_set_decorated(false)`。代价是 Linux 上多一行 GTK/窗口管理器的标题栏，它的配色来自
桌面主题（GTK3 主题），与终端配色是两套：切深浅色、换配色方案时这一行都不跟。

**决定**：

- Linux 与 Windows/macOS 一致：`TitleBarHidden` + `HideWindowButtons: true`
  （`internal/app/titlebar_linux.go`）。GTK 于是不建 `GtkHeaderBar`，主题碰不到标题栏；
  Wayland 上 GTK 仍自装饰（阴影、缩放边），X11 上窗口只保留内容区。
- 按钮位置与顺序跟随桌面：`mygo.Theme.WindowButtons()` 读 `gtk-decoration-layout`，
  `ui/tabs.go` 的 `windowButtonGroup` 按 `Start`/`End` 在标签条两端画；最大化状态由
  `Window.OnMaximize`/`OnUnmaximize` 回填（最大化时方框变成两个错开的方框），布局变了由 `mygo.Theme.OnUpdated` 刷新。
  图标用盒子画（横线 / 方框 / 两个错开的方框 / `×`），不用字形：这几个字形在部分字体里缺失或过宽；
  每个按钮用 `Label` 命名，读屏与测试都按名字找。
- 按钮经 `ui.Actions`（`MinimizeWindow`/`ToggleMaximizeWindow`/`CloseWindow`）回到 `app`，
  用 `a.Do` 排到下一轮主循环：`onMain` 在主线程是内联的（`mygo` 的 `loop.go`），
  在绘制帧里直接 `Close()` 会在帧内销毁窗口（sshyard 用 `go w.Close()` 是同一个坑）。

**理由**：标题栏是窗口的一部分，配色该跟终端走。v0.8.3 的 `HideWindowButtons` 让 GTK 干脆不建
header bar，比“留 GTK 标题栏再用 CSS 压掉 80px 空标题”（sshyard 的 `gtkfix`）干净，也不用 purego
手写 GTK 调用。按钮跟桌面布局，是让平铺/左侧按钮的用户拿到自己熟悉的位置，而不是硬编一套。

**代价**：`gtk-decoration-layout` 被清空（平铺窗口管理器常见）与“合成器自己画标题栏”（KWin Wayland）
在 API 上不可区分，两种情形都返回空，于是都不画按钮——平铺 WM 下窗口没有按钮，只能靠快捷键或
窗口管理器。真有人要，再给一个环境变量兜底（kero 用 `KERO_WINDOW_CONTROLS`）。
窗口按钮不是“动作”：没有 id、不进 `internal/action`、不可绑键位。

---

## ADR-030 字体族不是等宽就回落到 `monospace`

**背景**：`font.family` 写了一个这台机器没有的族名（Linux 上常见：`Cascadia Mono`），
fontconfig 按兜底链挑中一个**比例字体**（更纱黑体 SC）。mygo 的终端网格取“可打印 ASCII 里
最宽字形的前进量”当格宽（`mygo/plugins/terminal/paint.go:93`），比例字体里最宽的字 17px、
小写只有 3–8px，于是每格 17px、字形 7px、左右各空 5px：字距极宽、列数暴跌、本该一行的话提前折行。
全程不报错，是静默降级。

**决定**：

- `internal/app/font.go` 的 `monospace(family, size)` 用 `ui.Shape` 量一遍可打印 ASCII，
  前进量按 `Glyph.Runes` 归一化（`->` 这类连字是一个字形占两格），极差 < 0.5 DIP 才算等宽；
  空族名按 `monospace` 算，与 mygo 的 `fontKey.key()` 一致（它把空族名与 0 号都换成默认值）。
- `terminalFont()` 在不等宽时换成 `monospace`（每个系统都认这个通用族名），
  并记一条 `E204`（`fault.BadFont`），同一个族名只记一次（`App.fontWarned`）。
- 校验放在“把设置交给终端”的地方：开标签（`termOptions`）与改设置（`apply`）都走这一条路。

**理由**：终端只能用等宽字体，比例字体下的网格必然是错的；与其让人面对错位，不如换成能用的字体
并把原因写进日志（`wt-go diag` 的 `last fault:` 会抄出来）。不为这一个判断开新包：只有一处调用，
放 `app` 里，`ui.State` 不必多字段。

**代价**：设置里写的族名“看着生效了但没生效”，只能从日志/`diag` 看出来。

**后续**：设置页的字体栏改成从列表里挑（`monospaceFamilies`：`ui.FontFamilies()` 里引擎说是等宽、
量出来也等宽的那些），用户不必猜本机装了哪个族名；逗号列表仍可在 `settings.json` 里写。
Windows 的默认值也随之改成 `Cascadia Mono,monospace`：没装 Cascadia 的机器自己落到等宽，
不再靠 `E204` 救。

---

## ADR-031 设置页的控件用现成的，不手搓

**背景**：设置页原先自己拼控件：下拉是 `myui.MenuButton` 打开的桌面原生菜单，字体栏是
`myui.Combobox`（失焦会把没匹配到选项的输入退回原值），数字是 `-` / 输入框 / `+` 三个独立方块，
标签左对齐固定 200 宽，说明文字单行截断。mygo v0.8.3 已经带了主题化的 `myui.Select`、
`myui.NumberInput`、`myui.Form`、`myui.Field`。

**决定**：

- 控件先看有没有现成的：`internal/ui/form.go` 只留薄封装（`field` / `section` / `row` /
  `choice` / `count` / `measure` / `ratio`），调用点签名不变，底下换成 `myui.Form`、
  `myui.Field`、`myui.Select`、`myui.NumberInput`。
- “现成的”不限于 mygo 自带：`D:\Sou\MyGO参考` 里能直接用的组件同样可用（`mygo-kit`
  自己写 `Select` 是因为 v0.7.3 没有，v0.8.3 有了就不必再搬）。前提是不引新依赖、
  不越过纯绘制边界（`internal/ui` 不碰 OS）。

**理由**：自带的控件跟主题、键盘、读屏、`myui.NewTester` 一起演进，手搓的每一份都要自己
维护这些；设置页的控件又不特殊。

**代价**：`myui.Select` 的值与显示名是同一个字符串，`choice` 于是用显示名往返映射设置值
（值不在选项里时原样显示，例如手写的字体族列表）。`myui.Field` 把标签排成右对齐的一列、
说明移到控件下方：原来单行截断的长说明现在完整显示，但左侧标签列比原来的 200 宽窄一些。

---

## ADR-032 界面主题从配色方案推导

**背景**：`theme` 原先只有 `system` / `light` / `dark`，整个窗口是 mygo 的两套默认 token，
与终端里选的配色方案无关——用 Dracula 的终端配 VS Code 灰的标签栏。mygo v0.8.3 的
`ui.Theme` 有完整 token 集、`Context.SetTheme` 可以整棵替换，`mygo.Theme.SetSource` 只管
原生窗口（Linux 的 CSD、Windows 的标题栏）。参考项目里 kero 就是这么做的：拿终端方案的
背景与光标推导整套 UI 颜色（`kero/theme.go:334 ghosttyLook`）；dgopher 用 JSON 主题文件，
Mayoi 用 13 个写死的主题表。

**决定**：

- `theme` 可以是 `system` / `light` / `dark`，也可以是任意方案名（内置、Ghostty 617 套、
  `schemes` 里自己写的）。方案名时整个窗口换成那个方案的颜色。
- 界面 token 不手写，从方案的背景 `bg`、前景 `fg`、强调色推导（`internal/theme/ui.go`
  `window()`）：表面色是 `bg` 往 `fg` 混 7%/11%/15%，边框 20%，次要文字 38%；强调色取
  方案调色板第 5 色（各家的蓝），没有就用光标，再没有用前景；强调色上的文字按 WCAG 对比度
  取 `bg` 或 `fg`。
- 明暗不由 `theme` 字段的字面值决定，由方案自己的背景亮度决定：选了 `Catppuccin Latte`
  就是浅色窗口，终端没指定 `colorScheme` 时也跟着落到浅色底。
- 设置页的下拉列 11 个挑过的方案（`UIThemes()`）；写别的名字照样生效，只是不在下拉里。

**理由**：一份配色方案已经写好了「这个主题该长什么样」，界面再挑一遍颜色只会打架；推导出来的
窗口与终端同源，换方案时两边一起变。挑 11 个而不是 617 个，是因为下拉里滚 617 项没人受得了，
且多数方案的背景对比度撑不起一套可用的 UI。

**代价**：推导只能保证「能看」，不保证与方案作者给的 UI 版一致（比如 Catppuccin 官方给 UI 的
表面色是另算的）。强调色固定取调色板第 5 色，是因为几乎所有方案的光标都是灰的，先取光标会让
强调色变成无彩度。`mygo.WindowOptions.BackgroundColor` 在主题生效时取该主题的底色，页面画满
后看不到差别。

---
## 待定

- 会话常驻：阶段 5 评估是否采用 ADR-009 所述的会话服务模型。
- 验证 mygo `plugins/terminal` 是否支持外部字节流作为终端连接（gorex 即如此接入）。注意：gorex 依赖的是 mygo v0.3.5 伪版本，wt-go 锁定 v0.8.3，API 未核对，需阶段 0 实测。
- 多窗口：独立进程还是单进程多窗？前者简单，后者才能共享设置热重载。
- 是否内置 SSH（参考 sshterm 可直接拿）：暂定不做，外部调用 `ssh` 即可。
- 控制接口（HTTP/MCP）供自动化测试：阶段 3 之后仍未决定。已部分回答——诊断入口选了 MCP/stdio（ADR-022）；**给自动化测试用的控制接口仍未决定**（需要能驱动标签/窗格、读终端内容）。


