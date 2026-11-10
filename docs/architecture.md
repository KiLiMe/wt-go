# 架构

## 技术栈

| 层 | 选择 | 理由 |
| --- | --- | --- |
| UI 框架 | mygo v0.8.3 | Go 原生窗口，无 webview；sshterm 已验证 |
| 终端内核 | mygo `plugins/terminal`（libghostty-vt） | 完整的 VT 实现；ConPTY 管理；选中/搜索/链接/剪贴板协议齐全 |
| 语言 | Go 1.27 | 与 sshterm 同栈，便于复用经验 |

**不自己实现 VT 解析器**：终端模拟器的正确性坑极多（宽字符、reflow、键盘协议、
OSC 52……），libghostty-vt 已经做过，自研等于重造轮子还会退步。

## 目录

```
wt-go/
  cmd/wtgo/            程序入口：调用 app.Run()
  internal/app/        应用状态与动作的唯一持有者：开/关标签、切分、写设置；
                       唯一可以 import internal/sys 的包
  internal/ui/         纯绘制：标签栏（兼作标题栏）、终端挂载点、设置页、对话框；
                       只读 ui.State，动作经 ui.Actions 回传
  internal/theme/      配色方案解析：设置里写的方案 > 内置 > Ghostty 的 617 套（纯函数）
  internal/pane/       窗格树（Tab / Node）：切分、焦点、关闭补位与几何；
                       持有 terminal 句柄，但不关闭进程
  internal/layout/     窗格树的几何计算（矩形与分割线；不依赖任何内部包，不认识终端）
  internal/profile/    profile 模型，以及从已读入的数据解析出 profile 列表（纯函数）
  internal/store/      settings.json 的解析、序列化与默认值（纯函数，字节由 app 读写）
  internal/sys/        全部 OS 副作用的适配：文件读写与监听、注册表读取、Job Object
  internal/i18n/       界面语言（先做 zh/en）
```

## 依赖方向（硬规则）

```
cmd/wtgo → internal/app
internal/app → { internal/ui, internal/pane, internal/profile, internal/store, internal/theme, internal/i18n, internal/sys }
internal/ui → { internal/pane, internal/layout, internal/profile, internal/store, internal/theme, internal/i18n, mygo/ui, mygo/plugins/terminal }
internal/theme → internal/store
internal/pane → { internal/layout, mygo/plugins/terminal }
internal/profile → internal/store
internal/store → （无内部依赖）
internal/sys → （无内部依赖；标准库与 golang.org/x/sys/windows）
internal/layout → （无内部依赖）
internal/i18n → （无内部依赖）
```

规则：

- `internal/ui` 绝不能 import `internal/app` 或 `internal/sys`：界面只画，动作通过 `ui.Actions` 接口回传。
- `internal/layout` 不认识"终端"，只算矩形。
- `internal/pane` 不做 IO：移除窗格时返回被移除的句柄，由 `internal/app` 调用 `Terminal.Close()`。
- `internal/store`、`internal/profile` 只做纯解析与序列化，不 import `os` 或注册表相关包；数据由 `internal/app` 读入后注入。
- `internal/theme` 只把设置翻译成 `terminal.Theme` 与颜色，不读文件、不认识窗口；颜色解析（`#rrggbb` 等）只在它这里做一次，`internal/ui` 复用。
- `internal/sys` 只被 `internal/app` import。其他包需要 OS 能力时，在使用方定义接口，由 `app` 注入实现。
- 以上依赖规则由测试机械校验（`internal/app/deps_test.go`：用 `go list` 读每个包的直接 import，比对上面的依赖图，并检查纯函数包不 import `os`/`os/exec`/`path/filepath`/注册表）。

## 阶段 3 当前边界

阶段 1 的标签与 profile 已就位：`cmd/wtgo/main.go` 只调用 `app.Run()`；`internal/app` 持有状态、在 UI 线程上改（`Do`），`internal/ui` 只读 `ui.State` 并回传 `ui.Actions`。

标签栏兼作标题栏：窗口用 `mygo.TitleBarHidden` + `TitleBarHeight = ui.TitleBarHeight`（36dp），标签栏自己就是拖窗区域（空白处 `DragWindow`），激活标签用终端背景色 + 上圆角，未激活融入标题栏底色。

窗格树：一个标签是一棵 `pane.Node` 树（`Term` 非空即窗格），切分把焦点窗格原地变成 split，旧终端作第一格、新终端作第二格；关闭一个窗格时兄弟节点接替它的位置（保持指针不变），最后一个窗格关闭即关闭标签。几何由 `internal/layout` 算：分割线 4dp、每个窗格至少 40dp，比例 `Share` 在拖动时被夹在两端之内。焦点移动按几何判定（同方向、跨度重叠最多者优先，其次最近），因为树序不等于视觉相邻。

指针：每个窗格是一个绝对定位的 `Box`，`ObserveInput` 收到左键按下就 `FocusPane`；焦点窗格每帧调 `Element.Focus()` 保持键盘焦点（`AutoFocus` 只在一帧生效，移动焦点必须用 `Focus`），对话框打开时不抢。分割线先画线（`theme.Border`），再叠一条 8dp 宽的透明 handle（`Cursor` 改成 resize 光标、`Dragged()` 取每帧增量）。

