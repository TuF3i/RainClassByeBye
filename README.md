# RainClassByeBye

长江雨课堂 AI 自动答题 CLI，基于 [Auto-CQUPT-Plan/RainClassSDK](https://github.com/Auto-CQUPT-Plan/RainClassSDK) 构建。

## 免责声明

本项目仅供学习和研究使用。请自行评估使用风险，并遵守学校、课程平台和相关法律法规。

## 功能

- `cobra` 构建 CLI
- `fatih/color` 彩色日志
- RainClassSDK cookie 持久化，默认保存在 `cache/cookies.json`
- 基于 Eino 的 OpenAI-compatible ChatModel 自动答题
- 默认模型为 `qwen3.7-plus`
- 支持任务中断恢复
- 默认使用 20 个 goroutine worker 并发求解
- 提供用户、课程、作业信息获取命令

## 依赖

- Go
- 雨课堂账号
- DashScope API Key，默认从环境变量 `DASHSCOPE_API_KEY` 读取

## 快速开始

```bash
git clone <your-repo>
cd RainClassByeBye

export DASHSCOPE_API_KEY="your_api_key"

go run . login
```

登录成功后，cookie 会持久化到：

```text
cache/cookies.json
```

## 构建

```bash
go build -o rainclass-bye-bye .
```

也可以直接用：

```bash
go run . <command>
```

## 命令总览

```bash
rainclass-bye-bye login
rainclass-bye-bye info ...
rainclass-bye-bye run ...
rainclass-bye-bye resume ...
rainclass-bye-bye status ...
```

全局参数：

```bash
--cookie-path cache/cookies.json
--state-dir cache/state
```

## 登录

终端打印微信二维码，扫码后保存登录态：

```bash
go run . login
```

## 信息获取

这部分命令是按 `dev/unity_test.go` 里已有的 SDK 能力整理出来的。

### 用户信息

```bash
go run . info user
```

### 课程列表

```bash
go run . info courses
```

### 作业列表

```bash
go run . info homework --cid <classroom_id>
```

### 作业详情

```bash
go run . info homework --cid <classroom_id> --leaf-id <leaf_id>
```

## 自动答题

开始新的自动答题任务：

```bash
go run . run --cid <classroom_id> --exam-id <exam_id>
```

常用参数：

```bash
--api-key <key>
--api-key-env DASHSCOPE_API_KEY
--base-url https://dashscope.aliyuncs.com/compatible-mode/v1
--model qwen3.7-plus
--workers 20
--temperature 0.1
--request-timeout 2m
--max-completion-tokens 2048
--submit-paper
```

默认行为：

- 会先调用 `StartExam`
- 会并发请求模型，但逐题提交答案
- 默认只提交每道题，不自动交卷
- 只有显式加 `--submit-paper` 才会最终交卷

示例：

```bash
go run . run \
  --cid 24211265 \
  --exam-id 1945538 \
  --workers 20 \
  --submit-paper
```

## 恢复任务

如果程序异常退出、网络中断或你手动 `Ctrl+C`，可以恢复：

```bash
go run . resume --cid <classroom_id> --exam-id <exam_id>
```

状态文件默认保存在：

```text
cache/state/<cid>_<exam_id>.json
```

状态文件会记录：

- 已提交题目
- 失败题目
- 试卷标题
- 最后错误
- 是否已经交卷

## 查看任务状态

```bash
go run . status --cid <classroom_id> --exam-id <exam_id>
```

## 模型配置

当前默认使用 DashScope 的 OpenAI-compatible 接口：

```text
Base URL: https://dashscope.aliyuncs.com/compatible-mode/v1
Model:    qwen3.7-plus
```

如果你要切换模型或兼容接口，可以直接覆盖：

```bash
go run . run \
  --cid <classroom_id> \
  --exam-id <exam_id> \
  --base-url <your-openai-compatible-base-url> \
  --model <your-model> \
  --api-key <your-api-key>
```

## 测试

```bash
go test ./...
```

`dev/unity_test.go` 是参考资料，默认通过 build tag 排除；如果你要单独编译那份参考测试：

```bash
go test -tags devref ./dev
```

## 目录说明

```text
cmd/                 CLI 命令
internal/logging/    彩色日志
internal/runner/     自动答题执行器
internal/solver/     Eino 模型求解器
internal/state/      状态持久化与恢复
dev/                 参考资料
```
