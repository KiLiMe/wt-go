# 快捷键

默认值与 Windows Terminal 对齐；可被 `settings.json` 的 `keybindings` 覆盖。

已生效（阶段 3）：`tab.new` / `tab.close` / `tab.next` / `tab.prev` / `tab.duplicate` / `tab.goto.N`，
`pane.split.auto` / `pane.split.right` / `pane.split.down` / `pane.focus.left|right|up|down` /
`pane.close` / `pane.closeOthers`，以及 `settings.open` / `profile.N`；复制粘贴由终端插件自带的默认键位处理。
其余键位在对应阶段接入。

## 标签

| 键 | 动作 |
| --- | --- |
| `Ctrl+Shift+T` | 新建标签（默认 profile） |
| `Ctrl+T` | 新建标签（默认 profile，备用键） |
| `Ctrl+Tab` / `Ctrl+Shift+Tab` | 下一个 / 上一个标签 |
| `Ctrl+1` … `Ctrl+9` | 切到第 n 个标签 |
| `Ctrl+Shift+D` | 复制当前标签（同 profile） |
| `Ctrl+Shift+1` … `Ctrl+Shift+9` | 用第 n 个 profile 新建标签（与 ▾ 菜单里的顺序一致） |

## 设置

| 键 | 动作 |
| --- | --- |
| `Ctrl+,` | 打开设置标签页（已打开则切过去） |

设置页「操作」分组可以录制新键：点一下键位框，按下组合键即可；与其它动作冲突时那一行标红。
录制会监听整张键名表（字母、数字、`F1`–`F12`、方向键、`comma`/`minus`/`equal` 等），
所以 `ctrl+comma` 这种键也能录。

## 分栏

`Ctrl+Shift+W` 默认绑的是 `pane.close`（关窗格，WT 的行为），最后一个窗格关掉就关标签；
关整个标签走标签的 ✕ 与右键菜单。要改回“一键关整个标签”，在 `keybindings` 里把
`ctrl+shift+w` 绑到 `tab.close`（同一按键以 `keybindings` 里的为准）。

| 键 | 动作 |
| --- | --- |
| `Alt+Shift+D` | 自动切分（先竖后横） |
| `Alt+Shift+Right` | 纵向切分（左右） |
| `Alt+Shift+Down` | 横向切分（上下） |
| `Alt+左/右/上/下` | 焦点移到相邻窗格 |
| `Ctrl+Shift+W` | 关闭当前窗格（最后一个窗格时关标签） |
| `Alt+Shift+W` | 关闭标签内其它窗格 |

## 复制与粘贴

| 键 | 动作 |
| --- | --- |
| `Ctrl+Shift+C` / `Ctrl+Insert` | 复制选中 |
| `Ctrl+Shift+V` / `Shift+Insert` | 粘贴 |
| `Ctrl+Shift+A` | 全选 |
| `Ctrl+C` | 有选中时复制，否则发给程序（可配） |

## 视图

| 键 | 动作 |
| --- | --- |
| `Ctrl+滚轮` / `Ctrl+=` / `Ctrl+-` / `Ctrl+0` | 放大 / 缩小 / 重置字号 |
| `Ctrl+Shift+K` | 清屏 |
| `Shift+PageUp` / `Shift+PageDown` | 翻页 |
| `Ctrl+Shift+Home` / `Ctrl+Shift+End` | 滚到顶 / 底 |
| `Ctrl+Shift+F` | 搜索（阶段 4） |
| `Ctrl+Shift+P` | 命令面板（阶段 4） |
| `Ctrl+,` | 设置 |
| `F11` / `Alt+Enter` | 全屏 |
| `Ctrl+Shift+Space` | 打开下拉菜单 |

## 保留给程序

不在表内的键一律发给终端里的程序。`Ctrl+Shift+*` 组合默认由 wt-go 处理，
需要交给程序时（如 vim 的某些映射）用 `keybindings` 把它绑到 `none`。

## 字符串写法

`ctrl`（或 `control`）、`shift`、`alt`、`win`（或 `super`/`meta`）修饰键 + 键名：
`a`–`z`、`0`–`9`、`f1`–`f12`、`left` `right` `up` `down` `home` `end` `pageup` `pagedown`
`insert` `delete` `space` `tab` `enter`（或 `return`）`escape` `backspace` `minus`，
例如 `"ctrl+shift+t"`、`"alt+shift+d"`。

字符串解析不了（拼错、或用了尚未支持的 `plus`/`comma`/`period`/`f13`+）时，该条被忽略，
对应动作退回默认键，不影响启动。
