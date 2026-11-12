# 设计决策

按时间记录，后加的决定追加在末尾，不覆盖旧的。状态只有三种：**有效**（照此实现）、
**已被 ADR-NN 修订**（新条目说明了变在哪）、**归档**（记录当时的结论或数字，代码已不在或不再变）。

| # | 主题 | 状态 |
| --- | --- | --- |
| ADR-01 | 用 mygo 而不是 WPF/Electron/Tauri | 有效 |
| ADR-02 | 终端内核用 libghostty-vt，不自研 VT | 有效 |
| ADR-03 | 只接管自己创建的进程 | 有效 |
| ADR-04 | 配置格式自定，不兼容 WT 的 settings.json | 有效 |
| ADR-05 | 界面与动作分离 | 有效 |
| ADR-06 | 先不到 sshterm 里改 | 有效 |
| ADR-07 | 窗格树独立为 `internal/pane` | 有效 |
| ADR-08 | OS 副作用集中到 `internal/sys`，纯解析包不做 IO | 有效 |
| ADR-09 | 会话常驻（应用退出后进程不结束）暂不采用 | 有效 |
| ADR-10 | 配色方案：自定义 > 内置 > Ghostty 主题 | 有效（内置方案现为 7 套） |
| ADR-11 | 标签栏兼作标题栏 | Windows/macOS 有效；Linux 例外已由 ADR-28 取消 |
| ADR-12 | 窗格几何与焦点移动用矩形判定，不用树序 | 有效 |
| ADR-13 | 键盘焦点每帧用 `Element.Focus()` 保持 | 有效 |
| ADR-14 | 设置页实时预览：改了就应用，保存只负责写盘 | 有效 |
| ADR-15 | profile 编辑用工作副本，保存时才写回 `profiles` | 有效 |
| ADR-16 | 热重载以文件字节为准，有未保存改动时不覆盖 | 有效 |
| ADR-17 | 多行粘贴确认在 `terminal.Options.OnPaste` 里拦 | 有效 |
| ADR-18 | 启动大小（列×行）用两次测量，不估算单元格尺寸 | 已被 ADR-46 取代 |
| ADR-19 | 子进程收尾用“把本进程加进 Job Object” | 有效（`BREAKAWAY_OK` 已用于重置，见 ADR-35） |
| ADR-20 | 崩溃日志写在配置目录，写完后重新 panic | 有效 |
| ADR-21 | 失败给编号，日志写文件，`diag` 一次说完 | 有效 |
| ADR-22 | 排查接口用 MCP 走 stdio，不做 HTTP | 有效 |
| ADR-23 | 设置页列出动作目录的每一条，录制以动作 id 为键 | 有效 |
| ADR-24 | 标签宽度用 `Context.MeasureText` 量，不再估算 | 有效 |
| ADR-25 | 关闭确认：按进程判断，三个答案不是一个 | 有效 |
| ADR-26 | 标签条自己滚动：溢出了才出现的箭头 + 视口接管滚轮 | 有效 |
| ADR-27 | `diag --redact`：报告要能贴出去，但不说明这台机器属于谁 | 有效 |
| ADR-28 | Linux 也用自绘标题栏，按钮跟随桌面布局 | 有效 |
| ADR-29 | 字体族不是等宽就回落到 `monospace` | 有效 |
| ADR-30 | 设置页的控件用现成的，不手搓 | 有效 |
| ADR-31 | 界面主题从配色方案推导 | 有效 |
| ADR-32 | Linux 用 fontconfig 规则保住文字颜色 | 有效 |
| ADR-33 | 界面图标用打进二进制的 SVG，不用字体字形 | 有效 |
| ADR-34 | 自定义配色按「基名 + (custom)」命名，行尾给删除 | 有效 |
| ADR-35 | 启动时发现的 shell 写进 settings.json，探测只在写第一份文件时 | 有效 |
| ADR-36 | 字体下拉的族列表按需枚举，等宽判定按 `(family, size)` 记忆化 | 有效 |
| ADR-37 | 不做进程内 `malloc_trim` | 结论：不做（别再试） |
| ADR-38 | 设置页一帧取一次快照，配色列表随 `apply` 失效 | 有效 |
| ADR-39 | 等 shell 出现时读内核的 children 文件，不再全量扫进程表 | 有效 |
| ADR-40 | 性能基线（2026-11-11 单机实测，一次性记录） | 归档（数字为单机实测） |
| ADR-41 | 文案交给 mygo 的 catalog，`internal/i18n` 只剩一个薄门面 | 有效 |
| ADR-42 | 目录留空时回落到进程 cwd，而不是家目录 | 有效 |
| ADR-43 | 设置页画成卡片行，弹菜单画在窗口里，「关于」只读编译期信息 | 有效 |
| ADR-44 | 默认配置文件留空＝跟随系统，不是列表第一个 | 有效 |
| ADR-45 | 标签前是一块方块，不是色点 | 有效 |
| ADR-46 | 窗口开在上次的位置和大小，`startSize` 删掉 | 有效 |
| ADR-47 | 把 Ghostty 的四个行为做成设置：滚轮倍数、最小对比度、内边距颜色、剪贴板读写 | 有效 |

---

## ADR-01 用 mygo 而不是 WPF/Electron/Tauri

**决定**：Go + mygo。

**理由**：用户要求用 mygo；单文件分发、库体积小、启动快；与 sshterm 同栈，
验证过的做法能直接沿用。

**代价**：mygo 生态较新，遇到框架缺陷需要绕或提交上游；无 cgo 意味着某些
Win32 API 要手写 syscall。

---

## ADR-02 终端内核用 libghostty-vt，不自研 VT

**决定**：`github.com/egoist/mygo/plugins/terminal`。

**理由**：终端模拟器正确性极难（宽字符、reflow、Kitty 键盘协议、OSC 52、
bracketed paste、鼠标上报……）。libghostty-vt 是 Ghostty 的成熟实现，且已解决
Windows ConPTY 接入。

**代价**：受限于它的 Options 暴露面；需要的功能（如某些自定义键位）要通过
`Keybinds` / `OnKey` 绕。

---

## ADR-03 只接管自己创建的进程

**决定**：不做"附加到已存在的终端窗口/ConPTY 会话"。

**理由**：Windows 没有官方的"终端复用"接口，ConPTY 管道的句柄只对创建者有效；
第三方接管要靠注入，不稳定且触发安全软件。Windows Terminal 本身也只管理自己
创建的子进程。

**推论**：想"在 wt-go 里看到别的终端"只能新开一个进程，而不是搬过来。

---

## ADR-04 配置格式自定，不兼容 WT 的 settings.json

**决定**：wt-go 用自己的扁平结构。

**理由**：WT 的配置是 `defaults.json` + `settings.json` 两层合并、`profiles.list`
数组、`actions` 命令对象，实现与解释成本都对不上当前规模。扁平结构更好读、好写、
好热重载。

**代价**：从 WT 迁移要手抄一次。

---

## ADR-05 界面与动作分离

**决定**：`internal/ui` 只实现 `Actions` 接口的调用方，不直接改状态里的进程/文件。

**理由**：沿用 sshterm 的经验——界面单测可以用假 Actions（`idle{}`），不必起真进程；
异步回调统一 `Do(fn)` 回主线程，避免竞态。

---

## ADR-06 先不到 sshterm 里改

**曾被尝试**：在 sshterm 里给本地 shell 加"管道模式"跑 pi。

**放弃原因**：偏离用户目标（要的是一个终端软件，不是 pi 的容器）；且管道模式
引入大量非标准代码（隐藏控制台、光标探针记历史），收益不抵风险。代码保留在
sshterm 分支未提交，不影响 wt-go。

---

## ADR-07 窗格树独立为 `internal/pane`

**决定**：`Tab` / `Node` / `Pane` 与切分、焦点、关闭补位逻辑放在 `internal/pane`，不放进 `internal/ui`，也不放进 `internal/app`。

**理由**：`ui.State.Tabs` 需要 `*pane.Set`，而 `ui` 不能 import `app`；`layout` 只算矩形、不认识终端。`pane` 持有 `*terminal.Terminal` 句柄，因此单独成包，依赖 `layout` 与 mygo `terminal`。

**约定**：`pane` 不做 IO，不关闭进程。移除窗格时返回被移除的句柄，由 `app` 关闭。

**参考**：`egoist/gorex` 把窗格树放在 `package main`（`state.go`），与 UI 混在一起。wt-go 为了可单测，拆出独立包。

---

## ADR-08 OS 副作用集中到 `internal/sys`，纯解析包不做 IO

**背景**：原文档“只有 `internal/app` 能碰文件系统和进程”与 `store`（读写 `settings.json`、热重载）、`profile`（阶段 1 读注册表发现 WSL）的职责冲突。

**决定**：
- `internal/sys` 封装文件读写与监听、注册表读取、Job Object，只被 `internal/app` import。
- `store`、`profile` 只做纯函数的解析与序列化，字节或数据由 `app` 读入后注入。
- 其他包需要 OS 能力时，在使用方定义接口，由 `app` 注入实现。

**参考**：`egoist/gorex` 的 `internal/rex`（会话服务、客户端、ConPTY、进程查询）只依赖标准库与 `golang.org/x/sys/windows`，不依赖 mygo，也不依赖 UI。UI 通过 `rex.Stream` 把字节流接给 mygo `terminal`。这证明了“协议与进程层”可以与 UI 层干净分离。

---

## ADR-09 会话常驻（应用退出后进程不结束）暂不采用

