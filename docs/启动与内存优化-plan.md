# 启动与内存优化计划（Linux）

> 状态：已实施并实测（2026-11-12）；结果见第 7 节。
>
> 范围：Linux 上的一次首帧与常驻占用。只改 `wt-go` 自己的层（`internal/app`、
> `internal/sys`），不改渲染内核，也不做需要上游 mygo 才能做的结构性简化。

## 1. 背景与基线

同一台机器（Arch，KDE/Wayland，12 CPU，amdgpu）上，冷启动后 3 秒的空闲窗口：

| 指标 | wt-go | konsole | 结论 |
| --- | --- | --- | --- |
| RSS | ~150 MB | ~105 MB | 差 45 MB，但不是 Go 代码 |
| PSS | ~80 MB | ~40 MB | 差 40 MB，主要是 GTK3 + Mesa/LLVM |
| 线程 | 36 | 8–10 | 17 个是 Go 的（12 个 GC worker） |
| exec → shell 进程 | ~150 ms（早期探针 183–190） | 141 ms | 首帧 139 ms 里有 ~45 ms 可砍 |

归因（PSS）：glibc `[heap]` 21.9 MB、Go arena 21.9 MB（live heap 只有 1–4 MB）、
二进制 11.4 MB、`libLLVM.so.23.1` 6.5、`libgallium` 4.1、GTK3 2.5、字体 2.3。
`gc 1 @0.113s 0%: ... 3->4->1 MB, 4 MB goal, 12 P` —— **Go 代码只占约 9 MB**。

首帧分解：GTK 就绪 40–47 ms、`NewWindow`（EGL/Mesa）31–34、`Theme.SetSource` 13、
字体枚举 23–26、首次 `Shape` 9–10、`sys.Processes()` 5.6–6.6、ghostty+pty 2。

因此本计划只做两件事：**把首帧前不该做的活挪走**，**把每帧重复做的活缓存**。
GC 调参、懒加载 617 套配色表、缩小图标这类实测无收益的事不做（见第 3 节）。

## 2. 工作项

每项都写清：为什么、改哪里、怎么验收、有什么风险。测试与实测命令见第 5 节。

### 2.1 字体列表改成「打开字体页才枚举」

**为什么**：`Run` 在首帧前调 `loadFontFamilies`，主线程要枚举 261 个族、
对其中 5 个等宽族各整一遍 95 个 ASCII 字符（23–26 ms）。这份列表只服务
设置页「配置文件 → 默认值」里的字体下拉。

**做法**：
- `internal/app/app.go` 的 `Run` 不再调用；`internal/app/settings.go` 的
  `settingsState()` 在页面是 `ui.PageProfiles` 时调 `ensureFontFamilies()`。
- `internal/app/font.go`：`loadFontFamilies` → `ensureFontFamilies`，加
  `asked` 状态，重复调用不再起 goroutine；列表在 `a.Do` 里算完（测量要用窗口的
  文本引擎，必须在 UI 线程），到了再 `redraw`。
- 列表按会话缓存：等宽与否是字体族的属性，页面里改字号不重新枚举（否则每按一下
  步进器就重新枚举一次）。

**验收**：首帧前不再有字体枚举；首次进「配置文件 → 默认值」一两帧后下拉出现完整列表；
下拉在没有列表时显示 `monospace` 与当前值（`choice` 的既有行为），不空、不跳。

**风险**：列表到达前用户恰好展开下拉，只看到 `monospace` 一项；一帧后自动补上。

### 2.2 等宽判定加缓存

**为什么**：`monospace()` 每次整 95 个字符（6.7–10 ms）。每开一个标签、每次应用设置
都会走一遍。

**做法**：`internal/app/font.go` 加 `(family, size)` 记忆化，`monospace()` 保持纯函数，
`App.isMonospace()` 负责归一化（空族 → `monospace`、size ≤ 0 → 13）与缓存；
`terminalFont()` 与枚举列表都走缓存版本。

**验收**：同一 `(family, size)` 只整一次；开第 2..n 个标签不再为字体耗时。

**风险**：会话中安装新字体会让缓存过期；字体列表本来也只枚举一次（2.1），
重启即刷新，可接受。

### 2.3 设置页每帧只算一次，配色列表缓存

