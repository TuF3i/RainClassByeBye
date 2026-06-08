package solver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type stubChatModel struct {
	responses []*schema.Message
	calls     [][]*schema.Message
}

func (m *stubChatModel) Generate(_ context.Context, in []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	copied := make([]*schema.Message, len(in))
	copy(copied, in)
	m.calls = append(m.calls, copied)
	if len(m.responses) == 0 {
		return &schema.Message{}, nil
	}
	resp := m.responses[0]
	m.responses = m.responses[1:]
	return resp, nil
}

func TestExtractJSONObject(t *testing.T) {
	raw := "```json\n{\"problem_id\":1,\"result\":[\"A\"]}\n```"
	got := extractJSONObject(raw)
	if got != "{\"problem_id\":1,\"result\":[\"A\"]}" {
		t.Fatalf("unexpected json block: %s", got)
	}
}

func TestParseAnswerFallbackString(t *testing.T) {
	answer, err := parseAnswer(`{"problem_id":999,"result":"A,B"}`, models.ProblemsEntity{ProblemId: 123})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if answer.ProblemID != 123 {
		t.Fatalf("expected forced problem id 123, got %d", answer.ProblemID)
	}
	if len(answer.Result) != 2 || answer.Result[0] != "A" || answer.Result[1] != "B" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestParseAnswerEscapedJSONString(t *testing.T) {
	answer, err := parseAnswer(`"{\"problem_id\":999,\"result\":[\"C\"]}"`, models.ProblemsEntity{ProblemId: 123})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "C" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestParseAnswerBareKeysAndBareArrayValues(t *testing.T) {
	answer, err := parseAnswer("```json\n{problem_id: 999, result: [A, B],}\n```", models.ProblemsEntity{ProblemId: 123})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if len(answer.Result) != 2 || answer.Result[0] != "A" || answer.Result[1] != "B" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestParseAnswerSingleQuotedObject(t *testing.T) {
	answer, err := parseAnswer(`{'problem_id': 999, 'result': ['D']}`, models.ProblemsEntity{ProblemId: 123})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "D" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestParseAnswerLooseResultField(t *testing.T) {
	answer, err := parseAnswer("answer = (A/C)", models.ProblemsEntity{ProblemId: 123})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if len(answer.Result) != 2 || answer.Result[0] != "A" || answer.Result[1] != "C" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestParseAnswerChineseOptionCue(t *testing.T) {
	answer, err := parseAnswer("分析后可知，答案：C", models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
			{Key: "C"},
			{Key: "D"},
		},
	})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "C" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestParseAnswerRejectsNarrativeJSONResultForChoiceQuestion(t *testing.T) {
	_, err := parseAnswer(`{"problem_id":123,"result":["用户要求我解答一道数学题,并以特定的JSON格式返回答案。"]}`, models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
			{Key: "C"},
			{Key: "D"},
		},
	})
	if err == nil {
		t.Fatal("expected parseAnswer to reject narrative choice result")
	}
}

func TestNormalizeForSubmissionExtractsChoiceKeys(t *testing.T) {
	answer, err := NormalizeForSubmission(models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
			{Key: "C"},
			{Key: "D"},
		},
	}, Answer{
		ProblemID: 123,
		Result:    []string{"答案应该选 c"},
	})
	if err != nil {
		t.Fatalf("NormalizeForSubmission returned error: %v", err)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "C" {
		t.Fatalf("unexpected normalized result: %#v", answer.Result)
	}
}

func TestRandomChoiceFallbackReturnsLegalOption(t *testing.T) {
	answer, ok := RandomChoiceFallback(models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
			{Key: "C"},
			{Key: "D"},
		},
	})
	if !ok {
		t.Fatal("expected fallback answer")
	}
	if len(answer.Result) != 1 {
		t.Fatalf("unexpected fallback result: %#v", answer.Result)
	}
	switch answer.Result[0] {
	case "A", "B", "C", "D":
	default:
		t.Fatalf("fallback result is not a legal option: %#v", answer.Result)
	}
}