**背景**：`egoist/gorex` 用独立的会话服务进程（经 Unix socket 通信，JSON 行协议）持有 ConPTY，应用退出后进程继续运行，重开窗口时恢复标签与窗格。

**决定**：阶段 0–4 不做，保持 ADR-03 的“自己创建、自己管理生命周期”模型，应用退出 = 关闭所有标签。

**理由**：会话服务带来多进程、版本匹配（gorex 为此做了 `staleServer` 检测）、Windows 上分离进程等复杂度，超出阶段 0–4 的范围。ADR-03 不冲突（进程仍是自己创建的），但会改变生命周期模型。

**约束**：`internal/sys` 与 `internal/pane` 的接口不得假设“进程随窗口结束”，以便阶段 5 的“启动恢复”可以再评估。

---

## ADR-10 配色方案：自定义 > 内置 > Ghostty 主题

**背景**：`plugins/terminal` 自带 Ghostty 的 617 套主题（`GhosttyTheme(name)`），但 WT 的
Campbell / Solarized / Tango / Vintage 都不在其中。

**决定**：`internal/theme` 按 `settings.schemes` → 内置（手写 7 套 WT 取值：Campbell 与它的 Powershell 变体、Solarized Dark/Light、Tango Dark/Light、Vintage）→ Ghostty →
`theme` 对应的默认配色（`terminal.DarkTheme()` / `LightTheme()`）的顺序解析；自定义方案只写
部分字段时，未写的字段沿用同名内置/Ghostty 方案的值。

**理由**：不自己维护调色板库（Ghostty 已覆盖绝大部分需求），同时保证 WT 用户的
`colorScheme` 名字能直接用；局部覆盖让“只改背景色”这类需求不用抄一整套 16 色。

**代价**：名字写错静默回退到 `theme` 的默认配色（不报错、不崩溃），排错要看 `docs/使用说明.md` 的顺序说明。

---

## ADR-11 标签栏兼作标题栏

**决定**：Windows / macOS 上窗口用 `mygo.TitleBarHidden` + `TitleBarHeight = ui.TitleBarHeight`（40 DIP），
标签栏自己就是拖窗区域（空白处 `DragWindow`），不再单独画一条标题栏。

**理由**：与 Windows Terminal 一致（用户提供的参考图即如此），省掉一行高度；
`mygo` 的 `DragWindow` 放在容器上不影响子元素点击（`ui/input.go:481`）。

**代价**：标签栏右侧必须留空白给拖窗；标签放不下时没有横向滚动（后来由 ADR-26 的滚动视口补上）。

**状态**：Windows / macOS 有效。最初 Linux 例外地交给平台画标题栏，那条例外已由 ADR-28 取消。

---

## ADR-12 窗格几何与焦点移动用矩形判定，不用树序

**背景**：切分出来的窗格树里，树的顺序不等于视觉相邻：`A | (B1 / B2)` 中 B2 的“左边”
在视觉上是 A，在树里却先遇到 B1。

**决定**：`internal/layout` 只算矩形（`Split` 给出两格与分割线，`Share` 把比例夹在两端之内）；
`pane.Tab.Move` 按几何选目标：同方向的候选中取跨度重叠最多者，其次取最近者。
窗格上限（`maxPanes`）是策略而非几何，因此 `Tab.Split` 不收上限参数，由 `app` 拦。

**理由**：树序解法要在每个 split 上做方向归一，代码量大且易错；矩形判定只用现成的
几何，行为与用户看到的一致。

**代价**：每次移动焦点都重算一遍布局（窗格数 ≤ `maxPanes`，成本可忽略）。

---

## ADR-13 键盘焦点每帧用 `Element.Focus()` 保持

**背景**：`mygo` 的 `AutoFocus()` 只在元素“出现的那一帧”生效；窗格切分后旧窗格已经在树上，
只给它 `AutoFocus` 并不会把焦点从别的窗格拿回来（实测：`Alt+←` 后输入仍进原窗格）。

**决定**：焦点窗格每帧调 `Element.Focus()`（文档：“called in every frame, it keeps it there”）；
对话框（重命名）打开时不调，让 `Modal` 里的输入框拿到焦点。

**理由**：状态里只有一个“焦点窗格”，界面按它决定谁拿键盘，不需要额外的焦点请求队列。

**代价**：每帧都发一次焦点请求；如果将来有多个可输入控件（如设置页）同时存在，
这条规则要重新考虑。（阶段 3 落地后仍然成立：设置标签页没有窗格，两者不会同时出现；
对话框打开时 `st.Rename`/`st.Confirm` 非空，窗格那一路本来就不画。）

---

## ADR-14 设置页实时预览：改了就应用，保存只负责写盘

**背景**：WT 的设置页是“改完保存才生效”，但它能预览是因为它的设置模型是“重启式”的；
而字体、配色、透明度这些项的预览价值很高，等保存才看到效果会很别扭。

**决定**：表单直接改当前生效的设置（`a.settings`），每帧与 `a.applied` 比较，不同就调用一次幂等的
`apply()`（语言、窗口明暗、键位、profile、每个终端的字体/配色/选中即复制、窗口透明度/底色/置顶）。
「保存」= 写文件并把 `a.saved` 更新为当前值；「放弃」= 把 `a.settings` 换回 `a.saved` 再 `apply()`。

**理由**：只有一条“当前设置”通路，不需要“预览层 + 生效层”两套状态，也不需要为每个控件写预览逻辑。

**代价**：不点保存就关掉程序，改动会丢（和 WT 一致）；`apply()` 必须幂等且便宜，
新增设置项时要记得把它加进 `apply()`，否则只会写进文件而不生效。

---

## ADR-15 profile 编辑用工作副本，保存时才写回 `profiles`

**背景**：设置页要能看/改自动发现的 profile，但自动发现的结果只在内存里（不落盘）。
若直接绑到 `a.settings.Profiles[i]`，翻一遍页面就会把发现项写成用户配置。

**决定**：`internal/app` 持有 `a.editing *store.Profile` + `a.editingArgs string`（多行文本框的原始文本），
打开某页时从现有 profile 复制；`SaveSettings` 时才 upsert 进 `a.settings.Profiles`。

**理由**：读写分离清楚；`args` 用文本保存，避免每帧 `join`→`split` 把回车产生的空行吃掉。

**代价**：多一份状态（`editing`/`editingFrom`/`editingArgs`），dirty 判定要看三处。

---

## ADR-16 热重载以文件字节为准，有未保存改动时不覆盖

**背景**：设置页自己也会写文件，若把“文件变了”当成外部修改，会和自己的保存打架。

**决定**：`internal/app` 每秒读一次 `settings.json`，与上次读到的字节（`a.savedData`）比较；
不同且当前没有未保存改动 → 解析并应用；有未保存改动 → 不动，在设置页底部提示“放弃当前修改才能应用”；
解析失败 → 保留旧值并把错误写进提示。

**理由**：比“监听文件事件”简单（不依赖平台 API），一秒的延迟对配置文件完全够用；
比较字节而不是时间戳，避免程序自己写入后又被自己触发。

**代价**：每秒一次读文件（很小）；外部改动在有一秒内的延迟。

---

## ADR-17 多行粘贴确认在 `terminal.Options.OnPaste` 里拦

**背景**：要问“确定粘贴 N 行？”，需要拿到粘贴内容并在写入前取消它。

**决定**：`OnPaste func(transfer.Data) bool`（主线程、粘贴前调用，返回 true 表示已接管）：
按行数判断，需要问就弹确认框，确认时调用 `term.PasteData(d)`，并返回 true。

**理由**：插件已经提供了这个钩子，且在主线程上调用，可以直接改应用状态（不需要 `Do`）。

**代价**：粘贴内容要读两次（一次数行、一次写入），大段粘贴会有少量重复拷贝。

---

## ADR-18 启动大小（列×行）用两次测量，不估算单元格尺寸

> **已被 ADR-46 取代**：`startSize` 已删除，窗口几何由记住的窗口状态决定。以下保留为历史。

**背景**：终端只报告“能放下多少格”，不报告一格多宽；mygo 也没有字体度量接口。
第一版用“内容区宽 ÷ 列数”估算，实测 100×30 得到 97×29（还漏算了 36dp 标题栏）。

**决定**：`internal/app/size.go` 的状态机：先把窗口撑大 120px 并读两次稳定后的格数，
用差值算出 `cellW/cellH`，再按 `startSize` 反推窗口尺寸（窗口尺寸 = 单元格区 + 窗口与客户区的差值，
其中客户区高度 = 内容区 + 标题栏）。终端尺寸比窗口晚一帧，所以每次改窗口后都要求“连续两次读数相同”才算稳定。

**理由**：只用终端自己报的数据，不依赖字体文件解析，换字体/DPI 都自然成立。

**代价**：启动时窗口会先大一点再收回来（几帧，肉眼可见一瞬）；设置 `startSize` 时才有这段逻辑。

---

## ADR-19 子进程收尾用“把本进程加进 Job Object”

**背景**：窗格的 shell 是本进程的子进程。窗口被强杀（任务管理器、崩溃、`Stop-Process -Force`）时，
`Terminal.Close()` 不会跑，`pwsh.exe` 与它起的进程会留在系统里。

