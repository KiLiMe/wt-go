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
| ADR-42 | 目录留空时回落到进程 cwd，而不是家目录 | 已被 ADR-53 修订（Linux/macOS 部分） |
| ADR-43 | 设置页画成卡片行，弹菜单画在窗口里，「关于」只读编译期信息 | 有效 |
| ADR-44 | 默认配置文件留空＝跟随系统，不是列表第一个 | 有效 |
| ADR-45 | 标签前是一块方块，不是色点 | 有效 |
| ADR-46 | 窗口开在上次的位置和大小，`startSize` 删掉 | 有效 |
| ADR-47 | 把 Ghostty 的四个行为做成设置：滚轮倍数、最小对比度、内边距颜色、剪贴板读写 | 有效 |
| ADR-48 | 标签宽度按份额分，强调线删掉 | 有效 |
| ADR-49 | 起始目录认两个词：`current` 与 `home` | 有效 |
| ADR-50 | `-d`：这一次运行从哪开（不写文件） | 有效 |
| ADR-51 | Windows 上「没配」就是家目录 | 已被 ADR-53 取消分平台 |
| ADR-52 | 标签先说自己启动的 shell，再说 shell 的标题 | 有效 |
| ADR-53 | 起始目录不再分平台：没配就是家目录 | 有效 |
| ADR-54 | 新建标签只认 `defaultProfile`，不另设一项 | 有效 |
| ADR-55 | 「操作」页：一个动作一个键，改没改看得见 | 有效 |
| ADR-56 | 没有可用 profile 就是「跟随系统」，不在菜单里留空位 | 有效 |
| ADR-57 | ConPTY 载不动的键位，在 mygo 的编码器里换成老式字节 | 有效 |
| ADR-58 | fork mygo：一条回复一次写，ConPTY 载不动的颜色回复不发 | 已被 ADR-72/73 修订（病因、范围） |
| ADR-59 | 程序自己 fork 的剪贴板工具不归 wt-go 管，装 `xclip` 即可 | 有效 |
| ADR-60 | 设置页整页一个底色，左栏与底栏不铺 `Surface` | 有效 |
| ADR-61 | 「显示隐藏的配置文件」是页上的一行开关，默认开着，左栏始终只列没隐藏的 | 已被 ADR-71 取代 |
| ADR-62 | 「跟随系统」改成「跟随设备」：设备显式解析 | 有效 |
| ADR-63 | 关窗口前问一句：窗口带着多个标签时，×、Alt+F4、关掉最后一个标签都算 | 有效 |
| ADR-64 | 主题与终端配色分开：窗口照 `theme`，只有活动标签跟着窗格 | 有效 |
| ADR-65 | 浮层同时只有一个：开新的先关旧的，按键只问 `top()` | 有效 |
| ADR-66 | 名字写错回落到默认、未知键读入即保留 | 有效 |
| ADR-68 | 脱敏按文件系统的大小写规则匹配，WSL 列表跳过程序自用的发行版 | 有效 |
| ADR-69 | 设置页按用户心智分页：General 收拢启动与标签，Terminal & Input 收拢鼠标与缓冲 | 有效 |
| ADR-70 | 「外观与排版」的预览吸在页面顶上，不随字段滚走 | 有效 |
| ADR-71 | 隐藏跟着配置文件走：字段里设，列表照样列出并标「已隐藏」，探测到的条目只给隐藏 | 有效 |
| ADR-72 | ConPTY 剥掉 OSC 回复的头：ADR-58 的病因改写，决定不变 | 有效 |
| ADR-73 | ConPTY 一条回复都不发：ADR-58 的拆条作废，过滤面扩到全部答复 | 有效 |
| ADR-74 | 按住不放只连发「走一步」的动作：新建标签一次按下一次 | 有效 |

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
`maxTabName=172`（= `maxTabWidth` 240 减 `tabChrome` 68）之间，超出由 `Ellipsis("…")` 截断；
标签本身的宽度见 ADR-48 的分法。

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
- 与 shell **同名**的子孙不算程序：scoop 在 `PATH` 上放的是同名 shim（`…\shims\pwsh.exe` 启动
  `…\apps\pwsh\current\pwsh.exe`），跟随失败时标签会永远显示「还在运行：pwsh」。名字相同就分不出
  “启动器”与“用户又在 pwsh 里开了一个 pwsh”，宁可少问一次；那层之下的程序照旧点名（shell 是 shim、
  真 pwsh 里跑着 pi 时，问的仍是 pi）。

**理由**：产品要回答的是“关了会不会杀掉用户还在用的东西”，这个问题有三个答案，不是一个。
把 `Unknown` 单独拿出来，既不会把误报推给用户（假阳性），也不会在真的不知道时说“没事”（假阴性）。
纯函数化之后，macOS 的采集错误不可能再伪装成“有程序在跑”；点名程序则把“为什么”直接写在提示里。

**代价**：每次关标签枚举一次进程（几百条，微秒级）；快照拿不到 PID 时退回“总是问”（宁可多问，不可误杀）。
shell 内置的长命令（`Start-Sleep 300`）看不到——这是进程树的固有盲区（插件内部的 OSC 133 提示符状态没有导出），
只在文档里说明，不为此引入新的采集层；启动器若在 `shellSettle` 之后才启动子进程，仍会被当成 shell 本身
（现在会落进「同名子孙不算程序」，于是标签判为空闲——漏报一个同名程序换取不再误报整个标签）。

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
  上方没有一条空带），相邻两个标签之间不留空、画一条 1 DIP 宽 16 高的竖线（`stripDivider`，与 `+` 两侧
  同款、居中）——两个不在前面的标签底色都是那层 `t.Surface`，光留空会糊成一整条；在前的那个标签是深色圆角，
  它两侧不画线。活动标签紧贴窗口上沿、距左边 8、四角圆润，它本来就是窗格底色，靠这层底色与周围的
  表面色区分，不再另画一道强调线（Windows Terminal 也没有，量过：活动标签顶上没有更亮的线）。
  活动标签的名字用粗体（`nameBold`，量宽度时也按粗体量，否则最宽的标签会被截成 `…`），
  与 Windows Terminal 一样一眼看出哪个在前。不在前面的标签只淡图标底色，名字与 `×` 都用正文色
  （`t.Text`）：Windows Terminal 不给后面的标签压暗（量出来两边都是满亮度），压暗到 `TextMuted` 会让整条
  读成一排空格子。文字仍是系统 ClearType（放大看字母两侧有蓝/橙彩边）：
  mygo 的 Windows 文本引擎按系统设置取渲染方式，`text.SetFontRendering` 只有 Pango 引擎认，
  拿不到灰度抗锯齿，只能等上游。
  `+`、`▾`、滚动按钮、标签的 `×` 悬停时浮出 `SurfaceHover`；窗口的关闭按钮悬停用 `t.Danger`。
  `+`（两侧各一条 1px 竖分隔线，与 Windows Terminal 一样；左侧那条也算那条「在前的标签两侧不画」的线——
  `+` 紧贴着在前的标签时（`newTabPosition` 是 `last` 的常态，只剩一个标签时必然如此）它不画，
  量 WT 只剩一个标签时确实只有 `+` 与 `▾` 之间那一条）、`▾` 移到滚动视口之外、`❯` 右边：
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

> **Linux/macOS 部分已被 ADR-53 修订**：目录不再跟随进程 cwd，默认也是家目录。以下保留当时的推理。

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

**代价**：每个标签宽约 10 DIP（`tabChrome` 68 里含记号那 16），一屏能放下的标签少半个；
`inkOn` 只是启发式，极端的自定义颜色仍可能不够对比。

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

## ADR-48 标签宽度按份额分，强调线删掉

**背景**：`tabWidth` 一直是定宽——`equal` 是常量 176，`compact` 是名字宽夹 48–240、不随窗口变。
最大化后 1936 宽的窗口里三个标签只占 528，右边空 1370；十一个标签才开始滚动。Windows Terminal
量出来不是这样：标签随窗口变宽，标签多了各自变窄，窄到「记号 + 名字 + ×」这个下限就不再窄，再多就滚动。
另外活动标签顶上那道 2 DIP 强调线 WT 没有（量过 WT 活动标签顶上与周围同亮度），
它本来就是一块窗格底色，与周围的表面色已经分得开。

**决定**：

1. 删掉强调线（`tabAccent` 与 `tabLine`/`tabLineInset`），活动标签只留 `t.Background` 的深色圆角。
2. 每帧先算标签条能给标签的宽 `tabsRoom`（窗口宽减标题栏 `Left`/`Right`、`stripGutter`、拖拽区、
   `+`/`▾` 那组、窗口按钮组，滚动中再减两个 `❮`/`❯`），再按标签数与「两侧都不是在前标签」的
   竖线条数平分出每格份额 `tabsShare`。
3. 标签宽 = 份额夹在 `minTabWidth`（`tabChrome` 68 + `minTabName` 48 = 116）与 `maxTabWidth` 240 之间；
   `compact` 再夹一层「名字自己要的宽」（`measuredName + tabChrome`），`equal` 不夹，所有标签一样宽。
4. 份额低于 116 时宽度不再缩，总宽超出 `room` 由已有的 `viewport`（`ScrollHorizontal`）挡住并滚动，
   两个 `❯` 出在标签之后。
5. `maxTabName` 跟着改成 `maxTabWidth - tabChrome` = 172。

**理由**：与 Windows Terminal 对齐——用户看到的是「标签条是窗口的标签条」，不是一排固定宽度的按钮。
上限 240 是为了长名字的标签不至于霸占整条，下限 116 是「名字至少还有 48」的硬底线（再窄 `×` 与记号
就要挤掉名字）；压到下限还不放就滚动，而不是继续压——压到零宽比定位滚动难用，WT 也是这个选择。

**代价**：每帧多算一次布局（`c.Size()` 与几次加法，可忽略）；`equal` 的语义从「固定宽」变成「平分」，
旧行为（176 常量）本来就是权宜；窗口极窄时标签到 116 就滚动，能同时看到的标签比从前少。

**测试**：`internal/ui/tabs_test.go` 的 `TestTabWidths`（1 标签在 2400 宽落在 116–240；6 标签在 900 宽
严格窄于 1 标签；40 标签在 900 宽恰为 116、`MaxX > 0` 且名字没被 `…`）、`TestTabsShare`
（4 标签在前的是第 2 个、301 宽 → 75；在前的是第 1 个 → 74.75；没标签 → 0）、
`TestStripFits`（一个标签在窄条里仍与独自时同宽）。

## ADR-49 起始目录认两个词：`current` 与 `home`

