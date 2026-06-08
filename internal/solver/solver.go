package solver

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
	openaiModel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
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
	model        chatModel
	repairModel  chatModel
	modelName    string
	requestTTL   time.Duration
	systemPrompt string
	logger       *logging.Logger
}

type Answer struct {
	ProblemID int64    `json:"problem_id"`
	Result    []string `json:"result"`
}

type chatModel interface {
	Generate(ctx context.Context, in []*schema.Message, opts ...model.Option) (*schema.Message, error)
}

var (
	trailingCommaRE = regexp.MustCompile(`,(\s*[}\]])`)
	bareJSONKeyRE   = regexp.MustCompile(`([{,]\s*)([A-Za-z_][A-Za-z0-9_]*)(\s*:)`)
	resultFieldRE   = regexp.MustCompile(`(?is)(?:["']?(?:result|answer)["']?)\s*[:=]\s*(\[[^\]]*\]|"(?:\\.|[^"])*"|'(?:\\.|[^'])*'|[^,\r\n}]+)`)
	boxedAnswerRE   = regexp.MustCompile(`\\boxed\s*\{([^{}]+)\}`)
)

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

	chatModel, err := newChatModel(cfg, &openaiModel.ChatCompletionResponseFormat{
		Type: openaiModel.ChatCompletionResponseFormatTypeJSONObject,
	})
	if err != nil {
		return nil, err
	}

	repairModel, err := newChatModel(cfg, nil)
	if err != nil {
		return nil, err
	}

	return &Solver{
		model:       chatModel,
		repairModel: repairModel,
		modelName:   cfg.Model,
		requestTTL:  cfg.Timeout,
		logger:      cfg.Logger,
		systemPrompt: strings.Join([]string{
			"你是一个雨课堂考试自动答题代理。",
			"你会收到题目 JSON 和若干题面图片 URL。",
			"你必须只返回一个 JSON 对象，不能返回 Markdown、代码块、解释、题目复述、LaTeX 推导或任何前后缀。",
			"输出的第一个字符必须是 {，最后一个字符必须是 }。",
			`返回格式固定为 {"problem_id":123,"result":["A"]}。`,
			"problem_id 必须填写题目里的 problem_id。",
			"不要把 JSON 再包成字符串，也不要省略 key 或字符串的双引号。",
			"选择题请优先返回选项 key，例如 A/B/C/D。",
			"多选题返回多个选项 key。",
			"填空题或主观题只返回最终答案，返回字符串数组，每个字符串对应一个答案，不要输出解题过程。",
			"如果你无法解析，就尽量给出最可能答案，但仍然必须输出合法 JSON。",
		}, "\n"),
	}, nil
}

func newChatModel(cfg Config, responseFormat *openaiModel.ChatCompletionResponseFormat) (chatModel, error) {
	temperature := cfg.Temperature
	maxTokens := cfg.MaxCompletionTokens
	return openaiModel.NewChatModel(context.Background(), &openaiModel.ChatModelConfig{
		APIKey:              cfg.APIKey,
		BaseURL:             cfg.BaseURL,
		Model:               cfg.Model,
		Timeout:             cfg.Timeout,
		Temperature:         &temperature,
		MaxCompletionTokens: &maxTokens,
		ResponseFormat:      responseFormat,
	})
}

func (s *Solver) ModelName() string {
	return s.modelName
}

