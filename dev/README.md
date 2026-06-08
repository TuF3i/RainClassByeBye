# RainClassSDK - 长江雨课堂SDK

> **重庆邮电大学-长江雨课堂SDK**
>
> 本项目为 *Auto CQUPT Plan* 的一部分

> [!warning]
>
> **本项目仅供学习研究，请勿用于非法用途，否则后果自负！！！**

![cover](./images/cover.png)

## 一、项目架构概览

```
RainClassSDK (SDK)
  ├── client  (*client.Client)   — HTTP 请求层，负责所有 API 调用
  ├── decoder (*decoder.Decoder) — 解析层，从 HTML/文本中提取关键字段
  └── utils   (*utils.Utils)     — 工具层，终端渲染二维码
```

**核心类关系：**

| 类型 | 文件 | 说明 |
|------|------|------|
| `SDK` | `root.go:15` | 对外门面，聚合 client/decoder/utils |
| `Client` | `client/root.go:9` | HTTP 客户端，内置 cookie jar |
| `Settings` | `root.go:11` | 配置项（仅 Path 字段） |
| `Options` | `root.go:9` | 函数式选项 `func(s *Settings)` |

---

## 二、所有测试函数的参数与返回值

### 2.1 基础功能测试

| 测试函数 | 行号 | 调用的 SDK/Client 方法 | 参数 | 返回值 |
|----------|------|------------------------|------|--------|
| `TestSDK_QRLogin` | 20 | `NewSDK(WithCookiePath(...))` | `"./test_cookies.json"` | `(*SDK, error)` |
| | | `sdk.QRLogin()` | 无 | `error` |
| | | `sdk.GetUserInfo()` | 无 | `(*models.UserInfo, error)` |
| | | `sdk.Close()` | 无 | `error` |
| `TestSDK_GetUserInfo` | 48 | 同上 | 无 | `(*models.UserInfo, error)` |
| `TestSDK_GetCourseInfo` | 71 | `sdk.GetCourseInfo()` | 无 | `(*models.CourseInfo, error)` |
| `TestSDK_GetHomeWorkInfo` | 94 | `sdk.GetHomeWorkInfo(cid)` | `cid int64 = 24211265` | `(*models.HomeWorkInfo, error)` |
| `TestSDK_GetHomeWorkDetails` | 117 | `sdk.GetHomeWorkDetails(cid, leafID)` | `cid=24211265, leafID=45505698` | `(*models.HomeWorkDetails, error)` |
| `TestSDK_GetHomeWorkCover` | 140 | `sdk.GetHomeWorkCover(cid, examID)` | `cid=24211265, examID=1945460` | `(*models.HomeWorkCover, error)` |

### 2.2 考试流程底层测试（直接调用 client）

| 测试函数 | 行号 | 调用的 Client 方法 | 参数 | 返回值 |
|----------|------|---------------------|------|--------|
| `TestClient_ExamGenToken` | 163 | `sdk.client.ExamGenToken(cid, examID)` | `cid=24211265, examID=1945538` | `(*models.ExamGenTokenResp, error)` |
| `TestClient_ExamLogin` | 186 | `sdk.client.ExamLogin(examID, userID, token)` | `examID=1945538, userID=57510820, token="H1tf/67TN..."` | `error` |
| `TestClient_StartExam` | 203 | `sdk.client.StartExam(examID)` | `examID=1945538` | `error` |
| `TestClient_StartExamPaper` | 220 | `sdk.client.StartExamPaper(examID)` | `examID=1945538` | `(*models.StartExamPaperResp, error)` |
| `TestClient_GetExamPaperCover` | 243 | `sdk.client.GetExamPaperCover(examID)` | `examID=1945538` | `(*models.ExamPaperCover, error)` |
| `TestClient_GetExamPaperQuestion` | 266 | `sdk.client.GetExamPaperQuestion(examID)` | `examID=1945460` | `(*models.ExamPaperQuestions, error)` |
| `TestClient_RefreshTimeRemaining`| 294 | `sdk.client.RefreshTimeRemaining(examID)` | `examID=1945538` | `(*models.TimeRemaining, error)` |
| `TestClient_SubmitAnswer` | 317 | `sdk.client.SubmitAnswer(examID, ans)` | `examID=1945538, ans={ProblemId:92134642, Result:["A"], Time}` | `(*models.SubmitAnswerResp, error)` |
| `TestClient_SubmitPaper` | 346 | `sdk.client.SubmitPaper(examID, ansList)` | `examID=1945538, ansList=[...]` | `(*models.SubmitPaperResp, error)` |

### 2.3 高层考试流程测试 & AI 测试

| 测试函数 | 行号 | 说明 |
|----------|------|------|
| `TestSDK_StartExam` | 380 | SDK 层一键开始考试 `sdk.StartExam(cid=24211265, examID=1945538) → error` |
| `Test_GetImgB64` | 485 | 获取试卷题目中的图片 URL 并转 Base64 |
| `Test_AIFuck` | 510 | 完整 AI 做题流程：开始考试 → 获取题目 → Qwen 多模态答题 → 提交答案 |

---

## 三、API 调用一览

所有 API 定义在 `consts/consts.go`，共 16 个端点：

