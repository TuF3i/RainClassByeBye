package runner

import (
	"encoding/json"
	"testing"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
)

func TestBuildSubmitPaperPayloadUsesSnakeCaseFields(t *testing.T) {
	jsonBytes, err := buildSubmitPaperPayload(42, []models.SubmitPaperPostResultsEntity{
		{
			ProblemId:  1,
			Result:     []string{"A"},
			Time:       123,
			ShowAnswer: "",
			IsAnswered: true,
			IsSave:     true,
		},
	})
	if err != nil {
		t.Fatalf("buildSubmitPaperPayload returned error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(jsonBytes, &got); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}
	if got["exam_id"] != "42" {
		t.Fatalf("unexpected exam_id: %#v", got["exam_id"])
	}
	results, ok := got["results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("unexpected results payload: %#v", got["results"])
	}
	result, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected result item: %#v", results[0])
	}
	if _, ok := result["show_answer"]; !ok {
		t.Fatalf("missing show_answer: %#v", result)
	}
	if _, ok := result["is_answered"]; !ok {
		t.Fatalf("missing is_answered: %#v", result)
	}
	if _, ok := result["is_save"]; !ok {
		t.Fatalf("missing is_save: %#v", result)
	}
	if _, ok := result["showAnswer"]; ok {
		t.Fatalf("unexpected camelCase showAnswer: %#v", result)
	}
	if _, ok := result["isAnswered"]; ok {
		t.Fatalf("unexpected camelCase isAnswered: %#v", result)
	}
	if _, ok := result["isSave"]; ok {
		t.Fatalf("unexpected camelCase isSave: %#v", result)
	}

	answer, ok := result["result"].([]any)
	if !ok || len(answer) != 1 || answer[0] != "A" {
		t.Fatalf("unexpected answer payload: %#v", result["result"])
	}
}
