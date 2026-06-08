//go:build devref
// +build devref

package RainClassSDK

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
)

func TestSDK_QRLogin(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	err = sdk.QRLogin()
	if err != nil {
		t.Errorf("func sdk.QRLogin error: %v", err.Error())
	}

	data, err := sdk.GetUserInfo()
	if err != nil {
		t.Errorf("func sdk.GetUserInfo error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestSDK_GetUserInfo(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.GetUserInfo()
	if err != nil {
		t.Errorf("func sdk.GetUserInfo error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestSDK_GetCourseInfo(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.GetCourseInfo()
	if err != nil {
		t.Errorf("func sdk.GetCourseInfo error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestSDK_GetHomeWorkInfo(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.GetHomeWorkInfo(24211265)
	if err != nil {
		t.Errorf("func sdk.GetHomeWorkInfo error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestSDK_GetHomeWorkDetails(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.GetHomeWorkDetails(24211265, 45505698)
	if err != nil {
		t.Errorf("func sdk.GetHomeWorkDetails error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestSDK_GetHomeWorkCover(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.GetHomeWorkCover(24211265, 1945460)
	if err != nil {
		t.Errorf("func sdk.GetHomeWorkCover error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_ExamGenToken(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.client.ExamGenToken(24211265, 1945538)
	if err != nil {
		t.Errorf("func sdk.client.ExamGenToken error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_ExamLogin(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	err = sdk.client.ExamLogin(1945538, 57510820, "H1tf/67TNyDZkhZjQERTTLJtZUE1oQWUL5HpmMltzZR+GyeYX56u6vPDjTo+4Gmhn9kAcO+NDSNNk0qIMuTboA==")
	if err != nil {
		t.Errorf("func sdk.client.ExamLogin error: %v", err.Error())
	}

	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_StartExam(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	err = sdk.client.StartExam(1945538)
	if err != nil {
		t.Errorf("func sdk.client.ExamLogin error: %v", err.Error())
	}

	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_StartExamPaper(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.client.StartExamPaper(1945538)
	if err != nil {
		t.Errorf("func sdk.client.StartExamPaper error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_GetExamPaperCover(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.client.GetExamPaperCover(1945538)
	if err != nil {
		t.Errorf("func sdk.client.GetExamPaperCover error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_GetExamPaperQuestion(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.client.GetExamPaperQuestion(1945460)
	if err != nil {
		t.Errorf("func sdk.client.GetExamPaperQuestion error: %v", err.Error())
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err = enc.Encode(data)
	jsonBytes := buf.Bytes()
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_RefreshTimeRemaining(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	data, err := sdk.client.RefreshTimeRemaining(1945538)
	if err != nil {
		t.Errorf("func sdk.client.RefreshTimeRemaining error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_SubmitAnswer(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	ans := models.SubmitAnswerResultsEntity{
		ProblemId: 92134642,
		Result:    []string{"A"},
		Time:      time.Now().UnixMilli(),
	}

	data, err := sdk.client.SubmitAnswer(1945538, ans)
	if err != nil {
		t.Errorf("func sdk.client.RefreshTimeRemaining error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestClient_SubmitPaper(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	ansList := []models.SubmitPaperPostResultsEntity{
		models.SubmitPaperPostResultsEntity{
			ProblemId:  92134642,
			Result:     []string{"A"},
			Time:       time.Now().UnixMilli(),
			ShowAnswer: "",
			IsAnswered: true,
			IsSave:     true,
		},
	}

	data, err := sdk.client.SubmitPaper(1945538, ansList)
	if err != nil {
		t.Errorf("func sdk.client.RefreshTimeRemaining error: %v", err.Error())
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

func TestSDK_StartExam(t *testing.T) {
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
	}

	err = sdk.StartExam(24211265, 1945538)
	if err != nil {
		t.Errorf("func sdk.GetHomeWorkCover error: %v", err.Error())
	}

	err = sdk.Close()
	if err != nil {
		t.Errorf("func sdk.Close error: %v", err.Error())
	}
}

// ######## AI TEST ######## //
type DeepSeekChatCompletionResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
		ReasoningTokens  int `json:"reasoning_tokens"`
	} `json:"usage"`
}

type Images struct {
	OriginalUrl string
	Base64      string
}

type DeepSeekAnswer struct {
	ProblemId int64    `json:"problem_id"`
	Result    []string `json:"result"`
}

func ImageURLToBase64(url string) (string, error) {
	client := http.Client{
		Timeout: 10 * time.Second, // 设置超时，避免长时间等待
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("图片下载失败，状态码: %d", resp.StatusCode)
	}

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %w", err)
	}

	// 标准 Base64 编码（不含换行）
	base64Str := base64.StdEncoding.EncodeToString(imgBytes)
	return base64Str, nil
}

func GenImageB64(data models.ProblemsEntity) []Images {
	var res []Images

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(data)

	if err != nil {
		return nil
	}

	jsonBytes := buf.Bytes()

	imgRegex := regexp.MustCompile(`(?i)<img[^>]*src\s*=\s*\\?["']([^"']*?)\\?["']`)
	urlMatches := imgRegex.FindAllStringSubmatch(string(jsonBytes), -1)

	for _, m := range urlMatches {
		if len(m) >= 2 {
			b64, err := ImageURLToBase64(m[1])
			if err != nil {
				continue
			}
			res = append(res, Images{OriginalUrl: m[1], Base64: b64})
		}
	}
	return res
}

func Test_GetImgB64(t *testing.T) {
	var res []Images

	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	if err != nil {
		panic(err)
	}

	data, err := sdk.client.GetExamPaperQuestion(1945538)
	if err != nil {
		panic(err)
	}

	for _, problem := range data.Data.Problems {
		res = append(res, GenImageB64(problem)...)
	}

	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Errorf("func json.MarshalIndent error: %v", err.Error())
	}

	t.Logf("%v", string(jsonBytes))
}

func Test_AIFuck(t *testing.T) {
	// 使用阿里云 DashScope API Key（请从环境变量读取，不要硬编码）
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		t.Fatal("DASHSCOPE_API_KEY environment variable not set")
	}

	var cid int64 = 24211265
	var examid int64 = 1945460

	var ansList []models.SubmitPaperPostResultsEntity

	appendAnswer := func(problemID int64, result []string) {
		ans := models.SubmitPaperPostResultsEntity{
			ProblemId:  problemID,
			Result:     result,
			Time:       time.Now().UnixMilli(),
			ShowAnswer: "",
			IsAnswered: true,
			IsSave:     true,
		}
		ansList = append(ansList, ans)
	}

	genAns := func(problemID int64, result []string) models.SubmitAnswerResultsEntity {
		return models.SubmitAnswerResultsEntity{
			ProblemId: problemID,
			Result:    result,
			Time:      time.Now().UnixMilli(),
		}
	}

	// 合并系统提示为一条消息
	systemPrompt := strings.Join([]string{
		"你是重庆邮电大学研制的数学题解题机器人，接下来，你要用你的智慧解决数学题",
		"现在，你会收到一些题目，你做完后要严格返回{\"problem_id\":这题的problem_id(number类型), \"result\":这题的答案(一个字符串数组)}, 不要有Markdown格式",
		"这里是一个例子: { \"problem_id\": 92133234, \"result\": [\"A\"] }",
		"如果出现你无法处理的错误就返回: { \"problem_id\": -1, \"result\": [] }",
	}, "\n")

	// 初始化 SDK（保持不变）
	sdk, err := NewSDK(WithCookiePath("./test_cookies.json"))
	defer sdk.Close()
	if err != nil {
		t.Errorf("func NewSDK error: %v", err.Error())
		return
	}

	err = sdk.StartExam(cid, examid)
	if err != nil {
		t.Errorf("func sdk.GetHomeWorkCover error: %v", err.Error())
		return
	}

	data, err := sdk.GetExamPaperQuestion(examid)
	if err != nil {
		t.Errorf("func sdk.GetExamPaperQuestion error: %v", err.Error())
		return
	}

	if data == nil {
		t.Errorf("data is nil")
		return
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	for i, problem := range data.Data.Problems {
		t.Logf("Running: %v / %v", i, len(data.Data.Problems))

		// 获取图片信息（包含 URL 和 Base64，但只使用 URL）
		images := GenImageB64(problem) // 返回 []Images
		problemContent, _ := json.Marshal(problem)

		// 构建多模态 content 数组
		contentParts := make([]map[string]interface{}, 0)

		// 添加所有图片（直接使用原始 URL）
		for _, img := range images {
			if img.OriginalUrl != "" {
				contentParts = append(contentParts, map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]string{
						"url": img.OriginalUrl,
					},
				})
			}
		}

		// 添加文本（题目元数据）
		contentParts = append(contentParts, map[string]interface{}{
			"type": "text",
			"text": fmt.Sprintf("这是题目和一些其他元数据: %s", string(problemContent)),
		})

		// 构建 messages
		messages := []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": contentParts,
			},
		}

		// 构建请求体
		requestBody := map[string]interface{}{
			"model":    "qwen3.6-plus", // 通义千问视觉语言模型
			"messages": messages,
			"stream":   false,
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			t.Errorf("json.Marshal error: %v", err.Error())
			continue
		}

		// 创建请求（使用阿里云端点）
		ctx, cancle := context.WithTimeout(context.Background(), 2*time.Minute)
		req, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Errorf("http.NewRequest error: %v", err.Error())
			cancle()
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("client.Do error: %v", err.Error())
			cancle()
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("io.ReadAll error: %v", err.Error())
			continue
		}

		cancle()

		// 解析响应
		var chatResp DeepSeekChatCompletionResponse
		err = json.Unmarshal(body, &chatResp)
		if err != nil {
			t.Errorf("json.Unmarshal error: %v, body: %s", err.Error(), string(body))
			continue
		}

		if len(chatResp.Choices) == 0 {
			t.Errorf("chatResp.Choices is empty, full response: %s", string(body))
			continue
		}

		content := chatResp.Choices[0].Message.Content
		t.Logf("Qwen Response: %v", content)

		var ansJSON DeepSeekAnswer
		err = json.Unmarshal([]byte(content), &ansJSON)
		if err != nil {
			t.Errorf("json.Unmarshal error: %v, content: %s", err.Error(), content)
			continue
		}

		t.Logf("ProblemID: %v \n", ansJSON.ProblemId)
		t.Logf("Img: %v \n", images)
		t.Logf("Ans: %v \n", ansJSON.Result)

		d1, err := sdk.SubmitAnswer(examid, genAns(ansJSON.ProblemId, ansJSON.Result))
		jsonBytes, _ := json.Marshal(d1)
		t.Logf("SubmitAnswer response: %v", string(jsonBytes))

		t.Log("======================================================")

		appendAnswer(ansJSON.ProblemId, ansJSON.Result)
	}

	// 最后提交试卷
	//jsonBytes, _ := json.MarshalIndent(ansList, "", "  ")
	//t.Logf("All answers: %v", string(jsonBytes))
}