| 常量名 | 完整 URL | 方法 | 用途 |
|--------|----------|------|------|
| `GET_USER_INFO_URL` | `https://changjiang.yuketang.cn/v/course_meta/user_info` | GET | 触发 CSRF Token 设置 |
| `GET_WX_OAUTH_INFO_URL` | `https://changjiang.yuketang.cn/api/v3/user/login/wechat-auth-param` | POST | 获取微信 OAuth 参数 |
| `GET_QRCODE_UID_URL` | `https://open.weixin.qq.com/connect/qrconnect` | GET | 获取二维码内容（含 UID） |
| `GET_QRCODE_PNG_URL` | `https://open.weixin.qq.com/connect/qrcode/{uid}` | GET | 下载二维码 PNG |
| `GET_WXCODE_URL` | `https://lp.open.weixin.qq.com/connect/l/qrconnect` | GET | 轮询扫码状态 |
| `GET_SESSION_URL` | `https://changjiang.yuketang.cn/api/v3/user/login/wechat-web-callback` | GET | 用 wxcode 换取 Session |
| `GET_BASIC_INFO_URL` | `https://changjiang.yuketang.cn/api/v3/user/basic-info` | GET | 获取用户基本信息 |
| `GET_COURSE_INFO_URL` | `https://changjiang.yuketang.cn/v2/api/web/courses/list?identity=2` | GET | 获取课程列表 |
| `GET_HOME_WORK_INFO_URL` | `https://changjiang.yuketang.cn/mooc-api/v1/lms/learn/course/chapter` | GET | 获取作业列表 |
| `GET_HOME_WORK_DETAIL_URL` | `https://changjiang.yuketang.cn/mooc-api/v1/lms/learn/leaf_info/{cid}/{leafID}/` | GET | 获取作业详情 |
| `GET_HOME_WORK_COVER_URL` | `https://changjiang.yuketang.cn/v/exam/cover` | GET | 获取作业封面 |
| `GET_EXAM_GEN_TOKEN_URL` | `https://changjiang.yuketang.cn/v/exam/gen_token` | POST | 生成考试 Token |
| `EXAM_LOGIN_URL` | `https://changjiang-exam.yuketang.cn/login` | GET | 考试系统登录 |
| `START_EXAM_PAPER_URL` | `https://changjiang-exam.yuketang.cn/exam_room/start_paper` | POST | 开始考试答卷 |
| `GET_EXAM_PAPER_COVER_URL` | `https://changjiang-exam.yuketang.cn/exam_room/cover` | GET | 获取试卷封面 |
| `GET_EXAM_PAPER_QUESTION_URL` | `https://changjiang-exam.yuketang.cn/exam_room/show_paper` | GET | 获取试卷题目 |
| `GET_EXAM_TIME_REAMINING_URL` | `https://changjiang-exam.yuketang.cn/exam_room/refresh_time` | GET | 刷新剩余时间 |
| `SUBMIT_ANSWER_URL` | `https://changjiang-exam.yuketang.cn/exam_room/answer_problem` | POST | 提交单题答案 |
| `SUBMIT_PAPER_URL` | `https://changjiang-exam.yuketang.cn/exam_room/submit_paper` | POST | 提交整卷答案 |

**请求头约定：**
- 长江雨课堂主站 (`changjiang.yuketang.cn`)：`xtbz: ykt`，部分接口需要 `X-CSRFToken`
- 考试系统 (`changjiang-exam.yuketang.cn`)：`xtbz: cloud`，`x-client: web`

---

## 四、调用链流程图

### 4.1 QRLogin 完整调用链

```
TestSDK_QRLogin
  │
  ├─ NewSDK(WithCookiePath("./test_cookies.json"))
  │    └─ client.NewClient(path) → 创建带 cookie jar 的 HTTP 客户端
  │
  ├─ sdk.QRLogin()
  │    │
  │    ├─ [1] client.GetCRSFTokenResponse()
  │    │       └─ GET https://changjiang.yuketang.cn/v/course_meta/user_info
  │    │          (触发服务端设置 csrftoken cookie)
  │    │
  │    ├─ [2] client.GetWxOauthInfo()
  │    │       └─ POST https://changjiang.yuketang.cn/api/v3/user/login/wechat-auth-param
  │    │          → 返回 OAuthInfo{AppId, State, RedirectUri}
  │    │
  │    ├─ [3] client.GetQRCodeContent(info)
  │    │       └─ GET https://open.weixin.qq.com/connect/qrconnect?appid=...&state=...&...
  │    │          → 返回 HTML 页面内容 (string)
  │    │
  │    ├─ [4] decoder.GetQRCodeUIDFromContent(content)
  │    │       └─ 正则提取 HTML 中的 UUID
  │    │          → "xxxxxxxxxxxxxxxxxxxx"
  │    │
  │    ├─ [5] client.GetQRCodePNG(uid)
  │    │       └─ GET https://open.weixin.qq.com/connect/qrcode/{uid}
  │    │          → 返回 PNG 图片字节 ([]byte)
  │    │
  │    ├─ [6] utils.PrintQRFromBytes(pngBytes)
  │    │       └─ 终端打印二维码 ASCII 图形
  │    │
  │    ├─ [7] 轮询循环 (每 5 秒)
  │    │       │
  │    │       ├─ client.GetWxCodeContent(uid)
  │    │       │    └─ GET https://lp.open.weixin.qq.com/connect/l/qrconnect?uuid={uid}
  │    │       │       → 返回 JSON: {wx_code: "xxx"} 或 {wx_code: ""}
  │    │       │
  │    │       └─ decoder.GetWxCodeValueFromContent(content)
  │    │            └─ 正则提取 wx_code 字段
  │    │               → 有值: 退出循环; 无值: 继续等待
  │    │
  │    └─ [8] client.GetSession(wxcode, info)
  │           └─ GET https://changjiang.yuketang.cn/api/v3/user/login/wechat-web-callback
  │              ?path=...&code={wxcode}&state={state}
  │              → 期望 HTTP 302 (Found)，完成登录
  │
  ├─ sdk.GetUserInfo()
  │    └─ client.GetUserInfo()
  │         └─ GET https://changjiang.yuketang.cn/api/v3/user/basic-info
  │            → *models.UserInfo
  │
  └─ sdk.Close()
       └─ client.ShutdownClient()
            └─ 持久化 cookie jar 到文件
```