**决定**：`sys.HoldChildren()` 在 `Run` 开头建一个带 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` 的作业，
把**当前进程**加进去，句柄不关。子进程自动继承作业成员资格，本进程一结束（无论怎么结束）
Windows 关句柄 → 作业里的进程全部终止。

**理由**：比“每开一个窗格就枚举子进程再 AssignProcessToJobObject”简单得多，
也不会漏掉 shell 后来自己起的子进程；一个调用覆盖所有窗格、所有切分。

**代价**：本进程自己也在作业里，所以“脱离父作业”这类能力（`BREAKAWAY_OK`）用不上；
如果以后要把某个窗格做成“关窗后继续跑”（会话常驻，ADR-09），这条规则要重新考虑。
实测：强杀 wt-go 后 `pwsh.exe` 与两个 `conhost.exe` 全部退出。

---

## ADR-20 崩溃日志写在配置目录，写完后重新 panic

**背景**：上一阶段的空指针崩溃只能靠 `Start-Process -RedirectStandardError` 才看到，
用户那边就是一个窗口消失了。

**决定**：`Run` 与 `watch` 协程各带一个 `recover`，调 `sys.WriteCrashLog`（时间、原因、栈）
写到 `%LOCALAPPDATA%\wt-go\crash-<时间>.log`；`Run` 写完后 `panic(r)` 继续，保留非零退出码与原有栈。

**理由**：写文件不依赖 stderr 被重定向；日志与 `settings.json` 并排，用户找得到。

**代价**：`watch` 协程的 panic 只记录不退出（该协程停掉，热重载失效），下次启动看日志才知道。

---

## ADR-21 失败给编号，日志写文件，`diag` 一次说完

**背景**：用户看到的症状是“窗口不开”“某个快捷键没反应”；打包的 exe 是 GUI 子系统，
没有控制台，`log.Printf` 就这么消失了。设备上的事实（哪个 dll、哪个 shell、哪行键位）
只有机器自己知道。

**决定**：
1. `internal/fault` 定义 16 个码：`E101` 无可用 profile、`E102` 找不到 libghostty-vt、`E103` 配置目录不可用；
   `E201` 设置不是合法 JSON、`E202` 字段越界、`E203` 键位字符串读不懂、`E204` 字体不是等宽、`E205` 写不进配置、
   `E206` 删不掉配置、`E207` 重启失败、`E208` 打不开目录；`E301` shell 起不来、`E302` panic、`E303` 读不到进程列表、
   `E304` 没认出标签的 shell、`E305` 找不到 fontconfig，
   每码一行英文含义；日志里用 `fault.Line(code, detail)` 打印，不再写自由文本。
2. `app.Run` 一开始把日志接到 `%LOCALAPPDATA%\wt-go\wt-go.log`（超过 1 MiB 下次启动轮换成 `.log.1`）。
3. `wt-go diag` 打印同一套事实：版本/提交、exe、设置文件、原生库、每个 profile 能不能启动、
   日志与崩溃日志位置、检查项；`--json` 供机器读。

**理由**：重复的排查对话全部变成“先跑 `diag`”。码与文档一对一，用户报 `E203` 时不需要
解释现象，也不需要在聊天记录里找文件；测试也能断言“这个配置应得到 E201 而不是 E202”。

**代价**：多了一层间接（码 → 含义 → 文档）；新增失败路径必须三处同步（`fault`、`cli.checks`、
`docs/故障排查.md`），漏一处就会“有码没解释”。

---

## ADR-22 排查接口用 MCP 走 stdio，不做 HTTP

**背景**：“把问题交给 AI”需要机器可读的现场。候选是本地 HTTP 端口与 MCP（Model Context Protocol）。

**决定**：实现 `wt-go mcp`：JSON-RPC 2.0 走 stdin/stdout，五个只读工具
（`wtgo_diag`/`wtgo_status`/`wtgo_settings`/`wtgo_profiles`/`wtgo_log`）与两个资源，
实际协议实现自写（`internal/mcp`，只用标准库）。

**理由**：编辑器与 agent 已经支持 MCP，用户只需在配置里加一段 `command`+`args`；
没有端口就没有监听、防火墙弹窗、端口冲突与鉴权问题；stdio 进程自然随调用方退出。
上层（`internal/cli`）复用 `diag` 的数据，两个入口不会各说一套。

**代价**：没有“一直开着的接口”，也不能从另一台机器访问（不需要）；协议版本需跟进。

---

## ADR-23 设置页列出动作目录的每一条，录制以动作 id 为键

**背景**：旧实现只渲染 `settings.json` 里已有的键位，于是没有默认键的动作（例如「关闭标签」）
在设置页里根本不存在，永远无法绑定；录制状态还是以“行号”记的，而行的意义来自文件数组。

**决定**：「操作」分组遍历 `action.All()`，每个动作一行（`entry{action,row}`，`row == -1` 表示
文件里没绑定，显示 `无`）；录制/删除都按动作 id；`SetKeybinding` 找不到就 `append` 一条。

**理由**：动作的“全集”本来就只有 `action` 一个来源，界面不列它就会出现“看得到的才是存在的”这种错误；
以 id 为键让状态在文件变化（新增/删除键位、恢复默认）后仍然对得上，而行号会错位。
删除按动作过滤，也顺手修了“一个动作有多个键时只删掉一条”的问题。

**代价**：设置页的行数不再等于文件的键位数，新增动作会让这一页变长（目前 35 行，可接受）。

---

## ADR-24 标签宽度用 `Context.MeasureText` 量，不再估算

**背景**：标签条按 `float32(220)` 换算宽度，中文名、长短不一的名字都会留下过大或过小的空隙。

**决定**：`measuredName` 用 `Context.MeasureText` 取文本宽度，夹在 `minTabName=48` 与
`maxTabName=240` 之间，超出由 `Ellipsis("…")` 截断；等宽模式仍是定宽的 `tabWidth - 60`。

**理由**：mygo 已经提供文本度量，估算等于把排版交给常量；夹上下限防止单字标签太窄、长名字挤掉其它标签。

**代价**：每帧多一次文本测量（每标签一次，可忽略）；宽度现在依赖字体，换字体后布局会跟着变——
这正是想要的。

---

## ADR-25 关闭确认：按进程判断，三个答案不是一个

**背景**：最初关标签只看终端本身有没有结束（`pane.Tab.Running()`），停在提示符的 shell 也算“在跑”，
于是关任何一个活标签都弹「这个标签里的程序还在运行」——用户一眼就看穿：“实际上并没有程序”。
插件不给 PID（`terminal.Terminal` 没有进程句柄，`internal/pty` 是 internal 且 Windows 上返回 0），
所以“程序在跑”只能自己看进程树。

落地一版之后又发现三处说不通：macOS 没有 `/proc`，采集必然失败，退回“终端还在”于是空闲 shell 也提示；
两个标签同时启动时按“最大 Pid”认领会把两个标签认成同一个 shell；“读不到进程列表”和“确实没程序”
共用一句话，用户没法区分。另外 scoop 这类包管理器在 `PATH` 上放的是 **shim**
（`…\shims\pwsh.exe`），它再启动真正的 `…\apps\pwsh\current\pwsh.exe`，于是每个标签底下永远有一个
同名子进程，空闲标签也被判成“忙”。

**决定**：

- `internal/sys` 把判断做成纯函数：`Proc{Pid, Parent, Name, State}`、`ShellProc{Pid, Name}`、
  `ShellProc.Activity(all []Proc) Activity`，三态 `Idle` / `Busy` / `Unknown`。shell 自己的 Pid 找不到
  （已退出）或名字对不上（PID 被复用）→ `Idle`；有非僵尸、非 pty 宿主的子孙 → `Busy`；
  拿不到 shell 的 Pid → `Unknown`。
- 采集分平台，`Activity` 自己不认识平台：Windows 走 Toolhelp32，Linux 读 `/proc/<pid>/stat`，
  macOS 走 `unix.SysctlKinfoProcSlice("kern.proc.all")`（`P_stat` 映射成 Linux 风格的 `State` 字母）。
- `Spawned(before, after, parent, taken)` 在两次快照里找出“父进程是本进程、且还没被认领”的新进程，
  取**最小 Pid**；`taken` 是 `map[int]*terminal.Terminal`，多个标签同时启动时先看的先认领，
  `forget` 按终端清掉它的全部 Pid（否则 shim 的 Pid 会永久占位，将来被系统复用后新标签就认不出自己的 shell）。
- 认领时跟着“同名重执行链”往下走一层：`sys.Reexec(all, p)` 找出 p 唯一的同名子进程，认领后再看最多
  250ms（`shellSettle`），有就 `reclaim` 到子进程上。跳过 pty 的后台宿主（`conhost.exe` / `OpenConsole.exe`）：
  ConPTY 会把它和 shell 一起作为本进程的子进程拉起，pid 更大，不排除就会认错。
- 插件在 goroutine 里启动程序，`terminal.New` 返回时进程还没起来，所以等待放到后台：
  3 秒截止（`start.Add(3 * time.Second)`），10ms 起每次翻倍、上限 100ms 的自适应轮询，结果经 `App.Do` 回 UI 线程。
- `confirmClose` 三态：`running`（默认，`Activity != Idle` 才问）/ `always` / `never`；
  `UnmarshalJSON` 仍接受旧的 `true`（= `running`）与 `false`（= `never`）。`Unknown` 时说另一句话
  （“wt-go 无法判断这个标签里是否还有程序在运行”），读不到进程列表时打 `E303`，认不到 shell 打 `E304`。
- 提示点名：`ShellProc.Programs(all)` 列出让标签忙的程序名（去重、去掉 `.exe`、跳过僵尸与 pty 宿主），
  `Activity` 就是 `len(Programs) > 0`；`app` 用 `tabActivity{act, programs}` 把两者一起交给对话框，
  文案是「还在运行：PING」。用户问的是“为什么 pwsh 会提示”，点名程序比任何解释都短。

**理由**：产品要回答的是“关了会不会杀掉用户还在用的东西”，这个问题有三个答案，不是一个。
把 `Unknown` 单独拿出来，既不会把误报推给用户（假阳性），也不会在真的不知道时说“没事”（假阴性）。
纯函数化之后，macOS 的采集错误不可能再伪装成“有程序在跑”；点名程序则把“为什么”直接写在提示里。

**代价**：每次关标签枚举一次进程（几百条，微秒级）；快照拿不到 PID 时退回“总是问”（宁可多问，不可误杀）。
shell 内置的长命令（`Start-Sleep 300`）看不到——这是进程树的固有盲区（插件内部的 OSC 133 提示符状态没有导出），
只在文档里说明，不为此引入新的采集层；启动器若在 `shellSettle` 之后才启动子进程，仍会被当成 shell 本身。

---

## ADR-26 标签条自己滚动：溢出了才出现的箭头 + 视口接管滚轮

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

## ADR-27 `diag --redact`：报告要能贴出去，但不说明这台机器属于谁

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

## ADR-28 Linux 也用自绘标题栏，按钮跟随桌面布局

**背景**：ADR-11 让 Linux 把标题栏交给平台，因为 mygo 的 `TitleBarHidden` 在 X11 上只做
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

## ADR-29 字体族不是等宽就回落到 `monospace`

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

## ADR-30 设置页的控件用现成的，不手搓

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

## ADR-31 界面主题从配色方案推导

**背景**：`theme` 原先只有 `system` / `light` / `dark`，整个窗口是 mygo 的两套默认 token，与终端里选的
配色方案无关——用 Dracula 的终端配 VS Code 灰的标签栏。mygo v0.8.3 的 `ui.Theme` 有完整 token 集、
`Context.SetTheme` 可以整棵替换，参考项目里 kero 就是这么做的：拿终端方案的背景与光标推导整套 UI 颜色
（`kero/theme.go:334 ghosttyLook`）；dgopher 用 JSON 主题文件，Mayoi 用 13 个写死的主题表。
更要命的是默认配置：`theme` 为 `system` / `light` / `dark` 时 `theme.UITheme` 返回 `nil`，
`app.appearance()` 就让窗口落回 mygo 内置的 `DarkTheme`——栏 `#272729`、活动标签 `#18181a`、
终端底 `#0c0c0c`，标签与它下面的窗格之间一道色阶断层。2026-11-11 拿实机截图与 Windows Terminal 逐像素比，
还量出标签与栏同高、贴死窗口左上沿（`Radius(8,8,0,0)` 的两个圆角被窗口边缘裁掉），`+` 与 `▾` 被 `Grow(1)`
推到最右紧贴窗口按钮，`▾` 还在 `+` 前面、中间没有分隔线。

