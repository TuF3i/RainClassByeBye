package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
)

const (
	StatusPending       = "pending"
	StatusRunning       = "running"
	StatusInterrupted   = "interrupted"
	StatusPartial       = "partial"
	StatusReadyToSubmit = "ready_to_submit"
	StatusCompleted     = "completed"
)

type ExamState struct {
	Version        int                       `json:"version"`
	CID            int64                     `json:"cid"`
	ExamID         int64                     `json:"exam_id"`
	ExamTitle      string                    `json:"exam_title"`
	CookiePath     string                    `json:"cookie_path"`
	Status         string                    `json:"status"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
	TotalProblems  int                       `json:"total_problems"`
	SubmittedPaper bool                      `json:"submitted_paper"`
	LastError      string                    `json:"last_error,omitempty"`
	Answered       map[string]AnsweredRecord `json:"answered"`
	Failed         map[string]FailedRecord   `json:"failed"`
}

type AnsweredRecord struct {
	ProblemID         int64    `json:"problem_id"`
	ProblemIndex      int64    `json:"problem_index"`
	ProblemType       string   `json:"problem_type"`
	Result            []string `json:"result"`
	Model             string   `json:"model"`
	ModelRawOutput    string   `json:"model_raw_output"`
	SubmittedAtUnixMs int64    `json:"submitted_at_unix_ms"`
}

type FailedRecord struct {
	ProblemID int64     `json:"problem_id"`
	Attempts  int       `json:"attempts"`
	LastError string    `json:"last_error"`
	UpdatedAt time.Time `json:"updated_at"`
}

func DefaultPath(dir string, cid, examID int64) string {
	return filepath.Join(dir, fmt.Sprintf("%d_%d.json", cid, examID))
}

func New(cid, examID int64, cookiePath string) *ExamState {
	now := time.Now().UTC()
	return &ExamState{
		Version:    1,
		CID:        cid,
		ExamID:     examID,
		CookiePath: cookiePath,
		Status:     StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
		Answered:   make(map[string]AnsweredRecord),
		Failed:     make(map[string]FailedRecord),
	}
}

func Load(path string) (*ExamState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var st ExamState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	if st.Answered == nil {
		st.Answered = make(map[string]AnsweredRecord)
	}
	if st.Failed == nil {
		st.Failed = make(map[string]FailedRecord)
	}
	return &st, nil
}

func Save(path string, st *ExamState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	st.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func ProblemKey(problemID int64) string {
	return strconv.FormatInt(problemID, 10)
}

func (s *ExamState) MarkStarted(total int, title string) {
	s.Status = StatusRunning
	s.TotalProblems = total
	if title != "" {
		s.ExamTitle = title
	}
}

func (s *ExamState) MarkAnswered(problem models.ProblemsEntity, result []string, modelName, rawOutput string) {
	key := ProblemKey(problem.ProblemId)
	s.Answered[key] = AnsweredRecord{
		ProblemID:         problem.ProblemId,
		ProblemIndex:      problem.Index,
		ProblemType:       problem.TypeText,
		Result:            append([]string(nil), result...),
		Model:             modelName,
		ModelRawOutput:    rawOutput,
		SubmittedAtUnixMs: time.Now().UnixMilli(),
	}
	delete(s.Failed, key)
	s.LastError = ""
}

func (s *ExamState) MarkFailure(problemID int64, err error) {
	key := ProblemKey(problemID)
	rec := s.Failed[key]
	rec.ProblemID = problemID
	rec.Attempts++
	if err != nil {
		rec.LastError = err.Error()
		s.LastError = err.Error()
	}
	rec.UpdatedAt = time.Now().UTC()
	s.Failed[key] = rec
}

func (s *ExamState) MarkInterrupted(err error) {
	s.Status = StatusInterrupted
	if err != nil {
		s.LastError = err.Error()
	}
}

func (s *ExamState) MarkPartial() {
	s.Status = StatusPartial
}

func (s *ExamState) MarkReadyToSubmit() {
	s.Status = StatusReadyToSubmit
}

func (s *ExamState) MarkCompleted() {
	s.Status = StatusCompleted
	s.SubmittedPaper = true
}

func (s *ExamState) IsAnswered(problemID int64) bool {
	_, ok := s.Answered[ProblemKey(problemID)]
	return ok
}

func (s *ExamState) Pending(problems []models.ProblemsEntity) []models.ProblemsEntity {
	pending := make([]models.ProblemsEntity, 0, len(problems))
	for _, problem := range problems {
		if s.IsAnswered(problem.ProblemId) {
			continue
		}
		pending = append(pending, problem)
	}
	return pending
}

func (s *ExamState) AnsweredCount() int {
	return len(s.Answered)
}

func (s *ExamState) FailedCount() int {
	return len(s.Failed)
}

func (s *ExamState) BuildSubmitPaperResults(problems []models.ProblemsEntity) ([]models.SubmitPaperPostResultsEntity, error) {
	results := make([]models.SubmitPaperPostResultsEntity, 0, len(problems))
	for _, problem := range problems {
		record, ok := s.Answered[ProblemKey(problem.ProblemId)]
		if !ok {
			return nil, fmt.Errorf("problem %d 尚未完成，无法交卷", problem.ProblemId)
		}
		results = append(results, models.SubmitPaperPostResultsEntity{
			ProblemId:  problem.ProblemId,
			Result:     append([]string(nil), record.Result...),
			Time:       record.SubmittedAtUnixMs,
			ShowAnswer: "",
			IsAnswered: true,
			IsSave:     true,
		})
	}
	return results, nil
}
