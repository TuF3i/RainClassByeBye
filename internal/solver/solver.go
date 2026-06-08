package solver

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
	openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"RainClassByeBye/internal/logging"
)

type Config struct {
	APIKey              string
	BaseURL             string
	Model               string
	Timeout             time.Duration
	Temperature         float32
	MaxCompletionTokens int
	Logger              *logging.Logger
}

type Solver struct {
	model        *openaiModel.ChatModel
	modelName    string
	requestTTL   time.Duration
	systemPrompt string
	logger       *logging.Logger
}

type Answer struct {
	ProblemID int64    `json:"problem_id"`
	Result    []string `json:"result"`
}

func New(cfg Config) (*Solver, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key 不能为空")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("model 不能为空")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2 * time.Minute
	}

	temperature := cfg.Temperature
	maxTokens := cfg.MaxCompletionTokens
	chatModel, err := openaiModel.NewChatModel(context.Background(), &openaiModel.ChatModelConfig{
		APIKey:              cfg.APIKey,
		BaseURL:             cfg.BaseURL,
		Model:               cfg.Model,
		Timeout:             cfg.Timeout,
		Temperature:         &temperature,
		MaxCompletionTokens: &maxTokens,
	})
	if err != nil {
		return nil, err
	}

	return &Solver{
		model:      chatModel,
		modelName:  cfg.Model,
		requestTTL: cfg.Timeout,
		logger:     cfg.Logger,
		systemPrompt: strings.Join([]string{
			"你是一个雨课堂考试自动答题代理。",
			"你会收到题目 JSON 和若干题面图片 URL。",
			"你必须只返回 JSON，不能返回 Markdown、代码块或解释。",
			`返回格式固定为 {"problem_id":123,"result":["A"]}。`,
			"选择题请优先返回选项 key，例如 A/B/C/D。",
			"多选题返回多个选项 key。",
			"填空题或主观题返回字符串数组，每个字符串对应一个答案。",
			"如果你无法解析，就尽量给出最可能答案，但仍然必须输出合法 JSON。",
		}, "\n"),
	}, nil
}

func (s *Solver) ModelName() string {
	return s.modelName
}

func (s *Solver) Solve(ctx context.Context, problem models.ProblemsEntity) (Answer, string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, s.requestTTL)
	defer cancel()

	questionJSON, err := json.Marshal(problem)
	if err != nil {
		return Answer{}, "", err
	}

	parts := make([]schema.MessageInputPart, 0, 1+len(extractImageURLs(problem)))
	parts = append(parts, schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeText,
		Text: fmt.Sprintf("题目 JSON 如下，请直接作答：%s", string(questionJSON)),
	})
	for _, imageURL := range extractImageURLs(problem) {
		parts = append(parts, schema.MessageInputPart{
			Type: schema.ChatMessagePartTypeImageURL,
			Image: &schema.MessageInputImage{
				MessagePartCommon: schema.MessagePartCommon{URL: ptr(imageURL)},
				Detail:            schema.ImageURLDetailAuto,
			},
		})
	}

	resp, err := s.model.Generate(reqCtx, []*schema.Message{
		{
			Role:    schema.System,
			Content: s.systemPrompt,
		},
		{
			Role:                  schema.User,
			UserInputMultiContent: parts,
		},
	})
	if err != nil {
		return Answer{}, "", err
	}

	raw := strings.TrimSpace(resp.Content)
	answer, err := parseAnswer(raw, problem.ProblemId)
	if err != nil {
		return Answer{}, raw, err
	}
	return answer, raw, nil
}

func extractImageURLs(problem models.ProblemsEntity) []string {
	seen := make(map[string]struct{})
	urls := collectImageURLs(problem.Body, seen, nil)

	buf, err := json.Marshal(problem)
	if err != nil {
		return urls
	}
	return collectImageURLs(string(buf), seen, urls)
}

func parseAnswer(raw string, expectedProblemID int64) (Answer, error) {
	content := extractJSONObject(raw)
	var direct struct {
		ProblemID int64    `json:"problem_id"`
		Result    []string `json:"result"`
	}
	if err := json.Unmarshal([]byte(content), &direct); err == nil && len(direct.Result) > 0 {
		return Answer{
			ProblemID: expectedProblemID,
			Result:    normalizeResult(direct.Result),
		}, nil
	}

	var fallback struct {
		ProblemID int64 `json:"problem_id"`
		Result    any   `json:"result"`
	}
	if err := json.Unmarshal([]byte(content), &fallback); err != nil {
		return Answer{}, fmt.Errorf("解析模型 JSON 失败: %w", err)
	}

	normalized, err := normalizeAnyResult(fallback.Result)
	if err != nil {
		return Answer{}, err
	}
	return Answer{
		ProblemID: expectedProblemID,
		Result:    normalized,
	}, nil
}

func extractJSONObject(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func normalizeAnyResult(v any) ([]string, error) {
	switch val := v.(type) {
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			out = append(out, fmt.Sprint(item))
		}
		return normalizeResult(out), nil
	case string:
		return normalizeResult(splitLoose(val)), nil
	default:
		return nil, fmt.Errorf("result 字段格式不支持: %T", v)
	}
}

func normalizeResult(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, item := range in {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, "\"")
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func splitLoose(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			return arr
		}
	}
	if strings.Contains(raw, ",") {
		return strings.Split(raw, ",")
	}
	if strings.Contains(raw, " ") {
		return strings.Fields(raw)
	}
	return []string{raw}
}

func ptr[T any](v T) *T {
	return &v
}

func collectImageURLs(content string, seen map[string]struct{}, urls []string) []string {
	re := regexp.MustCompile(`(?i)<img[^>]*src\s*=\s*\\?["']([^"']*?)\\?["']`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		url := strings.TrimSpace(match[1])
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		urls = append(urls, url)
	}
	return urls
}