**决定**：

- `theme` 可以是 `system` / `light` / `dark`，也可以是任意方案名（内置、Ghostty 617 套、`schemes` 里自己写的）。
  方案名时整个窗口换成那个方案的颜色。
- 界面 token 不手写，从方案的背景 `bg`、前景 `fg` 推导（`internal/theme/ui.go` 的 `window()`）：
  表面 / 悬停 / 按下 / 边框 = `bg` 往 `fg` 混 16% / 22% / 28% / 32%；次要文字 38%；
  强调色取方案调色板第 5 色（各家的蓝），没有就用光标、再没有用前景；强调色上的文字按 WCAG 对比度取 `bg` 或 `fg`。
- 新增 `theme.Chrome(s)`：没选方案时窗口也走同一个 `window()`，只是底色换成 `theme.Terminal(s)` 的。
  `UITheme` 对 `""` / `system` / `light` / `dark` 仍返回 `nil`（设置页的下拉与 `mygo.Theme.SetSource` 的明暗判断
  都靠它），只是 `appearance()` 在 `nil` 时不再留空，而是取 `Chrome`。
- 明暗不由 `theme` 字段的字面值决定，由方案自己的背景亮度决定：选了 `Catppuccin Latte` 就是浅色窗口，
  终端没指定 `colorScheme` 时也跟着落到浅色底。
- 设置页的下拉列 11 个挑过的方案（`UIThemes()`）；写别的名字照样生效，只是不在下拉里。
- 标签条的几何与配色：栏高 30 DIP，标签与栏同高（都是 Windows Terminal 量出来的 30，标签不再比栏矮、
  上方没有一条空带），标签之间留 4 DIP，活动标签紧贴窗口上沿、距左边 8、四角圆润；活动标签顶边画一道
  2 DIP 的强调线
  （`t.Accent` 往文字色混 20%，两端各内缩 4 DIP，避开圆角）——它本来就是窗格底色，靠这道线才一眼分得清。
  活动标签的名字用粗体（`nameBold`，量宽度时也按粗体量，否则最宽的标签会被截成 `…`），
  与 Windows Terminal 一样一眼看出哪个在前。文字仍是系统 ClearType（放大看字母两侧有蓝/橙彩边）：
  mygo 的 Windows 文本引擎按系统设置取渲染方式，`text.SetFontRendering` 只有 Pango 引擎认，
  拿不到灰度抗锯齿，只能等上游。
  `+`、`▾`、滚动按钮、标签的 `×` 悬停时浮出 `SurfaceHover`；窗口的关闭按钮悬停用 `t.Danger`。
  `+`（两侧各一条 1px 竖分隔线，与 Windows Terminal 一样）、`▾` 移到滚动视口之外、`❯` 右边：
  标签放得下时仍紧贴最后一个标签，放不下时固定在右端
  （最初它们跟着标签一起滚动，标签一多就滚出视口，新建标签得先滚到末尾）。

**理由**：一份配色方案已经写好了「这个主题该长什么样」，界面再挑一遍颜色只会打架；推导出来的窗口与终端同源，
换方案时两边一起变——默认配置尤其如此，否则用户看到的恰好是两套颜色。挑 11 个而不是 617 个，是因为
下拉里滚 617 项没人受得了，且多数方案的背景对比度撑不起一套可用的 UI。

**代价**：推导只能保证「能看」，不保证与方案作者给的 UI 版一致（Catppuccin 官方给 UI 的表面色是另算的）；
强调色固定取调色板第 5 色，是因为几乎所有方案的光标都是灰的。`theme: light` 但 `colorScheme` 是深色方案时
窗口是深色的——栏与内容同源优先于 `theme` 字段的字面值；原生菜单与对话框仍按 `SetSource` 跟随系统，
极端组合下会一深一浅。

---

## ADR-32 Linux 用 fontconfig 规则保住文字颜色

**背景**：Linux 上（X11 与 Wayland 都会）文字整体发白、发绿，浅色底与提示框上几乎看不清。
原因是 mygo 的 Linux 文字后端先把字形按灰度光栅化，再把「通道值不一致」的字形判为彩色字形
（emoji）原样绘制（`mygo@v0.8.3 internal/text/pango_linux.go:1200` 的 `b.color` 分支）；
而 cairo 1.17.8 之前会把灰度改成次像素光栅化——只要 fontconfig 匹配到的字体带次像素顺序
（`rgba`），这次改成次像素就发生在 mygo 眼皮底下，于是每个字形都是白底绿 alpha。
`settings.ini` 的 `gtk-xft-rgba` 会被 XSETTINGS 盖掉；`xrdb` 只喂读 X 资源的 Xft 客户端；
`FONTCONFIG_FILE` 只影响新创建的配置，GTK 早就初始化完了——三条都不管用。
十个参考项目里只有 sshyard 处理了这件事（`sshyard/internal/gtkfix`），其余九个没有相关代码。

**决定**：启动时（`mygo.App.WhenReady` 的第一件事，早于第一次排版）用 purego 从已加载的库里
取 `FcConfigGetCurrent` 与 `FcConfigParseAndLoadFromMemory`，把两条 `rgba = none` 规则
（`target="pattern"` 与 `target="font"`）追加到**当前**的 fontconfig 配置
（`internal/sys/textcolor_linux.go`）。拿不到符号就记一条 E305 并继续，窗口照常开；
非 Linux 是空实现。

**理由**：改的是进程内已经生效的配置，所以不需要重启、不写用户目录、不受 XSETTINGS 影响。
只去掉字体的次像素顺序：mygo 自己要次像素的字形仍会显式设 `subpixel_order`，
所以不会把「本来就该次像素」的路径打死。

**代价**：这些字体从此只有灰阶抗锯齿，文字观感略软——这是为「文字颜色正确」付的价。
上游把 `b.color` 的判定改成「先强制灰度、再判颜色」之后（sshyard 的注释里写的是
"Drop it once MyGo forces grayscale itself"），这段可以整块删掉。

---

## ADR-33 界面图标用打进二进制的 SVG，不用字体字形

**背景**：标签条上的 `+`、`▾`、标签的 `×`、滚动标签的 `❮`/`❯`、查找栏的 `▲`/`▼`/`×`
原本都是 `myui.Text` 画的字形。字形不可靠：有些字体没有这些码位，按钮就空白；有的字体把它们
画得很宽，记号会顶出按钮的框（`internal/ui/tabs.go` 里 `windowButton` 的注释早为此写过，
最小化/最大化/还原才改用盒子画）。`+` 在不同字体里宽窄差一倍，`▾` 在等宽字体里常被拉扁。
十个参考项目里 sshyard 的做法是：把图标做成 SVG 打进二进制，用 `ui.Icon` 画（`ui.MustParseSVG`
在启动时解析，`fill` 取当前色，颜色因此跟随周围文字），一个字形都不用。

