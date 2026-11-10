# Linux（实验性）

现状：Linux/macOS 的平台分片（pty、配置目录）已交叉编译通过（`GOOS=linux GOARCH=amd64|arm64`、`go vet` 均无错），但**从未在 Linux 上运行过**，也没有 Linux CI。本文是首轮真机验证的手册：依赖、启动方式、逐条待验证的已知缺陷、以及结果记录方式。

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
| L-8 | `Ctrl+滚轮` 改的是字号、不是滚动 | 任意 | `internal/ui/keys.go:255-263`（`Cmd` 在 Linux = Ctrl） | **设计如此**，记录手感即可 |
| L-9 | 标签上的 profile 圆点没有颜色 | 任意 | `internal/profile/profile.go:116-128`（只认 pwsh7/powershell/cmd/wsl） | 预期复现 |
| L-10 | 设置里显示 `%USERPROFILE%`、默认 shell 是 `pwsh.exe` | 设置页 | `internal/ui/settings.go:151`、`:294`、`:300` | 预期复现 |
| L-11 | 主进程被 `kill -9` 后 shell 仍在运行 | 强杀 | `internal/sys/job_other.go:7-8`（空实现） | 预期复现；正常关窗应无残留 |
| L-12 | 保存设置后 `settings.json` 权限变 0644；软链接被替换成普通文件 | 设置里改任意项 | `internal/sys/write.go:12-45` | 预期复现 |
| L-13 | 窗口尺寸/位置漂移，或平铺 WM 下反复跳位 | 自定义/平铺 WM | `internal/app/size.go:112`、`:15` | 可能复现 |
| L-14 | 「窗口置顶」无效果 | Wayland | — | 预期无效（合成器限制） |

## 7. 记录方式

每条一次，填完即是一份可执行的验收清单：

| 编号 | 环境（发行版 / 后端 / 桌面） | 结果 | 证据 | 备注 |
| --- | --- | --- | --- | --- |
| L-1 |  | 复现 / 不复现 | 截图 / log 行 |  |

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
