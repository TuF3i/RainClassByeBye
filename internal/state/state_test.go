package state

import (
	"testing"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
)

func TestPending(t *testing.T) {
	st := New(1, 2, "cookies.json")
	st.MarkAnswered(models.ProblemsEntity{ProblemId: 10}, []string{"A"}, "model", `{"result":["A"]}`)

	problems := []models.ProblemsEntity{
		{ProblemId: 10},
		{ProblemId: 20},
	}
	pending := st.Pending(problems)
	if len(pending) != 1 || pending[0].ProblemId != 20 {
		t.Fatalf("unexpected pending problems: %#v", pending)
	}
}

func TestBuildSubmitPaperResults(t *testing.T) {
	st := New(1, 2, "cookies.json")
	st.MarkAnswered(models.ProblemsEntity{ProblemId: 10}, []string{"A"}, "model", `{"result":["A"]}`)
	st.MarkAnswered(models.ProblemsEntity{ProblemId: 20}, []string{"B"}, "model", `{"result":["B"]}`)

	results, err := st.BuildSubmitPaperResults([]models.ProblemsEntity{
		{ProblemId: 10},
		{ProblemId: 20},
	})
	if err != nil {
		t.Fatalf("BuildSubmitPaperResults returned error: %v", err)
	}
	if len(results) != 2 || results[0].ProblemId != 10 || results[1].ProblemId != 20 {
		t.Fatalf("unexpected results: %#v", results)
	}
}
