# MCP

`wt-go mcp` 是一个 Model Context Protocol 服务器：把 `wt-go diag` 看到的那份报告按 MCP 的
工具与资源交出去，让编辑器或 agent 自己读现场，不用人复制粘贴窗口。

## 1. 它是什么，不是什么

- **只读**：五个工具、两个资源读的都是同一份 `cli.Collect()`，与 `wt-go diag` 完全同源；
  不写文件、不改设置、不驱动窗口。
- **不碰窗口**：不要求 wt-go 正在运行，可以和窗口同时跑，也可以完全不开窗口。
- **不走网络**：JSON-RPC 2.0 走 stdin/stdout，没有端口，也没有 socket。
- **不加依赖**：协议层 `internal/mcp` 只用标准库，exe 仍是单文件。

## 2. 接到客户端

```json
{
  "mcpServers": {
    "wt-go": {
      "command": "D:\\path\\to\\wt-go.exe",
      "args": ["mcp"]
    }
  }
}
```

Linux / macOS 把 `command` 换成 wt-go 的路径，`args` 不变。Claude Desktop、Cursor、VS Code 等
认 `mcpServers` 这一段的客户端通用；`wt-go mcp` 后面的参数一律忽略。

`initialize` 的应答里 `serverInfo` 是 `{"name":"wt-go","version":"<报告的 version 字段>"}`
（不带 `wt-go version` 里 `(packaged …)` 那一段），另有一段 `instructions`：先调 `wtgo_diag`，
失败项带 E1xx / E2xx / E3xx 编号，编号的解释在 [故障排查.md](故障排查.md)。

## 3. 五个工具

| 工具 | 输入 | 回答 |
| --- | --- | --- |
| `wtgo_diag` | 无 | `wt-go diag` 的文本：报告的行，加上每个失败检查的 `fault.Line`（编号 + 细节）。**先调这个** |
| `wtgo_status` | 无 | 整份报告的 JSON：`program` / `version` / `revision` / `packaged` / `go` / `platform` / `executable` / `settings` / `native` / `shells` / `profiles` / `crash_logs` / `log` / `checks` |
| `wtgo_settings` | 无 | `settings` 与头两项检查（配置目录、`settings.json`）：路径、是否存在、能否解析、语言、主题、配色、默认 profile、键位数 |
| `wtgo_profiles` | 无 | `shells` 与合并后的 `profiles`：这台机器有哪些 shell，每个 profile 用什么命令、能不能真正启动 |
| `wtgo_log` | `lines`（1–2000，默认 120）、`path`（`wtgo_status` 里看到的某份日志，留空取最新的崩溃日志） | 日志末尾若干行；文件超过 1 MiB 时只读尾部并标一行 `(only the end of the file)` |

工具自己失败（比如一条崩溃日志都没有）时按 MCP 的约定回 `isError: true` 的 content，
而不是 JSON-RPC 错误，例如 `no crash log: wt-go has not stopped with a panic`。

## 4. 两个资源

| URI | MIME | 内容 |
| --- | --- | --- |
| `wtgo://diag` | `text/plain` | 同 `wtgo_diag` |
| `wtgo://status` | `application/json` | 同 `wtgo_status` |

不支持 `resources/subscribe`，也不发 `listChanged`：工具与资源是固定的，与窗口开没开无关。

## 5. 协议细节

- 版本 `2025-06-18`：`initialize` 里客户端说自己是谁都行，服务器一律答自己的版本。
- 能力：`tools`；有资源时再加 `resources`（`subscribe:false`、`listChanged:false`）。
- 方法：`initialize`、`ping`、`tools/list`、`tools/call`、`resources/list`、`resources/read`、
  `shutdown`；通知没有 `id`，一律不回包（`notifications/initialized`、`notifications/cancelled`
  直接吞掉）。`shutdown` 只回一个空对象，不结束进程，收尾靠客户端关掉 stdin。
- 报文是**一行一个 JSON 对象**（新行分隔），不是 LSP 那套 `Content-Length` 头。
- 错误码：方法不认识 `-32601`、参数不合法或工具/资源名不认识 `-32602`、读资源出错 `-32603`。
- 一行报文坏了只在 stderr 记一行（`wt-go mcp: bad message: …`）然后继续，结束会话的是 stdin 的 EOF。
- 退出码：stdin 关闭返回 0；写回 stdout 出错返回 1。诊断出问题不算命令失败，所以仍是 0。

## 6. 手测

不接客户端也能问它，一个请求一行：

```pwsh
'{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | .\build\windows-amd64\wt-go.exe mcp
```

```sh
printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"wtgo_diag"}}' |
  ./build/linux-amd64/wt-go mcp
```

## 7. 隐私

工具读的是这台机器上的真实路径：`--redact` 只作用在 `wt-go diag` 上，MCP 的输出不脱敏。
贴进 issue 时用 `wt-go diag --redact`；把 `wt-go mcp` 接到远端 agent 之前先想清楚这些路径会去哪。

排查流程见 [故障排查.md](故障排查.md) 第 2 节，代码位置见 [代码导读.md](代码导读.md)。