### 4.2 StartExam 完整调用链（考试流程核心）

```
TestSDK_StartExam / Test_AIFuck
  │
  ├─ sdk.StartExam(cid=24211265, examID=1945538)
  │    │
  │    ├─ [1] client.ExamGenToken(cid, examID)
  │    │       └─ POST https://changjiang.yuketang.cn/v/exam/gen_token
  │    │          Body: {"exam_id":"1945538","classroom_id":"24211265"}
  │    │          Headers: xtbz: ykt, X-CSRFToken: {从cookie jar提取}
  │    │          → *models.ExamGenTokenResp{Token, ExamHost, UserId}
  │    │
  │    ├─ [2] client.ExamLogin(examID, userId, token)
  │    │       └─ GET https://changjiang-exam.yuketang.cn/login
  │    │          ?exam_id=1945538&user_id=57510820&crypt={token}&next=...
  │    │          → 期望 HTTP 302，设置考试系统 cookie
  │    │
  │    ├─ [3] client.StartExam(examID)
  │    │       └─ GET https://changjiang-exam.yuketang.cn/start/1945538?isFrom=2
  │    │          → HTTP 200，确认进入考场
  │    │
  │    └─ [4] client.StartExamPaper(examID)
  │           └─ POST https://changjiang-exam.yuketang.cn/exam_room/start_paper
  │              Body: {"exam_id":"1945538"}
  │              → *models.StartExamPaperResp{HasLimit, TimePast, TimeLeft}
  │
  ├─ sdk.GetExamPaperQuestion(examID)
  │    └─ GET https://changjiang-exam.yuketang.cn/exam_room/show_paper?exam_id=1945538
  │       → *models.ExamPaperQuestions{Problems: [...]}
  │
  ├─ sdk.SubmitAnswer(examID, answer)   ← 逐题提交
  │    └─ POST https://changjiang-exam.yuketang.cn/exam_room/answer_problem
  │       Body: {"results":[{"problem_id":...,"result":["A"],"time":...}],"exam_id":...,"record":[]}
  │       → *models.SubmitAnswerResp
  │
  └─ sdk.SubmitPaper(examID, allAnswers)  ← 整卷提交
       └─ POST https://changjiang-exam.yuketang.cn/exam_room/submit_paper
          Body: {"results":[...],"exam_id":"1945538"}
          → *models.SubmitPaperResp
```

### 4.3 各测试函数调用关系总图

```
                              ┌─────────────────────────┐
                              │    NewSDK(cookiePath)    │
                              └────────────┬────────────┘
                                           │
              ┌────────────────────────────┼────────────────────────────┐
              │                            │                            │
              ▼                            ▼                            ▼
  ┌───────────────────────┐   ┌──────────────────────┐   ┌──────────────────────────┐
  │   TestSDK_QRLogin     │   │  TestSDK_GetUserInfo  │   │   TestSDK_GetCourseInfo   │
   │   └─ QRLogin()        │   │   └─ GetUserInfo()    │   │   └─ GetCourseInfo()       │
   │   └─ GetUserInfo()    │   └──────────────────────┘   └──────────────────────────┘
  └───────────────────────┘
              │
  ┌───────────────────────┐   ┌───────────────────────────┐   ┌───────────────────────────┐
  │TestSDK_GetHomeWorkInfo │   │TestSDK_GetHomeWorkDetails  │   │ TestSDK_GetHomeWorkCover   │
  │ └─ GetHomeWorkInfo(cid)│   │ └─ GetHomeWorkDetails(...) │   │ └─ GetHomeWorkCover(...)   │
  └───────────────────────┘   └───────────────────────────┘   └───────────────────────────┘
              │
              │  ┌────────────────────────────────────────────────────────────┐
              │  │               考试流程（低级 API 逐个测试）                    │
              │  │                                                            │
              │  │  TestClient_ExamGenToken                                    │
              │  │      │                                                     │
              │  │      ▼                                                     │
              │  │  TestClient_ExamLogin                                      │
              │  │      │                                                     │
              │  │      ▼                                                     │
              │  │  TestClient_StartExam                                      │
              │  │      │                                                     │
              │  │      ├──── TestClient_StartExamPaper                       │
              │  │      ├──── TestClient_GetExamPaperCover                    │
              │  │      ├──── TestClient_GetExamPaperQuestion                 │
              │  │      ├──── TestClient_RefreshTimeRemaining                 │
              │  │      ├──── TestClient_SubmitAnswer                         │
              │  │      └──── TestClient_SubmitPaper                          │
              │  └────────────────────────────────────────────────────────────┘
              │
              │  ┌────────────────────────────────────────────────────────────┐
              │  │              高级考试流程 & AI 做题                          │
              │  │                                                            │
              │  │  TestSDK_StartExam                                         │
              │  │   └─ sdk.StartExam(cid, examID)  ← 封装了上面 4 个低级调用    │
              │  │                                                            │
              │  │  Test_GetImgB64                                            │
              │  │   └─ GetExamPaperQuestion → GenImageB64 (提取图片转Base64)   │
              │  │                                                            │
              │  │  Test_AIFuck                                               │
              │  │   └─ StartExam → GetExamPaperQuestion                      │
              │  │   └─ 逐题: Qwen3.6-Plus 多模态推理 → SubmitAnswer           │
              │  │   └─ (SubmitPaper 已注释)                                   │
              │  └────────────────────────────────────────────────────────────┘
```

