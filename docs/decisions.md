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

**代价**：名字写错静默回退到 `theme` 的默认配色（不报错、不崩溃），排错要看 `docs/settings.md` 的顺序说明。

---

## ADR-011 标签栏兼作标题栏

**决定**：窗口用 `mygo.TitleBarHidden` + `TitleBarHeight = ui.TitleBarHeight`（36dp），
标签栏自己就是拖窗区域（空白处 `DragWindow`），不再单独画 40dp 标题栏。

**理由**：与 Windows Terminal 一致（用户提供的参考图即如此），省掉一行高度；
`mygo` 的 `DragWindow` 放在容器上不影响子元素点击（`ui/input.go:481`）。

**代价**：标签栏右侧必须留空白给拖窗，标签放不下时没有横向滚动（已知退化，待做文本测量）。

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

## 待定

- 会话常驻：阶段 5 评估是否采用 ADR-009 所述的会话服务模型。
- 验证 mygo `plugins/terminal` 是否支持外部字节流作为终端连接（gorex 即如此接入）。注意：gorex 依赖的是 mygo v0.3.5 伪版本，wt-go 锁定 v0.8.3，API 未核对，需阶段 0 实测。
- 多窗口：独立进程还是单进程多窗？前者简单，后者才能共享设置热重载。
- 是否内置 SSH（参考 sshterm 可直接拿）：暂定不做，外部调用 `ssh` 即可。
- 控制接口（HTTP/MCP）供自动化测试：阶段 3 之后仍未决定。

