# 交接：终端里的链接是怎么识别的

> 起因：用户问「`echo 1.exe` 为什么 `1.exe` 也会被匹配到（下划线）」，并要求对照 Windows Terminal
> 「它不会把 `1.exe` 当可点击」。本文件记录已核实的事实、代码位置、以及若要改动的三条路线。
> 读者：下一个接手这个仓库的 AI/人。核实时间 2026-11-17。

## 0. 一句话结论

- **识别在 mygo 终端插件里**，本项目（wt-go）只决定「打开的到底是 URL 还是路径、从哪个目录开」。
- 插件有三种来源，按优先级：**OSC 8 超链接 → 文本 URL → 文本路径**（路径这一档只在宿主设了
  `Options.OpenLink` 时启用）。
- **Windows Terminal 根本没有「文本路径」这一档**：它只有 OSC 8 + 一条必须带 scheme 的 URL 正则。
  所以 `1.exe` 不是被它「过滤」掉的，是从来没进过候选集。
- `1.exe` 在 wt-go 里会下划线，是因为 `pathOf` 的无斜杠判据只要求「名字含字母 + 末段像扩展名」，
  不要求首字符是字母。点击时 `os.Stat` 失败 → 静默什么也不做（`internal/app/links.go:43`）。

## 1. 版本基线（看代码前先确认这三处）

| 项 | 值 |
| --- | --- |
| 插件来源 | `go.mod` 里 `replace github.com/egoist/mygo => github.com/KiLiMe/mygo v0.9.1-0.20261116134437-f8b821f95ae1` |
| 模块缓存路径 | `C:\Users\mk\go\pkg\mod\github.com\!ki!li!me\mygo@v0.9.1-0.20261116134437-f8b821f95ae1\` |
| 本项目侧 | `internal/app/links.go`、`internal/app/proc.go:39`、`internal/app/apply.go:123` |
| WT 对照源码副本 | `tmp_wt_src\`（main 分支 raw 下载，可删；重新拉取见 §5 末尾） |

插件关键文件：`plugins/terminal/links.go`（纯函数：正则与判定）、`plugins/terminal/view.go`（命中与手势）、
`plugins/terminal/terminal.go`（`Options.OpenLink`、`SetLinks`、`SetLinkHoverHint`）。

## 2. 插件侧：三种来源（`view.linkAt`，view.go:1451）

顺序固定，先命中先返回：

1. **OSC 8 超链接** — `v.hyperlinkAt`（view.go:1502）读 `term.CellAt(col,row)` + `term.Hyperlink(ref)`。
   命中则整段超链接（含软换行跨行）为链接，`path=false`。
2. **文本 URL** — `urlAt` → `matchAt(urlPattern, …)`（links.go:259 / links.go:263）。正则（links.go:15）：
   ```
   (?:(?:https?|ftp|file)://|mailto:)[^\s"'<>`\x00-\x1f\x7f-\x{10ffff}]+
   ```
   含义：非 ASCII（CJK、全角标点）处截断；`trimURL`（links.go:281）再去掉句末标点与 URL 自己没开过的 `)]}`。
3. **文本路径** — 仅当 `t.opts.OpenLink != nil` 时调用 `pathAt`（links.go:42）。

### 2.1 `pathAt` 的判定链（这是 `1.exe` 的成因所在）

| 步骤 | 位置 | 规则 |
| --- | --- | --- |
| 引号优先 | `quotedAt` links.go:96 | `"…"` / `'…'` 内可含空格；`don't` 的撇号不算引号；可吞掉后置 `"main.py":12` |
| 切词 | `pathToken` links.go:23 | 排除空白、引号、`<>`、中文标点 `，。；：！？、（）【】《》「」『』` |
| 整词试 | `pathOf` links.go:152 | 见下表 |
| 词内抠 | `pathInToken` links.go:28 | `main.go:12:code`、`--out=build/app.js`、`PATH=/usr/bin:/bin`；左右不得贴 `wordByte`（`-`/`_`/`@`/`+`/字母数字），故 `OpenConsole.exe-amd64` 不算 |
| 剥壳 | `pathOf` links.go:152 | 前导 `([{*`、`--opt=` / `VAR=` 前缀、`write(notes.md)` 外层括号、尾随 `,;.!?]`；`balanced`（links.go:226）处理 `(image.png)已保存` |
| 位置后缀 | `pathPlace` links.go:36 | `:line`、`:line:col`、`(line,col)`、`#LlineCcol`，统一重写成 `path:line:col` |
| 硬否 | links.go:200-203 | 含 `:`、`\x00`、换行 → 否；无任何字母数字 → 否 |
| 单词兜底 | links.go:205-213 | **无 `/` 的名字**：必须含字母（不是「以字母开头」），且 行号存在 或 末段通过 `kind()` |
| 扩展名形状 | `kind` links.go:132 | 首字符是字母、长度 ≤8、余下全字母数字。所以 `1.2.3`、`12:30`、`exe-amd64` 都不算 |