已生效的设置项：全部，除下表列出的。`language`、`theme`（界面明暗）、`defaultProfile`、`startDirectory`、`startSize`、`font`、`colorScheme`、`schemes`、`padding`、`backgroundOpacity`（窗口整体透明度）、`cursorShape`、`cursorBlink`、`scrollbackMB`、`copyOnSelect`、`rightClick`、`pasteMultilineConfirm`、`closeOnLastTab`、`maxPanes`（切分上限，0 为不限）、`newTabPosition`、`tabWidth`、`alwaysOnTop`、`titleFromActiveTerminal`、`profiles`、`keybindings`。

只解析、未生效的设置项：`startOn`。`settings.json` 缺失时写入默认值，之后被监视并热重载（每秒比字节，解析失败保留旧值并把原因写进设置页底部）；自动发现的 profile 只补进内存，不写回文件。

设置标签页：`settings.open`（默认 `Ctrl+,`）开一个没有窗格的标签（`pane.Tab.Root == nil`），`ui.Tab.Page` 让它画设置页而不是窗格树。表单直接改 `a.settings`，每帧与 `a.applied` 比较，不同就 `apply()` 一次（幂等：语言、窗口明暗、键位、profile、每个终端的字体/配色/选中即复制、窗口透明度与底色与置顶）。「保存」把设置写回文件并更新 `a.saved`，「放弃」回到 `a.saved`；有未保存的改动时关这个标签先问一句（`ui.Confirm`）。配置文件页编辑的是副本（`a.editing` + `a.editingArgs`），保存时才 upsert，所以只是看一眼不会把自动发现的 profile 写进文件。

终端的标题与退出回调统一经 `App.Do` 回到 UI 线程；窗口关闭后 `App` 丢弃这些回调，不再触碰已销毁的窗口。

## 状态

单实例 `ui.State`，由 `internal/app` 持有并在 UI 线程上修改（`Do(fn)` 回主线程）。

```go
type State struct {
    Tabs      []Tab        // 每个标签的名字、颜色与窗格树
    Active    int          // 在前面的标签
    Profiles  []Profile    // 新建标签与 ▾ 菜单里的候选
    Rename    *Rename      // 重命名对话框，关闭时为 nil
    Settings  *Settings    // 设置标签页，没打开时为 nil
    Confirm   *Confirm     // 先问一句的对话框，没有时为 nil
    EqualTabs bool         // 标签等宽
}

type Tab struct {
    Name   string
    Color  string
    Panes  *pane.Tab      // 窗格树，焦点在 Panes.Focus
    Mount  func(c *myui.Context, n *pane.Node) myui.Element
    Page   func(c *myui.Context) myui.Element  // 设置标签页画它，有它就不画窗格
}
```

`ui` 不认识终端：它只按 `Panes.Layout(room)` 的矩形摆窗格，`Mount` 由 `app` 给。

## 渲染流程

1. `mygo` 每帧调用根 `View` → `internal/ui.View(c, st, act)`，整棵树重建。
2. 标签栏、窗格、对话框都是这一帧的产物；状态只在 `internal/app` 改，改完 `Invalidate()` 请求下一帧。
3. 先画标签栏（36dp，兼作标题栏），再画窗格树（窗格、分割线与 handle）或设置页，最后画覆盖层（重命名、先问一句、toast）。
4. 每个终端只有一个 `terminal.View` 挂载点（libghostty 的约束）。

## 为什么不用"接管已有 ConPTY 进程"

Windows 上 ConPTY 会话的输入输出是给创建者的管道，外部进程无法安全接管；
"附加到已有终端"需要注入或 debug API，不稳定且会被杀毒误报。wt-go 自己创建、
自己管理生命周期，与 Windows Terminal 的做法一致。

## 进程与生命周期

- 每个窗格一个 `*terminal.Terminal`，一个子进程（ConPTY）；设置标签页没有窗格，也没有终端。
- 关窗格 = `Terminal.Close()`；关标签 = 关它所有窗格；关窗口 = 关所有标签。
- 子进程树靠 Windows 的 Job Object 收尾：`Run` 一开始就把本进程加进一个 `KILL_ON_JOB_CLOSE` 的作业，
  之后所有窗格的 shell（以及它们自己起的进程）都在作业里，窗口崩了或被强杀也不会留下 `pwsh.exe`。
  因为子进程自动继承作业，不需要逐个挂。
- 主 goroutine 与 `watch` 协程的 `panic` 都会写一份 `crash-<时间>.log` 到配置目录（与 `settings.json` 并排）再退出。
- 应用退出 = 关闭所有标签，最后落盘设置。

## 与 sshterm 的复用边界

可直接借鉴（复制后改名、按需裁减）：

- tab 样式（选中底色 + 上圆角、未激活融入标题栏）、四边面板、rail
- 主题与配色方案映射、字体/字号设置
- 冒烟脚本（临时配置目录 + `ui.capture` 截图）
- 控制接口（HTTP/MCP）用于测试与自动化 —— 可选，第二阶段再决定

不继承：SSH/telnet/serial/隧道/保管库/传输/文件面板。