---

## 五、每个单元测试函数详解

### 5.1 `TestSDK_QRLogin` (行 20-46)

**功能：** 测试完整的微信扫码登录流程。

**流程：**
1. 创建 SDK 实例（cookie 持久化到 `./test_cookies.json`）
2. 调用 `QRLogin()` 执行完整的扫码登录流程
3. 扫码成功后调用 `GetUserInfo()` 验证登录状态
4. 打印用户信息 JSON
5. 关闭 SDK（持久化 cookies）

**关键点：** 这是唯一一个需要人工交互的测试 —— 会在终端打印二维码，等待用户用微信扫码。如果已登录（cookie 有效），`QRLogin` 可能仍然会尝试重新走一遍流程。

---

### 5.2 `TestSDK_GetUserInfo` (行 48-69)

**功能：** 获取当前登录用户的基本信息。

**调用链：** `sdk.GetUserInfo()` → `client.GetUserInfo()` → `GET /api/v3/user/basic-info`

**返回值 `*models.UserInfo`：**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 12345678,
    "avatar": "https://...",
    "name": "张三",
    "school": "重庆邮电大学",
    "school_number": "2022210000",
    "nickname": "...",
    "role": 0,
    ...
  }
}
```

**前置条件：** 需要 `test_cookies.json` 中存在有效的登录 cookie。

---

### 5.3 `TestSDK_GetCourseInfo` (行 71-92)

**功能：** 获取当前用户的课程列表。

**调用链：** `sdk.GetCourseInfo()` → `client.GetCourseList()` → `GET /v2/api/web/courses/list?identity=2`

**返回值 `*models.CourseInfo`：** 包含按学期分组的课程列表 `CourseDatas.List []CourseNode`，每个 `CourseNode` 包含课程名称、教师信息、班级 ID (`ClassroomId`) 等。

---

### 5.4 `TestSDK_GetHomeWorkInfo` (行 94-115)

**功能：** 获取指定课程的作业/章节列表。

**参数：** `cid = 24211265`（课程/班级 ID，来自 `GetCourseInfo` 返回的 `ClassroomId`）

**调用链：** `sdk.GetHomeWorkInfo(cid)` → `client.GetHomeWorkList(cid)` → `GET /mooc-api/v1/lms/learn/course/chapter?cid=...&classroom_id=...`

**返回值 `*models.HomeWorkInfo`：** 课程章节树，每个章节下包含多个 `SectionLeafListNode`（作业、考试等叶子节点），其中的 `Id` 字段可作为后续 `leafID` 参数。

---

### 5.5 `TestSDK_GetHomeWorkDetails` (行 117-138)

**功能：** 获取某个作业/考试的详细信息。

**参数：**
- `cid = 24211265` — 课程 ID
- `leafID = 45505698` — 叶子节点 ID（来自 `GetHomeWorkInfo` 返回的 `SectionLeafListNode.Id`）

**调用链：** `sdk.GetHomeWorkDetails(cid, leafID)` → `client.GetHomeWorkDetails(cid, leafID)` → `GET /mooc-api/v1/lms/learn/leaf_info/{cid}/{leafID}/`

**返回值 `*models.HomeWorkDetails`：** 包含发布人、发布时间、截止时间、内容信息（`ContentInfoEntity`）、分数评估配置等。

**注意：** 请求头需要 `xtbz: ykt` 和 `classroom-id: {cid}`。

---

### 5.6 `TestSDK_GetHomeWorkCover` (行 140-161)

**功能：** 获取作业/考试的封面信息（考试入口页面数据）。

**参数：**
- `cid = 24211265`
- `examID = 1945460`（考试 ID，来自 `GetHomeWorkInfo` 中叶子节点的 `LeafinfoId`）

**调用链：** `sdk.GetHomeWorkCover(cid, examID)` → `client.GetHomeWorkCover(cid, examID)` → `GET /v/exam/cover?exam_id=...&classroom_id=...`

**返回值 `*models.HomeWorkCover`：** 包含考试标题、总分、截止时间、题目数量、用户信息、成绩等。

---

### 5.7 `TestClient_ExamGenToken` (行 163-184)

**功能：** 生成考试 Token —— 考试流程的第一步。

**参数：** `cid = 24211265`, `examID = 1945538`

**调用链：** `client.ExamGenToken(cid, examID)` → `POST /v/exam/gen_token`

**请求体：**
```json
{"exam_id": "1945538", "classroom_id": "24211265"}
```

**关键实现细节：** 该接口需要 CSRF Token。Client 从 cookie jar 中提取 `changjiang.yuketang.cn` 域下的 `csrftoken` cookie 值，通过 `X-CSRFToken` 请求头传递。

**返回值 `*models.ExamGenTokenResp`：**
```json
{
  "success": true,
  "data": {
    "token": "H1tf/67TNyDZkhZjQERTTLJtZUE1oQ...",
    "exam_host": "changjiang-exam.yuketang.cn",
    "user_id": 57510820
  }
}
```

---

### 5.8 `TestClient_ExamLogin` (行 186-201)

**功能：** 用生成的 Token 登录考试系统 —— 考试流程的第二步。

**参数：**
- `examID = 1945538`
- `userID = 57510820`（来自 `ExamGenTokenResp.Data.UserId`）
- `token = "H1tf/67TNyDZkhZjQERTTLJtZUE1oQWUL5HpmMltzZR+GyeYX56u6vPDjTo+4Gmhn9kAcO+NDSNNk0qIMuTboA=="`（来自 `ExamGenTokenResp.Data.Token`，是加密后的凭证）

**调用链：** `client.ExamLogin(examID, userID, token)` → `GET /login?exam_id=...&user_id=...&crypt=...&next=...`

**关键点：** 期望返回 HTTP 302 (Found)，服务端会设置考试系统 (`changjiang-exam.yuketang.cn`) 的 cookie，这是后续所有考试 API 的认证凭证。

---

### 5.9 `TestClient_StartExam` (行 203-218)

**功能：** 进入考试页面 —— 考试流程的第三步。

**参数：** `examID = 1945538`

**调用链：** `client.StartExam(examID)` → `GET /start/{examID}?isFrom=2`

**关键点：** 这一步相当于"点击进入考场"，确认考生就位。

---

### 5.10 `TestClient_StartExamPaper` (行 220-241)

**功能：** 正式开始答卷 —— 考试流程的第四步（最后一步）。

**参数：** `examID = 1945538`

**调用链：** `client.StartExamPaper(examID)` → `POST /exam_room/start_paper`

**请求体：** `{"exam_id": "1945538"}`

**返回值 `*models.StartExamPaperResp`：**
```json
{
  "errcode": 0,
  "data": {
    "has_limit": true,
    "time_past": 0,
    "time_left": 7200
  }
}
```
- `has_limit` — 是否限时
- `time_past` — 已用时间（秒）
- `time_left` — 剩余时间（秒）

---

### 5.11 `TestClient_GetExamPaperCover` (行 243-264)

**功能：** 获取试卷封面信息。

**参数：** `examID = 1945538`

**调用链：** `client.GetExamPaperCover(examID)` → `GET /exam_room/cover?exam_id=...`

**返回值 `*models.ExamPaperCover`：** 与 `HomeWorkCover` 结构类似但包含考试特有的字段（如 `FaceAuthStatus` 人脸认证状态），包括试卷标题、总分、题目数量、倒计时等。

---

### 5.12 `TestClient_GetExamPaperQuestion` (行 266-292)

**功能：** 获取试卷的全部题目。

**参数：** `examID = 1945460`

**调用链：** `client.GetExamPaperQuestion(examID)` → `GET /exam_room/show_paper?exam_id=...`

**返回值 `*models.ExamPaperQuestions`：**
```json
{
  "errcode": 0,
  "data": {
    "problems": [
      {
        "problem_id": 92134642,
        "body": "题目内容（可能含HTML图片标签）",
        "type": "single_choice",
        "score": 5,
        "options": [
          {"key": "A", "value": "选项A内容"},
          {"key": "B", "value": "选项B内容"}
        ],
        "problem_type": "choice",
        ...
      }
    ],
    "has_problem_dict": false,
    "title": "2024年XXX考试"
  }
}
```

**注意：** 该测试使用 `json.Encoder` 并关闭 HTML 转义（`SetEscapeHTML(false)`），因为题目内容中可能包含 HTML 标签（如图片 `<img>`）。

---

### 5.13 `TestClient_RefreshTimeRemaining` (行 294-315)

**功能：** 刷新考试剩余时间。

**参数：** `examID = 1945538`

**调用链：** `client.RefreshTimeRemaining(examID)` → `GET /exam_room/refresh_time?exam_id=...`

**返回值 `*models.TimeRemaining`：**
```json
{
  "errcode": 0,
  "data": {
    "has_limit": true,
    "time_past": 120,
    "time_left": 7080
  }
}
```

**用途：** 用于客户端实时更新倒计时，也可作为心跳保活。

---

### 5.14 `TestClient_SubmitAnswer` (行 317-344)

**功能：** 提交单道题的答案。

**参数：**
- `examID = 1945538`
- `answer = SubmitAnswerResultsEntity{ProblemId: 92134642, Result: ["A"], Time: 当前时间戳(ms)}`

**调用链：** `client.SubmitAnswer(examID, answer)` → `POST /exam_room/answer_problem`

**实际请求体（Client 层包装后）：**
```json
{
  "results": [{"problem_id": 92134642, "result": ["A"], "time": 1710000000000}],
  "exam_id": 1945538,
  "record": []
}
```

**注意：** SDK 层的 `SubmitAnswer` 接收单个 `SubmitAnswerResultsEntity`，Client 层会将其包装为数组放入 `results` 字段。

---

### 5.15 `TestClient_SubmitPaper` (行 346-378)

**功能：** 批量提交整卷答案（交卷）。

**参数：**
- `examID = 1945538`
- `ansList = []SubmitPaperPostResultsEntity` — 所有题目的答案列表

**单条答案结构：**
```go
SubmitPaperPostResultsEntity{
    ProblemId:  92134642,
    Result:     []string{"A"},
    Time:       当前时间戳(ms),
    ShowAnswer: "",      // 空字符串
    IsAnswered: true,
    IsSave:     true,
}
```

**调用链：** `client.SubmitPaper(examID, ansList)` → `POST /exam_room/submit_paper`

**请求体：**
```json
{
  "results": [
    {"problem_id": 92134642, "result": ["A"], "time": ..., "show_answer": "", "is_answered": true, "is_save": true}
  ],
  "exam_id": "1945538"
}
```

**与 `SubmitAnswer` 的区别：** `SubmitAnswer` 逐题提交（答题过程中使用），`SubmitPaper` 一次性提交全部答案（交卷时使用），且后者多了 `ShowAnswer`、`IsAnswered`、`IsSave` 字段。

---

### 5.16 `TestSDK_StartExam` (行 380-395)

**功能：** SDK 层的高级封装 —— 一键开始考试。

**参数：** `cid = 24211265`, `examID = 1945538`

**调用链（内部分 4 步）：**
```
sdk.StartExam(cid, examID)
  ├─ client.ExamGenToken(cid, examID)       // 生成 Token
  ├─ client.ExamLogin(examID, userId, token) // 登录考试系统
  ├─ client.StartExam(examID)                // 进入考场
  └─ client.StartExamPaper(examID)           // 开始答卷