**背景**：`startDirectory` 与 profile 的 `dir` 只能写路径。要「开在主目录」得写 `%USERPROFILE%`（Windows）
或 `~`（Linux/macOS）——同一个意思两套写法；而「跟随启动目录」只能靠留空，与「没配」共用一个值，
用户没法定它。ADR-42 把第三级回落定成进程 cwd，于是 Windows 上双击 exe 时 cwd 是 exe 所在目录，
「什么都没配」就等于「开在 build 目录」，与本意（跟随调用方的 chdir）只在从文件管理器启动时一致。

**决定**：

1. `internal/sys` 加 `Home()`（`os.UserHomeDir()`，失败返回 `""`）与 `StartDir(pref string) string`：
   `home` → 主目录，`current` 与 `""` → `WorkDir()`，其余原样返回。
2. `app.openTab` 在合并 profile 的 `dir` 与 `startDirectory` 之后、`sys.ExpandPath` 之前调用一次
   `sys.StartDir`，所以 profile 的 `dir` 也认这两个词。
3. 关键词判定放 `internal/sys` 而不是 `internal/store`：`store` 是纯函数包，不 import `os`（AGENTS.md）。
4. 本步不改任何现有行为：`""` 仍等于 `current`，两平台默认仍是进程 cwd。Windows 是否把 `""` 改成
   `home`，是后续 ADR 的事（它必须同时给出「显式传目录」的入口，否则桌面「在此打开」会失效）。

**理由**：路径写法已由 `sys.ExpandPath` 做到跨平台（`~`、`~/`、`~\`、`%NAME%`、`$NAME`、`${NAME}`），
缺的只是「平台中立地表达主目录」。用词而不是路径，也让将来「默认值」成为词汇的一部分，而不是又加一个
环境变量：ADR-42 否决 `PWD` 的理由（Windows 上没有、可被父进程伪造、`chdir` 后不更新会过期）一条不少。
参考同类实现（kero 的 `newSessionDir`：`current`/`project`/`home`/`custom`）也是这个词表。

**代价**：`current`、`home` 成了保留词，真叫这两个名字的相对目录要写 `./current`、`./home`；两个字段的
取值不再是单一类型（词或路径），设置页的两句说明得说清三态。

**测试**：`internal/sys/workdir_test.go` 的 `TestStartDir`：`""`/`current` 跟随 `os.Chdir`（用既有的
`resolved()` 比符号链接），`home` 等于 `os.UserHomeDir()`，`C:\code`、`/opt/src`、`./current`、`sub`
原样返回（带盘符的路径不被误认成词）。

## ADR-50 `wt-go -d <目录>`：命令行比设置更近

**背景**：ADR-49 把「开在哪」变成词，但 Windows 上「什么都没配」仍等于进程 cwd；一旦默认改成 `home`
（下一步），文件管理器与桌面快捷方式的「在此打开」就失效，而 wt-go 没有别的入口能说「开在这里」：
`internal/cli.Run` 只认 `diag`/`mcp`/`version`/`help`，其余算未知命令（退出码 2）。Windows Terminal 的
答案是 `wt -d <目录>`，`TerminalDirCommand` 那类集成也是靠参数把目录传进来。

**决定**：

1. `wt-go -d DIR`（同义 `--dir DIR`）只在命令行最前面解析：`cli.windowFlags` 取走 `-d`/`--dir` 与它的值，
   其余仍是命令；`-d` 与命令同时出现是用法错误（退出码 2），因为 `-d` 要的是窗口，命令不开窗口。
2. `cli.Run` 返回 `Result{Handled, Code, StartDir}`，`StartDir` 就是 `-d` 说的目录；`app.Run` 把它交给
   `newApp(startDir)`，存在 `App.startDir`。
3. 取值链变成 profile 的 `dir` → `App.startDirectory()`（`-d` 非空则用它，否则用 `startDirectory`）→
   `sys.StartDir`（`current`/`home`）→ `sys.ExpandPath`。`-d` 不写文件，不改设置页看到的值，只影响
   这一次运行，包括之后新开的标签。
4. 词也照收：`-d home`、`-d current` 与写在设置里同一个意思（同一个 `sys.StartDir`），路径则经
   `sys.ExpandPath` 展开 `~` 与环境变量。

**理由**：文件管理器集成靠「chdir 后 exec」传目录，Windows 上还有一类启动方式（Start 菜单、快捷方式、
`wt.exe` 转发）拿不到调用方的 cwd，只能靠参数说；参数还说清了「这一次」与「以后每次」的区别，设置里
不该为一次运行留下痕迹。优先级取「profile 的 `dir` 最高」：profile 是用户点名要的程序加上它固定的落脚点，
比一个笼统的窗口参数更具体。

**代价**：`Run` 的签名从 `(handled bool, code int)` 变成 `Result`，`internal/cli` 的测试助手跟着改；
`-d` 只在最前面生效（`wt-go diag -d x` 里那个 `-d` 会被 `diag` 当成未知开关），文档要写清。

**测试**：`internal/cli/window_test.go` 的 `TestDirFlagNamesTheStart`（`-d`/`--dir`、后一个生效、
无参数时 `StartDir` 为空）与 `TestDirFlagMustStandAlone`（缺值、空值、与 `diag` 同用都是退出码 2 且有
提示）；`internal/app/startup_test.go` 的 `TestStartDirectoryPrefersTheFlag`（`-d` 压过设置、清掉后回到设置）。

## ADR-51 Windows 上「没配」就是家目录

> **已被 ADR-53 修订**：`defaultStartDir` 不再分平台，两处都返回家目录。以下保留当时的推理。

**背景**：ADR-42 让「什么都没配」回落到进程 cwd，理由是「文件管理器只能用 chdir+exec 传目录」。这在
Linux/macOS 成立（`TerminalDirCommand`、`xdg-terminal-exec --working-directory`），在 Windows 不成立：
资源管理器双击 exe 时 cwd 是 exe 自己的文件夹，开始菜单/任务栏启动是 `%SystemRoot%\System32`；
Windows 传目录靠的是 `wt -d` 那样的参数（ADR-50 已给 wt-go 补上）。于是「跟随启动目录」在 Windows 上
跟随的是启动器：双击 `build\windows-amd64\wt-go.exe` 就落在 build 目录。Windows Terminal 的默认
（`"startingDirectory": "%USERPROFILE%"`）与此一致地选了家目录。

**决定**：

1. `sys.defaultStartDir()` 按平台分叉（`startdir_windows.go` / `startdir_other.go`）：Windows 返回
   `Home()`（拿不到主目录时才回落到 `WorkDir()`），其他平台返回 `WorkDir()`。
2. `sys.StartDir("")` 返回它；`"current"` 仍是 `WorkDir()` 且两个平台同义，所以想保留旧行为的人写
   `"startDirectory": "current"` 即可，不必换平台配置。
3. profile 的 `dir` 与命令行 `-d` 的优先级不变；`""` 仍然只是「没写」，`settings.json` 不因此改写。

**理由**：默认值应该答「用户最可能想要的地方」。Windows 上双击启动不带任何目录信息，家目录是有信息的
猜测；Linux/macOS 上 cwd 是从启动器继承的意图，换掉它才是丢信息。同一个字段两平台默认不同，但两平台
都能用词把它写成确定值（`home`/`current`），这正是 ADR-49 留出词汇的用处。

**代价**：`startDirectory` 的默认值不再是一个词，文档、设置页说明与故障排查都要讲平台差异；从旧版
升级的人会发现双击落点从 exe 目录变成家目录（想回去写 `current`，不改写设置文件，只在文档里说明）。
ADR-42 的回落理由在 Windows 上被本 ADR 取代，Linux/macOS 部分仍然有效。

**测试**：`internal/sys/workdir_test.go` 的 `TestStartDir` 里 `""` 的期望按 `runtime.GOOS` 取家目录或
chdir 后的目录（取不到家目录时断言回落到 cwd）；`current` 的断言两平台不变。

## ADR-52 标签先说自己启动的 shell，再说 shell 的标题

**背景**：ADR-44 让「跟随系统」没有 profile 名，标签名于是落到最后一级的词「跟随系统」。Windows 上看
不出来：conhost 启动时把标题设成程序路径，`programName` 取成 `pwsh`。POSIX 上 `bash -l` 默认不发
OSC 0/2，标签就一直写着「跟随系统」，而设置页里同一个词说的是「没指定 profile」，两处含义不同。

**决定**：

1. `sys.Discover()` 的第一项（POSIX 是 `$SHELL`，Windows 是 `pwsh.exe`）经 `programName` 变成
   `bash`/`pwsh`：`newApp` 存进 `App.runs`，`openTab` 只在 profile 没写命令时写进 `pane.Tab.Runs`。
2. `label` 在「标题里的程序名」之后、「跟随系统」之前加一级：名字 > 非路径标题 > profile 名 >
   标题里的程序名 > `Runs` > 「跟随系统」。
3. 「跟随系统」只剩在机器上探测不到任何 shell 时用。

**理由**：一个空白标签条和一个写着「跟随系统」的标签都不回答问题，而「这个标签跑的是什么」在启动
瞬间就有答案，只是 mygo 不把它告诉我们。这一级排在标题之后而不是之前，是因为标题是程序自己说的，
比我们按 `Discover()` 的顺序猜准。

**为什么不显式传命令**：mygo 的 `command()`（`plugins/terminal/terminal.go:769-793`）对显式命令按
`argv = 命令本身`，POSIX 上会从 `-bash`（登录 shell，读 `.bash_profile`）变成 `bash`。为一个标签名
改掉启动语义不划算。

**为什么不设 `Tab.Name` 或 `ProfileName`**：这两级在 `label` 里是常驻分支，设了之后就再也跟不了 shell
的标题（`titleFromActiveTerminal` 一并失效）。新增一级才保得住「标题一到就跟随」。

**代价**：`runsOf` 复制了 mygo 选默认 shell 的规则（取第一项），两家不一致时标签会在首个标题到达前
说错一小会儿，`runsOf` 的注释点名了这个依赖。测试：`TestLabel`（两级新用例）、`TestRunsOf`
（Windows/POSIX 路径与空列表）、`TestTabNamesTheShellItRuns`（标题前、以及有 profile 名时各叫什么）。

## ADR-53 起始目录不再分平台：没配就是家目录

**背景**：ADR-42 让「什么都没配」回落到进程 cwd，理由是 Linux/macOS 的文件管理器只能用 chdir+exec
传目录；ADR-51 又把 Windows 单独改成家目录。实机反馈（2026-11-12）：在文件管理器里打开
`build/linux-amd64/wt-go`，新标签落在 `build/linux-amd64` —— cwd 这条通道里混着启动器自己的习惯，
人无法从「点了个可执行文件」推出「终端开在它旁边」。同一个字段两套默认，还得在文档、设置页说明与
测试里各讲一遍。

**决定**：

1. `sys.defaultStartDir()` 只剩一份（`internal/sys/startdir.go`，没有 build tag）：`Home()`，拿不到
   主目录时才回落 `WorkDir()`；`startdir_windows.go` 与 `startdir_other.go` 删除。
2. `"startDirectory": ""` 与 profile 的空 `dir` 在两个平台都等于 `home`；`current` 仍然可用，
   只是要明说。
3. 「在此打开终端」改由参数表达：`wt-go -d "$PWD"`。`-d` 与保留词（`current`/`home`）的语义不变
   （ADR-49、ADR-50）。

**理由**：cwd 里没有「是谁要这个目录」的信息。资源管理器给 exe 目录、开始菜单给 `%SystemRoot%\System32`、
桌面会话给它自己被启动的地方，而 Linux 上同一个 chdir+exec 既可能来自文件管理器的「在此打开」，
也可能来自双击一个可执行文件。既然分不出，默认就答一个稳定的地方，把「在此打开」留成必须明说的意图。
家目录也与 Windows Terminal 的 `startingDirectory` 默认一致，两平台一个答案。

**代价**：把 wt-go 设成桌面终端的 Linux 用户，文件管理器的「在此打开终端」不再落在目标目录；要旧行为
就写 `"startDirectory": "current"`，或在集成处传 `-d`（`xdg-terminal-exec --working-directory` 这类
调用方本来就在传参数）。ADR-42 的 Linux/macOS 部分与 ADR-51 的第 1 条由本条取代。

**测试**：`internal/sys/workdir_test.go` 的 `TestStartDir` 去掉 `runtime.GOOS` 分支：`""` 期望家目录
（拿不到家目录时期望 `os.Chdir` 到的目录），`current` 仍是 chdir 后的目录，路径原样返回。

## ADR-54 新建标签只认 `defaultProfile`，不另设一项

**背景**：`settings.json` 写着 `"defaultProfile": "zsh"` 时，启动的首个标签是 zsh，但 `Ctrl+T` 开的标签
跑 bash。根因是 `internal/app/tabs.go` 的 `NewTab("")` 把空串一路传到 `profile.Default`，空 id 得到零值
profile 即「跟随系统」——`defaultProfile` 只被 `internal/app/app.go` 的 `openFirst()` 读过（ADR-44 之后
这里是全仓库仅有的读取处）。文案本来也不是这么写的：`locales/zh-Hans.json` 的 `defaultProfileNote`
早就说「新标签默认运行的配置文件」，代码没照它做。

**曾考虑的方案**：在「默认配置文件」下面再加一项「新建标签」，可选「默认配置文件」或某个 profile，
即 `settings.json` 多一个 `newTabProfile`，空值指向 `defaultProfile`。

**决定**：不另设。`NewTab` 自己解析：

```go
func (a *App) NewTab(profileID string) {
	a.openTab(a.tabProfile(profileID), "")
}

