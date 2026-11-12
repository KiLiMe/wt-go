# Linux / macOS（实验性）

Windows 是主要支持与验证目标；Linux 与 macOS 有 pty 与平台路径实现，可交叉编译，属实验性。
这里讲 Linux 上的依赖、装法与已知限制；Windows 的用法见 [使用说明.md](使用说明.md)，
构建与发布见 [代码导读.md](代码导读.md) 末节。


首轮真机验证在 Arch Linux + KDE Plasma（Wayland `wayland-1`，X11 走 XWayland `:1`）完成：
构建、`go test -race`、`mygo build` 打包（deb / tar.gz / PKGBUILD）、启动、终端与设置页交互均可用。
发布只取其中的便携 `zip`，`.deb` 与 `PKGBUILD` 不随发布走。

## 1 运行库依赖

mygo 在 Linux 上通过 purego 直接调用已加载的 GTK 3 / Pango，**不用 cgo，也不需要 WebKitGTK**。

| 库 | Debian / Ubuntu 包 | 缺失后果 |
| --- | --- | --- |
| `libgtk-3.so.0`、`libgdk-3.so.0` | `libgtk-3-0` | 窗口打不开，进程退出 |
| `libglib-2.0.so.0`、`libgobject-2.0.so.0`、`libgio-2.0.so.0` | `libglib2.0-0` | 同上 |
| `libcairo.so.2` | `libcairo2` | 同上 |
| `libgdk_pixbuf-2.0.so.0` | `libgdk-pixbuf-2.0-0` | 同上 |
| `libpango-1.0.so.0`、`libpangocairo-1.0.so.0` | `libpango-1.0-0`、`libpangocairo-1.0-0` | **窗口能开，但没有任何文字** |
| `libharfbuzz.so.0`、`libpangoft2-1.0.so.0` | `libharfbuzz0b`、`libpangoft2-1.0-0` | 可选；缺则退回简单排版 |

另外需要 **CJK 字体**（Debian `fonts-noto-cjk`、Fedora `google-noto-sans-cjk-fonts`、
Arch `noto-fonts-cjk`），否则中文与框线字符是豆腐块；**glibc ≥ 2.28**；
输入法要设 `GTK_IM_MODULE`（fcitx5 用 `fcitx`，ibus 用 `ibus`），否则中文打不进去。

## 2 安装与运行

```sh
# Debian / Ubuntu
sudo apt install libgtk-3-0 libpango-1.0-0 libpangocairo-1.0-0 fonts-noto-cjk
# Fedora
sudo dnf install gtk3 pango pangocairo google-noto-sans-cjk-fonts
# Arch
sudo pacman -S gtk3 pango noto-fonts-cjk

go build -o wt-go .         # 裸二进制：没有图标、版本号
./Scripts/build.sh          # 日常构建走它：与发布版一致，mygo 注入图标、版本与 libghostty-vt.so
./wt-go                     # 前台跑：原生库失败只写 stdout 一行 log
GDK_BACKEND=x11 ./wt-go     # 指定后端；不设则自动选，两种都要试
xvfb-run -a ./wt-go         # 无头冒烟：只验证"窗口能开、不崩"
```

数据目录 `~/.config/wt-go/`（`settings.json`、`wt-go.log`、`crash-<yyyyMMdd-hhmmss>.log`）。

## 3 已知限制

Linux 上标题栏也由程序画（`TitleBarHidden` + `HideWindowButtons`），GTK 不再建 header bar，
所以桌面主题影响不到它；代价是 `gtk-decoration-layout` 被清空时（平铺窗口管理器常见）与
「合成器自己画标题栏」无法区分，两种情形都不画按钮，窗口只能靠快捷键或窗口管理器操作。
X11 与 Wayland 都用同一套自绘标题栏；X11 上窗口由 GTK 自装饰（阴影、缩放边）。
**自绘标题栏这次改动**只在 Windows 上编译并跑过单测，Linux 真机还有两项待验：按钮位置与拖动
在 X11、Wayland（GNOME 与 KWin）各一遍。Wayland 下「窗口置顶」受合成器限制可能无效。
`kill -9` 主进程后 shell 可能残留一次
（内核 hangup 与 mygo 关终端时的 SIGKILL 之外，mygo 没有注入 `Pdeathsig` 的扩展点）。

文字颜色靠启动时给 fontconfig 追加一条 `rgba=none` 规则保住（`internal/sys/textcolor_linux.go`）：
mygo 先把字形按灰度光栅化，再把「通道不一致」的字形当成彩色字形（emoji）原样绘制，
而 cairo 1.17.8 之前会在字体声明了次像素顺序时改做次像素光栅化，于是整片文字变成
白底绿 alpha。规则只去掉字体的次像素顺序，不改 GTK 设置、不写用户的 fontconfig 配置；
代价是这些字体不再做次像素抗锯齿（灰阶，观感略软）。上游自己强制灰度之后这段可以删掉。

## 4 反馈时附上

```sh
uname -a
echo "$XDG_SESSION_TYPE $XDG_CURRENT_DESKTOP $WAYLAND_DISPLAY"
echo "SHELL=$SHELL"
ldd --version | head -1
fc-list :lang=zh | head -3
ls -l ~/.config/wt-go 2>/dev/null
cd wt-go && ./wt-go            # 前台跑一遍，把输出一起发
```