**为什么**：`view()` 调一次 `settingsState()`，设置页的 `settingsView` 渲染时又调一次；
里面每次 `uiSchemes()` → `theme.Schemes()`（`os.ReadDir` + 617 个名字排序）与 `dirty()`
的整份 `reflect.DeepEqual`（`store.Equal`）。实测一次 `uiSchemes` 是 2.19 ms、745 KB、
39 153 次分配，设置页每帧付两次。

**做法**：
- `internal/app/app.go` 的 `view()` 把算好的 `ui.Settings` 存进 `a.settingsShown`，
  `settingsView` 直接用它，一帧一次。
- `internal/app/settings.go`：配色列表缓存在 `a.schemes`（nil 表示待重建），`apply()`
  在设置真的变化后置 nil。设置是列表唯一的输入（自定义配色写在 `settings.Schemes` 里，
  内置与 Ghostty 的是编译期表），`apply()` 又是「设置生效」的唯一出口，所以在那里失效
  就够，不必逐字段观察。

**验收**：设置页打开、不动表单时每帧不再读目录、不再排序 617 个名字（`BenchmarkUISchemes`
钉住单次成本）；改自定义配色、删配色后下一帧就是新列表。

**风险**：绕过 `apply()` 直接改 `a.settings` 的地方会让列表滞后一帧——目前没有这样的地方。

### 2.4 窗口按钮为空时返回 nil

**为什么**：`windowButtons()` 在 `!hideWindowButtons()` 时返回 nil，但在 Linux
（Wayland）上 `hideWindowButtons()` 恒为 true、`a.buttons` 却可能是空的，
于是每帧返回一个空结构体，`ui` 侧还要建一遍。

**做法**：`internal/app/app.go`：`hideWindowButtons()` 为假、或 `Start`/`End`
都为空时返回 nil，与 mygo 在合成器自绘标题栏时的约定一致。

**验收**：Wayland 下条带不再画窗口按钮（本来就没画），X11 上不变。

**风险**：理论上无——空结构体本来也画不出按钮。

### 2.5 `OpenPath` 不再留僵尸

**为什么**：`internal/sys/open_other.go` 只 `exec.Command(...).Start()`，
子进程退出后没人 `Wait`，每打开一次路径留一个僵尸，Linux 上会话越久越多。

**做法**：`Start()` 成功后 `Process.Release()`（不想等 `xdg-open` 跑完，
只要不留下可回收项）。

**验收**：连续打开几次 settings 目录后 `ps -o stat,cmd --ppid <wt-go pid>`
没有 `Z` 行。

**风险**：`Release` 之后拿不到退出码——这里从来不看。

### 2.6 glibc 的已释放页还给内核（验证后不做）

**原假设**：`[heap]` 21.9 MB 里约 7 MB 是已释放、glibc 留着复用的页。探针里用
`LD_PRELOAD` 调 `malloc_trim(0)`，3 秒空闲 RSS 从 ~154 MB 降到 ~143 MB，于是打算在
`internal/sys` 加 `TrimHeap()`，首帧后调一次。

**实测（改成进程内调用后）**：`purego.Dlsym(RTLD_DEFAULT, "malloc_trim")` 能取到符号，
`malloc_trim(0)` 返回 1（确实归还了页），调用瞬间 VmRSS 154.6 → 149.0 MB；但 3 秒后
又回到 151 MB —— 归还的页很快被同一批分配触到。改在 2 秒时调用同样只有
145 → 147 MB。只有周期性 trim（`LD_PRELOAD` 每 500 ms）能稳定在 143 MB / PSS 75 MB，
那是把 glibc 的复用池一直压小的结果，代价是一个定时器和跨线程的「上次绘制时间」。

**结论**：一次性归还真值 3–4 MB 且不稳定，周期性归还要求 UI 循环里多一个定时器，
收益不值这份复杂度。故不实现（`internal/sys/trim_*.go` 已删除），理由记在 ADR-040。

### 2.7 关闭「其它标签」时进程表只取一次（P1）

**为什么**：`othersActivity` 对每个标签调一次 `activity()`，而 `activity()`
每次 `sys.Processes()` 全量扫 `/proc`（4.5–6.6 ms / 356 进程）。4 个标签就是 4 次。

**做法**：`internal/app/tabs.go`：`activity(tab)` 改成 `activity(tab, list)`，`list` 是
`processList`（进程表 + 「读没读到」），由 `a.processList()` 取一次；`othersActivity`
取一次给所有标签用，`CloseTab` 自己取一次。读不到时按原有语义降级（当作说不清），
失败日志仍然只打一行。