func (a *App) tabProfile(profileID string) profile.Profile {
	if profileID == "" {
		profileID = a.settings.DefaultProfile
	}
	return profile.Default(a.profiles, profileID)
}
```

1. 首个标签与新建标签同一个答案，与 Windows Terminal 的「默认配置文件」一致（WT 的说明是「单击 `+`
   图标或通过键入新的选项卡键绑定时打开的配置文件」）。
2. 解析放在 `NewTab` 而不是 `act()` 里：传空串的入口有三个（`internal/app/tabs.go` 的 `act`、
   `internal/ui/tabs.go` 的标签栏 `+`、`internal/ui/panes.go` 空窗格的「打开标签」），改一处即可三处一致。
3. `defaultProfile` 之外不再增字段：`store`、`locales`、设置页的「启动」页都不动。
4. 复制标签（`DuplicateTab`）与拆窗格仍按**当前标签**的 `ProfileID`，它们说的是「把眼前这个再来一份」。

**理由**：多一个开关等于把同一件事问两遍——两份设置可以互相矛盾（首标签 zsh、新标签 bash），
矛盾时用户还得自己推「为什么这一个是这个」。WT 只有一处默认，用户不需要学第二种心智模型。

**代价**：想「首标签一个 profile、新标签另一个」就做不到了。真需要时再按 ADR-44 的办法加一项，
而不是先加一个没人会用的下拉。

**测试**：`internal/app/tabs_test.go` 的 `TestTabProfile`：`defaultProfile="pwsh7"` 且不指定 → `pwsh7`；
指定 `zsh` → `zsh`（压过设置）；`defaultProfile=""` → 零值 profile（系统 shell）；id 不存在或已被隐藏
→ 零值 profile，不换成列表里另一个（ADR-44）。

## ADR-55 「操作」页：一个动作一个键，改没改看得见

**背景**：`settings.json` 的 `keybindings` 是**整张表**（写了就取代默认值，见 ADR-23 与
`docs/代码导读.md`），默认键由 `store.Default()` 按 `action.All()` 物化写入，`tab.new` 因此默认有
`ctrl+shift+t` 与 `ctrl+t` 两行。旧界面对此有两个毛病：`SetKeybinding` 只改该动作**第一条**匹配行，
录了新键后旧键还留着跑（页面上看不见）；页面也没有「哪些改过」「回到默认」的区分，18 个编号动作
（`tab.goto.1..9`、`profile.1..9`）把其他分组挤出屏幕。

**决定**：

1. **一个动作一个键**：`SetKeybinding` 把该动作的键全部换成新键（第一条原地改，其余删掉）。
   格子里列出该动作**当前的全部键**（`ctrl+shift+t · ctrl+t`，窄了打省略号、`Tooltip` 给全），
   所以「页面上看不见的别名」已经不存在；仍然只留一个键，是因为录一次只收一个键，留下来的其余键
   既不能单独删，也和按集合比较的「已修改 / 重置」对不上——等于两个按钮只管一半。要「多个键跑同一个
   动作」，就写两个动作或手改文件，界面不假装支持。
2. **「已修改」＝ 文件绑的键集合 ≠ 目录默认键集合**（不比行号、不比顺序）；窗口不认识的 id
   恒为已修改——它没有默认可回。
3. **「重置」按动作恢复目录默认键**，放在该动作原来第一条键的位置，文件其余部分与顺序不动；
   **「删除」是空集合**。两者不同不是冗余：文件整张取代默认值，删了就是真的不跑。
4. **编号动作每组只露前三项**，其余折起来（`展开其余 n 项` / `收起`）；搜索中或「仅看已修改」时
   全展开。理由：编号项是「第几个」，人一次只需要头几个，全列会把其他分组挤出屏幕。
5. **冲突只标红、不自动解**：自动解要先选「谁赢」，猜错比红字更糟。顶栏同时给一句总结，
   滚动到看不见的行也知道有冲突。

**放弃的部分**：原型里的字体缩放、`ctrl+l` 清屏、`alt+enter` 全屏等键位是生成时编的，wt-go 没有
这些动作；不为了对齐原型新增动作。界面照做，键位表只认 `internal/action`。

**测试**：`internal/ui/keybindings_test.go` 的 `TestChangedActions`（默认键不算改、另一个键算改、
未知动作算改）、`TestClashingKeys`（单次绑定的键不算冲突、空键不算冲突）、`TestListedRows`
（折叠 3 后、展开全出、仅看已修改与搜索不过滤）；`internal/ui/settings_test.go` 的
`TestActionsPageToolbar`（点「仅看已修改」真的回调、折叠条按 6 项出现、全默认时不出现「已修改」）、
`TestActionsPageNamesOnlyRealClashes`（默认键位下顶栏不出现红字，人为重复一个键才出现）、
`TestActionsPageShowsEveryKeyOfAnAction`（`tab.new` 的一格写着两个键）；
`internal/app/settings_test.go` 的 `TestResetKeybinding`（恢复默认且不动邻居）、
`TestSetKeybindingDropsTheOtherKeys`（`tab.new` 录 `f2` 后只剩 `f2`）。

## ADR-56 没有可用 profile 就是「跟随系统」，不在菜单里留空位

**背景**：`settings.json` 的 `profiles` 条目全部被删掉或标成 `hidden` 之后，运行层其实已经回到系统
shell（`profile.Default` 认不出 id 或认到 `Hidden` 就返回零值 `Profile`，`CommandLine()` 为 `nil`，见
ADR-44），但界面没说这件事：`▾` 菜单的 profile 项全被滤掉，只剩「设置」「关于」，看上去像窗口坏了；
设置页的「默认配置文件」下拉又把隐藏的 profile 照旧列出来，选了它却跑系统 shell。

**决定**：

1. **`▾` 菜单在可用 profile 为空时列一项「跟随系统」**（`ui/menuProfiles`），点了走
   `App.NewSystemTab()`（`openTab(profile.Profile{}, "")`）。理由：那一项就是窗口此时会跑的东西，
   菜单把它说出来比空着强；不为它新建 profile 条目，也不给空 id 编快捷键（它没有编号）。
2. **隐藏的 profile 不进这两处列表**（`app.knownProfiles` 跳过 `Hidden`）。理由：列出来是一个不成立的
   选项——选中也会回落系统 shell，「跟随系统」已经能表达同一件事。
3. **`defaultProfile` 指向已不存在或已隐藏的 id 时，下拉显示为「跟随系统」**，且只在用户真的选了才写进草稿
   （画一帧不改文件）。理由：保存时的死引用与「没写」在运行时是同一件事，界面上不该出现两个意思。

**否决的方案**：菜单里画一个灰掉的「没有配置文件」条目——它不能点，等于把问题推回给人；给零值 profile
新增一个 `System` 字段——零值已经是「跟随系统」，多一个字段要同步 profile 解析、序列化与合并三处。

**测试**：`internal/ui/tabs_test.go` 的 `TestMenuProfiles`（空列表得到唯一一项「跟随系统」、有 profile 时
原样返回）；`internal/app/settings_test.go` 的 `TestKnownProfilesSkipsHidden`；`internal/ui/settings_test.go` 的
`TestStartupPageShowsTheSystemForAProfileThatIsGone`（页面不画出 `gone`，也不为显示它改草稿）。
## ADR-57 ConPTY 载不动的键位，在 mygo 的编码器里换成老式字节

**背景**：在 wt-go 里跑 pi，Ctrl+C 清不掉输入框、Ctrl+U 删不掉行首、方向键与 Home 也不动；同一个 pi 在
Windows Terminal 里一切正常。用 `OnKey` 探针记录到：这些键**都到达**了 mygo 终端视图（mods 正确），
丢失发生在 `encodeKey` 之后。原因是程序（pi）先用 `CSI > 7 u` 打开了 Kitty 键盘协议，Ghostty 的编码器
随即把 Ctrl+字母编成 `CSI 99;5 u`、方向键编成 `CSI 1;1:1 A` 这类 CSI-u，而 ConPTY 的输入解析不认 CSI-u，
直接丢弃；纯文本键走文本分支，所以「打字正常」。mygo v0.9.0 只给 Ctrl+I/M/[ 补了 legacy 控制字节，
Kitty 分支与 ConPTY 判断都没动，上游 `main` 也一样（git issue #1203 修的是 AltGr）。

**决定**：兜底放进 mygo 的 fork（`KiLiMe/mygo` 的 `wtgo` 分支，见 ADR-58）：`plugins/terminal/view.go` 的
`legacyKey` 在 `(*Terminal).conpty()` 且程序开着 Kitty 协议时替掉编码器的结果——Ctrl+字母（含 Ctrl+Shift）→
`0x01..0x1a`、Shift+Tab → `\x1b[Z`、方向键/Home/End/Insert/Delete/PageUp/PageDown → `\x1b[…`，
程序设了 DECCKM 时方向键与 Home/End 改读 `\x1bO…`。wt-go 这边先前的 `OnKey` 桥（`internal/app/termkeys.go`）
已删除。

**理由**：①`OnKey` 只是把同样的字节再写一遍，却要 wt-go 自己判平台、自己维护键表，且拿不到 DECCKM——
方向键只能放着（旧写法就是这么留下的坑）；②fork 里就在编码器旁，`t.term.Mode(vt.ModeCursorKeys)` 可读、
`conpty()` 判据已有，测试能直接写；③不改协议宣告，程序不必知道自己的协议被降级。

**代价与边界**：程序以为 Kitty 开着、收到的是老式字节，所以 Shift+方向键、Ctrl+方向键这类组合会退化成无修饰的
形式（`legacyKey` 对它们返回 nil，交给编码器，仍会被 ConPTY 丢掉）；Alt 组合不动——编码器本来就发 `ESC` 前缀。

**测试**：fork 的 `TestLegacyKey`（12 组普通形式）、`TestLegacyKeyApplicationCursorKeys`（DECCKM 开后的
`\x1bO…`）、`TestLegacyKeyLeavesTheRest`（文本、release、Alt、Ctrl+Alt、Ctrl+方向键、裸 Tab 都返回 nil）。

## ADR-58 fork mygo：一条回复一次写，ConPTY 载不动的颜色回复不发

**背景**：pi 启动时问终端 18 个颜色（`OSC 10`、`OSC 11`、16 个 `OSC 4`）再问一次 DA1，并且**按读块整块
匹配**（`pi-tui` 的 `/^\x1b\](?:(1[01])|4;(\d{1,3}));([^\x07\x1b]*)(?:\x07|\x1b\\)$/`、
`/^\x1b\[\?[\d;]*c$/`）。在 wt-go 里，这一整批回复落进 pi 的输入框，屏幕上留下
`b:7676/7676/7676d` 一类残片，而同一个 pi 在 Windows Terminal 里一切正常。

**查到的事实**（本地 fork 的主机侧实验，`plugins/terminal` 的 `TestSkipReply` 与一次性探针）：
1. mygo 把那批回复交给 `Effects.WritePTY` 时是一次写，`inputQueue` 原样一次 `w.Write` 进 pty；
2. ConPTY 会解析**终端写进它输入端**的字节，只保留能翻成按键的东西：文本、控制字节、CSI；`OSC` 整段丢掉。
   实测：`S1` 到了、其后的 `\x1b]4;1;rgb:…\x07` 与 `\x1b[?62;22c` 都没到，`abc\x1b]4;1;rgb:…\x07def`
   只到 `abcdef`；`\x1b[A`、`\x03` 正常到；
3. `PSEUDOCONSOLE_PASSTHROUGH_MODE`（0x8）试过，行为不变，ConPTY 仍自己答 DA1（`\x1b[?1;0c`）；
4. 上游没修：`main` 到 `v0.9.1-0.20261113060250-614ea7a6af6e`，`plugins/terminal` 最近 12 个提交没有一条
   碰回复或是 ConPTY 输入，tag 仍停在 `v0.9.0`。

**决定**：fork `egoist/mygo` 到 `KiLiMe/mygo`，分支 `wtgo`，一处两个动作：
1. `WritePTY` 的回调里用 `splitReplies` 把一次写里的多条回复拆开，逐条 `t.in.push`；
2. 同一处用 `t.skipReply` 丢掉 ConPTY 载不动的颜色回复（`OSC 4/10/11/12`）——只在
   `Options.Conn == nil && runtime.GOOS == "windows"`（即有 ConPTY）时生效，`Conn` 场景与其它平台照旧全答。

`go.mod` 用 `replace github.com/egoist/mygo => github.com/KiLiMe/mygo v0.9.1-0.20261113062854-43b180c1f204`
（基线是 `v0.9.1-0.20261113060250-614ea7a6af6e`，比 `v0.8.3` 多 68 个提交；`wtgo` 分支 rebase 上去后是
两个提交：本条的回复拆条/跳过，加 ADR-57 的键位兜底）。

**理由**：规范上这是 pi 的健壮性缺陷（读块边界从来不是协议保证），但 wt-go 侧没有别的落点：`Options`
里没有「少答/分批答」的开关，想自己控 pty 就得用 `Options.Conn` 把 mygo 的 `internal/pty` 重做一遍
（sshyard 就是自己写了 72KB 的 vt 与 ConPTY 层）。拆条让「能送达的地方」一条一条到；跳过让「送不到的地方」
不再假装答过——pi 拿不到回复就用自己的默认配色，输入框也干净了（fork 里用真 pi 跑过：残片消失）。

**否决的方案**：① 只拆条不跳过——ConPTY 仍会把整批丢进输入框（实测 orig 的 511 字节就是漏进来的）；
② 不答 `OSC 4/10/11` 之外的也一并砍掉（DA1 仍要答，程序靠它认能力，ConPTY 自己也答）；
③ 手改本地 `pi-tui/dist` —— bun 一升级就被覆盖，且不解决问题本身；④ 用 `Options.Conn` 自管 pty ——
等于换掉 mygo 的终端。**待办**：已提交 `earendil-works/pi#12126`（2026-11-13）；上游回复否掉了「读块匹配」这个病因，见 ADR-72。

**代价**：多一个要跟着上游走的依赖。升级流程＝从新 tag 起把 `wtgo` 分支 rebase 上去，改 pseudo-version；
本地差异是 `terminal.go` 与 `view.go` 的三处（约 180 行）加四个测试文件。ConPTY 的输入解析在 Windows 上无法绕开，
所以「Windows 上程序问不到终端的配色」是这条路的固有代价，不是回归。判据收在 `(*Terminal).conpty()` 一处
（`Options.Conn == nil && runtime.GOOS == "windows"`），回复与键位两条路径共用。

**测试**：fork 里 `plugins/terminal/reply_split_test.go`（19 条拆成 19 份、单条不拆、写到一半的保持整块）与
`reply_skip_test.go`（Windows 无 `Conn` 时 `OSC 4/10/11/12` 跳过、标题与 CSI 照发、有 `Conn` 一律不跳）；
键位兜底在 `legacykey_test.go`（ADR-57）。

**后续**：`splitReplies` 已移除、`skipReply` 扩成「一条回复都不发」，见 ADR-73。

## ADR-59 程序自己 fork 的剪贴板工具不归 wt-go 管，装 `xclip` 即可

**背景**：X11（`kwin_x11`，非 Wayland、非 SSH）上，在 wt-go 的标签里跑 `pi`，选中复制后界面闪出
`Copied!`，到别处粘贴却是空的或上一次的旧内容；`xsel --clipboard --output` 同刻读到的也是旧内容。
加 `PI_CLIPBOARD_OSC52=1` 启动则复制与粘贴都正常。

**查到的事实**：
1. `xsel --clipboard --input` 是 double fork：父进程 `connect` 到 X 后 0.5 ms 就 `exit_group(0)`，
   `XSetSelectionOwner` 与后来的 `SelectionRequest` 应答都在孙进程里，与调用者无父子关系；本机观察到
   孙进程 31 s 后自行退出，也见过 0.6 s 内就没了。
2. 用探针把 `xsel` 包一层（记 argv / 字节数 / sha256 / 写入前后内容）做两次真实复制：都是 exit 0，
   一次 0.6 s 后回读为空，一次正常。退出码 0 与「能粘出新内容」没有因果关系。
3. pi 的 `dist/utils/clipboard.js` `copyToClipboard`：Linux 桌面依次试 `xclip -selection clipboard`、
   `xsel --clipboard --input`，只看 `runClipboardCommand` 的返回值（即 `code === 0`）设 `copied = true`；
   OSC 52 兜底的条件是 `isRemoteSession(env) || (!copied && headless)`，`copied` 为真时桌面**不再发
   OSC 52**。`isRemoteSession` 只认 `SSH_CONNECTION` / `SSH_CLIENT` / `MOSH_CONNECTION` /
   `PI_CLIPBOARD_OSC52 === "1"`。
4. 排除 wt-go：祖链 `wt-go → bash → …` 里连写 `xsel` 20 次、复刻 pi 的 spawn 选项（`stdio[0]=pipe`、
   `stdout/stderr=ignore`、`cwd:/`、`detached`）20 次，均 0 失败；wt-go 空闲时不持有 `CLIPBOARD`。
5. 本机 `CLIPBOARD_MANAGER` 所有者为 `None`（没有 parcellite / copyq / clipman），X11 的内容只活在
   持有者内存里，持有者一死就静默丢失。
6. **实测**：把 `xsel` 换成 `xclip`（pi 的候选表里 `xclip` 本就在前）之后问题消失，复制与粘贴都正常。

**决定**：不碰 pi，不给 pane 里的程序注入环境变量，wt-go 也不为第三方命令行工具兜底；用户侧装 `xclip`
即可。

**理由**：`xsel` 的退出码语义决定了这条路无法验证，`xclip` 在同样场景下稳定。wt-go 完全不在
`pi → 剪贴板工具 → X server` 这条链上——它既不持有也不转发那段内容，任何「拦一层」的做法都得自己去做
X11 剪贴板管理器。

**否决的方案**：① 给 pane 注入 `PI_CLIPBOARD_OSC52=1`——有效，但要为一个程序加一条 `env` 设置并长期
维护；② 改 pi 的 dist（`clipboard.js` 加一行「桌面也发 OSC 52」）——bun 一升级就覆盖，而且改第三方不是
wt-go 的事；③ wt-go 自己当 X11 `CLIPBOARD_MANAGER`（xgb + XFixes 监听所有者变更、缓存内容后接管，
约 300 行）。

**代价**：无代码改动。用户机器上多一个 `xclip`；超过 OSC 52 上限（`MAX_OSC52_ENCODED_LENGTH`
= 100 000 个 base64 字符，约 75 KB 原文）的复制仍没有兜底。

**待办**：是否把「`xsel` 退出码不可信」报给 pi 上游。

---

## ADR-60 设置页整页一个底色，左栏与底栏不铺 `Surface`

**背景**：设置页左栏铺的是 `Theme.Surface`（Campbell 下 `#252525`），右栏是窗口底色
`Theme.Background`（`#0c0c0c`），一块灰挨着一块近黑，中间一条硬边；底栏同样是 `Surface`，
于是窗口里上下各浮着一条灰带，看上去像两个窗口拼在一起。

**决定**：设置页只用一个底色（`Theme.Background`）：左栏（`nav`）与底栏（`footer`）都去掉
`Background(t.Surface)`，改用 1 DIP 的 `cardLine` 发丝线分开——左栏右缘、底栏上缘。
选中与悬停仍然是 `SurfaceHover` 的药丸，跟着指针走的反馈不变。

**理由**：当前标签在标签条里就是用 `Theme.Background` 画的（`internal/ui/tabs.go` 里活动标签那对
圆角上沿，让标签和下面的窗格连成一片），标签体再铺 `Surface` 会和标签条撞色。反过来把两半统一到
底色：卡片（`cardFace`，背景向文字混 7%）仍是唯一「抬起来」的面，层次反而更清楚；`Surface` 回到
它在 ADR-43 里的角色——按钮与输入框的脸。

**否决的方案**：把右栏改成 `Surface`。那样标签条（也是 `Surface`）与标签体同色，活动标签和栏
分不开，比原来更花。

---

## ADR-61 「显示隐藏的配置文件」是页上的一行开关，默认开着，左栏始终只列没隐藏的

**背景**：「默认值」页的「配置文件」卡片末尾有一个「显示隐藏的配置文件」勾选框，默认没勾。两个问题：
列表长的时候（这台机器上有十几个 shell）要滚到卡片底部才看得见它；而且勾选框本身就在这张列表里，
默认关着等于把隐藏条目连「显示」按钮一起藏了起来——隐藏的条目正是「重新装回来点显示就回来」的那个
入口（见 `docs/使用说明.md` 第 7.2 节）。另外左栏用的是同一份过滤列表，关掉开关时左栏也跟着少一行，
也就是「默认值」页的一个视图开关改掉了导航的样子。

**决定**：

1. 勾选框改成卡片第一行的开关，用 `field`（左字右控件，跟「窗口置顶」一样），说明一行写在名字下。
2. 默认开着：`App.showHidden` 在 `newApp` 里置 `true`。
3. 左栏（`nav`）只列没隐藏的配置文件，跳过 `Hidden` 的；这个开关只管「默认值」页那张列表。

**理由**：开关跟着它过滤的那张列表走，放在列表上方，列表再长也挤不掉它（它就在卡片开头）。默认开着是
因为它守着「找回来」的入口：关掉时的代价是看不见隐藏条目、也就看不见「显示」按钮，而左栏已经替页面
隐去了这些条目，一处边界足够。左栏与列表分家，是因为左栏是导航：一行左栏对应一个能打开的页面，把
隐藏条目列进去，就是给用户一堆「本来就打算看不见」的页面。

**否决的方案**：把开关放到左栏底部、跟「＋ 添加新配置文件」并排。左栏只有 220 DIP 宽，一句话的开关
要么折行要么缩成图标，而它管的并不是左栏。

## ADR-62 「跟随系统」改成「跟随设备」：设备显式解析

**背景**：`defaultProfile` 为空时，`profile.Default` 返回零值 profile（没有程序、没有名字、没有颜色），
把「跑什么 shell」交给终端自己找（Windows 上 `pwsh.exe` → `powershell.exe` → `%COMSPEC%` → `cmd.exe`，
POSIX 用 `$SHELL`）。结果窗口对这件事一无所知：标签没有 profile 名与颜色，只能事后从标题或
`runsOf` 猜一个程序名（截图里新标签只写 `pwsh`）；设置页也看不出接下来会开什么。另一边，wt-go
自己的探测顺序是 `pwsh7` → `powershell` → `bash` → `cmd` → `nu` → `wsl:*`，跟终端的兜底顺序
并不一致（机器上有 Git Bash、却没有 pwsh/powershell 时，两者跑的不是同一个）。

**决定**：

1. `profile.Device(found)` 就是设备：取探测到的第一个 shell 的程序与参数，名字与颜色取内置表；
   `found` 为空才回到零值 profile（终端自己找）。
2. `profile.Default(profiles, id, found)`：id 命中且未隐藏 → 那条；否则 → `Device(found)`。
   启动的首个标签、`Ctrl+T`/`Ctrl+Shift+T`、`+`、空窗口的「打开标签」、复制标签、拆窗格都走它。
3. 设备**不读** `settings.Profiles` 的任何字段：隐藏条目的名字、颜色、参数一律不借。
4. 设置页下拉第一项叫「跟随设备（`PowerShell 7`）」，括号里是设备 shell 在内置表里的名字
   （`ui.Settings.Device`）；标签名走同一张表，所以下拉、标签、profile 列表三处写的是同一个名字。
5. 设备每次启动都探一次（`newApp` 里 `shells := found()` → `a.device`），哪怕 `settings.json`
   已经在；写进文件的 `a.found` 仍只在第一份文件时赋值，两者分开：`profile.Resolve` 要用文件里
   改过的命令行，而「这台机器是什么」与文件无关。

**理由**：跟随设备是这台机器的事实，不是列表里的某个位置——「全部隐藏也照样开 pwsh」要靠解析做到，
而不是靠「碰巧列表第一项还是它」。从设置里借字段会让它变成「跟随某一条设置」：隐藏的含义是
「别在菜单里提供它」，不是「机器上没有它」，借了就等于让一个被隐藏的条目继续决定新标签的
名字、颜色与命令行。明说程序名（而不是含糊的「系统」）也让下拉与标签对得上：用户问的正是
「打开 pwsh 就显示 pwsh」。

**否决的方案**：让 `Default` 退回「列表里第一条没隐藏的 profile」。那会把「默认」变成一个可以
被隐藏操作改掉的列表位置，且读 `settings.json` 之前就已经知道设备是什么，没有理由绕这一圈。

---

## ADR-63 关窗口前问一句，按窗口问而不是按标签问

**背景**：关掉最后一个标签（`closeOnLastTab`）、标题栏的 `×`、`Alt+F4` 都会让窗口消失，把它跑着的
shell 一起结束，此前没有任何一问。「标签里没有程序在跑」恰恰是最容易顺手点成退出的状态（shell
停在提示符，看起来什么都没有）。

**决定**：

1. `settings.json` 加 `confirmCloseWindow`（布尔，默认 `true`；旧文件不写这一项，关窗口就要问了）。
2. 判定点只有一个：窗口自己的 `OnClose`（`App.onClose`）。标题栏的 `×`、`Alt+F4`、以及任何一处的
   `win.Close()` 都到这里；`PreventDefault()` 让窗口留着并弹问，答「关闭」才 `closeWindow()`
   （先置 `closing` 再关窗口，于是这一次关闭不再问第二遍）；`closeWindow()` 就是「绕过这一问」的
   唯一入口，它自己不再问。
3. 问的是「窗口带着的标签」：`asksCloseWindow()` 要求窗口里有**两个及以上**标签，所以问只从窗口
   自己的关闭传来。只开着一个标签时关窗口就是关那个标签，标签级那一问（`confirmClose`）已经管过，
   窗口这一问只会是同义反复；关掉最后一个标签时窗口里就只剩它一个，于是那一条路自然不问。
   问的时候说出会跟着走的标签数（`confirmCloseWindowText` 的 `{count}`）。
4. 设置页「启动」页的开关跟在「关最后一个标签时关窗口」下面：它管的是窗口，与这一行相邻。
5. 标签级的 `confirmClose` 仍先问：先问「标签里还有程序在跑」，答是之后才轮到窗口这一问。
6. 「删除设置并重启」不再问第二遍（见理由）。

**理由**：`OnClose` 是唯一能拦下 `×` 与 `Alt+F4` 的位置——它们在窗口里，不经过 `CloseTab`；反过来说，
拦在它之外的任何一处都拦不全。窗口这一问与标签那一问是两件事，不借用 `confirmClose` 的 `always`：
`always` 说的是「每个标签都问」，而有人想每个标签都不问、只在退出前拦一下。默认开着，是因为它挡的
是一次带走几个 shell 的、不可撤销的误操作，而代价只是在真要走时多点一下「关闭」；只开着一个标签时
不问，是因为那一问给不出选择：关窗口与关那个标签是同一件事，标签级那一问该问的已经问过。
「删除设置并重启」的关闭发生在 `sys.RemoveSettings` 与 `relaunch` 之后：这时再问，答「取消」会留下
一个设置已删、又已经起了第二个实例的窗口，所以那一路直接用 `closeWindow()`；它自己的确认框问在动手之前，
任何会让窗口消失的路径都至少问过一次。

**否决的方案**：

- 给 `confirmClose` 加第四个值（如 `window`）：三态变四态，两种问答挤进一个字段，触发点与文案都不同。
- 只在 `closeTab` 的最后标签分支拦：`×` 与 `Alt+F4` 仍会一声不响地退出，而这正是最常见的关窗口方式。
- 给 `closeTab` 的最后标签分支也加一问：能走到那里时窗口里只剩一个标签，问题给不出选择。
- 给「删除设置并重启」再叠一问窗口：见理由，答「取消」之后的状态无法收拾。
- 默认关着：那这道护栏只帮到特意打开它的人，而它挡的正是「顺手一关」这种不看设置的人常犯的错。
- 只开着一个标签时也问：这时「关掉」与「关掉那个标签」是一回事，问题给不出选择，只多一次点击。

## ADR-64 主题与终端配色分开：窗口照 `theme`，只有活动标签跟着窗格

**背景**：[ADR-31](#adr-31-界面主题从配色方案推导) 让界面从配色方案推导，解决了「默认配置下标签与终端
两套颜色」的断层，代价是把两把旋钮绑在一起：`theme` 是 `system` / `light` / `dark` 时窗口的颜色来自
`colorScheme`（`theme.Chrome`），于是「深色窗口 + 浅色终端」这样的组合做不到。

**决定**：

1. 删 `theme.Chrome`。没写方案名时窗口用程序自带的浅色 / 深色（新 `theme.Native(dark)` 拷
   `myui.DarkTheme()` / `myui.LightTheme()` 的 token），明暗由 `theme` 的字面值定：`light` / `dark`
   照字面，`system` 问 `mygo.Theme.IsDark()`。这个解析集中在新 `theme.Dark(s)`，返回 `(dark, known)`：
   `known` 为假表示没人说得准，交给 app 问 mygo。
2. 写了方案名仍走 `window()`：那是用户点名要的配色，与 ADR-31 一致。
3. 终端配色只看 `colorScheme`。`background()` 恒取 `theme.Terminal(s).Background`：窗口还没被铺上的
   地方就是终端的面，其余都会被标签栏或窗格盖住。终端没写 `colorScheme` 时的默认明暗不再单独看
   `theme` 的字面值，走同一个 `Dark()`。
4. 唯一跟着终端走的是活动标签。`ui.State` 新增 `Pane Face{Background, Foreground}`（零值表示这个窗口
   不画窗格，回落到主题自己的底色），标签栏用它在活动标签上铺一块与窗格同色的面、名字也用它的前景，
   `×` 的悬停同样从这块面派生。app 在 `apply()` 里算一次存进 `App.pane`（解析 Ghostty 主题不能每帧做）；
   哪块面不该铺窗格由 `internal/ui` 自己判断——零值，或最前面那个标签有 `Page`（设置页是窗口自己画的）。

**理由**：拆开之后窗口与终端是两件东西，两把旋钮各管各的；但活动标签与它下面的窗格必须是一块，
否则每切一次标签都在阅读区顶上看一道接缝，Windows Terminal 与 iTerm 都是这么画的。把「面」做成
`State` 的显式输入、而不是让 `internal/ui` 自己去解析配色，保持 `internal/ui` 纯绘制、
`internal/theme` 纯函数的分层。

**代价**：默认配置（`theme: dark` + `colorScheme: Campbell`）的窗口从「按 Campbell 推导的灰」变成 mygo
内置深色，不再与终端底色同源；浅色窗口配深色终端时，标签栏上那块深色面贴着浅色栏，能看见一条边界——
这是分立的必然而不是缺陷。想两边同色，得自己把 `theme` 与 `colorScheme` 设成同一套。

**否决的方案**：

- 只删 `Chrome`、让标签栏照主题画（不引入 `Pane`）：浅色窗口配深色终端时，活动标签是浅色的，
  下面的窗格是深色的，每个标签顶上都一道接缝；设置页的卡片底（`t.Background` 的混合色）也会与终端底色串色。
- 给 `settings.json` 再加一个手写 UI 颜色的对象：617 套方案之外再让人逐个填十来个 token，
  换方案就全废，与自己配色的初衷矛盾。
- 保留 `Chrome` 作默认、只给方案名放开：默认配置仍是两套颜色绑死，用户报的正是这个。

---

## ADR-65 浮层同时只有一个：开新的先关旧的，按键只问 `top()`

**背景**：命令面板、查找栏、重命名、确认框、「关于」各自一个字段，各自在入口处零散地关掉别人
（`OpenPalette` 关查找栏、`OpenAbout` 关面板、`BeginRename`/`ask` 谁也不关）。于是「先问了再开面板」
会留下两个浮层：`ui.View` 按构造顺序画，后画的盖住前一个，可前一个的 `Modal` 仍拿着窗口快捷键，
而查找栏画在模态之上、谁也不让路。进入设置页时（设置标签不是终端）查找栏也要等下一帧 `findState`
才发现没有终端。

**决定**：

1. 新增 `internal/app/overlay.go`：`overlayKind`（`overlayNone`/`overlayRename`/`overlayConfirm`/
   `overlayAbout`/`overlayPalette`/`overlayFind`）与四个入口——`top()` 说现在开的是哪个，
   `openOverlay(kind)` 先 `closeOpen()` 再开，`closeOverlay(kind)` 只在该层正显示时关，
   `closeOverlays()` 全收。所有打开/关闭都走它们，别处不再直接赋值这些字段。
2. `closeOverlay(kind)` 带条件是为了迟到的回调：确认框已经随新浮层关掉后，一个晚到的 `ConfirmYes`
   不会关掉顶上的新浮层，也不会执行旧问句的动作。
3. 按键分派唯一依据 `top()`：`bind` 在任一浮层打开时不挂窗口快捷键，`run` 再挡一次（确认框、重命名、
   「关于」期间窗口动作不生效）；查找栏只挡窗格的键，窗口的键位照旧可用。命令面板本身就是从窗口打开的，
   它自己负责关自己再 `act`。
4. 离开终端视图（`OpenSettings`）调 `closeOverlays()`，查找栏的搜索同时结束。
5. `view` 每帧开头把面板/重命名/确认/关于里 `Open == false` 的那些丢掉（`Modal` 点背景或 `Escape`
   会把 `Open` 置 false），查找栏不参与——它没有 `Open`，靠 `findState` 与 `EndFind` 管。

**代价**：`ask` 与 `OpenFind` 不能直接走 `openOverlay` 的通用分支（问句要带回调，查找栏要先确认有终端），
于是它们先调 `openOverlay` 关掉旧的、再自己赋值；`openOverlay` 里没有 `overlayConfirm` 分支，读代码时
需要知道这一点（注释里写了）。

**否决的方案**：

- 让 `internal/ui` 自己保证只画一个：它每帧拿到的 `State` 里已经有多个非 nil，挑哪一个画就把
  「谁是顶」的判断搬进了纯绘制层，而 app 侧仍能同时持有两份状态、迟到的回调照样会串。
- 用 `overlay` 接口（`overlayKind()` 方法）+ 断言来消除重复的 switch：Go 不允许给外部包的类型加方法，
  `internal/ui` 的 `Palette`/`Find` 等方法只能加在 ui 里，为一个簿记接口去改纯绘制层不划算。

---

## ADR-66 名字写错回落到默认、未知键读入即保留

**背景**：`store.Validate` 只管范围（字号、透明度、窗格数），枚举字段（`theme`、`cursorShape`、
`tabWidth`、`rightClick`、`confirmClose` …）写错时没有任何检查：窗口悄悄画成别的样子，设置页的下拉框
显示空。另一半是写文件：`Parse` 只认 `Settings` 声明的键，`Marshal` 也只写它们，于是手写的扩展键
（以及以后版本才有的键）在设置页改一项再保存时就没了。

**决定**：

1. 名字写错与数值越界**分开处理**：越界仍然整个文件被拒（`docs/配置参考.md` 第 1 节的归档流程不变），
   名字写错只把那一项换成 `Default()` 的值，见 `internal/store/validate.go` 的 `fallbackUnknown`。
   每个枚举的正规取值写成一张表（`enumCursorShapes` 等），默认值取自 `Default()`，不在两处各写一遍。
   这一步只在 `Parse` 里跑一次，窗口自己从不写出错的名字，所以保存前不必再跑，跑也不会改动任何东西。
2. `theme` 与 `colorScheme` 不能逐字对照一张表：方案名是任何名字（自定义、内置、Ghostty 617 套），
   而内置与 Ghostty 的名单在 `internal/theme` 里，`store` 不许 import 它。于是 `theme` 只认
   `system` / `light` / `dark` 与文件自己 `schemes` 里的名字，其余（包括没写，即 `""`）回到 `dark`；
   内置与 Ghostty 的方案名在 `internal/theme` 里同样会认不出而回落到默认深色（`theme.Dark` 的
   `known == false`），所以两边各自回落，结果一致。`colorScheme` 不做校验，仍然交给 `internal/theme`。
3. `language` 只校验形状：不是 BCP 47 标签（含空格、路径）的回到 `""`（跟随系统）；是标签而没有译文的
   （如 `fr-FR`）保留，界面显示原文语言——与 `i18n.Valid` 的取舍一致，也不把人写下的语言标签悄悄抹掉。
4. `Settings` 增加 `Extra map[string]json.RawMessage`（`json:"-"`），存 `Parse` 时认不出的顶层键；
   `Marshal` 先按字段写出对象，再把 `Extra` 里不与该对象冲突的键并进去，然后统一 `MarshalIndent`
   （先合并再缩进，是为了键按字母序排、文件每次写出来一样，而不是已知键在前、未知键在后）。
   冲突时以字段为准。`knownKeys` 是这些键名的唯一一份清单，`TestKnownKeysAreTheFields` 用反射把它和
   `Settings` 的 json tag 对齐，新加字段忘了登记会直接失败。
5. `Extra` 不参与 `Equal`：`dirty()` 问的是「窗口有没有没保存的设置」，而键都不是设置，谁也没改过它们。
   `Clone` 原样带过去（这个 map 不算设置，没人写它）；`Validate` 不看它。
6. 嵌套对象（`font`、`profiles[i]`、`schemes[i]`）里的未知键**不保留**：那个对象整个由窗口重写，
   把它里面的键留下来需要每个子类型都带一份 `Extra`，收益远小于代价。

**理由**：`docs/配置参考.md` 早就对 `cursorShape` 写了「认不出按 `bar`」，把这条推广到所有枚举字段，
用户看到的就是设置页显示的值；拒绝整个文件只为「窗口跑不动」的情况保留，所以范围留在 `Validate`，
名字走回落。未知键读入即保留，才能让手写的文件经过一次 GUI 保存之后仍然是原来那份文件。

**代价**：`Parse` 要 `json.Unmarshal` 两遍（一次进 `Settings`，一次进 `map`），`Marshal` 在 `Extra` 非空时
多一次 `Unmarshal` 与一次 `Indent`；文件通常只有几十 KB，这部分可以忽略。另外「名字写错就回落」意味着
写错的人不会看到任何提示——文件里的值仍在设置页底部显示的不是这个名字，只有对比 `settings.json` 才能发现。

**否决的方案**：

- 名字写错也算越界、整个文件被拒：一个人把 `tabWidth` 打成 `titleLength`，设置页就整份变成默认值并被归档，
  代价与收益不成比例；`Confirm` 已经按「旧文件写 `true`/`false` 也读得懂」的思路做了回落。
- 让 `Settings` 自己实现 `UnmarshalJSON`/`MarshalJSON` 来保留未知键：字段一多在两处都要维护，
  而且 `Confirm` 已经用着 `UnmarshalJSON`，再套一层容易出错。
- 在 `store` 里照抄一份 Ghostty 主题名来校验 `theme`：名单在 `internal/theme`（来自 mygo），两份必然走散。
- `Extra` 参与 `Equal`：那么只要文件里有一个未知键，`Parse` 之后 `dirty()` 就永远为真，设置页一打开就问
  「要保存吗」。

---

## ADR-67 保存被校验挡下时，原因留在页面上，草稿不动

**背景**：`App.SaveSettings` 校验失败时只把错误写进底栏那条 `note`（`a.setNote(err.Error())`）。
那条 note 有三处不够用：它和「已保存」「settings.json 被外部改过」共用同一个位置，下一条提示就把它顶掉；
它在底栏里 `SingleLine().MaxWidth(420)`，多条越界被 `strings.Join(bad, "; ")` 连成一行，最要紧的那条可能
正好被截掉；放弃草稿（`DiscardSettings`）会把 note 清空，于是「为什么没保存」和草稿一起消失。
另外 profile 页的命令行为空时只有占位符，没有任何文字说明「留空会跑系统 shell」，也没有提示该填什么。

**决定**：

1. `App` 增加 `validation string`，只在保存被 `Validate` 挡下时写入，在下一次保存成功、`DiscardSettings`、
   `closeSettingsTab` 时清空。它与 `note`（上一次动作的结果）、`configErr`（文件本身不是设置）各管一件事：
   `note` 会被任何动作覆盖，`validation` 在被改正或放弃前一直在。
2. `ui.Settings` 增加 `Validation` 字段。`SettingsTab` 在左栏右侧、页头上方画一条红底横幅（`Theme.Danger`
   文字、`Danger` 混 10% 的底色，与 `dangerSection` 同一套颜色），全文 `Wrap()`，不截断。
   底栏在 `Validation` 非空时显示 `saveRefusedNote`（「没有保存：改正后再存，或放弃当前修改」）并用
   `Danger` 色，而不是「有未保存的更改」——原因在横幅里，底栏回答「现在该怎么办」。
   判断顺序是 `Note` → `Validation` → `Dirty`：刚发生的动作仍然先说。
3. 校验失败**不改当前页、不改草稿、不写文件**，`dirty` 保持为真，保存/放弃按钮照常可用。
4. profile 页：非只读（`Discovered == false`）的 profile 命令行为空时，字段下方给一行内联提示
   `commandEmptyHint`（说明留空会跑系统默认 shell，并给出 `store.DefaultShellHint()` 的例子）；
   带「内置」的（`Discovered == true`）命令行为只读，说明换成 `discoveredProfileNote`（探测来的路径不在这里改）。
5. 文案只加在 `locales/en.json` 与 `locales/zh-Hans.json`，`go tool mygo generate` 重写 `catalog_gen.go`；
   `commandEmptyHint` 带占位符 `{hint}`，两边都声明 `@commandEmptyHint.placeholders`。

**理由**：校验失败是「你还得改，然后才能存」的状态，不是一次性事件，所以它需要一块不随下一次动作消失的
地方；同时它必须贴着表单，因为要改的是表单。草稿不动是原有的正确行为（`SaveSettings` 失败时直接 `return`），
这里只是让它看得见。发现来的 profile 命令行改成只读是为了让「留空跑默认 shell」这条提示只说真话：
探测来的条目由 `apply` 重新解析，页面上改它的命令行在下次启动会被覆盖。

**代价**：多一份与 `note` 并行的状态（`validation`），要在保存成功、放弃、关标签三处记得清空；
`SettingsTab` 多一层 `Column` 包住滚动区。带「内置」的 profile 在页面上不能再改命令行，要改只能改文件
或重新探测——与 `docs/使用说明.md` 7.2 里「重新探测会把这些命令行换成现在装的那份」一致。

**备选**：
- 只加长、改成多行底栏提示：底栏高度固定 48，改成多行会让每一页都矮一截，而且它离越界的字段很远。
- 把越界项按字段标红：需要把 `Validate` 的英文串解析回字段名，`store` 与 `ui` 之间多一层约定；
  横幅先解决「看得见」，按字段标红留到有需要时再做。

## ADR-68 脱敏按文件系统的大小写规则匹配，WSL 列表跳过程序自用的发行版

**背景**：`wt-go diag --redact` 原来用 `strings.ReplaceAll` 逐条替换，是字节相等比较。
Windows 上同一个目录会被写成 `C:\Users\Foo`（环境变量）与 `c:\users\foo`（某些程序报回来的路径），
后者于是原样留在报告里——脱敏最需要生效的场合正好漏掉了。另一边 WSL 的发行版列表整个从注册表
`Lxss` 读出，Docker Desktop 运行期间写入的 `docker-desktop`、`docker-desktop-data` 也在里面，
在它们上面开标签是一个没有可交互内容的容器宿主。

**决定**：
1. `internal/cli` 的 `redact` 改成 `replacePath`：`foldCase`（`runtime.GOOS == "windows" || "darwin"`）
   为真时把 needle 转小写、逐次 `strings.Index` 找并跳过 `len(needle)` 字节，只把匹配到的路径换成
   占位符，周围文本的大小写不动；为假时仍是 `ReplaceAll`。仍然不 import `os`/`path/filepath`。
2. `internal/sys` 的 `userDistribution` 按 `internalDistros`（`docker-desktop`、`docker-desktop-data`）
   过滤，用 `strings.EqualFold`（WSL 自己也不分大小写），过滤在 `discover` 里做，不只在读注册表时做。
3. `job_other.go` 的 `HoldChildren` 仍是无操作，注释写清它为什么无操作：POSIX 上没有等价的作业对象，
   进程组清理由 mygo 关终端时完成，剩下的缺口（本进程未运行清理就死掉）写在这里，不假装它不存在。

**理由**：文件名不分大小写的系统上，两个只差大小写的路径就是同一个文件，替换其中一个是为另一个
编造一个不存在的路径；反过来在 Linux 上做不敏感匹配会把真实路径改错。过滤放在 `discover` 而不是
注册表读取处，是因为它是「哪些发行版是给人用的」的判断，不是「怎么读注册表」的细节。

**代价**：Windows/macOS 上 `replacePath` 每个匹配位置多一次 `strings.ToLower` 的扫描，报告只有几十条
路径，代价可忽略。`foldCase` 是包级变量，测试用 `withFold` 临时改它；按真实 GOOS 运行时它不会变。

## ADR-69 设置页按用户心智分页：General 收拢启动、标签页与关闭，Terminal & Input 收拢鼠标、剪贴板与缓冲

**背景**：原来的设置左栏把「启动」「交互」「外观」平铺，「关闭确认」与「每标签窗格上限」在「交互」页，
「新标签位置」「标签宽度」「窗口置顶」「标题跟随窗格」在只画主题与透明度的「外观」页，而「回滚缓冲」在
「默认值」页——同一件事（标签页怎么开、窗口怎么关）散在三个页面上，找到它得先猜它被归给了哪一页。

**决定**：
1. 页面 id `general` 取代 `startup`：`internal/ui` 的 `PageGeneral = "general"`，`PageStartup` 保留为它的
   别名常量（同一个值，不是两个分支），`page`/`pageIntro` 的 `switch` 只认 `PageGeneral`，
   `internal/app.OpenSettings` 的默认页改为 `ui.PageGeneral`。别名而不是双分支，是因为两个名字指的是同一张
   页面，留一份 `case` 就不会出现只改一处的分裂。
2. 「常规与启动」分三张卡片：**语言与启动**（`DefaultProfile`、`Language`、`StartDirectory`）、
   **窗口与标签页**（`NewTabPosition`、`TabWidth`、`TitleFromActiveTerminal`、`AlwaysOnTop`、`MaxPanes`）、
   **关闭行为**（`CloseOnLastTab`、`ConfirmClose`、`ConfirmCloseWindow`）。`ConfirmCloseWindow` 仍紧跟
   `CloseOnLastTab`：后者关掉时前者没有窗口关闭可问。
3. 「终端与交互」分三张卡片：**鼠标与滚动**（`CopyOnSelect`、`RightClick`、`ScrollMultiplier`）、
   **剪贴板与安全**（`ClipboardRead`、`ClipboardWrite`、`PasteMultilineConfirm`）、
   **缓冲区**（`ScrollbackMB`）。`ScrollbackMB` 从「默认值」页移出：它是终端的会话设置，与该页的字体、
   光标不同——后者是 profile 的默认值，前者不问 profile。
4. 分组标题与页名只加在 `locales/en.json` / `locales/zh-Hans.json`（`general`、`generalNote`、
   `generalStartup`、`generalWindow`、`generalClose`、`interactionMouse`、`interactionClipboard`、
   `interactionBuffer`），`interaction` 的文案改成「Terminal & Input / 终端与交互」。

**理由**：左栏的一页应当回答一个问句。「窗口怎么开、标签怎么摆、关之前问不问」是同一个问句；
「终端怎么答鼠标与剪贴板、留多少历史」是另一个。按归属摆位之后，`NewTabPosition` 与它管的那排标签在同一页，
`ConfirmCloseWindow` 与它跟随的那只开关相邻，两个「关闭前问一句」也终于挨在一起。

**代价**：`settings_test.go` 与 `internal/app/settings_test.go` 里 `PageStartup` 的引用改成 `PageGeneral`；
`TestGeneralPageAsksBeforeTheWindowCloses` 的两只开关从同一张卡片挪到了第一张卡片与第三张卡片，
900×700 的测试窗口里第三张卡片要滚动才进入视野，所以这条测试先断言「提问在关标签开关之下」，
再 `tt.Scroll` 到底后取开关坐标点击——与 `TestNavAddsProfiles` 已经用的做法一致。
## ADR-70 「外观与排版」的预览吸在页面顶上，不随字段滚走

**背景**：ADR-69 把主题、透明度、字体排版、内边距、光标、对比度与配色方案收进同一页，顶上那块实时预览
就是这一页的意义——改哪一项都要看着它。但它在滚动的页面里：滚到「字体」「最小对比度」这些排得最靠后的
字段时，预览已经出了视口，改一行往上滚一次，等于没有预览。

**决定**：预览那一节用 mygo 的 `Sticky().Top(0)` 吸在滚动区顶上（`internal/ui/form.go` 的
`pinnedSection`），往下滚时字段从它下面穿过。`internal/ui/settings.go` 的 `appearancePage` 用它画
`locales.AppearancePreview()` 那一节；`PageSchemes`（旧的「配色方案」页名）走同一个函数，行为一样。

**吸顶的那一条要自己铺页面底色**：`Sticky` 的元素画在同级之上，而卡片与它的小标题之间是透明的，
不铺底色就会看到滚过去的字从缝里透出来。横条下缘留 12px 底色、上缘留 10px，再加一条 `cardLine`
发丝线：上缘那 10px 是吸到页顶时标题与窗格边缘之间的距离，下缘那条线把「压在上面的一层」与
「从下面滚过的一层」分开。

**为什么不做成滚动区之外的常驻页头**：那样得连页面标题与说明一起固定，否则静止时「预览」会跑到
「外观与排版」标题上面；而固定页头会长期占掉约 200px，比吸顶贵得多。吸顶在页面顶部时只多约 20px 留白，
往下滚才有区别，也就没有为此加 `settings.json` 开关的必要。

**代价**：窗口很矮时预览一直占着约 130–150px，字段区仍可滚，但一屏看到的行数变少；滚动时预览下沿
会切掉半行字段——这是吸顶的常态，用不着为此改布局。

## ADR-71 隐藏跟着配置文件走：字段里设，列表照样列出并标「已隐藏」，探测到的条目只给隐藏

**背景**：ADR-61 把「显示隐藏的配置文件」做成「配置文件」页卡片第一行的开关，默认开着，列表因此列全部条目。
这个开关放在外面有三个问题：它管的是**一条**配置文件的属性，却摆在页面顶上；它是页面记住的视图开关，不进
`settings.json`，关掉时隐藏条目与它的回程入口一起消失；ADR-61 里说的「每条带『显示』」在实际界面上早已
没有按钮，回来的路只剩「点开条目、关掉开关」。

**决定**：

1. 删掉那个开关（`Actions.ShowHiddenProfiles`、`Settings.Hidden`、`App.showHidden`、`fileProfiles` 的
   过滤参数一起去掉）：配置文件列表**永远列出全部**条目。
2. 隐藏只在页内两处设：右侧字段里的「隐藏」开关（草稿里的一项，保存生效），与卡片底部的「隐藏」按钮
   （`Actions.HideProfile`，同样写草稿：`a.settings` 里那条置 `hidden`，正在编辑的副本同步置上，于是
   「保存」可以做）。字段与按钮指同一件事，按钮只是少一次滚动；正在看的条目留在页上，不跳走。
3. 列表里隐藏的条目仍是普通行：名字变灰，后面加「已隐藏」徽标（新 key `hiddenProfile`，与「内置」同款
   灰底小字），并且仍可点开——点开在字段里关掉开关就回来了。
4. 卡片底的按钮按条目来源分：探测到的（`discovered`）只给「复制」「隐藏」，`RemoveProfile` 对它的置
   `hidden` 分支留作兜底（它仍然防住「删掉一条下次又会被找到的条目」）；自己写的给「复制」「隐藏」「删除」。
5. 删掉 `Actions.ShowProfile`：ADR-61 之后它已经没有界面调用点，只剩测试在用。

**理由**：隐藏是这条 profile 的属性，不是页面的一种看法，所以它该跟其它字段在一起、跟着保存走。列表不能
藏起隐藏项，因为列表是这个页面唯一能把它找回来的地方：留着它、标出来，「怎么回来」就在眼前一步。

**代价**：`locales` 少 `showHiddenProfiles` / `showHiddenProfilesNote`、多 `hiddenProfile`，要跑
`go tool mygo generate` 与 `go tool mygo i18n check`；`internal/app/removeprofile_test.go` 里关于
`showHidden` 的断言换成 `HideProfile` 的，`TestProfileListLeavesHiddenProfilesOut` 改名为
`TestProfileListKeepsHiddenProfiles`、断言反过来。列表里混着隐藏项之后，「在不在列表里」不再等于「能不能
开新标签」，这件事由徽标与本条 ADR 说。

## ADR-72 ConPTY 剥掉 OSC 回复的头：ADR-58 的病因改写，决定不变

**背景**：ADR-58 把病因写成「pi 按读块整块匹配」（`pi-tui` 的两条正则对每个读块整体匹配），据此定性为
上游的健壮性缺陷，待办是「把 ConPTY 丢 OSC、pi 读块匹配写成 upstream issue」。issue 已提交：
`earendil-works/pi#12126`（2026-11-13，标签 `no-action`，自动关闭后由维护者回复）。

**上游的结论**（维护者 davidbrai 的回复，注明 AI 起草、经人审）：

1. **不是读块问题**。pi 的 `StdinBuffer` 按 BEL/ST/CSI 终止符把一批读拆成单条回复；当前 `main` 上一次写
   含 `OSC 10`、`OSC 11`、16 条 `OSC 4` 加 DA1 也能干净消费，同一条回复被切成两次读同样干净，编辑器
   两种情况都保持空。
2. **真因是头被剥掉**。漏进输入框的 `b:7676/7676/7676…`、`gb:e7e7/…` 从回复**中间**开始、没有
   `ESC ]4;N;` 头——Windows 的 conhost（内置或较旧的捆绑版）剥掉或切断了 OSC 回复的头，尾巴作为普通
   文本进 pi。由 `#10256`（open，25 条评论）跟踪；pi 侧「丢掉无头尾巴」的缓解在 `#11642`（open，
   零评论，未合）。
3. **出路**：用 Windows Terminal，或自带当前 conhost/OpenConsole 的终端；模拟器自带
   `conpty.dll`/`OpenConsole.exe` 而不走系统 conhost 也能避开（WezTerm nightly 在 #10256 里就是这样好的）。

**对决定的影响**：没有。ADR-58 的两处改动分开看——`splitReplies`（拆条）只是让「送得到的」一条一条到；
`skipReply`（`Options.Conn == nil` 且 Windows 时不答 `OSC 4/10/11/12`）才是关键：既然 conhost 会剥头，
分条也救不回尾巴，只有「不发」能让输入框干净。ADR-58 的「代价」段照旧成立——Windows 上程序问不到终端
配色是这条路的固有代价；`PSEUDOCONSOLE_PASSTHROUGH_MODE`（0x8）无效的实测也与上游结论相容。

**保留的分歧**：ADR-58 把这件事定性为 pi 的健壮性缺陷，上游不接受，他们的落点是 pi 侧丢弃无头尾巴
（#11642）。#11642 合进去之后残片在 pi 侧也会消失，但 `skipReply` 仍要留——它防的不是残片，是
「答了但送不到」的假装回复。

**留待评估**：自带 `conpty.dll`/`OpenConsole.exe` 是 ADR-58 从未评估过的分支（能让 Windows 上的程序
重新问到配色），记在 `docs/plan.md` 的「待评估」，不与本条一起决定。

**后续**：这里留着的两处过滤已并成一个入口、拆条一并删掉，见 ADR-73。

## ADR-73 ConPTY 一条回复都不发：ADR-58 的拆条作废，过滤面扩到全部答复

**背景**：ADR-58 的决定有两个假设没验过：一是「拆成逐条写」对送不到的答复无害也有益，二是
`skipReply` 只挡 `OSC 4/10/11/12` 就够——注释里写的是「ConPTY 保留 CSI」。

**实测**（同一个子进程用 `ENABLE_VIRTUAL_TERMINAL_INPUT` 进 raw 模式，只换 ConPTY 实现；子进程问
`OSC 10`、`OSC 11`、16 条 `OSC 4` 加 DA1，父进程把答复写进 pty 输入端）：

| ConPTY | 一次写 19 条 | 逐条写 | 答复里夹纯文本 |
| --- | --- | --- | --- |
| 系统 conhost（`kernel32.CreatePseudoConsole`） | 子进程收到 7 字节，全是 conhost 自答的 `ESC[?1;0c` | 同样 7 字节 | 还是 7 字节，`A`/`E` 也没到 |
| 自带 `conpty.dll` 1.25 + `OpenConsole.exe` | 454 字节，18 条完整 | 同样 454 字节 | 37 字节，逐字节完整 |

两个假设都不成立：①拆条对系统 conhost **毫无影响**（一次写与逐条写都是 0 字节到达）；②CSI 答复同样被吃
——单独写 `ESC[?62;22c` 时子进程只收到前后的 `Z`、`Y`，应答没了。上游也印证了后者：
`earendil-works/pi#10717` 里 Windows + WezTerm 漏出的碎片是 `ESC[6;21;10t`，那是 CSI 的窗口报告。

**决定**：把「不发」收成一个入口 `(*Terminal).writeReply`。`conpty()` 为真时**什么都不写**（`OSC`、
`CSI`、`DCS` 一体对待），否则整批一次 `push`。三处都走它：
1. `WritePTY` 回调——终端对程序答复的常规路径；
2. `schemeChanged` 的 `?997;n`——此前绕过过滤直接 `t.in.push`；
3. `sendReply`（剪贴板的整段答复，走 `replying` 标记）——`conpty()` 时直接返回。

`splitReplies`、`replyLen`、`skipReply` 一并删除：上游 `main` 本来就成批写，pi 也早已能整批消费
（ADR-72 的上游结论），拆条既无用又多一处与上游的分歧。

**理由**：①ConPTY 保留哪些字节不可预测——同一实现下连普通文本都不一定到，只有「不发」是确定的；
②过滤面按「是不是终端→程序的答复」划，比按具体操作数划好维护，也不会漏掉下一个漏进输入框的字节；
③fork 因此更薄：`wtgo` 分支只剩两件事——键位降级（ADR-57）与不写回复。

**代价**：与 ADR-58 相同。Windows 上程序问不到终端配色与深浅（`OSC 4/10/11/12`、`?996n`），读剪贴板
（OSC 52 read、Kitty 的 5522）也拿不到答复；macOS/Linux 与 `Options.Conn`（程序在另一台机器）不受影响。
判据仍收在 `(*Terminal).conpty()` 一处。

**否决的方案**：①保留拆条——实测对系统 conhost 零效果；②只堵 `OSC` 不堵 `CSI`——DA1、DSR 同样到不了，
且「CSI 能过」这个前提已被实测推翻；③探测程序是否 VT 输入模式——那是 hyla 自带 ConPTY 的前提，不是
「发不发」的前提。

**测试**：fork 的 `plugins/terminal/reply_conpty_test.go`——`TestWriteReply` 在 `conpty()` 下队列为空、
有 `Conn` 时整批一次写入（非 Windows 跳过）。


## ADR-74 按住不放只连发「走一步」的动作：新建标签一次按下一次

**背景**：按住 `Ctrl+T` 不放，几秒内开出几十个标签，每个标签各起一个 shell 加一个 ConPTY，窗口卡住。

**查到的事实**：绑定这条路的代码从不看「这是不是按住不放的连发」——`internal/app/app.go` 的 `bind` 把每个
动作原样挂到 `c.OnShortcut`。mygo 一直给得出这个信息：Windows 后端按 `lp&(1<<30)` 判定重复并填进
`platform.SurfaceEvent.Repeat`（`internal/windows/surface.go`），UIKit 用 `Context.ShortcutRepeated()`
暴露它，文档里的例子也正是「开关只认按下、走步每一下都认」。

**决定**：动作目录加一个字段 `Action.Repeats`，只有「走一步」的六个打开它——`tab.next`、`tab.prev`、
`pane.focus.left/right/up/down`。其余动作（新建/关闭/复制标签、切分、关窗格、查找、命令面板、设置、
关于、`tab.goto.N`、`profile.N`）按住只执行第一次。实现上 `binding` 带一个 `repeats`，`bind` 在
`c.ShortcutRepeated() && !b.repeats` 时直接返回；查表仍只有一份，在 `internal/action`。

**理由**：①系统的连发速度与节奏是键盘的事，不是用户的意图，一个动作按一次就该发生一次；②「这个动作
值得连发吗」是动作自己的性质，放在动作目录里，菜单、命令面板与设置页读的仍是同一份；③不写死在 `app`，
以后加动作不会漏。

**代价**：按住 `Ctrl+Shift+T` 也不再连发——想开多个标签得一次次按，这是要的。文件里给不认识的 id
（`GroupOther`）绑的键一律不连发。

**测试**：`internal/action/action_test.go` 的 `TestRepeats`（连发集合正好是那六个，未知 id 为假）与
`internal/app/bind_test.go` 的 `TestHeldKeys`（`store.DefaultKeybindings()` 编成的键位里，默认键绑到的
动作各自该不该连发）。
