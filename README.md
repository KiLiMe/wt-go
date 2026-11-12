# wt-go

Windows Terminal 风格的终端：多标签、分栏、profile、配色方案与可录制快捷键，设置是一份 `settings.json`。

## 快速开始

```pwsh
git clone https://github.com/KiLiMe/wt-go.git; cd wt-go   # 1 取源码
pwsh ./Scripts/build.ps1                                  # 2 构建（Linux/macOS：./Scripts/build.sh）
./build/windows-amd64/wt-go.exe                           # 3 运行
```

发布页每个平台一个便携 zip：Windows 是 `wt-go.exe`、`ghostty-vt.dll` 与 `LICENSE`，Linux 是 `wt-go`、`libghostty-vt.so` 与 `LICENSE`；解压即用，没有安装器、也没有 `.deb`。
构建脚本参数：`-Test` 先跑 vet 与测试，`-Clean` 清掉 `build/`，`-Package` 只出 `dist/`。

## 依赖环境

| 项 | 要求 | 出处 |
| --- | --- | --- |
| Go | 1.27.3 | `go.mod` 的 `go` 行，CI 用 `go-version-file: go.mod` |
| mygo | v0.8.3（提供 `go tool mygo`、UI 与终端插件） | `go.mod` 的 `require` 与 `tool` 行 |
| purego / golang.org/x/sys | v0.11.1 / v0.48.0 | `go.mod` 的 `require` |
| Windows | 10 1809+ / 11，需要 ConPTY | 主要支持与验证目标 |
| Linux | glibc ≥ 2.28；GTK 3、Pango、Cairo 运行库；CJK 字体 | [docs/Linux.md](docs/Linux.md) 第 1 节 |
| macOS | 能编译，CI 只跑测试、不打包 | `.github/workflows/ci.yml` 的构建矩阵 |
| 构建工具 | 无额外依赖；PowerShell 7+（Windows）或 bash（Linux/macOS） | `Scripts/build.ps1` / `Scripts/build.sh` |

## 贡献指引

- 提交前本地跑一遍：`gofmt -l .`（必须无输出）、`go vet ./...`、`go test ./...`；
  或 `pwsh ./Scripts/build.ps1 -Test` 一次做完。
- 改行为的提交要同时改文档：`docs/使用说明.md`、`docs/配置参考.md`、`docs/代码导读.md`、
  `docs/decisions.md`（新决策写 ADR）、`docs/plan.md`。
- 包依赖方向是硬规则，由 `internal/app/deps_test.go` 用 `go list` 校验；新增包或改 import 要同步那里的规则表
  和 `docs/代码导读.md` 的依赖图，否则 `go test ./internal/app` 失败。
- 文案只有 `locales/en.json` 与 `locales/zh-Hans.json` 两份，改完跑 `go tool mygo generate`（写出
  `locales/catalog_gen.go`，生成物提交）与 `go tool mygo i18n check`。
- 注释写英文、文档写中文；提交信息用 `feat:` / `fix:` / `docs:` 前缀。
- 报 issue 附上 `wt-go diag --redact` 的输出：个人路径会换成 `%USERPROFILE%` 之类的占位。
- 其余约定见 [AGENTS.md](AGENTS.md)。