跨行：`linkAt` 先把该行所在的软换行组拼成一条 `text` + 累计列号 `cols`，长 URL 折行也能整段命中；
`linkSpan` / `spanOf` 把命中区间还原成矩形（view.go:1335 `HoveredLink` 用它画预览）。

### 2.2 `1.exe` 逐步

1. `echo 1.exe` 切成 `echo`、`1.exe`。
2. `echo`：无斜杠无点 → `pathOf` 拒，`pathInToken` 两个分支都要 `/` 或 `.` → 无命中。正确。
3. `1.exe`：整词命中 `pathInToken` 第三分支，边界检查通过 → 进 `pathOf`。
4. `pathOf`：`name="1.exe"`、`line=0`、`dot=1`；`IndexFunc(name, unicode.IsLetter) >= 0`（有 `e/x/e`）
   且 `kind("exe") == true` → **放行**，返回 `1.exe`。
5. 点击 → 本项目 `openLink`（`internal/app/links.go:24`）：相对路径 join `linkDir` → `os.Stat` 失败 →
   `return true`（什么都不开，也不报错）。若该目录真有 `1.exe`，则正常用系统默认程序打开。

## 3. 触发条件（什么时候才「看起来是链接」）

- 位置来源 `view.linkHere`（view.go:1369）：以下任一为真则不识别——`t.noLinks`、正在选择、程序上报鼠标
  （`v.reporting`）、光标在右侧滚动条带上（`overScrollbar`）、`cellW == 0`。
- 悬停：`view.linkUnder`（view.go:1359）要求**按住 `ui.Cmd`**（Windows 上即 Ctrl）**或** `hoverHint` 为真
  （`SetLinkHoverHint(true)`）才下划线/手型。
- 点击（带修饰键）：Ctrl+左键按下即 `v.opening`，事件被吞，锁外调 `OpenLink`（view.go:1174、view.go:642）。
- 点击（不带修饰键）：按下只是「armed」，**必须在同一单元格松开**才算点击，拖走就变成选择
  （view.go:1195、view.go:1254、view.go:1268）。
- 回落：`open := t.opts.OpenLink; (open == nil || !open(link)) && !v.openingPath` 才 `services.OpenURL`。
  即**路径**宿主不接就什么都不做；**URL** 才交给系统（view.go:644）。

## 4. 本项目侧（`internal/app/links.go`）