```

**与低级 API 测试的对应关系：** 这一个测试等价于前面 `ExamGenToken → ExamLogin → StartExam → StartExamPaper` 四个测试的串联。

---

### 5.17 `Test_GetImgB64` (行 485-508)

**功能：** 测试题目图片提取与 Base64 转换。

**流程：**
1. 获取试卷题目 (`GetExamPaperQuestion`)
2. 对每道题调用 `GenImageB64()` 提取 `<img>` 标签中的 URL
3. 通过 `ImageURLToBase64()` 下载图片并转为 Base64 编码

**辅助函数：**
- `GenImageB64(data ProblemsEntity) []Images` — 用正则 `(?i)<img[^>]*src\s*=\s*\\?["']([^"']*?)\\?["']` 从 JSON 序列化的题目中提取图片 URL
- `ImageURLToBase64(url string) (string, error)` — HTTP GET 下载图片，返回标准 Base64 编码

**返回值：** `[]Images`，每个元素包含 `OriginalUrl`（原始 URL）和 `Base64`（Base64 编码）。

---

### 5.18 `Test_AIFuck` (行 510-698)

**功能：** 完整的 AI 自动做题测试 —— 使用阿里云 DashScope 的通义千问 (Qwen) 多模态模型解答数学题。

**前置条件：** 需要设置环境变量 `DASHSCOPE_API_KEY`。

**常量参数：**
- `cid = 24211265` — 课程 ID
- `examid = 1945460` — 考试 ID

**完整流程：**

```
1. NewSDK + defer Close
2. sdk.StartExam(cid, examid)       ← 开始考试
3. sdk.GetExamPaperQuestion(examid)  ← 获取所有题目
4. 逐题循环:
   │
   ├─ [1] GenImageB64(problem)        ← 提取题目中的图片
   │
   ├─ [2] 构建多模态 messages:
   │      - system: 角色提示词（数学解题机器人 + JSON 输出格式要求）
   │      - user:   [图片URL数组, 题目JSON文本]
   │
   ├─ [3] POST https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
   │       Body: {"model":"qwen3.6-plus", "messages":[...], "stream":false}
   │       Header: Authorization: Bearer {DASHSCOPE_API_KEY}
   │
   ├─ [4] 解析 Qwen 返回: DeepSeekAnswer{ProblemId, Result}
   │
   ├─ [5] sdk.SubmitAnswer(examid, {problemId, result, time})
   │
   └─ [6] 记录答案到 ansList