**决定**：新增 `internal/icons`——`//go:embed svg/*.svg` 带进 6 个 **Remix Icon v4.8.0** 的 SVG
（`add-line` `arrow-down-s-line` `arrow-up-s-line` `arrow-left-s-line` `arrow-right-s-line`
`close-line`），`load(name)` 解析成 `*ui.SVG`，以语义命名（`Add`/`Chevron`/`Up`/`Left`/`Right`/`Close`）。
`internal/ui` 的 `+`、`▾`、标签的 `×`、滚动按钮、查找栏三个按钮、窗口的关闭按钮都改用它。
版本钉在 **v4.8.0**：Remix Icon 从 v4.9.0（2026-01）起改用 “Remix Icon License v1.0”，
v4.8.0 是最后一个 **Apache-2.0** 版本，与 wt-go 的许可证相容；`svg/` 原样拷贝，同目录放
`LICENSE`（Apache-2.0 全文）与 `NOTICE`（出处、版本、换证说明）。

**理由**：矢量、抗锯齿、按按钮要的尺寸缩放，不受用户装的字体影响；颜色走 `TextColor`，
主题一换就跟着换，不需要第二套图。`ui.Icon`/`ui.MustParseSVG` 是 mygo 自带的，没有新增依赖，
也不需要图片文件或 webview。

**代价**：二进制多约 1 KB 的 SVG 与一张解析表；每个图标要在 `internal/icons` 里有一行变量
（加图标 = 拷一个 SVG + 加一行）。最小化/最大化/还原/关闭**四个窗口按钮仍用线画**：
它们要的是 1px 直线，与系统画法一致；24 网格的图标笔画在小尺寸上会被抗锯齿磨成灰糊、
还比旁边的 1px 横线粗一倍（先试过 `icons.Close`，截图里那一个叉明显发胖），所以关闭按钮
用 `Painter.Line` 画两道 1px 斜线（`internal/ui/tabs.go` 的 `windowButton`）。

---

## ADR-34 自定义配色按「基名 + (custom)」命名，行尾给删除

**背景**：设置页「配色方案」页的「复制为自定义」早先拿**当前方案名**当前缀
（`a.settings.ColorScheme + " (custom)"`）。当前方案可能就是一份副本，于是每复制一次多套一层：
截图里出现 `Campbell (custom)`、`Campbell (custom) (custom)` … 十来个同色方案，下拉里串成一长条；
页面又只有「选」没有「删」，只能手改 `settings.json`。

**决定**：
- 复制时先剥掉名字里历史留下的 `(custom)` / `(custom N)` 尾巴（`schemeBaseName`），再接上一次，
  重名才编号：`Campbell` → `Campbell (custom)` → `Campbell (custom 2)`。
- 列表行尾留 20px 给 `×`，只有 `schemes` 里自己写的方案有；指针停在行上或该行被选中时出现，
  隐藏时 `Opacity(0).PassThrough()` 占位，行不跳动。删除只改草稿，点「保存」才写回文件。
- 删的正是当前方案时，终端回到它复制自的基名方案；基名也没了则回默认 `Campbell`。
- 内置方案与 Ghostty 的 617 套只读：它们不是文件里的行，没有可删的东西。

**理由**：名字以「基名」为准，方案表才不会膨胀成一串同色副本；删除放在行尾、与选中态共用显隐，
既不占默认视野（列表要能一口气扫几百个名字），也不用第二套右键菜单。
`internal/ui` 只拿 `ui.Scheme{Name, Custom}`，不认哪个来自文件——`internal/app` 用
`theme.Schemes` 与 `settings.Schemes` 对比得出 `Custom`，依赖方向不变。

**代价**：老文件里的长名字不会自动改名（复制一次才缩短）；`×` 与行点击同处一行，
靠 mygo 把点击给最深的那一层；`schemes_test.go` 用 `Tester` 钉住了「隐藏时点不到、点 × 不改选中」。

---

## ADR-35 启动时发现的 shell 写进 settings.json，探测只在写第一份文件时

**背景**：最早 `sys.Discover()` 每次启动都重新探测，名字（`"PowerShell 7"`）和颜色硬编码在 `internal/sys`
与 `internal/profile` 两处，结果只在内存里，`settings.json` 的 `profiles` 一直是空的。于是改名/改色下次启动
被探测结果盖住、菜单里永远无法真正删掉一个条目、用户自己装的 shell（Scoop 之外的 `nu.exe` 之类）不会出现。
改成“每次启动都并进文件”之后又有两个新问题：一次启动就可能覆盖用户改过的命令行（只靠“只加不改”挡着），
且 `settings.json` 已是唯一来源了，再每次探测等于把“文件说了算”打折。另外用户需要两个维护动作：
机器上装了新 shell 想让它出现、设置改坏了想从头来；底栏的「打开 settings.json」在 Linux 上要靠桌面有 JSON 关联。

**决定**：

1. 探测只在**写第一份 `settings.json`** 时做：`load()` 多返回“文件原本不存在”，`newApp()` 只在这个条件下
   `found()` + `adopt` + 回写；此后 `a.found` 为空，`profile.Resolve` 全按文件。探测结果作为带
   `"discovered": true` 的普通条目写进 `profiles`，插在最后一个 `discovered` 条目之后（没有则插在最前）。
2. **只加不改**：文件里已有同 `id` 的条目原样保留，改过的名字、颜色、命令行不会被覆盖。
3. 发现只查 **PATH**（`sys.LookCommand`）与 WSL 注册表，不猜安装位置：猜错的路径比没有更糟。
   名字与颜色上移到纯包 `internal/profile` 的一张表，`sys.Shell` 只剩 `id`/`command`/`args`。
4. 「删除」对 `discovered` 条目是**置 `hidden`**：下次探测还会找到它，真删也会被加回；设置页给这些行
   「内置」徽标与「隐藏」，并加「显示隐藏的配置文件」+「显示」。程序文件不在了（卸载/挪走）则置 `hidden`。
5. 设置页加 **高级** 页（左栏在「操作」之后），改的是**文件**而不是草稿，所以没有保存、都先二次确认：
   - **重新探测**（`adoptAgain`）：把内置条目的 `command`/`args` 换成现在探测到的（这是唯一一处会覆盖
     用户改过的命令行的路径），探不到的置 `hidden`、回来的取消 `hidden`；`name`/`color`/`dir` 保留；
     新增的按 `adopt` 的位置插入；立即写文件。
   - **重置**：`sys.RemoveSettings()` → `sys.Relaunch()` → 关窗；新窗口找不到文件，写入默认值并探测，
     等于第一次启动。
   - 重启的新进程要活过旧窗口：进程本身在 `KILL_ON_JOB_CLOSE` 的作业里，所以作业加
     `JOB_OBJECT_LIMIT_BREAKAWAY_OK`，新进程用 `CREATE_BREAKAWAY_FROM_JOB` 起。
     别的子进程（panic 里的 shell）不请求 breakaway，仍随窗口结束。
   - 底栏按钮「打开 settings.json」改成「打开设置目录」，另有「设置目录」「程序目录」：打开**目录**在
     各桌面都能用，打开文件则要关联的编辑器。

**理由**：文件是唯一来源这件事要彻底；探测变成一个显式动作，用户改过的命令行不会在某次重启里被悄悄换掉；
重置是唯一的“从头来”出口，删文件最直白，不需要写一套默认值合并逻辑；`sys` 只做 OS 事实，`profile` 负责叫法。

**代价**：首次启动会写文件（配置目录不可写时记 `E205`，本次仍按内存里的列表工作）；新装的 shell 不会自己出现，
要手动点一次「重新探测」；卸载的 shell 不再自动隐藏（`diag` 里仍显示 `missing`）；自动隐藏是单向的，
装回来要在设置页点一次「显示」；`reset` 先删文件再启动，启动失败时文件已经没了（旧窗口留着，提示 `E207`），
顺序换过来则会启动一个读到旧设置的窗口，更糟；命令写成裸名字（`pwsh.exe`，靠 PATH 解析）的条目不会被判为“不见了”。

**测试**：`internal/sys/shell_windows_test.go` 用注入的 `lookPath` 钉住候选顺序、`System32\bash.exe` 不算 Bash、
WSL 项在最后；`internal/app/adopt_test.go` 钉住插入位置、只加不改、程序不见了就隐藏、第二次启动不改动；
`internal/app/removeprofile_test.go` 钉住“删除探测条目=隐藏、删除自己写的=真删、能恢复”；
`internal/app/advanced_test.go` 钉住“有文件就不探测、文件原样”、`adoptAgain` 的覆盖/隐藏/恢复/幂等、
`reset` 删文件并调用 `relaunch`；`internal/ui/advanced_test.go` 钉住两个按钮都先问再动、目录按钮的接线。

---

## ADR-36 字体下拉的族列表按需枚举，等宽判定按 `(family, size)` 记忆化

**背景**：`Run` 在首帧前调 `loadFontFamilies`：主线程枚举 261 个族，对其中 5 个等宽族各
整一遍 95 个 ASCII 字符，23–26 ms；而这份列表只服务设置页「配置文件 → 默认值」的字体下拉。
另外 `monospace()` 每次也要整一遍 95 个字符（6.7–10 ms），每开一个标签、每次应用设置都走一遍。

