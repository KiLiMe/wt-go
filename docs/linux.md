# Linux（实验性）

现状：已完成**首轮真机验证**（Arch Linux + KDE Plasma Wayland / XWayland，2026-11，结果见 §7）：构建、`go test -race`、`mygo build` 打包（deb / tar.gz / PKGBUILD）、启动、终端与设置页交互均可用；仍存在窗口边框（X11 下完全无边框）与若干 Windows 残留文案。本文是验证手册与结果记录。

配合阅读：[docs/architecture.md](architecture.md)（分层与依赖方向）、[docs/plan.md](plan.md)（阶段计划）。

## 1. 运行库依赖

mygo 在 Linux 上通过 purego 直接调用已加载的 GTK 3 / Pango，**不用 cgo，也不需要 WebKitGTK**（原生 UI，无 webview）。

| 库 | Debian / Ubuntu 包 | 缺失后果 |
| --- | --- | --- |
| `libgtk-3.so.0`、`libgdk-3.so.0` | `libgtk-3-0` | 窗口打不开，进程退出 |
| `libglib-2.0.so.0`、`libgobject-2.0.so.0`、`libgio-2.0.so.0` | `libglib2.0-0` | 同上 |
| `libcairo.so.2` | `libcairo2` | 同上 |
| `libgdk_pixbuf-2.0.so.0` | `libgdk-pixbuf-2.0-0` | 同上 |
| `libpango-1.0.so.0`、`libpangocairo-1.0.so.0` | `libpango-1.0-0`、`libpangocairo-1.0-0` | **窗口能开，但没有任何文字** |
| `libharfbuzz.so.0`、`libpangoft2-1.0.so.0` | `libharfbuzz0b`、`libpangoft2-1.0-0` | 可选；缺则退回简单排版 |

另外需要：

- **CJK 字体**：缺了中文与框线字符都是豆腐块。Debian 系 `fonts-noto-cjk`，Fedora `google-noto-sans-cjk-fonts`，Arch `noto-fonts-cjk`
- **glibc ≥ 2.28**（`libghostty-vt.so` 的要求；Debian 10 / Ubuntu 18.04 太旧）
- **输入法**：需设环境变量，fcitx5 用 `GTK_IM_MODULE=fcitx`，ibus 用 `GTK_IM_MODULE=ibus`；否则中文打不进去

## 2. 安装依赖

```sh
# Debian / Ubuntu
sudo apt install libgtk-3-0 libpango-1.0-0 libpangocairo-1.0-0 fonts-noto-cjk

# Fedora
sudo dnf install gtk3 pango pangocairo google-noto-sans-cjk-fonts

# Arch
sudo pacman -S gtk3 pango noto-fonts-cjk
```

## 3. 自检（先跑这些，再看程序）

```sh
ldd --version | head -1                                  # 需 >= 2.28
ldconfig -p | grep -E 'lib(gtk-3|gdk-3|pango-1\.0|pangocairo-1\.0|cairo|glib-2\.0|gdk_pixbuf-2\.0)\.so'
fc-list :lang=zh | head -3                               # 需有输出
echo "SHELL=$SHELL"                                      # 见 L-1
echo "session=$XDG_SESSION_TYPE desktop=$XDG_CURRENT_DESKTOP"
```

## 4. 终端内核需要的原生库

终端由 mygo 的 `plugins/terminal`（libghostty-vt）提供，`libghostty-vt.so` **不随 Go module 分发**。查找顺序：

1. `$MYGO_GHOSTTY_VT`（指向 .so 的完整路径）
2. `mygo.PathResources`（`mygo build` 打包时放进去的）
3. **可执行文件所在目录**
4. `~/.cache/mygo/natives/<sha256>/`

前三个都没有时，它才会联网从 GitHub Releases 下载（2 分钟超时、SHA-256 校验）。校验值：

| 架构 | sha256 |
| --- | --- |
| linux/amd64 | `9dcd8af0de67778a561f436cb673a9fb1738be8fa00849dad49606baed1e57d7` |
| linux/arm64 | `fb163f8f123206956751f1c4f0303b40899513b7dc171a8ab8d85ba7cd541bc0` |

开发时最省事的做法是把它放在 `wtgo` 旁边，或导出 `MYGO_GHOSTTY_VT=/path/to/libghostty-vt.so`。发布包应走 `mygo build`，由它把库打进包里。

## 5. 构建与运行

```sh
go build -o wtgo ./cmd/wtgo
./wtgo                     # 必须在前台终端里跑：原生库失败只写 stdout 一行 log
```

- 后端：X11 `GDK_BACKEND=x11 ./wtgo`；Wayland `GDK_BACKEND=wayland ./wtgo`（不设则自动选，**两种都要试**）
- 数据目录：`~/.config/wt-go/`（`settings.json`、崩溃时的 `crash-<时间>.log`）
- 无头环境冒烟：`xvfb-run -a ./wtgo`（只验证「窗口能开、不崩」，文字与按钮要在真会话里看）