func TestParseAnswerBoxedNarrative(t *testing.T) {
	answer, err := parseAnswer("题目要求计算积分。\n最终可得 $\\boxed{0}$。", models.ProblemsEntity{ProblemId: 123})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "0" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestParseAnswerFreeformNarrativeFallback(t *testing.T) {
	raw := "题目要求计算曲线积分。\n沿边界方向积分后结果收敛到 0。"
	answer, err := parseAnswer(raw, models.ProblemsEntity{ProblemId: 123})
	if err != nil {
		t.Fatalf("parseAnswer returned error: %v", err)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "沿边界方向积分后结果收敛到 0。" {
		t.Fatalf("unexpected result: %#v", answer.Result)
	}
}

func TestExtractImageURLs(t *testing.T) {
	problem := models.ProblemsEntity{
		Body: `<p>test</p><img src="https://example.com/a.png">`,
	}
	urls := extractImageURLs(problem)
	if len(urls) != 1 || urls[0] != "https://example.com/a.png" {
		t.Fatalf("unexpected urls: %#v", urls)
	}
}

func TestSolveRetriesWhenFirstResponseIsEmpty(t *testing.T) {
	model := &stubChatModel{
		responses: []*schema.Message{
			{Content: ""},
			{Content: `{"problem_id":123,"result":["A"]}`},
		},
	}
	s := &Solver{
		model:        model,
		modelName:    "test-model",
		requestTTL:   time.Second,
		systemPrompt: "system",
	}

	answer, raw, err := s.Solve(context.Background(), models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
		},
	})
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if raw != `{"problem_id":123,"result":["A"]}` {
		t.Fatalf("unexpected raw: %q", raw)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "A" {
		t.Fatalf("unexpected answer: %#v", answer)
	}
	if len(model.calls) != 2 {
		t.Fatalf("expected 2 model calls, got %d", len(model.calls))
	}
	if len(model.calls[1]) != 3 {
		t.Fatalf("expected repair conversation to have 3 messages, got %d", len(model.calls[1]))
	}
}

func TestSolveRetriesWhenFirstResponseIsNonJSON(t *testing.T) {
	model := &stubChatModel{
		responses: []*schema.Message{
			{Content: "答案应该选 A"},
			{Content: `{"problem_id":123,"result":["A"]}`},
		},
	}
	s := &Solver{
		model:        model,
		modelName:    "test-model",
		requestTTL:   time.Second,
		systemPrompt: "system",
	}

	answer, _, err := s.Solve(context.Background(), models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
		},
	})
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "A" {
		t.Fatalf("unexpected answer: %#v", answer)
	}
	if len(model.calls) != 2 {
		t.Fatalf("expected 2 model calls, got %d", len(model.calls))
	}
	if len(model.calls[1]) != 4 {
		t.Fatalf("expected repair conversation to include assistant response, got %d messages", len(model.calls[1]))
	}
}

func TestSolveUsesPlainRepairModel(t *testing.T) {
	primary := &stubChatModel{
		responses: []*schema.Message{
			{Content: ""},
		},
	}
	repair := &stubChatModel{
		responses: []*schema.Message{
			{Content: `{"problem_id":123,"result":["A"]}`},
		},
	}
	s := &Solver{
		model:        primary,
		repairModel:  repair,
		modelName:    "test-model",
		requestTTL:   time.Second,
		systemPrompt: "system",
	}

	answer, raw, err := s.Solve(context.Background(), models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
		},
	})
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if raw != `{"problem_id":123,"result":["A"]}` {
		t.Fatalf("unexpected raw: %q", raw)
	}
	if len(answer.Result) != 1 || answer.Result[0] != "A" {
		t.Fatalf("unexpected answer: %#v", answer)
	}
	if len(primary.calls) != 1 {
		t.Fatalf("expected 1 primary call, got %d", len(primary.calls))
	}
	if len(repair.calls) != 1 {
		t.Fatalf("expected 1 repair call, got %d", len(repair.calls))
	}
}

func TestSolveReportsEmptyResponseMetadata(t *testing.T) {
	model := &stubChatModel{
		responses: []*schema.Message{
			{
				ResponseMeta: &schema.ResponseMeta{
					FinishReason: "stop",
					Usage: &schema.TokenUsage{
						CompletionTokens: 32,
					},
				},
				ReasoningContent: "thinking",
			},
			{
				ResponseMeta: &schema.ResponseMeta{
					FinishReason: "stop",
				},
			},
		},
	}
	s := &Solver{
		model:        model,
		modelName:    "test-model",
		requestTTL:   time.Second,
		systemPrompt: "system",
	}

	_, _, err := s.Solve(context.Background(), models.ProblemsEntity{
		ProblemId: 123,
	})
	if err == nil {
		t.Fatal("expected Solve to fail")
	}
	if !strings.Contains(err.Error(), "finish_reason=stop") {
		t.Fatalf("expected finish reason in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "reasoning_len=8") {
		t.Fatalf("expected reasoning length in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "completion_tokens=32") {
		t.Fatalf("expected token usage in error, got: %v", err)
	}
}

func TestSolveFallsBackToLegalChoiceOption(t *testing.T) {
	model := &stubChatModel{
		responses: []*schema.Message{
			{Content: `{"problem_id":123,"result":["用户要求我解答一道数学题,并以特定的JSON格式返回答案。"]}`},
			{Content: `{"problem_id":123,"result":["还是返回说明文本"]}`},
		},
	}
	s := &Solver{
		model:        model,
		modelName:    "test-model",
		requestTTL:   time.Second,
		systemPrompt: "system",
	}

	answer, _, err := s.Solve(context.Background(), models.ProblemsEntity{
		ProblemId: 123,
		Options: []models.OptionsEntity{
			{Key: "A"},
			{Key: "B"},
			{Key: "C"},
			{Key: "D"},
		},
	})
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if len(answer.Result) != 1 {
		t.Fatalf("unexpected answer length: %#v", answer.Result)
	}
	switch answer.Result[0] {
	case "A", "B", "C", "D":
	default:
		t.Fatalf("Solve fallback is not a legal option: %#v", answer.Result)
	}
	if len(model.calls) != 2 {
		t.Fatalf("expected 2 model calls, got %d", len(model.calls))
	}
}
