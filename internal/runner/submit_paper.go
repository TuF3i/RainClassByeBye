package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"
	persistentcookiejar "github.com/cdwiegand/persistent-cookiejar"
)

const defaultSubmitPaperURL = "https://changjiang-exam.yuketang.cn/exam_room/submit_paper"

var submitPaperURL = defaultSubmitPaperURL

type submitPaperRequest struct {
	Results []submitPaperRequestResult `json:"results"`
	ExamID  string                     `json:"exam_id"`
}

type submitPaperRequestResult struct {
	ProblemID  int64    `json:"problem_id"`
	Result     []string `json:"result"`
	Time       int64    `json:"time"`
	ShowAnswer string   `json:"show_answer"`
	IsAnswered bool     `json:"is_answered"`
	IsSave     bool     `json:"is_save"`
}

type submitPaperResponse struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func buildSubmitPaperPayload(examID int64, answerList []models.SubmitPaperPostResultsEntity) ([]byte, error) {
	reqResults := make([]submitPaperRequestResult, 0, len(answerList))
	for _, answer := range answerList {
		reqResults = append(reqResults, submitPaperRequestResult{
			ProblemID:  answer.ProblemId,
			Result:     append([]string(nil), answer.Result...),
			Time:       answer.Time,
			ShowAnswer: answer.ShowAnswer,
			IsAnswered: answer.IsAnswered,
			IsSave:     answer.IsSave,
		})
	}

	reqBody := submitPaperRequest{
		Results: reqResults,
		ExamID:  strconv.FormatInt(examID, 10),
	}

	return json.Marshal(reqBody)
}

func submitPaper(cookiePath string, examID int64, answerList []models.SubmitPaperPostResultsEntity) (*submitPaperResponse, error) {
	jar, err := persistentcookiejar.New(&persistentcookiejar.Options{Filename: cookiePath})
	if err != nil {
		return nil, err
	}

	jsonBytes, err := buildSubmitPaperPayload(examID, answerList)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, submitPaperURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xtbz", "cloud")
	req.Header.Set("x-client", "web")

	client := &http.Client{Jar: jar}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var submitResp submitPaperResponse
	if err := json.NewDecoder(resp.Body).Decode(&submitResp); err != nil {
		return nil, err
	}
	if err := jar.Save(); err != nil {
		return nil, fmt.Errorf("保存 cookie 失败: %w", err)
	}

	return &submitResp, nil
}