## 6. 已知缺陷：本轮要逐条验证

表里「预期」是我们根据代码推断的结论，测试就是用真机确认或推翻它。**不要修，先记录**。

| 编号 | 现象 | 触发条件 | 相关代码 | 预期 |
| --- | --- | --- | --- | --- |
| L-1 | 窗口开着，一个标签都没有，界面全空 | `$SHELL` 为空或不是绝对路径 | `internal/sys/shell_other.go:16-22`；`internal/app/app.go:218-225`、`:419-424` | 必然复现 |
| L-2 | 设置页点「打开 settings.json」无任何反应 | 任意 | `internal/app/settings.go:264`（`exec.Command("explorer.exe", path)`） | 必然复现 |
| L-3 | 启动即空白窗口 | 无本地 .so 且无网络/被墙 | `internal/app/app.go:445-448`（只 `log.Printf`） | 可在前台 log 看到原因 |
| L-4 | 标题栏没有最小化/最大化/关闭按钮 | Wayland 且合成器自装饰（KWin、Hyprland、Sway 默认） | `internal/app/app.go:115` 未设 `HideWindowButtons` | 预期复现，只能 Alt+F4 关窗 |
| L-5 | 窗口按钮左侧有一条约 80px 的死区，点它不响应 | X11 | `internal/ui/tabs.go:36`（`bar.Right` 让位给 GTK header bar 的空标题） | 预期复现 |
| L-6 | 所有文字发白/边缘发绿 | cairo < 1.17.8（Debian 12、Ubuntu 22.04、RHEL 8/9、麒麟） | 上游 `gtkfix.TextColor` 未应用 | 看发行版 |
| L-7 | 中文、框线字符显示成方块 | 未装 CJK 字体 | — | 装上字体即消失 |
| L-8 | `Ctrl+滚轮` 改的是字号、不是滚动 | 任意 | mygo 的 `ui/keys.go:257-262`（非 macOS 上 `Cmd == Ctrl`）与 `plugins/terminal/view.go:1396`（`zooming := ev.Mods&ui.Cmd != 0`） | **设计如此**，记录手感即可 |
| L-9 | 标签上的 profile 圆点没有颜色 | 任意 | `internal/profile/profile.go:116-128`（只认 pwsh7/powershell/cmd/wsl） | 预期复现 |
| L-10 | 设置里显示 `%USERPROFILE%`、默认 shell 是 `pwsh.exe` | 设置页 | `internal/ui/settings.go:151`、`:294`、`:300` | 预期复现 |
| L-11 | 主进程被 `kill -9` 后 shell 仍在运行 | 强杀 | `internal/sys/job_other.go:7-8`（空实现） | 预期复现；正常关窗应无残留 |
| L-12 | 保存设置后 `settings.json` 权限变 0644；软链接被替换成普通文件 | 设置里改任意项 | `internal/sys/write.go:12-45` | 预期复现 |
| L-13 | 窗口尺寸/位置漂移，或平铺 WM 下反复跳位 | 自定义/平铺 WM | `internal/app/size.go:112`、`:15` | 可能复现 |
| L-14 | 「窗口置顶」无效果 | Wayland | — | 预期无效（合成器限制） |
| L-15 | 窗口完全没有边框与标题栏，最小化/最大化/关闭点不到 | X11（XWayland，KWin 不装饰该窗口） | `internal/app/app.go:115`（`TitleBarStyle: mygo.TitleBarHidden`） | 预期复现（首轮验证中实测复现） |

## 7. 验证结果

首轮环境：Arch Linux（滚动）· 内核 7.2.15 · Go 1.27.3 · gtk3 1:3.24.52 · pango 1:1.58.2 · cairo 1.18.6 · glib2 2.90.1 · KDE Plasma（Wayland 会话 `wayland-1`，X11 走 XWayland `:1`）· noto-fonts-cjk / Source Han Sans。构建链路：`go build`、`go vet`、`gofmt`、`go test -race ./...`（9 个包全过，20s）与 `go tool mygo build -platform linux/amd64 -skip-dmg -skip-installer` 均通过。

