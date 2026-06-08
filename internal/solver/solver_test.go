package solver

import (
	"testing"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
)

func TestExtractJSONObject(t *testing.T) {
	raw := "```json\n{\"problem_id\":1,\"result\":[\"A\"]}\n```"
	got := extractJSONObject(raw)
	if got != "{\"problem_id\":1,\"result\":[\"A\"]}" {
		t.Fatalf("unexpected json block: %s", got)
	}
}

func TestParseAnswerFallbackString(t *testing.T) {
	answer, err := parseAnswer(`{"problem_id":999,"result":"A,B"}`, 123)
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

func TestExtractImageURLs(t *testing.T) {
	problem := models.ProblemsEntity{
		Body: `<p>test</p><img src="https://example.com/a.png">`,
	}
	urls := extractImageURLs(problem)
	if len(urls) != 1 || urls[0] != "https://example.com/a.png" {
		t.Fatalf("unexpected urls: %#v", urls)
	}
}