func (s *Solver) Solve(ctx context.Context, problem models.ProblemsEntity) (Answer, string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, s.requestTTL)
	defer cancel()

	baseMessages, err := s.buildQuestionMessages(problem)
	if err != nil {
		return Answer{}, "", err
	}

	raw, responseMeta, err := s.generateRaw(reqCtx, s.model, baseMessages)
	if err != nil {
		return Answer{}, "", err
	}

	answer, structuredErr := parseStructuredAnswer(raw, problem.ProblemId)
	if structuredErr == nil {
		answer, structuredErr = NormalizeForSubmission(problem, answer)
		if structuredErr == nil {
			return answer, raw, nil
		}
		structuredErr = fmt.Errorf("结构化答案不合法: %w", structuredErr)
	}
	structuredErr = annotateModelResponseError(structuredErr, raw, responseMeta)

	if s.logger != nil {
		if responseMeta != "" {
			s.logger.Warn("problem=%d 首次响应不是合法 JSON，准备请求模型重写；%s", problem.ProblemId, responseMeta)
		} else {
			s.logger.Warn("problem=%d 首次响应不是合法 JSON，准备请求模型重写", problem.ProblemId)
		}
	}

	repairedRaw, repairedMeta, repairErr := s.repairAnswer(reqCtx, baseMessages, problem, raw)
	if repairErr == nil {
		answer, repairedErr := parseStructuredAnswer(repairedRaw, problem.ProblemId)
		if repairedErr == nil {
			answer, repairedErr = NormalizeForSubmission(problem, answer)
			if repairedErr == nil {
				return answer, repairedRaw, nil
			}
			repairedErr = fmt.Errorf("结构化答案不合法: %w", repairedErr)
		}
		repairedErr = annotateModelResponseError(repairedErr, repairedRaw, repairedMeta)
		if answer, err = parseHeuristicAnswer(repairedRaw, problem); err == nil {
			answer, err = NormalizeForSubmission(problem, answer)
			if err == nil {
				return answer, repairedRaw, nil
			}
		}
		err = annotateModelResponseError(err, repairedRaw, repairedMeta)
	}

	if answer, err = parseHeuristicAnswer(raw, problem); err == nil {
		answer, err = NormalizeForSubmission(problem, answer)
		if err == nil {
			return answer, raw, nil
		}
	}
	err = annotateModelResponseError(err, raw, responseMeta)

	if fallback, ok := RandomChoiceFallback(problem); ok {
		if s.logger != nil {
			s.logger.Warn(
				"problem=%d 无法提取合法选项答案，使用随机合法选项兜底=%s",
				problem.ProblemId,
				strings.Join(fallback.Result, ","),
			)
		}
		if strings.TrimSpace(repairedRaw) != "" {
			return fallback, repairedRaw, nil
		}
		return fallback, raw, nil
	}

	if repairErr != nil {
		return Answer{}, raw, fmt.Errorf("%w；二次结构化失败: %v", structuredErr, repairErr)
	}
	return Answer{}, repairedRaw, fmt.Errorf("%w；二次结构化后仍无法解析: %v", structuredErr, err)
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

func parseAnswer(raw string, problem models.ProblemsEntity) (Answer, error) {
	answer, err := parseStructuredAnswer(raw, problem.ProblemId)
	if err == nil {
		return NormalizeForSubmission(problem, answer)
	}

	answer, heuristicErr := parseHeuristicAnswer(raw, problem)
	if heuristicErr == nil {
		return NormalizeForSubmission(problem, answer)
	}
	return Answer{}, err
}

