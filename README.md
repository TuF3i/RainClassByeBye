# RainClassByeBye

一个基于 Go 实现的长江雨课堂自动答题 CLI。项目负责登录、读取课程与作业信息、进入考试环境、调用兼容 OpenAI 接口的大模型生成答案，并按题提交到雨课堂考试系统。

本项目构建在 [Auto-CQUPT-Plan/RainClassSDK](https://github.com/Auto-CQUPT-Plan/RainClassSDK) 之上：`RainClassSDK` 负责雨课堂登录和接口调用，`RainClassByeBye` 在其上补齐命令行交互、LLM 求解、状态持久化、断点恢复和交卷流程。

## 项目介绍

项目当前提供 5 类能力：

- `login`：微信扫码登录，并把 cookie 持久化到本地。
- `info`：查看当前用户、课程列表、课程作业列表和单个作业详情。
- `run`：开始新的自动答题任务。
- `resume`：恢复异常中断或未完成的任务，并可选择最终交卷。
- `status`：查看状态文件中的执行进度、失败题目数和最后错误。

整体流程如下：

1. 通过 `RainClassSDK` 登录雨课堂并保存 cookie。
2. 读取课程和作业信息，拿到 `cid`、`leaf_id`、`exam_id`。
3. `run` 命令进入考试环境，拉取试卷题目。
4. `internal/solver` 把题目 JSON 和图片 URL 发给大模型，要求只返回结构化 JSON。
5. `internal/runner` 并发求解、逐题提交、落盘状态。
6. 全部题目提交完成后，根据参数决定是否调用 `submit_paper` 交卷。

## 项目结构

```text
.
├── main.go                  # CLI 入口
├── cmd/                     # Cobra 命令定义与参数绑定
│   ├── login.go
│   ├── info*.go
│   ├── run.go
│   ├── resume.go
│   ├── status.go
│   └── helpers.go
├── internal/
│   ├── logging/             # 彩色终端日志
│   ├── runner/              # 自动答题调度、提交与交卷
│   ├── solver/              # 大模型调用、答案解析与兜底修复
│   └── state/               # 状态文件读写与断点恢复
├── dev/                     # 参考测试与 RainClassSDK 调用样例
├── go.mod
├── go.sum
└── LICENSE
```

## 依赖

### 运行依赖

- Go
  - 当前仓库 `go.mod` 声明为 `go 1.26.4`，请以仓库里的 `go.mod` 为准。
- 长江雨课堂账号
- 一个兼容 OpenAI Chat Completions 的模型服务
  - 默认配置为阿里云 DashScope 兼容接口
  - 默认从环境变量 `DASHSCOPE_API_KEY` 读取 API Key

### 主要第三方库

- `github.com/Auto-CQUPT-Plan/RainClassSDK`
  - 雨课堂登录、课程/作业/考试接口封装
- `github.com/spf13/cobra`
  - CLI 命令组织
- `github.com/cloudwego/eino`
  - 模型调用抽象
- `github.com/cloudwego/eino-ext/components/model/openai`
  - OpenAI-compatible ChatModel 实现
- `github.com/fatih/color`
  - 彩色终端输出
- `github.com/scylladb/termtables`
  - 表格渲染

## 项目使用方法

### 1. 获取源码

```bash
git clone <your-repo>
cd RainClassByeBye
```

### 2. 配置模型 API Key

默认读取 `DASHSCOPE_API_KEY`：

```bash
export DASHSCOPE_API_KEY="your_api_key"
```

也可以在执行 `run` 或 `resume` 时通过 `--api-key` 显式传入。

### 3. 登录雨课堂

```bash
go run . login
```

登录成功后，cookie 默认保存到：

```text
cache/cookies.json
```

### 4. 查询课程和作业

查看当前用户：

```bash
go run . info user
```

查看课程列表：

```bash
go run . info courses
```

查看某门课的作业列表：

```bash
go run . info homework --cid <classroom_id>
```

查看单个作业详情：

```bash
go run . info homework --cid <classroom_id> --leaf-id <leaf_id>
```

### 5. 开始自动答题

```bash
go run . run --cid <classroom_id> --exam-id <exam_id>
```

常用参数：

```bash
--cookie-path cache/cookies.json
--state-dir cache/state
--base-url https://dashscope.aliyuncs.com/compatible-mode/v1
--model qwen3.7-plus
--api-key <key>
--api-key-env DASHSCOPE_API_KEY
--workers 20
--temperature 0.1
--request-timeout 2m
--max-completion-tokens 2048
--submit-paper
```

默认行为：

- 会先调用 `StartExam` 进入考试环境。
- 会并发请求模型，但按题逐个提交答案。
- 会把每次提交结果写入状态文件，便于恢复。
- 默认只提交每题答案，不会自动交卷。
- 只有显式加上 `--submit-paper` 才会在全部题目完成后调用最终交卷接口。

### 6. 恢复中断任务

如果程序退出、网络异常，或者有部分题目失败，可以继续恢复：

```bash
go run . resume --cid <classroom_id> --exam-id <exam_id>
```

如果题目已经全部提交完成，只差最终交卷：

```bash
go run . resume --cid <classroom_id> --exam-id <exam_id> --submit-paper
```

状态文件默认路径：

```text
cache/state/<cid>_<exam_id>.json
```

### 7. 查看任务状态

```bash
go run . status --cid <classroom_id> --exam-id <exam_id>
```

### 8. 构建二进制

```bash
go build -o rainclass-bye-bye .
```

之后可以直接执行：

```bash
./rainclass-bye-bye login
./rainclass-bye-bye info courses
./rainclass-bye-bye run --cid <classroom_id> --exam-id <exam_id>
```

### 9. 运行测试

```bash
go test ./...
```

`dev/unity_test.go` 带有 `devref` build tag，用来记录 `RainClassSDK` 的接口样例；默认不会参与普通测试。需要时可单独执行：

```bash
go test -tags devref ./dev
```

## 关键代码

### 1. CLI 入口与命令组织

- `main.go`
  - 统一执行 `cmd.Execute()`，并把错误格式化输出到终端。
- `cmd/root.go`
  - 定义根命令和全局参数 `--cookie-path`、`--state-dir`。
- `cmd/run.go` / `cmd/resume.go`
  - 负责收集模型参数，组装 `runner.Options`，然后调用执行器。

这部分决定了项目怎样从“命令行参数”进入“自动答题流程”。

### 2. 自动答题执行器

- `internal/runner/runner.go`

这是项目最核心的调度器，主要负责：

- 校验 `cid`、`exam_id`、worker 数量和 solver 配置。
- 加载或创建状态文件。
- 通过 `RainClassSDK` 进入考试环境并获取完整试卷。
- 建立 worker 池并发调用 LLM。
- 对每道题做答案合法化处理，再调用 `SubmitAnswer` 提交。
- 每处理一题就刷新状态文件。
- 所有题目完成后，根据 `--submit-paper` 决定是否继续交卷。

可以把它理解为“考试流程编排层”。

### 3. LLM 求解与答案修复

- `internal/solver/solver.go`

这一层做了几件很关键的事：

- 把题目对象序列化为 JSON，连同题面图片 URL 一起发给模型。
- 强约束模型只返回形如 `{"problem_id":123,"result":["A"]}` 的 JSON。
- 当模型返回空串、非 JSON、单引号 JSON、裸 key、字符串化 JSON 时，尝试解析或修复。
- 对选择题结果做标准化，只保留合法选项 key。
- 当模型输出不可用时，对选择题启用随机合法选项兜底。

这里是项目区别于“直接调一次模型接口”的关键部分，因为它处理了实际场景中最常见的脏输出问题。

### 4. 状态持久化与断点恢复

- `internal/state/state.go`

状态文件会记录：

- `cid`、`exam_id`
- 当前试卷标题
- 已提交题目
- 失败题目及重试次数
- 最后错误
- 当前状态
  - `pending`
  - `running`
  - `interrupted`
  - `partial`
  - `ready_to_submit`
  - `completed`

`run` 和 `resume` 都依赖这套状态模型。没有它，就无法做到安全恢复，也无法区分“题目已提交完成但未交卷”这类场景。

### 5. 交卷请求构造

- `internal/runner/submit_paper.go`

这个文件单独实现最终交卷，是因为交卷接口对 payload 字段命名比较敏感。项目专门把结果结构组装为 snake_case JSON，例如：

```json
{
  "exam_id": "42",
  "results": [
    {
      "problem_id": 1,
      "result": ["A"],
      "time": 123,
      "show_answer": "",
      "is_answered": true,
      "is_save": true
    }
  ]
}
```

`internal/runner/submit_paper_test.go` 就是在验证这里不会错误地发成 camelCase。

## Auto-CQUPT-Plan

本项目和 `Auto-CQUPT-Plan` 的关系很直接：

- 上游项目 `RainClassSDK` 来自 `Auto-CQUPT-Plan`。
- 当前仓库没有重复实现雨课堂底层登录和接口访问，而是在 SDK 之上实现 CLI 自动化能力。
- `dev/unity_test.go` 记录了不少 SDK 的原始调用样例，适合在排查接口行为时参考。

如果你要扩展本项目，通常有两条路径：

- SDK 层有问题：回到 `Auto-CQUPT-Plan/RainClassSDK` 排查。
- CLI、模型求解、状态恢复有问题：在当前仓库修改。

## 免责声明

本项目仅供学习、研究和接口分析使用。

使用者需要自行承担以下责任：

- 确认自己的使用行为符合学校规定、课程平台规则及当地法律法规。
- 自行评估账号风险、成绩风险、封禁风险和其他衍生后果。
- 不将本项目用于任何未经授权的考试、测验或其他违规场景。

作者和贡献者不对因使用、滥用或误用本项目造成的任何直接或间接损失负责。

另需注意：

- 项目默认支持自动提交单题答案。
- `--submit-paper` 会触发最终交卷，属于不可逆操作。
- 雨课堂平台可能存在时间限制、风控策略或接口变更，项目并不保证始终可用。

## 许可证

本项目采用 [MIT License](./LICENSE)。

你可以自由使用、修改、分发本项目，但必须保留原始许可证声明。完整条款见仓库根目录下的 `LICENSE` 文件。