5. (SubmitPaper 已注释)
```

**System Prompt 要点：**
- 角色：重庆邮电大学研制的数学题解题机器人
- 输出格式：严格 JSON `{"problem_id": number, "result": [string数组]}`
- 错误兜底：`{"problem_id": -1, "result": []}`

**关键设计：**
- 使用阿里云 DashScope 兼容 OpenAI 格式的 API 端点（`/compatible-mode/v1/chat/completions`）
- 使用 `qwen3.6-plus` 多模态模型，支持图片 + 文本输入
- 图片直接传递原始 URL（非 Base64），减少 token 消耗
- 响应结构复用 `DeepSeekChatCompletionResponse`（兼容 OpenAI Chat Completions 格式）
- 做题范围：遍历所有题目，逐题解答、逐题提交
- 超时设置为 2 分钟/题（`context.WithTimeout`）

**测试中未启用的部分：** 最后的 `SubmitPaper` 被注释掉了，所以这个测试目前是"做完题但暂不交卷"的状态。

---

## 六、测试依赖关系总结

```
test_cookies.json (cookie 持久化文件)
        │
        ├── 必须项：所有测试都需要有效的登录 cookie
        │   获取方式：先跑通 TestSDK_QRLogin 扫码登录
        │
        ├── TestSDK_QRLogin           ← 产生 cookie（首次）
        ├── TestSDK_GetUserInfo       ← 依赖 cookie
        ├── TestSDK_GetCourseInfo     ← 依赖 cookie
        ├── TestSDK_GetHomeWorkInfo   ← 依赖 cookie + cid
        ├── TestSDK_GetHomeWorkDetails← 依赖 cookie + cid + leafID
        ├── TestSDK_GetHomeWorkCover  ← 依赖 cookie + cid + examID
        ├── TestClient_ExamGenToken   ← 依赖 cookie + csrftoken
        ├── TestClient_ExamLogin      ← 依赖 ExamGenToken 的输出
        ├── TestClient_StartExam      ← 依赖 ExamLogin 成功
        ├── TestClient_StartExamPaper ← 依赖 StartExam 成功
        ├── TestClient_GetExamPaperCover    ← 依赖考试系统 cookie
        ├── TestClient_GetExamPaperQuestion ← 依赖考试系统 cookie
        ├── TestClient_RefreshTimeRemaining ← 依赖考试系统 cookie
        ├── TestClient_SubmitAnswer   ← 依赖考试系统 cookie
        ├── TestClient_SubmitPaper    ← 依赖考试系统 cookie
        ├── TestSDK_StartExam         ← 封装上述 4 步
        ├── Test_GetImgB64            ← 依赖 GetExamPaperQuestion
        └── Test_AIFuck               ← 依赖 StartExam + Qwen API Key