**决定**：

1. 枚举推迟到 `settingsState()` 里页面为 `ui.PageProfiles` 时（`ensureFontFamilies`），
   一次会话只做一次，反复进出页面不重来。
2. 等宽结论按 `(family, size)` 缓存（`App.isMonospace`），`monospace()` 保持纯函数；
   归一化（空族 → `monospace`、size ≤ 0 → 13）在缓存键上做，所以默认设置与显式的
   `"monospace"` 共用一条。

**理由**：字体族的等宽与否不随会话变化，字号影响度量——按两者缓存既够用又不用去猜哪些
设置会改变结论。实测首帧 153.5 → 124.5 ms。

**代价**：列表到达前展开下拉只看到 `monospace` 与当前值（`choice` 的既有行为），一帧后补全；
会话中新装的等宽字体要重启才进列表（缓存是有意的：按一下字号步进器不该重新枚举并逐族量一遍）。

**测试**：`internal/app/font_test.go` 的 `TestMonospaceFamilies` 走 `App`，
`TestIsMonospaceCaches` 钉住「`monospace` 与空族共用一条、换字号另算一条」。

---

## ADR-37 不做进程内 `malloc_trim`

**背景**：`[heap]` 21.9 MB 里约 7 MB 是已释放、glibc 留着复用的页。探针里用 `LD_PRELOAD`
调 `malloc_trim(0)`，3 秒空闲 RSS 从 ~154 MB 降到 ~143 MB，看起来值得在首帧后调一次。

**实测**（进程内用 purego 取 `malloc_trim` 调用）：符号能取到、返回 1（确实归还），
调用瞬间 VmRSS 154.6 → 149.0 MB，但 3 秒后回到 151 MB；改在 2 秒时调用是 145 → 147 MB。
只有周期性 trim（每 500 ms）能稳定在 143 MB，那是把复用池一直压小的结果。

**决定**：不做。一次性归还真值 3–4 MB 且不稳定；要稳定就得在 UI 循环里加定时器与跨线程的
「上次绘制时间」，为 glibc 的复用池不值得。代码已删，这条 ADR 记的是「别再试」。

**代价**：常驻 RSS 停在 ~150 MB。真需要时 `MALLOC_TRIM_THRESHOLD_` 仍可由用户设置。

**测试**：无（不留代码）。

---

## ADR-38 设置页一帧取一次快照，配色列表随 `apply` 失效

**背景**：`view()` 与 `settingsView()` 各算一次 `settingsState()`，每次含两件贵事：
`theme.Schemes()`（`uiSchemes`）与 `store.Equal` 的 `reflect.DeepEqual`。前者实测
2.19 ms / 745 KB / 39 153 次分配——设置页每帧两次。

**决定**：

1. `view()` 开头算一次，存进 `a.settingsShown`；`st.Settings` 与 `settingsView` 都用它。
2. 配色列表缓存在 `a.schemes`（nil 表示待建），`apply()` 在设置真的变化后置 nil。

**理由**：同一帧里没有别的东西会改设置，两次取值必然相同；`apply()` 是唯一「设置生效」的
出口，在那里失效最自然，也不需要观察 `Settings` 的每个字段。

**代价**：绕过 `apply()` 直接改 `a.settings` 的地方会让配色列表滞后一帧——目前没有这样的地方。

**测试**：`TestSchemeListIsBuiltOnce` 钉住「同一份设置只建一次」，`BenchmarkUISchemes`
钉住单次成本（2.19 ms / 745 KB，改动前设置页每帧付两次）。

---

## ADR-39 等 shell 出现时读内核的 children 文件，不再全量扫进程表

**背景**：`spawned` 每 10–100 ms 调一次 `sys.Processes()`（读整台机器 356 个
`/proc/<pid>/stat`），最久 3 秒；`reexec` 在 250 ms 里同样反复扫。启动后 4 秒内
`/proc/<pid>/stat` 被打开 2098 次——为了找一个刚起来的子进程。

**决定**：

1. `internal/sys` 加 `Subtree(root)`：Linux 读 `/proc/<pid>/task/<tid>/children`
   （内核为每个线程维护的子女表），只 stat 子树里的进程；Windows/macOS 读一次整张表，
   再用纯函数 `SubtreeOf` 筛。
2. `internal/app/tabs.go` 的 `spawned`（含 `before` 快照）与 `reexec` 走 `Subtree`，
   并在轮询里 `select` 上 `term.Done()`：程序起不来就不等满 3 秒。
3. 「一次问好几个标签」的 `othersActivity`/`CloseTab` 仍读一张整表。

**理由**：问题本身是「这个进程下面有什么」，而内核已经知道答案，问它比读一台机器便宜
两个数量级；跨标签的问题反过来，一张表比每个标签各读一棵树便宜（在 Windows/macOS 上
「读一棵树」就是「读整张表」）。所以两种形状各留一个入口，而不是把一种套到所有调用点。

**代价**：`children` 是 Linux 的接口，别的平台退化成整张表（与改动前同价）。进程在
「内核报道名字」与「读它的 stat」之间结束时会读失败——原有实现同样允许这种跳跃。

**测试**：`internal/sys/process_test.go` 的 `TestSubtreeOf` 钉住纯函数、未知根返回空；
`internal/sys/process_linux_test.go` 起一个真的子进程，钉住 `Subtree` 装在 Linux 上能看见它、
父 id 对得上、不存在的 pid 报错。

---

## ADR-40 性能基线（2026-11-11 单机实测，一次性记录）

**背景**：2026-11-12 做过一轮启动与内存优化（原计划文件已完成并删除），数字留在这里备查。
环境：Arch Linux + KDE Plasma（Wayland），12 CPU，amdgpu，同机对照 konsole。

**数字**：

- 常驻：RSS 约 150 MB（konsole 约 105），PSS 约 80 MB（konsole 约 40），线程 36（konsole 8–10；其中 17 个是 Go 运行时
  的，12 个是 GC worker）。
- 启动：exec → shell 可用约 150 ms（早期探针 183–190 ms；konsole 141 ms）；首帧 139 ms，分解为 GTK 初始化 40–47、
  `NewWindow` 31–34、`Theme.SetSource` 13、字体枚举 23–26、首次 `Shape` 9–10、`sys.Processes()` 5.6–6.6。
- PSS 归因：glibc `[heap]` 21.9 MB、Go arena 21.9 MB（live 1–4 MB）、二进制 11.4 MB、`libLLVM.so.23.1` 6.5、
  `libgallium` 4.1、GTK3 2.5、字体 2.3。
- `theme.Schemes()`（`uiSchemes`）单次 2.19 ms / 745 KB / 39 153 次分配，设置页每帧付两次；启动后 4 秒内
  `/proc/<pid>/stat` 被打开 2098 次。

**结果**：首帧 153.5 → 124.5 ms，常驻 RSS 不变；`/proc/<pid>/stat` 打开次数 2098 → 9。
逐条的取舍见 ADR-36（字体枚举记忆化）、ADR-37（不做 `malloc_trim`）、ADR-38（设置页一帧一次快照）、
ADR-39（等 shell 时读内核的 children）。

**代价**：这些数字绑在当时的 mygo v0.8.3 与这台机器上；换机器、换 mygo 版本要重测，不要把绝对毫秒当承诺。

**状态**：归档。方案文件已删，新的测量另开记录。

---

## ADR-41 文案交给 mygo 的 catalog，`internal/i18n` 只剩一个薄门面

**背景**：旧做法是「一条消息一个 `T("英文原句")`」加 `internal/i18n/i18n.go` 里一张 209 条的 `zhCN` map，
`Set()` 只认 `"zh-CN"`。问题有三类：英文原句既是 id 又是文案，改一个字就等于换 key；
没有参数类型与复数，`fmt.Sprintf(i18n.T(...))` 把占位符搬到调用点，`%s` 与 `%d` 写反只会在运行时现形；
第三门语言要复制一张 map，`zh-TW` 用户直接掉回英文。mygo 自带 i18n（ARB 目录、`mygo generate` 出 typed 访问器、
`mygo i18n check` 校验、`App.SetLocale` 切换并在各平台持久化），隔壁 6 个项目（Mayoi、kero、sshyard、onevra、
breeze、WinTray）里 5 个用 catalog、1 个用 struct 字面量，主流是前者。

**决定**：

1. 文案放根目录 `locales/`：`en.json`（`mygo.json` 的 `i18n.source`）与 `zh-Hans.json`，ARB 格式，
   key 是小驼峰加区域前缀（`closeTab`、`errSettingsInvalid`、`settingsFolderNote`）。
   `go tool mygo generate` 写 `locales/catalog_gen.go`（`Catalog` + `In(lang)` + 每 key 一个 typed 函数），生成物提交。
2. **一条文案一个 key**，按旧 map 一对一搬过来（192 条 + 新增的 `systemLanguage`），迁移可审计；
   旧 map 里 41 条既没有字面量也没被动作引用的废弃文案不再搬。
3. 静态调用点改成 typed 函数（`locales.CloseTab()`）：key 与参数类型在编译期检查，这是 mygo 与本仓库 6 个参考项目的正路。
   `internal/action` 是纯数据包、存不了函数，改存 key（`TitleKey`/`GroupKey`），由 `i18n.T`/`i18n.F` 在调用时解析——
   这是全仓库唯一按字符串 key 取文案的出口，`locales/locales_test.go` 的 `TestActionKeys` 盯着它。