| 编号 | 结果 | 证据 | 备注 |
| --- | --- | --- | --- |
| L-1 | 复现 → 已修 | 截图 | `SHELL=` 空时不自动开标签，但界面不是全空：显示「没有打开的标签」+「新建标签」按钮。已修：`internal/sys/shell_other.go` 现在按 `$SHELL` → `/etc/shells`（只收认识的 shell，跳过 git-shell/rbash 一类）→ `/bin/sh` 兜底选择 |
| L-2 | 复现 → 已修 | 设置页多出一行 `exec: "explorer.exe": executable file not found in $PATH` | 不是「无任何反应」，错误进了 note 行；功能仍不可用。已修：新增 `sys.OpenPath`（POSIX 用 xdg-open，Windows 用 explorer.exe），见 `internal/sys/open_other.go`、`internal/sys/open_windows.go` |
| L-3 | 未验证 | — | 本机有网，`libghostty-vt.so` 自动下载（1.6s）；断网/墙内场景未测 |
| L-4 | 不复现 | 截图 | Wayland 下 mygo 自绘标题栏，最小化/最大化/关闭都在 |
| L-5 | 复现 → 已修（随 L-15） | 同 L-15 | 条带右侧约 80px（`bar.Right` 让位）无任何可点/可拖内容；X11 下窗口按钮本身没画出来，整块成死区。已修：标题栏交给平台后 `c.TitleBar()` 归零，让位区随之消失 |
| L-6 | 不复现 | 截图 | cairo 1.18.6，文字颜色正常，无发白/绿边 |
| L-7 | 不复现 | 截图 | 中文与框线字符正常 |
| L-8 | 未验证 | — | 需有滚动内容时手测 |
| L-9 | 复现 → 已修 | `internal/profile/profile.go:116-128` | Linux 的 profile id（`zsh`）不在颜色表内，圆点为灰色。已修：`internal/profile/profile.go` 的 `colorOf` 现在认 bash/zsh/fish/sh/nu/dash/ksh/tcsh/pwsh |
| L-10 | 部分复现 → 已修 | 截图 | 「起始目录」占位符是 `%USERPROFILE%`（`internal/ui/settings.go:151` 硬编码）；「默认配置文件」显示「无」，未见 `pwsh.exe`。已修：新增 `internal/sys/expand.go` 的 `sys.ExpandPath`（`~`、`%VAR%`、`$VAR`、`${VAR}`），起始目录在启动时展开；设置页占位符改用 `internal/store/defaults_other.go` / `defaults_windows.go` 的 `DefaultHomeHint` / `DefaultFontFamily` / `DefaultShellHint` |
| L-11 | 复现（已说明，未修复） | `ps` 进程树 | `kill -9` 主进程后仍有 `wt-go` 进程存活（被 reparent 到 init），需再杀一次；`SIGTERM`/关窗则干净退出、无残留。现状：靠内核 hangup 与 mygo 关终端时对进程组发 SIGKILL，残留只发生在忽略 SIGHUP 或用 setsid 逃逸的进程；mygo 没有注入 `SysProcAttr`（Pdeathsig）的扩展点，见 `internal/sys/job_other.go` 注释 |
| L-12 | 复现 → 已修 | `ls -l ~/.config/wt-go/settings.json` → `0644` | 软链接场景未测。已修：`internal/sys/write.go` 现在保留既有文件权限，并跟随软链接写入目标 |
| L-13 | 不复现 | `xdotool windowsize 1200 700` → 重启仍 1200x700 | 窗口位置/尺寸存于 `~/.config/wt-go/window-state.json`；平铺 WM 未测 |
| L-14 | 未验证 | — | Wayland 下未测置顶 |
| L-15 | 复现 → 已修 | 截图 | `GDK_BACKEND=x11` 下窗口完全无边框无标题栏；a11y 树里存在最小化/最大化/关闭，但屏幕上不可见；拖动条带也不移动窗口。已修：`TitleBarStyle` 改为按平台选择（`internal/app/titlebar_linux.go` / `titlebar_other.go`），Linux 交给平台画标题栏（Wayland 由合成器、X11 由窗口管理器），条带右侧让位区随之消失；X11 后端仍不是主要支持目标。修复后复测：X11 窗口有了 KWin 的标题栏与按钮（`xdotool getwindowgeometry` 报 960x572 的内容区，位置 0,28）；用 `xdotool` 合成的拖动在本会话里既拖不动标题栏也拖不动条带，属 XTEST→KWin 的限制，真鼠标拖动未验证 |

复现时的补充（非缺陷清单内）：Wayland 与 X11 两种后端共存于同一会话时窗口外观差异很大，X11 下几乎不可用（无边框、无法拖动、按钮不可见），这条已随 L-15 一起处理：Linux 改用 `mygo.TitleBarDefault`，由平台画标题栏。X11 仍是次要后端，值得在 X11 会话里再跑一遍上面这张表。

## 8. 反馈时附上这些

```sh
uname -a
echo "$XDG_SESSION_TYPE $XDG_CURRENT_DESKTOP $WAYLAND_DISPLAY"
echo "SHELL=$SHELL"
ldd --version | head -1
fc-list :lang=zh | head -3
ls -l ~/.config/wt-go 2>/dev/null
cd wt-go && ./wtgo            # 前台跑一遍，把输出一起发
```