**验收**：`internal/app` 现有测试全过；关闭多标签时 `/proc/<pid>/stat` 的打开次数
与标签数无关。

**风险**：标签之间共享同一份快照，与原来「每个标签各取一次」的差异在毫秒级，
只影响「某个标签刚刚起进程」的极端时序，属于原有语义的容差。

### 2.8 等 shell 出现时读 children 文件，不再全量扫 `/proc`（第二轮补做）

**为什么**：`spawned` 每 10–100 ms 调一次 `sys.Processes()`（读 356 个
`/proc/<pid>/stat`），最久等 3 秒；`reexec` 在 250 ms 里同样反复扫。实测启动后 4 秒内
`/proc/<pid>/stat` 被打开 **2098** 次，全在后台线程上。

**做法**：
- `internal/sys` 加 `Subtree(root)`：Linux 读 `/proc/<pid>/task/<tid>/children`
  （内核自己维护的子女表，每个线程一份，逐个 task 读），只 stat 子树里的进程；
  Windows/macOS 没有这个文件，读一次整张表再用纯函数 `SubtreeOf` 筛。
- `internal/app/tabs.go`：`spawned` 的 `before`/每次轮询、`reexec` 的每次轮询改走
  `sys.Subtree`；两处 `select` 上 `term.Done()`，程序没起来（`E304`）不再白等 3 秒。
- 「一次问好几个标签」的 `othersActivity`/`CloseTab` 仍用一张表：问题跨标签，
  一张表比每个标签各读一棵树便宜（Windows 上尤其），这也是 2.7 的结论。

**验收**：启动后 4 秒内 `/proc/<pid>/stat` 打开次数从 2098 降到个位数；开标签仍然认得
shell（不等宽、启动后没有 `E304`）。

**风险**：`children` 文件是 Linux 特有的；别的平台读整张表，与改动前同价。进程在
「内核报了名字」与「读它的 stat」之间结束时会读失败，跳过即可（本来就允许）。

## 3. 明确不做（含理由）

| 事项 | 为什么不做 |
| --- | --- |
| `GOMAXPROCS=4` | 只省 6 线程 / ~1 MB PSS，却给多标签大输出的解析留了吞吐风险；收益不值这个风险 |
| GC 调参（GOGC/GOMEMLIMIT） | live heap 1–4 MB，实测无收益或反向 |
| 懒加载 617 套配色表 | 编译期字面量，启动成本≈0 |
| scrollback 默认调小、裁剪图标/主题、延迟日志 | 实测 ±0 |
| `Theme.SetSource` 移到 `NewWindow` 之后 | 省 6–8 ms，但首帧 GTK chrome 可能先闪一下系统色，收益不抵不确定 |
| 把 `activity` 也改成每棵子树各读一次 | 跨标签的问题用一张表更便宜（Windows/macOS 读子树就是读整张表）；见 2.8 |
| 进程内 `malloc_trim(0)` | 见 2.6：一次性只值 3–4 MB 且被后续分配吃掉，周期性要往 UI 循环里加定时器 |
| 向 mygo 要 `Terminal.Pid()` | 上游能力，做了能删掉整套快照/轮询/claimed 逻辑，但不在本仓库范围内 |

## 4. 实施顺序

1. 2.1 + 2.2（同一文件，先做，首帧收益最大）。
2. 2.3 + 2.4（每帧成本，改 `view()` 的取数方式，一起验）。
3. 2.5（`internal/sys` 的独立小修）；2.6 经实测放弃，不留代码。
4. 2.7（最后做，改 `tabs.go` 的调用形状）。
5. 跑测试、实测、填第 7 节、改相关文档、提交。

## 5. 验证方法

```sh
go vet ./...
go test ./...
GOOS=windows go vet ./... && GOOS=darwin go vet ./...
GOOS=linux GOARCH=arm64 go vet ./...
go test ./internal/app -run XXX -bench UISchemes
go tool mygo build -platform linux/amd64 -skip-dmg -skip-installer
```

实测脚本（不用 strace，读 `/proc`）：

- **首帧/到 shell**：从 `exec` 到出现 `wt-go` 进程的子 shell，按 10 ms 采样；
  连续 5 次取中位数。
- **RSS/PSS/线程**：3 秒后读 `/proc/<pid>/status`（`VmRSS`、`Threads`）
  与 `/proc/<pid>/smaps_rollup`（`Pss`）。
