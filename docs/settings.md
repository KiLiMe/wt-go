# 设置

配置文件：Windows 为 `%LOCALAPPDATA%\wt-go\settings.json`，Linux/macOS 为系统配置目录下的 `wt-go/settings.json`（UTF-8、无 BOM）。
首次启动时若不存在则写入默认值。改完文件会被监视并热重载（约一秒内生效）。
写入采用临时文件、同步后原子替换，避免留下半截配置。解析失败或配置越界时继续使用当前可运行值，并将坏内容归档为 `settings.json.bad-<YYYYMMDD-HHMMSS>`（重名时追加序号），在设置页提示原因。设置界面（`Ctrl+,`）读写同一份文件，不另存一份。

下方示例中的 `//` 只是说明，实际 `settings.json` 必须是纯 JSON，不能直接复制带注释的内容。

## 完整示例

```json
{
  "language": "zh-CN",
  "theme": "dark",

  // 启动
  "defaultProfile": "pwsh7",        // profile 的 id
  "startOn": "default",             // default | last（尚未生效）
  "startDirectory": "",             // 空 = profile 自己的
  "startSize": { "columns": 0, "rows": 0 }, // 首屏网格；0 = 窗口默认大小

  // 外观
  "font": { "family": "Cascadia Mono", "size": 13, "lineHeight": 1.2 },
  "colorScheme": "Campbell",
  "padding": { "x": 12, "y": 8 },
  "backgroundOpacity": 1.0,         // 窗口整体透明度，0.3 – 1.0
  "cursorShape": "bar",             // bar | block | underline
  "cursorBlink": true,
  "scrollbackMB": 10,
  "newTabPosition": "last",         // last | first
  "tabWidth": "compact",            // compact（随名字）| equal（等宽）
  "alwaysOnTop": false,
  "titleFromActiveTerminal": false, // true = 标题跟随有焦点的窗格

  // 行为
  "copyOnSelect": false,
  "rightClick": "copyOrPaste",
  "pasteMultilineConfirm": true,    // 多行粘贴先问一句
  "confirmClose": true,             // 有子进程在跑时关标签先确认
  "closeOnLastTab": true,

  // 分栏
  "maxPanes": 4,                    // 每个标签最多的窗格总数

  // 配色方案（可覆盖内置）
  "schemes": [
    {
      "name": "My Dark",
      "background": "#0c0c0c",
      "foreground": "#cccccc",
      "cursor": "#ffffff",
      "selection": "#264f78",
      "palette": [
        "#0c0c0c", "#c50f1f", "#13a10e", "#c19c00",
        "#0037da", "#881798", "#3a96dd", "#cccccc",
        "#767676", "#e74856", "#16c60c", "#f9f1a5",
        "#3b78ff", "#b4009e", "#61d6d6", "#f2f2f2"
      ]
    }
  ],

  // profile
  "profiles": [
    {
      "id": "pwsh7",
      "name": "PowerShell 7",
      "command": "pwsh.exe",
      "args": ["-NoLogo"],
      "dir": "",
      "color": "#3b78ff",
      "hidden": false
    }
  ],

  // 快捷键：按键 → 动作，沿用 VS Code 风格字符串；同一动作写多次＝多个键
  "keybindings": [
    { "key": "ctrl+shift+t", "action": "tab.new" },
    { "key": "ctrl+shift+w", "action": "pane.close" },
    { "key": "alt+shift+d", "action": "pane.split.auto" },
    { "key": "alt+left", "action": "pane.focus.left" }
  ]
}
```

已实现的动作名：`tab.new`、`tab.close`、`tab.next`、`tab.prev`、`tab.duplicate`、`tab.goto.N`（N 从 1 起），
`pane.split.auto`、`pane.split.right`、`pane.split.down`、`pane.focus.left`、`pane.focus.right`、
`pane.focus.up`、`pane.focus.down`、`pane.close`、`pane.closeOthers`，
以及 `settings.open`（打开设置标签页）和 `profile.N`（开第 N 个 profile 的标签，N 从 1 起，按菜单里的顺序）。
写了未知动作或解析不了的按键，该条不会执行动作；未知动作仍可能接管按键。需要把按键交给终端时，请删除该条，而不是写 `none`。