| 符号 | 行 | 职责 |
| --- | --- | --- |
| `App.openLink(link, dir)` | 24 | `linkPath` 为 `""` → URL → `sys.OpenURL`；否则补相对路径、`os.Stat`、`sys.OpenPath`。失败只 `log.Printf` + `fault.Line(fault.OpenLink, …)`（`E209`，`internal/fault/fault.go:31`） |
| `setLinks(term, open)` | 54 | 一次设两个旋钮：`SetLinks` + `SetLinkHoverHint` |
| `linkDir(term, start)` | 66 | 优先 shell 通过 **OSC 7** 报的目录（必须绝对），否则终端启动目录；PowerShell 不报 OSC 7 |
| `linkPath(link)` | 80 | `mailto:` → URL；含 `://` 且 scheme 前无 `/\` → URL（**盘符的冒号不是 scheme**）；其余 `sys.ExpandPath` 后去位置 |
| `pathWithoutPlace(path)` | 93 | 反复剥 `:line(:col)`；`i <= 1` 保护盘符 |
| `allDigits` | 105 | 位置必须是数字串 |

注入点：`internal/app/proc.go:39`
`opts.OpenLink = func(link string) bool { return a.openLink(link, linkDir(term, opts.Dir)) }`

开关：`settings.json` 的 `links.open`（`internal/store/settings.go:110`，`Links` 结构体见 `:105`，
每个 profile 的 `links` 覆盖见 `:75`，默认值 `internal/store/defaults.go:71`），
改动时 `internal/app/apply.go:71` 比较、`:123` 对每个终端重放。用户文档：`docs/配置参考.md:88`。

**已知副作用**（测试里已明确接受）：`pathWithoutPlace("12:30") == "12"`（`internal/app/links_test.go` 的
`TestPathWithoutPlace` 有注释说明「A time is not a place, but it reads as one」）。

**不变量，别破坏**：
- 插件做纯渲染，不做 IO（不 `stat`、不查存在性）。
- `openLink` 对「路径不存在」静默返回 `true`，不弹 explorer、不弹错误框——终端里没有地方说「为什么没反应」。
- 只有 URL 打开失败才记 `E209`。

## 5. Windows Terminal 的对照（已核实，非推测）

点击一次 = 三步，全程没有路径检测：

1. `ControlInteractivity::_getHyperlinkForPointerEvent`（`tmp_wt_src\ControlInteractivity.cpp:266`）：
   只有 `isLeftButton && modifiers.IsCtrlPressed()` 才 `_core->GetHyperlink(pos)`，否则 `{}`。
2. `ControlCore::GetHyperlink`（`tmp_wt_src\ControlCore.cpp:925`）→ `Terminal::GetHyperlinkAtViewportPosition`
   → `Terminal::GetHyperlinkAtBufferPosition`（`tmp_wt_src\Terminal.cpp:539`），只有两个 case：
   - OSC 8：`attr.IsHyperlink()` → `GetHyperlinkUriFromId`
   - 自动检测：`_patternIntervalTree.findOverlapping(...)` 命中 `_hyperlinkPatternId` → `GetPlainText(start, stop)`
   **没有 case 3 去把周围字符拼成路径。**
3. 自动检测只有一条正则：`Terminal::_getPatterns`（`tmp_wt_src\Terminal.cpp:1422`），`patterns` 数组**只有一个元素**
   （`tmp_wt_src\Terminal.cpp:1425`）：
   ```
   \b(?:https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|$!:,.;]*[A-Za-z0-9+&@#/%=~_|$]
   ```
   `1.exe` 无 `://` → 不可能进树 → 不可点击。整条链受 `DetectURLs` 开关控制
   （`tmp_wt_src\Terminal.cpp:149`、`:1428-1431`）。

旁证：hover 下划线也走同一套——`ControlCore::_updateHoveredCell`（`tmp_wt_src\ControlCore.cpp:881`）只算
`GetHyperlinkIdAtViewportPosition` + `GetHyperlinkIntervalFromViewportPosition`；`HoveredUriText`
（`tmp_wt_src\ControlCore.cpp:931`）只回这两类；`TermControl::_hoveredHyperlinkChanged`
（`tmp_wt_src\TermControl.cpp:3377`）拿它做 tooltip，还带 IDN 同形字 Punycode 并排显示。

重新拉取（pwsh）：
```powershell
$b = "https://raw.githubusercontent.com/microsoft/terminal/main/src/cascadia"
Invoke-WebRequest "$b/TerminalCore/Terminal.cpp" -OutFile tmp_wt_src/Terminal.cpp
Invoke-WebRequest "$b/TerminalCore/Terminal.hpp" -OutFile tmp_wt_src/Terminal.hpp
Invoke-WebRequest "$b/TerminalControl/ControlCore.cpp" -OutFile tmp_wt_src/ControlCore.cpp
Invoke-WebRequest "$b/TerminalControl/ControlInteractivity.cpp" -OutFile tmp_wt_src/ControlInteractivity.cpp
Invoke-WebRequest "$b/TerminalControl/TermControl.cpp" -OutFile tmp_wt_src/TermControl.cpp
```

## 6. 能力对比表

| | Windows Terminal | mygo 插件（KiLiMe fork 新增） |
| --- | --- | --- |
| OSC 8 | ✅ | ✅ |
| 文本 URL | 一条 scheme 正则，字符集窄 | `urlPattern`（links.go:15），更宽 + `trimURL` |
| 文本路径 | ❌ 完全没有 | ✅ `pathAt`（links.go:42）三层启发式 |
| 判定依据 | 正则语法 | 语法启发式 + `kind()` 扩展名形状 |
| 存在性检查 | 无 | 无（在宿主 `internal/app/links.go:43`，且失败静默） |
| 触发键 | Ctrl+左键 / hover 即划线 | Ctrl+左键 / `hoverHint` / 不带键但同格松开 |

## 7. 若要「让 `1.exe` 也不可点」：三条路线

**路线 A（最小、只杀纯数字词干）** — 改插件 `pathOf` 的无斜杠分支（links.go:205-213）：

```go
if !strings.Contains(name, "/") {
    dot := strings.LastIndexByte(name, '.')
    if strings.IndexFunc(name, unicode.IsLetter) < 0 || line == 0 && (dot < 0 || !kind(name[dot+1:])) {
        return "", 0, 0
    }
}
```

加一条「词干不能全是数字」（`stem = name[:dot]`，`allDigits(stem)` → 否）。效果：杀 `1.exe`、`404.html`；
保住 `7z.exe`、`2to3.py`。**代价**：这是 fork 里的上游代码，必须改 `KiLiMe/mygo` 并升级 `go.mod` 的
`replace` 版本；本项目无法从外部覆盖（插件不暴露这个钩子）。

**路线 B（要求首字符是字母）** — 同一处把 `IndexFunc(...) < 0` 换成 `!isLetter(name[0])`。简单但误杀
`7z.exe`、`2to3.py`、`404.html` 一整个家族。**不推荐**。

**路线 C（存在性前移到 hover）** — 让 `linkUnder` 回调宿主做 `stat`。**不推荐**：破坏「插件纯渲染、无 IO」
的分层，且每次鼠标移动都要 stat，还要处理网络盘卡顿。

**路线 D（什么都不改）** — 现状代价很低：`1.exe` 只是被划线，点击无副作用（`os.Stat` 失败即静默）。
当前项目取舍是「识别宽、打开严」。若没有真实用户抱怨，倾向不改。

## 8. 验证

```powershell
cd D:/Sou/wt-go
go test ./internal/app -run 'TestLink|TestPathWithoutPlace' -v
go test ./internal/app            # linkDir 的 OSC 7 用例会起无头终端
```

插件侧行为（`pathAt` / `pathOf` / `kind`）的测试在 fork 仓库
`plugins/terminal/terminal_test.go`（`TestCommandClickLink*`、`TestSetLinksOff`、`TestSetLinkHoverHint`），
**在本仓库里跑不到**；改插件必须去 `KiLiMe/mygo` 跑。

## 9. 未决 / 待确认

- 用户尚未决定是否要改（A/B/C/D 未选）。上次对话末尾已问「需要我去 fork 里改并跑插件测试吗？」——**未答复**。
- 若改插件：需同步 `go.mod` 的 `replace` 版本号，并检查 `internal/app/links_test.go` 是否仍全绿
  （`pathWithoutPlace` 的 `12:30` 副作用与本问题无关，不要顺手「修」）。
- `tmp_wt_src\` 是临时下载，可随时删除；文档里引用的 WT 行号基于 2026-11-17 的 `main`。
