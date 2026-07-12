# go-agent

基于 Go + Anthropic Claude API 实现的命令行Coding Agent。

它在终端中与 Claude 进行多轮对话，模型可自主调用 `bash`、文件读写等工具来完成用户交办的编程任务——读取代码、执行命令、修改文件，直到任务完成。
## 快速开始

### 环境要求

- Go 1.26+
- 可访问的 Claude 兼容 API 端点与 API Key

### 配置环境变量

通过环境变量提供模型与鉴权信息：

```bash
export URL="https://api.anthropic.com"          # 或兼容端点
export API_KEY="sk-ant-..."                      # 你的 API Key
export MODEL="claude-3-5-sonnet-20241022"        # 模型名
export LOG_LEVEL=debug                           # 可选，开启调试日志
```

### 构建与运行

```bash
# 构建
go build -o build/go_agent .

# 运行
./build/go_agent
```

或直接运行：

```bash
go run .
```

### 使用

启动后进入交互式 REPL：

```
Welcome to Go Agent! Type `/exit` to quit.
User >> 帮我看看当前目录有哪些 Go 文件
Agent:
 ...
User >> /exit
Bye!
```

输入 `/exit` 退出。模型会自行决定调用哪个工具来完成任务。

## 配置项

常量定义于 `common/consts/`，可按需调整：

| 常量 | 默认值 | 说明 |
| --- | --- | --- |
| `MaxTokens` | 10000 | 单次响应最大 token |
| `RequestTimeout` | 90s | 单次 API 请求超时 |
| `MaxRequestTries` | 3 | 可重试错误的最大重试次数 |
| `RetryDelay` | 500ms | 重试基础延迟（按次数递增） |
| `BashTimeout` | 120s | bash 命令执行超时 |

## 路线图

见 [`docs/TODO.md`](docs/TODO.md)：

## 许可

本项目采用 [MIT License](LICENSE) 开源协议，版权所有 © 2026 Yujie Zhou。