`keybindings` 是整张表：写了就完全取代默认键位（不是“在默认之上叠加”）。
从旧版升级时，旧文件里没有 `pane.*` 那几行，删掉 `keybindings` 这一项或整个文件重新生成即可。

## 字段说明

### 路径与优先级

`maxPanes` 表示单个标签内允许存在的窗格总数上限；`4` 即最多 4 个窗格，不表示 4×4 网格；`0` 表示不限。

1. 命令行参数（目前尚未实现）
2. `settings.json` 显式值
3. 内置默认值

配置文件中的数值会经过范围校验；越界时整份文件被拒绝，继续使用当前值并提示原因。常用范围：字体大小 4–200、行高 0.5–4、内边距 0–256、透明度 0–1、滚动缓冲 0–1024 MB、窗格数 0–64、启动尺寸列/行 0–1000。设置页 +/- 控件只覆盖更窄的常用范围，手写超出控件范围的值后点击 +/- 可能将其夹到控件边界。

### 关闭时的确认

`confirmClose` 为真时，关标签（`Ctrl+Shift+W` 关到最后一个窗格、✕、右键菜单）会先问一句，
因为里面的 shell 还在跑。关窗格（标签里还有别的窗格时）不问；关其它标签一次问一句。
点窗口的 ✕ 目前不问，未保存的设置会丢弃；只有关闭设置标签时会询问。mygo 虽提供可取消关闭回调，但当前 wt-go 尚未接入。

### 启动大小

`startSize` 用列×行表示首屏网格。终端只报告“能放下多少格”而不报告一格多大，
所以窗口启动时先长一次、读出差值算出单元格尺寸，再设成恰好放下这些格的大小（约两三帧内完成）。
`0` 表示不设，用窗口默认大小。

### 自动发现的 profile

启动时扫描并补进 `profiles`（仅在用户没写同 `id` 时）：

- `pwsh.exe`：WHERE 找到的 PowerShell 7
- `powershell.exe`：Windows PowerShell 5.1
- `cmd.exe`：命令提示符
- `wsl -l -q` 的每个发行版 → `wsl.exe -d <distro>`

自动发现的 profile 有 `"discovered": true`，用户可改名字/颜色，删除后不再自动加回。
发现结果只补进内存（`internal/profile.Resolve`），不写回 `settings.json`。

### 保存与热重载

保存前会检查文件是否被外部修改；如果已修改，保存会被拒绝并提示，需先放弃当前草稿、重新载入文件后再保存。设置页有未保存改动时，外部文件变化不会覆盖草稿，只会提示需要先放弃改动。

字体、配色、选中即复制、窗口透明度与置顶、语言和快捷键对已打开的终端立即生效；内边距、光标形状/闪烁、滚动缓冲和右键行为主要在新建标签或窗格时生效。

### 配色方案

`colorScheme` 按以下顺序找名字：

1. `schemes` 里同名的自定义方案（只写部分字段时，没写的沿用同名内置方案的颜色）；
2. 内置方案（下表）；
3. Ghostty 的 **617 套**主题，名字照抄即可（`Dracula`、`Gruvbox Dark`、`One Dark Two`……）；
4. 都没有 → 回退到 `theme` 对应的默认配色（暗：Visual Studio Code 调色板；亮：同源亮色）。

`theme` 只管界面（标签栏、窗口底色），终端配色只看 `colorScheme`，与 Windows Terminal 的分工一致。

### 配色方案内置清单

Campbell、Campbell Powershell、One Half Dark、One Half Light、Solarized Dark、
Solarized Light、Tango Dark、Tango Light、Vintage。

其中 One Half Dark/Light 直接用 Ghostty 的同名主题；其余七套由 `internal/theme/builtin.go` 写入（取值同 Windows Terminal 的 `defaults.json`）。

## 与 Windows Terminal 的差异

- WT 的 `settings.json` 结构（`profiles.list` / `schemes` / `actions` 数组）不兼容，
  wt-go 用自己的扁平结构，迁移需手抄。
- 不做 `defaults.json` 两层合并。
- `rightClick` 的忽略值是 `nothing`，不是 `ignore`。
- `language` 当前支持 `zh-CN` 与 `en-US`；其它值按英文处理。