4. `internal/i18n` 变成门面：`T`/`F`/`Has`/`Valid`/`Locale`/`Languages`/`Autonym`，走 `locales.Catalog.Format(Locale(), …)`。
   它**不 import 根包 `mygo`**（`SetLocale` 会把窗口层拖进纯函数包），语言切换由 `internal/app` 的 `locale()` 收口。
5. 语言的存法是 BCP 47：`""`（跟随系统，新的默认值）、`zh-Hans`、`en`；`store.Parse` 把旧文件的 `zh-CN`/`en-US`
   归一成 `zh-Hans`/`en`，所以升级后语言不会跳。设置页下拉＝「跟随系统」+ `i18n.Languages()` 每种语言的自称（`Autonym`）。
   手写坏的值（`en_US` 之外还有 `zh CN` 这类）不会让 `SetLocale` panic：`Valid` 挡住，退回跟随系统。
6. `mygo generate` 出来的 catalog 只带译文，中文那份也要 `@key.placeholders`：ICU 参数名两边不一致时
   `i18n check` 与 `TestPlaceholders` 都会报。

**理由**：key 与文案分开后，改文案不再改 id；占位符有名字和类型，`{count, plural, …}` 这类复数交给 ICU；
加语言＝加一个 json，`go build` 与 `go test` 不依赖 CLI（生成物提交）。

**代价**：翻译文件与生成物要一起提交，忘了 `mygo generate` 时 `catalog_gen.go` 与 json 不一致（`i18n check` 会报）；
英文不再能直接当 key，`grep` 一句英文找不到调用点，得先查 json；`internal/action` 的 key 是字符串，
拼错了只能靠 `TestActionKeys` 兜；`SetLocale` 会把选择写进用户数据目录（`locale` 文件），与 `settings.json` 有两份记录，
好在两者都由 `app.locale()` 单点写入。

**测试**：`locales/locales_test.go`（key 覆盖率、占位符声明与消息正文一致、动作 key 都在目录里、语言列表以原文打头）；
`internal/i18n/i18n_test.go`（原文回退、缺 key 回显 key、`F` 填占位符、`Valid` 的边界、`Autonym`）；
`internal/action/action_test.go` 钉住编号动作的标题与未知 id；`internal/ui/keybindings_test.go` 钉住未知动作落在 `other` 组。

---

## ADR-42 目录留空时回落到进程 cwd，而不是家目录

**背景**：Linux 上把系统的终端（`/usr/bin/rminal`）换成 wt-go 后，在文件管理器里「在此打开终端」，
第一个 shell 总落在用户家目录。实机抓取（2026-11-12）：wt-go 进程 cmdline 零参数、进程 cwd 是
`/home/<user>/Applications`，而它拉起的 `/bin/bash` 的 cwd 是 `/home/<user>`。链路是
`internal/app/app.go` 的 `openTab`：profile 的 `dir` 与 `settings.json` 的 `startDirectory` 都是 `""`
时 `Dir: ""` 交给 mygo，`plugins/terminal/terminal.go` 的 `if dir == "" { dir, _ = os.UserHomeDir() }`
把 pty 开到家里。桌面与文件管理器只能用「chdir 后 exec」表达「开在这里」，进程 cwd 是它唯一的传递通道。
`PWD` 虽然也是继承来的（实测值正是目标目录），但它由 shell 维护、可能与真实 cwd 不符，Windows 上也没有。

**决定**：

1. `internal/sys` 新增 `WorkDir()`：`os.Getwd()`，出错返回 `""`。`internal/sys` 是全部 OS 副作用的出口，
   `internal/app` 因此仍不直接 import `os`。
2. `app.openTab` 的目录按 `profile.dir` → `startDirectory` → `sys.WorkDir()` 三级回落，之后才 `sys.ExpandPath`。
3. 不读 `PWD`，也不加环境变量/CLI 参数（读 `WT_GO_START_DIR` 那种方案留给需要区分「图标启动回家目录」的场合）。
4. 不改 mygo 侧：`Options.Dir == ""` → 家目录是它的文档契约。

**理由**：exec 式启动器没有别的办法说目录；这一级补上之后，文件管理器、桌面图标、脚本里 `cd` 后启动
都自然落在启动处，与多数 Linux 终端的默认一致。前两级（profile 的 `dir`、`startDirectory`）仍优先，
配置过的人不受影响。

**代价**：点图标/按快捷键启动时也跟随会话的 cwd（实测 NDE 会话是 `/home/<user>/Applications`），
不再固定家目录；要严格区分就得让 shim 在 `exec` 前导出 `WT_GO_START_DIR="$PWD"`，那是另一套功能。
`os.Getwd()` 在 cwd 已被删除时失败，这时返回空字符串，仍由 mygo 交给家目录。

**测试**：`internal/sys/workdir_test.go` 的 `TestWorkDirFollowsChdir` 钉住「跟随 chdir 到的目录」；
macOS 上 `t.TempDir()` 是 `/var/...` 而内核报 `/private/var/...`，两侧各套一次 `filepath.EvalSymlinks` 再比。


## ADR-43 设置页画成卡片行，弹菜单画在窗口里，「关于」只读编译期信息

**背景**：设置页原来用 `myui.Form`/`myui.Field`：标签左对齐成一列、控件拉满右侧整行，
一眼看过去全是按钮，和 Windows Terminal 的设置页差得远；弹菜单交给桌面画，深色窗口旁边浮出一块
系统灰的菜单，像另一个程序的。同时项目没有「关于」，版本、许可、项目地址只能从 README 里找。

**决定**：

1. 设置页的骨架改成自己画（`internal/ui/form.go`）：`section` 是一张卡片，`field` 是「左字右控件」的一行。
   卡片面 = `Theme.Background` 向 `Theme.Text` 混 7%（`cardFace`），边线混 14%（`cardLine`）。
   **不用 `Theme.Surface`**：它是按钮与输入框的脸（背景向文字混 16%），卡片用它会把里面的控件吃掉。
2. 危险操作（「重置」）用 `dangerSection`：面是背景向 `Danger` 混 10%，边线是 `Danger` 半透明；
   按钮红底白字。红卡片放在页面最后：误点的人得先滑到底。
3. 页头由 `intro` 写这一页的名字与一句说明；页名已经在页头，卡片就不再写一遍组名，
   只有一页里真的分成几组（如「高级」的 Shell / 目录 / 重置）才给卡片标题。
4. 不用 `myui.Field` 换来左右布局的代价是标签与控件不再自动成对：每行用 `Role(RoleGroup).Label(标题)`
   把一行说成一组，读屏与 `TestAdvancedPage` 这类按名字点控件的测试都还认得。
5. 弹菜单改为窗口内自绘：`app.view` 里 `c.SetMenuPresentation(myui.PopoverMenus)`。
   标签右键、`▾`、终端右键、编辑菜单一起换过来，键盘（上下、回车、`Esc`、首字母、子菜单）由 mygo 自带，
   不必自己画一套 `menuEntry`（sshyard 那样做是因为它要图标与多行）。
6. 「关于」的数据只读编译期信息（`internal/cli/about.go` 的 `About()`：版本、提交、Go、平台），
   不跑 `Collect()`：对话框不该顺手探测一遍 shell。打开项目主页失败时记 `E209`，并把同一行显示在对话框里
   ——按了没反应的按钮比一句实话更糟。

**理由**：卡片行把「这是什么 / 现在是什么」分开摆，按钮成为控件而不是背景；配色全部从主题推导，
换方案、换明暗都跟着走，不需要为卡片另配一套颜色。弹菜单跟着窗口走，视觉上是一体的。

**代价**：`myui.Form` 的对齐、标签点击聚焦、a11y 自动关联都放弃了，靠 `Role`/`Label` 与固定的
`controlWidth`（320）兜底；卡片底色是推导值，主题方案极端（纯黑底 + 纯黑文字）时对比会不够。

**测试**：`internal/ui/advanced_test.go` 仍按按钮文字定位（页头加了标题、卡片重排都没动它）；
`internal/ui/form_test.go` 覆盖 `choice`/`count`/`measure`/`ratio` 的取值映射。
## ADR-44 默认配置文件留空＝跟随系统，不是列表第一个

**背景**：设置页「启动」页的下拉里，空值那一项叫「无」，而它的实际含义是「列表里第一个可见的
profile」（`profile.Default` 的回落循环）。于是隐藏一个 profile 就会静默改掉所有新标签跑什么，
而「无」这个词说的是「什么都不跑」——都不对。POSIX 上它凑巧等于用户自己的 shell，只因为
`sys.Discover` 把 `$SHELL` 排在第一位；Windows 上则等于 `pwsh7`（候选顺序的第一位）。

**决定**：

1. 空值＝**跟随系统**：`profile.Default(profiles, "")` 直接返回零值 profile，不再查列表。
   零值 profile 的 `Command` 为空，`CommandLine()` 据此返回 `nil`，终端自己去找程序：
   Windows 依次 `pwsh.exe` → `powershell.exe` → `%COMSPEC%` → `cmd.exe`；POSIX 用 `$SHELL`
   （空则 `/bin/sh`，macOS `/bin/zsh`）并以登录 shell（`-zsh`）起。
2. `profile.Default` 不再返回 `(Profile, bool)`：它总能给出一个 profile，零值就是「没有名字的那个」。
   名字对不上（profile 删了、被隐藏了）也回零值，而不是换列表里的另一个。
