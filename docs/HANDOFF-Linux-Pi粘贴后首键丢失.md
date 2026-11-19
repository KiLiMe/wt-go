# 交接：Linux 上 Pi 粘贴后丢失首键

## 现象和进度

- 在 Linux 的 wt-go 中启动 Pi，从其他应用复制并粘贴到 Pi 后，第一个按键可能不生效。例如按两次空格，编辑器只插入一个。用户确认 Windows 无法复现。Linux 上的粘贴方式、文本行数、确认弹窗和输入法状态待记录。
- Windows 上用真实剪贴板分别测试了 `Ctrl+Shift+V` 和 `Ctrl+V`，Pi 都收到了第一个空格。另用 Pi 1.5.2 的输入解析器和编辑器测试了普通空格与两种 Kitty 空格按下序列，覆盖与粘贴同批、分批以及粘贴结束标记被拆分三种情况，9 组都通过。
- 根因待确认。当前缺少能复现的 Linux 桌面上的输入记录，尚未修改应用代码。

## 在 Linux 上复现和采集

1. 记录 `pi --version`、wt-go 版本或提交、`echo "$XDG_SESSION_TYPE"`、桌面环境和输入法。先直接启动 Pi，用平常的粘贴方式和节奏复现一次，记录是单行还是多行、是否弹出确认框，以及等待一秒再按首键是否也会丢键。
2. 在同一个 wt-go 窗格里用 `script` 启动 Pi。只使用测试文字，输入日志会记录按键和粘贴内容。

   ```sh
   log=$(mktemp)
   script -q -I "$log" -c 'pi --no-session'
   ```

3. 从另一个应用复制 `paste-probe` 并粘贴，按两次空格，记下 Pi 显示了几个空格，然后退出。若间隔一秒时仍能复现，用这个间隔方便区分事件。若需要多行或其他条件才能复现，按原条件记录。

   ```sh
   od -An -tx1 -v "$log"
   ```

4. 对照记录里的括号粘贴起止序列 `ESC [ 200 ~` 和 `ESC [ 201 ~`，检查结束标记后的两次空格。普通空格是 `20`，Kitty 空格按下可能是 `ESC [ 32 u` 或 `ESC [ 32 ; 1 : 1 u`。`ESC [ 32 ; 1 : 3 u` 是释放事件，不能算作按下。分享时只摘录测试粘贴与两次按键附近的字节。
5. 如果加上 `script` 后现象消失，记录这一差异，并在同一 Linux 桌面用另一个终端跑相同版本的 Pi 作对照。

## 根据记录定位

| 记录结果 | 下一步 |
| --- | --- |
| 只收到一次空格按下，或只有第一次释放 | 跟踪 mygo 的 GTK 按下事件、输入法提交、终端视图和 PTY 写入，找出第一次按下在哪一步消失。 |
| 收到两次有效按下，Pi 只插入一次 | 用记录里的字节序列和实际 Pi 版本重放，跟踪 `StdinBuffer`、按键解析和编辑器。若仅用户配置下复现，再对照禁用扩展的结果。 |
| 只在多行确认后丢键 | 优先检查确认框关闭、终端焦点恢复及 `PasteData` 调用顺序。 |

## 代码入口

- wt-go 在 [internal/app/open.go:179](../internal/app/open.go#L179) 注册 `OnPaste`。确认逻辑在 [internal/app/confirm.go:58](../internal/app/confirm.go#L58)，去掉末尾换行后至少两行才弹窗，确认时调用 `term.PasteData(d)`。
- [go.mod:32](../go.mod#L32) 固定 mygo fork `github.com/KiLiMe/mygo v0.9.1-0.20261116134437-f8b821f95ae1`。可用 `go list -m -f '{{.Dir}}' github.com/egoist/mygo` 找到当前依赖源码。
- mygo 的 `internal/linux/clipboard.go:136` 和 `internal/linux/clipboard.go:196` 同步等待 GTK 剪贴板的格式和内容。`internal/linux/surface.go:1048` 附近处理按键和输入法。同步等待期间是否处理了其他输入事件，需要靠 Linux 上的事件顺序验证。
- mygo 的 `plugins/terminal/view.go:680` 处理 `keyDown`，快捷键执行后会写入 `v.pending`。`plugins/terminal/view.go:861` 的 `typed` 会忽略被标为 `taken` 或 `sent` 的待处理按键。若首键在终端侧消失，记录粘贴前后 `keyDown`、`typed` 和 `keyUp` 的顺序及 `v.pending` 变化。
- Pi TUI 的 `dist/stdin-buffer.js:292` 在完成括号粘贴时清理 Kitty 可打印字符的去重状态。`dist/tui.js:777` 在向编辑器分发前过滤释放事件。行号对应本次检查的 Pi 1.5.2。