func parseStructuredAnswer(raw string, expectedProblemID int64) (Answer, error) {
	var lastErr error
	for _, candidate := range answerCandidates(raw) {
		answer, err := parseStructuredCandidate(candidate, expectedProblemID)
		if err == nil {
			return answer, nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("未找到可解析的 JSON 对象")
	}
	return Answer{}, fmt.Errorf("解析模型 JSON 失败: %w，原始响应片段=%q", lastErr, truncateForError(raw, 160))
}

func parseHeuristicAnswer(raw string, problem models.ProblemsEntity) (Answer, error) {
	expectedProblemID := problem.ProblemId

	if answer, ok := parseLooseAnswer(raw, expectedProblemID); ok {
		return answer, nil
	}

	if answer, ok := parseOptionAnswer(raw, problem); ok {
		return answer, nil
	}

	if answer, ok := parseNarrativeAnswer(raw, problem); ok {
		return answer, nil
	}

	return Answer{}, fmt.Errorf("未找到可用答案，原始响应片段=%q", truncateForError(raw, 160))
}

func (s *Solver) buildQuestionMessages(problem models.ProblemsEntity) ([]*schema.Message, error) {
	questionJSON, err := json.Marshal(problem)
	if err != nil {
		return nil, err
	}

	parts := make([]schema.MessageInputPart, 0, 1+len(extractImageURLs(problem)))
	parts = append(parts, schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeText,
		Text: fmt.Sprintf(
			"只允许返回一个 JSON 对象，且第一个字符必须是 {，最后一个字符必须是 }。不要解释，不要复述题目，不要输出代码块。题目 JSON 如下：%s",
			string(questionJSON),
		),
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

	return []*schema.Message{
		{
			Role:    schema.System,
			Content: s.systemPrompt,
		},
		{
			Role:                  schema.User,
			UserInputMultiContent: parts,
		},
	}, nil
}

func (s *Solver) generateRaw(ctx context.Context, chat chatModel, messages []*schema.Message) (string, string, error) {
	resp, err := chat.Generate(ctx, messages)
	if err != nil {
		return "", "", err
	}
	return extractResponseText(resp), summarizeResponseMeta(resp), nil
}

func (s *Solver) repairAnswer(ctx context.Context, baseMessages []*schema.Message, problem models.ProblemsEntity, raw string) (string, string, error) {
	messages := append([]*schema.Message(nil), baseMessages...)
	if raw != "" {
		messages = append(messages, &schema.Message{
			Role:    schema.Assistant,
			Content: raw,
		})
	}

	messages = append(messages, &schema.Message{
		Role: schema.User,
		Content: fmt.Sprintf(
			"你上一个回答%s，无法解析为合法 JSON。请基于同一道题重新输出一个 JSON 对象。只允许输出形如 {\"problem_id\":%d,\"result\":[\"A\"]} 的内容，不要解释，不要代码块，不要额外文本。",
			previousAnswerStatus(raw),
			problem.ProblemId,
		),
	})
	modelToUse := s.repairModel
	if modelToUse == nil {
		modelToUse = s.model
	}
	return s.generateRaw(ctx, modelToUse, messages)
}

func previousAnswerStatus(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "为空字符串"
	}
	return "不是纯 JSON"
}

func extractJSONObject(raw string) string {
	trimmed := stripCodeFence(normalizePunctuation(raw))
	for _, candidate := range extractJSONObjects(trimmed) {
		return candidate
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

func NormalizeForSubmission(problem models.ProblemsEntity, answer Answer) (Answer, error) {
	result := normalizeResult(answer.Result)
	if len(result) == 0 {
		return Answer{}, fmt.Errorf("答案为空")
	}

	normalized := Answer{
		ProblemID: problem.ProblemId,
		Result:    result,
	}

	if len(problem.Options) == 0 {
		return normalized, nil
	}

	choiceResult := normalizeChoiceResult(result, optionKeySet(problem.Options))
	if len(choiceResult) == 0 {
		return Answer{}, fmt.Errorf("选择题答案不包含合法选项: %#v", answer.Result)
	}
	normalized.Result = choiceResult
	return normalized, nil
}

func RandomChoiceFallback(problem models.ProblemsEntity) (Answer, bool) {
	keys := sortedOptionKeys(problem.Options)
	if len(keys) == 0 {
		return Answer{}, false
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano() ^ problem.ProblemId))
	return Answer{
		ProblemID: problem.ProblemId,
		Result:    []string{keys[rng.Intn(len(keys))]},
	}, true
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

func answerCandidates(raw string) []string {
	candidates := make([]string, 0, 8)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		for _, existing := range candidates {
			if existing == v {
				return
			}
		}
		candidates = append(candidates, v)
	}

	normalized := stripCodeFence(normalizePunctuation(raw))
	add(normalized)
	for _, obj := range extractJSONObjects(normalized) {
		add(obj)
	}

	if unquoted, ok := tryUnquote(normalized); ok {
		unquoted = stripCodeFence(normalizePunctuation(unquoted))
		add(unquoted)
		for _, obj := range extractJSONObjects(unquoted) {
			add(obj)
		}
	}

	repaired := repairJSONLike(normalized)
	add(repaired)
	for _, obj := range extractJSONObjects(repaired) {
		add(obj)
	}

	if unquoted, ok := tryUnquote(repaired); ok {
		unquoted = stripCodeFence(normalizePunctuation(unquoted))
		add(unquoted)
		for _, obj := range extractJSONObjects(unquoted) {
			add(obj)
		}
	}

	return candidates
}

func parseStructuredCandidate(content string, expectedProblemID int64) (Answer, error) {
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
		return Answer{}, err
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

func parseLooseAnswer(raw string, expectedProblemID int64) (Answer, bool) {
	normalized := stripCodeFence(normalizePunctuation(raw))
	match := resultFieldRE.FindStringSubmatch(normalized)
	if len(match) < 2 {
		return Answer{}, false
	}

	result := parseLooseResultValue(match[1])
	if len(result) == 0 {
		return Answer{}, false
	}
	return Answer{
		ProblemID: expectedProblemID,
		Result:    result,
	}, true
}

func parseOptionAnswer(raw string, problem models.ProblemsEntity) (Answer, bool) {
	keys := optionKeySet(problem.Options)
	if len(keys) == 0 {
		return Answer{}, false
	}

	normalized := strings.TrimSpace(stripCodeFence(normalizePunctuation(raw)))
	if normalized == "" {
		return Answer{}, false
	}

	if result := parseDirectOptionTokens(normalized, keys); len(result) > 0 {
		return Answer{ProblemID: problem.ProblemId, Result: result}, true
	}

	for _, line := range reverseNonEmptyLines(normalized) {
		if !containsAnswerCue(line) {
			continue
		}
		if result := extractOptionKeys(line, keys); len(result) > 0 {
			return Answer{ProblemID: problem.ProblemId, Result: result}, true
		}
	}

	return Answer{}, false
}

func parseNarrativeAnswer(raw string, problem models.ProblemsEntity) (Answer, bool) {
	normalized := strings.TrimSpace(stripCodeFence(normalizePunctuation(raw)))
	if normalized == "" {
		return Answer{}, false
	}

	if boxed := extractBoxedAnswers(normalized); len(boxed) > 0 {
		return Answer{
			ProblemID: problem.ProblemId,
			Result:    boxed,
		}, true
	}

	if marked := extractNarrativeAnswer(normalized); marked != "" {
		return Answer{
			ProblemID: problem.ProblemId,
			Result:    []string{marked},
		}, true
	}

	freeform := sanitizeNarrative(normalized)
	if freeform == "" {
		return Answer{}, false
	}
	return Answer{
		ProblemID: problem.ProblemId,
		Result:    []string{freeform},
	}, true
}

func parseLooseResultValue(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "{}")
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "[]()")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	if unquoted, ok := tryUnquote(raw); ok {
		raw = strings.TrimSpace(unquoted)
	}

	raw = strings.ReplaceAll(raw, "'", `"`)
	raw = strings.Trim(raw, `"`)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	splitter := strings.NewReplacer("，", ",", "、", ",", ";", ",", "/", ",", "\n", ",", "\r", ",")
	raw = splitter.Replace(raw)

	parts := strings.Split(raw, ",")
	if len(parts) == 1 {
		return normalizeResult([]string{raw})
	}

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return normalizeResult(out)
}

func optionKeySet(options []models.OptionsEntity) map[string]struct{} {
	keys := make(map[string]struct{}, len(options))
	for _, option := range options {
		key := strings.ToUpper(strings.TrimSpace(option.Key))
		if key == "" {
			continue
		}
		keys[key] = struct{}{}
	}
	return keys
}

func sortedOptionKeys(options []models.OptionsEntity) []string {
	keySet := optionKeySet(options)
	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func normalizeChoiceResult(result []string, keys map[string]struct{}) []string {
	out := make([]string, 0, len(result))
	seen := make(map[string]struct{}, len(result))
	for _, item := range result {
		for _, key := range extractOptionKeys(item, keys) {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	return out
}

func parseDirectOptionTokens(raw string, keys map[string]struct{}) []string {
	candidate := sanitizeNarrative(raw)
	if candidate == "" {
		return nil
	}
	if strings.Contains(candidate, " ") {
		return nil
	}
	return extractOptionKeys(candidate, keys)
}

func extractOptionKeys(raw string, keys map[string]struct{}) []string {
	raw = strings.ToUpper(raw)
	raw = strings.NewReplacer(
		"答案", " ",
		"最终", " ",
		"应选", " ",
		"故选", " ",
		"选择", " ",
		"选项", " ",
		"RESULT", " ",
		":", " ",
		"：", " ",
		"=", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"，", " ",
		",", " ",
		"、", " ",
		"/", " ",
		";", " ",
		"；", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(raw)

	tokens := strings.Fields(raw)
	out := make([]string, 0, len(tokens))
	seen := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, ok := keys[token]; !ok {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

func reverseNonEmptyLines(raw string) []string {
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func containsAnswerCue(raw string) bool {
	return strings.Contains(raw, "答案") ||
		strings.Contains(raw, "最终") ||
		strings.Contains(raw, "故选") ||
		strings.Contains(raw, "应选") ||
		strings.Contains(raw, "结果") ||
		strings.Contains(strings.ToLower(raw), "result")
}

func extractBoxedAnswers(raw string) []string {
	matches := boxedAnswerRE.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		value := sanitizeNarrative(match[1])
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return normalizeResult(out)
}

func extractNarrativeAnswer(raw string) string {
	for _, line := range reverseNonEmptyLines(raw) {
		if !containsAnswerCue(line) {
			continue
		}

		candidate := line
		for _, sep := range []string{"：", ":", "="} {
			if idx := strings.LastIndex(candidate, sep); idx >= 0 && idx+len(sep) < len(candidate) {
				candidate = candidate[idx+len(sep):]
				break
			}
		}
		if idx := strings.LastIndex(candidate, "为"); idx >= 0 && idx+len("为") < len(candidate) {
			candidate = candidate[idx+len("为"):]
		}

		candidate = sanitizeNarrative(candidate)
		if candidate != "" {
			return candidate
		}
	}

	return ""
}

func sanitizeNarrative(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'`)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return strings.Join(strings.Fields(raw), " ")
}

func stripCodeFence(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		if nl := strings.Index(trimmed, "\n"); nl >= 0 {
			trimmed = trimmed[nl+1:]
		}
	}
	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}

func tryUnquote(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) < 2 {
		return "", false
	}
	if trimmed[0] != '"' && trimmed[0] != '`' {
		return "", false
	}
	if trimmed[len(trimmed)-1] != trimmed[0] {
		return "", false
	}
	unquoted, err := strconv.Unquote(trimmed)
	if err != nil {
		return "", false
	}
	return unquoted, true
}

func repairJSONLike(raw string) string {
	repaired := strings.TrimSpace(raw)
	repaired = strings.ReplaceAll(repaired, `\"`, `"`)
	repaired = strings.ReplaceAll(repaired, `'`, `"`)
	repaired = bareJSONKeyRE.ReplaceAllString(repaired, `$1"$2"$3`)
	repaired = trailingCommaRE.ReplaceAllString(repaired, `$1`)
	return repaired
}

func extractResponseText(msg *schema.Message) string {
	if msg == nil {
		return ""
	}

	if text := strings.TrimSpace(msg.Content); text != "" {
		return text
	}

	parts := make([]string, 0, len(msg.AssistantGenMultiContent))
	for _, part := range msg.AssistantGenMultiContent {
		switch part.Type {
		case schema.ChatMessagePartTypeText:
			if text := strings.TrimSpace(part.Text); text != "" {
				parts = append(parts, text)
			}
		case schema.ChatMessagePartTypeReasoning:
			if part.Reasoning != nil {
				if text := strings.TrimSpace(part.Reasoning.Text); text != "" {
					parts = append(parts, text)
				}
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func summarizeResponseMeta(msg *schema.Message) string {
	if msg == nil {
		return "模型返回 nil 响应"
	}

	parts := make([]string, 0, 5)
	if msg.ResponseMeta != nil {
		if finishReason := strings.TrimSpace(msg.ResponseMeta.FinishReason); finishReason != "" {
			parts = append(parts, "finish_reason="+finishReason)
		}
		if msg.ResponseMeta.Usage != nil {
			parts = append(parts, fmt.Sprintf("completion_tokens=%d", msg.ResponseMeta.Usage.CompletionTokens))
		}
	}
	if reasoning := strings.TrimSpace(msg.ReasoningContent); reasoning != "" {
		parts = append(parts, fmt.Sprintf("reasoning_len=%d", len([]rune(reasoning))))
	}
	if len(msg.AssistantGenMultiContent) > 0 {
		parts = append(parts, fmt.Sprintf("output_parts=%d", len(msg.AssistantGenMultiContent)))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

func annotateModelResponseError(err error, raw string, responseMeta string) error {
	if err == nil {
		return nil
	}
	if strings.TrimSpace(raw) != "" || responseMeta == "" {
		return err
	}
	return fmt.Errorf("%w；模型响应元信息: %s", err, responseMeta)
}

func normalizePunctuation(raw string) string {
	replacer := strings.NewReplacer(
		"“", `"`,
		"”", `"`,
		"‘", `'`,
		"’", `'`,
		"：", ":",
		"，", ",",
		"（", "(",
		"）", ")",
		"【", "[",
		"】", "]",
		"｛", "{",
		"｝", "}",
	)
	return replacer.Replace(raw)
}

func extractJSONObjects(raw string) []string {
	objects := make([]string, 0, 2)
	start := -1
	depth := 0
	inString := false
	escaped := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}

		switch ch {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 && start >= 0 {
				objects = append(objects, raw[start:i+1])
				start = -1
			}
		}
	}

	return objects
}

func truncateForError(raw string, limit int) string {
	raw = strings.TrimSpace(raw)
	if limit <= 0 || len(raw) <= limit {
		return raw
	}
	return raw[:limit] + "..."
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