```

**注意：** 考试相关的测试（`TestClient_ExamLogin` 到 `TestClient_SubmitPaper`）需要按顺序执行，因为考试系统的 cookie 是在 `ExamLogin` 步骤中设置的。`TestSDK_StartExam` 和 `Test_AIFuck` 内部已封装这些步骤。

---

## 七、SDK 基础使用指南

### 7.1 初始化 SDK

```go
import "github.com/Auto-CQUPT-Plan/RainClassSDK"

sdk, err := RainClassSDK.NewSDK(
    RainClassSDK.WithCookiePath("./my_cookies.json"),
)
if err != nil {
    panic(err)
}
defer sdk.Close()
```

- `NewSDK` 接受零个或多个 `Options`，目前唯一选项是 `WithCookiePath`，用于指定 cookie 持久化文件路径
- 不传任何选项时，cookie 默认保存到 `./CookieJar.json`
- `defer sdk.Close()` 确保退出时将 cookie 写入文件，下一次启动无需重新登录

### 7.2 扫码登录（首次使用）

```go
err := sdk.QRLogin()
if err != nil {
    panic(err)
}
```

- 终端会打印一个二维码（ASCII art），用微信扫码即可完成登录
- 扫码成功后登录 cookie 自动持久化到指定的 cookie 文件
- 下一次启动 SDK 时，如果 cookie 文件存在且未过期，**无需再次调用 `QRLogin`**，直接调用后续 API 即可

### 7.3 获取用户信息

```go
userInfo, err := sdk.GetUserInfo()
if err != nil {
    panic(err)
}
fmt.Printf("姓名: %s, 学号: %s, 学校: %s\n",
    userInfo.Data.Name,
    userInfo.Data.SchoolNumber,
    userInfo.Data.School,
)
```

### 7.4 获取课程和作业信息

```go
// 获取课程列表
courses, err := sdk.GetCourseInfo()
// courses.CourseData.List 是课程数组，每个课程有 ClassroomId (班级ID)

// 获取某门课的作业/章节列表
homework, err := sdk.GetHomeWorkInfo(classroomId)
// homework.Data.CourseChapter[].SectionLeafList[].Id 是叶子节点 ID

// 获取某个作业的详细信息
details, err := sdk.GetHomeWorkDetails(classroomId, leafId)