3. `store.Settings.DefaultProfile` 加 `omitempty`：「跟随系统」在 `settings.json` 里根本不出现。
4. 文案用「跟随系统」（`systemProfile`，与语言栏的 `systemLanguage` 同词）；`▾` 菜单不显示解析结果
   ——菜单是选 profile 的地方，不是解释回落链的地方。

**理由**：把「默认」从**列表里的位置**变成**明文规则**，隐藏或重排 profile 不再改变新标签的行为；
POSIX 上还能拿到登录 shell 的 `PATH`（`.zprofile` 里的东西不会莫名其妙消失）；Windows 上仍是
`pwsh` 在前，不会因为「系统默认＝cmd」这个说法把用户推回 cmd.exe。

**代价**：Windows 上只装了 `bash.exe`、没装 `pwsh`/`powershell` 的机器，默认标签从 bash 变成
`%COMSPEC%`/`cmd.exe`；反过来，期待「跟随系统＝cmd」的人要在下拉里显式选 Command Prompt。
老文件里写着 `"defaultProfile": ""` 的语义不变，写着具体 id 的照旧。
零值 profile 没有名字可显示，标签名因此多了一级回落：标题是程序路径就取文件名去 `.exe`
（`pwsh`、`cmd`、`zsh`），连标题都没有才写「跟随系统」（ADR-44 补丁，`internal/app/app.go` 的
`label`/`programName`）。

**测试**：`internal/profile/profile_test.go` 的 `TestDefault` 覆盖空、未知、隐藏三种 id 都回零值；
`TestCommandLine` 覆盖零 profile 返回 `nil`。实机（2026-11-12，Windows）：默认标签的子进程是
`pwsh.exe`，与改造前一致。

## ADR-45 标签前是一块方块，不是色点

**背景**：标签名前只有一个 8 DIP 的色点（`myui.Box(c).Size(8,8).Radius(4)`），那个尺寸在 40 DIP
的栏里几乎看不出是什么；设置页那个标签根本没有点；profile 一多，作者只能靠名字分辨。

**决定**：

1. 色点换成 16 DIP 的圆角方块（半径 0.3×边长），中间是记号：终端标签用 `icons.Terminal`
   （Remix 的 `terminal-line`），设置页标签（`tb.Page != nil`）用 `icons.Settings`（`settings-3-line`）
   加强调色底。
2. 底色是该 profile 的颜色；**记号颜色按 WCAG 对比度选**（`internal/ui/colors.go` 的 `inkOn`：
   黑色只在对比度高出白色 1.8 倍时才用）。因为 wt-go 的 shell 颜色里有浅的（cmd `#cccccc`、
   zsh `#e5c07b`、sh `#b0b0b0`、bash `#4eaa25`），照参考实现那样一律白字会糊；但深蓝
   （`#3b78ff`）上黑字也不好，所以用「明显更好才换」的门槛，而不是亮度阈值。
3. 没配颜色的 profile（跟随系统、用户自建）用 `t.Border` 底 + `t.Text` 记号：方块是安静的，
   记号是亮的。第一版给的是 `t.TextMuted` 记号，截图里 `#545454` 压在 `#414141` 上，对比度约
   1.4:1，方块看上去像一块空的灰格子。
4. 不在前面的标签只给**底色** `Alpha(0.85)`，记号不变：记号跟底色一起压暗时同样读不出来。
5. 两个 SVG 从 Remix Icon v4.8.0 取（与 jsDelivr 上 `remixicon@4.8.0` 的原文逐字节一致，分类是
   `Development/terminal-line.svg`、`System/settings-3-line.svg`），`internal/icons` 的 LICENSE/NOTICE
   不用改。

**理由**：全部用现成的 `myui.Box` + `myui.Icon`（`Center()` 居中、`Radius` 圆角、`Size` 定尺寸），
不写自绘代码；参考实现（sshyard `internal/app/style.go:50` 的 `tile`）也是同一个写法，
但 `Size(20)` 在 wt-go 32 DIP 的标签里偏高，取 18。

**代价**：每个标签宽约 10 DIP，`EqualTabs` 时名字宽度从 `tabWidth-60` 改成 `tabWidth-70`，
一屏能放下的标签少半个；`inkOn` 只是启发式，极端的自定义颜色仍可能不够对比。

**测试**：`internal/ui/colors_test.go` 的 `TestInkOn` 覆盖 8 个内置 shell 颜色；实机截图（2026-11-12）：
跟随系统的 `pwsh` 灰方块白记号、PowerShell 7 蓝方块白记号、Command Prompt 浅灰方块黑记号、设置页
强调色方块配齿轮。

---

## ADR-46 窗口开在上次的位置和大小，`startSize` 删掉

**背景**：`startSize`（列×行）与“记住窗口几何”是同一件事的两种说法：前者要求用户按格数猜一个尺寸，
后者是用户调好窗口、关掉、下次还是那样。mygo 自带 `WindowOptions.StateKey` 与 `WindowStateStore`
（关窗和退出时保存，下次创建同名 key 的窗口时还原，并检查记下的位置是否落在
连着的屏幕上），我们早已在 `WindowOptions` 里给了 `StateKey: "main"`，所以几何一直在记 —— 只是记在
mygo 自己的 `%APPDATA%\wt-go\window-state.json`，与 `settings.json` 不在同一个目录，谁也没看见它。
`startSize` 那套“先撑大 120px 量两次格数再缩回来”（ADR-18）为此背着一整台状态机与启动瞬间的跳变。

**决定**：删掉 `startSize`（字段、校验、设置页那一行、`internal/app/size.go` 整份、文案、文档），
改为 `internal/app/windowstate.go` 实现 mygo 的 `WindowStateStore`，把 `window-state.json` 写到
`settings.json` 旁边（字节由 `internal/sys` 的 `WindowStatePath`/`ReadWindowState`/`WriteWindowState` 读写）。
没有记录时窗口按 `WindowOptions` 的 960×600 居中打开。不加“是否记住”的开关。

**理由**：两个设置管同一件事，必然有一个是错的提示；用户能看见的窗口就是它下次的样子，比按格数推测
更符合直觉（sshyard 也是记几何，只是它默认关、需要开关，而我们的默认就该是记住）。文件与设置同目录，
`wt-go diag`、备份、整目录拷贝才是一份东西，而不是一半在 Roaming、一半在 Local。

**代价**：改字体或 DPI 后，记下的像素尺寸会让格数变化（不是错，但用户可能觉得“列数变了”）；
`window-state.json` 的坐标是屏幕绝对像素，手写无意义；旧版留下的 `%APPDATA%\wt-go\window-state.json`
成为孤儿文件（不读、不删，用户可以自己删）。

**测试**：`go test ./...`；实机：移到 140,160 / 900×650 后正常关窗 → `%LOCALAPPDATA%\wt-go\window-state.json`
写入 `{"main":{"x":140,...}}`，再次启动窗口就在 140,160 / 900×650；删掉该文件后回到 960×600 居中。

## ADR-47 把 Ghostty 的四个行为做成设置：滚轮倍数、最小对比度、内边距颜色、剪贴板读写

**背景**：mygo 的终端把 Ghostty 的行为翻成 Go 选项，我们只接了 11 个。`scrollMultiplier`、
`minimumContrast`、`paddingColor`、`clipboardRead`/`clipboardWrite` 一直空着，而它们各解决一个
看得见的问题：滚轮一格太慢、浅色方案里某些颜色糊成一片、全屏程序自己画的背景铺不到留白里，
以及程序能不能悄悄读你复制的东西。

**决定**：

1. 五个都做成 `settings.json` 的字段与设置页的一行：`scrollMultiplier`（0.1–10，默认 1）、
   `minimumContrast`（0–21，默认 0 即不干预）、`paddingColor`（`background`/`extend`/`always`，
   默认 `background`）、`clipboardRead`（`deny`/`allow`/`ask`，默认 `deny`）、
   `clipboardWrite`（`allow`/`deny`，默认 `allow`）。
2. 滚轮倍数同时给触摸板（`Precision`）与滚轮（`Discrete = 3×`）：mygo 的 `Discrete` 把 3 当
   「按系统设置的行数」，乘同一个数，两种输入的手感才一致。
3. 剪贴板读取默认**拒绝**（Ghostty 默认「询问」）：要问就得有地方画那句询问，默认值取安全的那一侧；
   选 `ask` 时由 mygo 在视图里画自己的询问条。
4. `paddingColor` 与两个剪贴板策略只在**建终端时**给：mygo 没有对应的 setter，所以改完是对新标签生效，
   设置页那三行的说明里写明「新标签生效」；`scrollMultiplier` 与 `minimumContrast` 有 setter，
   进 `apply()`，当场生效。

**理由**：这些键都是 Ghostty 自己的名字与范围，照搬过来，用户从 Ghostty 或 Windows Terminal
迁过来时不用重新学一套；默认值全取「与改造前行为一致」的那一侧（滚轮 1、对比度 0、留白画窗口底色、
读拒绝写允许），升级后不会有任何可见变化。

**代价**：多五个字段、五行设置，`docs/配置参考.md` 与设置页都得跟着长；`paddingColor` 的
`extend` 在程序只画了一部分背景时会显得留白不一致（这正是 Ghostty 把它做成选择而不是默认的原因）。

**测试**：`internal/store/validate_test.go` 加了五个越界用例与「旧文件缺字段取默认值」；
`internal/app/options_test.go` 覆盖三个词到枚举的映射与未知词的兜底；实机截图确认五行都在
「交互」与「默认值」页，设置页改动后已有标签也拿到新的滚轮倍数与对比度（`apply()` 里的两个 setter）。