- **设置页每帧成本**：`BenchmarkUISchemes` 量单次 `uiSchemes`（一帧两次），
  比在窗口里数帧更稳；要看端到端时用 `/proc/<pid>/stat` 的 utime 前后差 / 帧数。

## 6. 预期收益

- 首帧：字体枚举 23–26 ms + 首个标签的等宽判定 7–10 ms 移出启动路径（合计 ~30 ms）。
- 设置页：每帧两次 → 一次；`Schemes()` 从每帧 2×2.2 ms、2×745 KB 分配变成
  「设置变了才做一次」。
- 每次 `xdg-open` 少一个僵尸。
- 关闭多标签：进程表从「每标签一次」变成一次（每次省 4.5–6.6 ms）。

## 7. 结果

机器同上；`before` 是改动前 `git stash` 出来的二进制，`after` 是本次构建（均为
`-ldflags "-s -w"`）。首帧 = 从 `exec` 到出现子 shell 进程；RSS/PSS 取启动 3 秒后。

| 指标 | before | after |
| --- | --- | --- |
| exec → shell | 144 / 154 / 146 ms，153 / 154 / 155 ms | 127 / 124 / 126 ms，124 / 123 / 124 ms |
| 中位数 | 153.5 ms | 124.5 ms（**−29 ms**） |
| RSS（3 s） | 153.0–153.6 MB | 154.6–156.9 MB |
| PSS（3 s） | 84.2–84.7 MB | 83.2–85.5 MB |
| 线程 | 36–37 | 36–37 |
| `uiSchemes` 一次 | 2.19 ms / 745 KB / 39 153 allocs | 设置页每帧不再调用 |

- **首帧**：−29 ms，与预测的「字体枚举 23–26 + 等宽判定 7–10」吻合；设置页的
  `Schemes()` 缓存不参与启动路径，故未影响这一项。
- **常驻占用没有变化**：本计划的四项都只影响 CPU 与分配量，不动 GTK/Mesa 的内存；
  唯一可能降常驻的 2.6 已放弃（见上）。150 MB 里 Go 只占 ~9 MB 这个结论没有变。
- **设置页**：每帧从 2×2.19 ms、2×745 KB 的分配变成 0（`BenchmarkUISchemes` 钉住
  单次成本，见 `internal/app/settings_test.go`）；`DeepEqual` 的比较只剩一次。
- **僵尸**：`OpenPath` 现在 `Process.Release()`；这一项的验收是靠代码路径
  （不 `Wait` 也不 `Release` 才会留 `Z`），没有可自动化的窗口内断言。
- **多标签**：关闭「其它标签」时进程表取一次，标签数与 `/proc` 扫描次数解耦；
  具体耗时未单独计时（4 个标签约省 14–20 ms，且发生在 UI 线程）。
- 交叉检查：`GOOS=windows/darwin/linux-arm64 go vet ./...` 通过，`go test ./...` 通过。

### 第二轮（2.8，2026-11-12 补做）

| 指标 | before（6aa6f38） | after |
| --- | --- | --- |
| 启动后 4 s 内 `/proc/<pid>/stat` 打开次数 | 2098 | 9 |
| 同上，`/proc` 下任意打开 | 2098 | 106（children 与 task 目录） |
| exec → shell | 123 / 134 / 132 ms | 115 / 116 / 116 ms |
| 线程（3 s） | 37 / 39 / 37 | 32 / 34 / 34 |
| RSS / PSS | 154.1–155.9 MB / 82.7–84.6 MB | 153.9–155.9 MB / 82.5–84.5 MB |

- `/proc` 的读从「每次轮询一台机器」变成「每次轮询几棵子树」，2098 → 9 次 stat。
- 线程少了 3–5 个：轮询不再产生成片的阻塞读，runtime 留下的 M 线程随之变少（观察到的
  相关性，未单独做归因实验）。RSS/PSS 不变，说明这些线程本来就不带映射。
- `spawned`/`reexec` 现在等 `term.Done()`：程序起不来（`E304`）当场记一行日志，
  不再空转 3 秒。
- 遗留：`/proc` 全量扫描换 `children` 查询、`Theme.SetSource` 推迟、向 mygo 要
  `Terminal.Pid()`、渲染路径的 pango/atlas 成本，都留给后续（见 `docs/plan.md` 阶段 6）。