// 获取某个作业/考试的封面
cover, err := sdk.GetHomeWorkCover(classroomId, examId)
```

### 7.5 考试流程

考试功能分两层：**底层 `client` API**（逐步调用）和 **SDK 高层封装**（一步到位）。

#### 方式一：SDK 高层封装（推荐）

```go
// 一行开始考试（内部自动完成 genToken → examLogin → startExam → startExamPaper）
err := sdk.StartExam(courseId, examId)

// 获取试卷题目
questions, err := sdk.GetExamPaperQuestion(examId)

// 逐题提交答案
for _, problem := range questions.Data.Problems {
    answer := models.SubmitAnswerResultsEntity{
        ProblemId: problem.ProblemId,
        Result:    []string{"A"},      // 单选题选 A
        Time:      time.Now().UnixMilli(),
    }
    resp, err := sdk.SubmitAnswer(examId, answer)
    _ = resp
}

// 交卷
allAnswers := []models.SubmitPaperPostResultsEntity{...}
submitResp, err := sdk.SubmitPaper(examId, allAnswers)
```

#### 方式二：底层 client 逐步调用（适合需要精细控制的场景）

```go
// 1. 生成 Token
tokenResp, err := sdk.Client().ExamGenToken(courseId, examId)

// 2. 登录考试系统
err = sdk.Client().ExamLogin(examId, tokenResp.Data.UserId, tokenResp.Data.Token)

// 3. 进入考场
err = sdk.Client().StartExam(examId)

// 4. 开始答卷
paperResp, err := sdk.Client().StartExamPaper(examId)

// 5. 获取题目
questions, err := sdk.Client().GetExamPaperQuestion(examId)

// 6. 逐题提交
for _, p := range questions.Data.Problems {
    ans := models.SubmitAnswerResultsEntity{...}
    sdk.Client().SubmitAnswer(examId, ans)
}

// 7. 交卷
sdk.Client().SubmitPaper(examId, allAnswers)
```

> **注意：** `sdk.Client()` 方法需要确认 SDK 是否暴露了 `client` 字段。如果未暴露，底层 API 仅能通过直接创建 `client.Client` 实例来使用。

### 7.6 完整示例：从登录到做完考试

```go
package main

import (
    "fmt"
    "time"

    "github.com/Auto-CQUPT-Plan/RainClassSDK"
    "github.com/Auto-CQUPT-Plan/RainClassSDK/models"
)

func main() {
    sdk, _ := RainClassSDK.NewSDK(
        RainClassSDK.WithCookiePath("./cookies.json"),
    )
    defer sdk.Close()

    // 如果 cookie 过期，需要扫码登录
    if _, err := sdk.GetUserInfo(); err != nil {
        fmt.Println("需要扫码登录，请用微信扫描终端中的二维码...")
        if err := sdk.QRLogin(); err != nil {
            panic(err)
        }
    }

    // 获取用户信息
    user, _ := sdk.GetUserInfo()
    fmt.Printf("欢迎, %s!\n", user.Data.Name)

    // 获取课程列表
    courses, _ := sdk.GetCourseInfo()
    for _, c := range courses.CourseData.List {
        fmt.Printf("课程: %s (班级ID: %d)\n", c.Name, c.ClassroomId)
    }

    // 假设已知 courseId=24211265, examId=1945538
    courseId := int64(24211265)
    examId := int64(1945538)

    // 开始考试
    if err := sdk.StartExam(courseId, examId); err != nil {
        panic(err)
    }

    // 获取题目
    paper, _ := sdk.GetExamPaperQuestion(examId)

    // 做题...（这里写 AI 做题逻辑或手动填答案）
    var allAnswers []models.SubmitPaperPostResultsEntity
    for _, p := range paper.Data.Problems {
        ans := models.SubmitPaperPostResultsEntity{
            ProblemId:  p.ProblemId,
            Result:     []string{"A"}, // 默认选 A，实际应替换为真实答案
            Time:       time.Now().UnixMilli(),
            IsAnswered: true,
            IsSave:     true,
        }
        allAnswers = append(allAnswers, ans)
    }

    // 交卷
    resp, _ := sdk.SubmitPaper(examId, allAnswers)
    fmt.Printf("交卷结果: errcode=%d, msg=%s\n", resp.Errcode, resp.Errmsg)
}
```

### 7.7 关键注意事项

| 要点 | 说明 |
|------|------|
| **Cookie 持久化** | `defer sdk.Close()` 必须调用，否则 cookie 不会写入磁盘 |
| **登录状态判断** | 调用 `GetUserInfo()` 检查返回值是否为 nil 和 err 来判断 cookie 是否有效 |
| **跨域 Cookie** | 考试系统 (`changjiang-exam.yuketang.cn`) 与主站 (`changjiang.yuketang.cn`) 是不同的域，考试流程中会自动在两个域之间切换 |
| **CSRF Token** | `ExamGenToken` 接口需要 CSRF Token，SDK 内部自动从 cookie jar 提取 |
| **超时控制** | Client 默认 HTTP 超时 60 秒（见 `Test_AIFuck` 中自定义 `http.Client{Timeout: 60s}`），实际 SDK 内部使用持久化 cookie jar 的 client，超时取决于初始化配置 |
| **API Key 安全** | AI 做题功能需要 `DASHSCOPE_API_KEY` 环境变量，不要硬编码在代码中 |
