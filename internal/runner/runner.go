package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	RainClassSDK "github.com/Auto-CQUPT-Plan/RainClassSDK"
	"github.com/Auto-CQUPT-Plan/RainClassSDK/models"

	"RainClassByeBye/internal/logging"
	"RainClassByeBye/internal/solver"
	"RainClassByeBye/internal/state"
)

type Options struct {
	CID         int64
	ExamID      int64
	CookiePath  string
	StatePath   string
	Workers     int
	SubmitPaper bool
	Resume      bool
	Logger      *logging.Logger
	Solver      *solver.Solver
}

type Runner struct {
	opts Options
}

type solveResult struct {
	Problem  models.ProblemsEntity
	Answer   solver.Answer
	Raw      string
	Duration time.Duration
	Err      error
}

func New(opts Options) *Runner {
	return &Runner{opts: opts}
}

func (r *Runner) Execute(ctx context.Context) error {
	if r.opts.CID == 0 || r.opts.ExamID == 0 {
		return fmt.Errorf("cid 和 exam-id 不能为空")
	}
	if r.opts.Workers <= 0 {
		r.opts.Workers = 20
	}
	if r.opts.Logger == nil {
		r.opts.Logger = logging.New(os.Stdout)
	}
	if r.opts.Solver == nil {
		return fmt.Errorf("solver 未初始化")
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	taskState, err := r.loadState()
	if err != nil {
		return err
	}

	sdk, err := RainClassSDK.NewSDK(RainClassSDK.WithCookiePath(r.opts.CookiePath))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sdk.Close(); closeErr != nil {
			r.opts.Logger.Warn("cookie 持久化失败: %v", closeErr)
		}
	}()

	userInfo, err := sdk.GetUserInfo()
	if err != nil {
		return fmt.Errorf("读取用户信息失败，请先执行 login: %w", err)
	}
	r.opts.Logger.Info("当前用户: %s (%s)", userInfo.Data.Name, userInfo.Data.SchoolNumber)
	r.opts.Logger.Info("状态文件: %s", r.opts.StatePath)
	r.opts.Logger.Info("模型: %s", r.opts.Solver.ModelName())
	r.opts.Logger.Info("Worker 池: %d", r.opts.Workers)

	r.opts.Logger.Step("进入考试环境")
	if err := sdk.StartExam(r.opts.CID, r.opts.ExamID); err != nil {
		return err
	}

	paper, err := sdk.GetExamPaperQuestion(r.opts.ExamID)
	if err != nil {
		return err
	}
	if paper == nil {
		return fmt.Errorf("试卷数据为空")
	}

	pending := taskState.Pending(paper.Data.Problems)
	if len(pending) > 0 {
		taskState.MarkStarted(len(paper.Data.Problems), paper.Data.Title)
		if err := state.Save(r.opts.StatePath, taskState); err != nil {
			return err
		}
	} else {
		taskState.TotalProblems = len(paper.Data.Problems)
		if paper.Data.Title != "" {
			taskState.ExamTitle = paper.Data.Title
		}
	}

	if len(pending) == 0 {
		r.opts.Logger.Success("没有待处理题目")
		return r.finishIfNeeded(sdk, taskState, paper.Data.Problems)
	}

	r.opts.Logger.Step("开始自动答题，剩余 %d / %d", len(pending), len(paper.Data.Problems))

	jobs := make(chan models.ProblemsEntity)
	results := make(chan solveResult)

	var wg sync.WaitGroup
	for i := 0; i < r.opts.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for problem := range jobs {
				started := time.Now()
				answer, raw, err := r.opts.Solver.Solve(ctx, problem)
				results <- solveResult{
					Problem:  problem,
					Answer:   answer,
					Raw:      raw,
					Duration: time.Since(started),
					Err:      err,
				}
				if err == nil {
					r.opts.Logger.Info("worker-%02d 完成 problem=%d index=%d，用时=%s", workerID, problem.ProblemId, problem.Index, time.Since(started).Round(time.Millisecond))
				}
			}
		}(i + 1)
	}

	go func() {
		defer close(jobs)
		for _, problem := range pending {
			select {
			case <-ctx.Done():
				return
			case jobs <- problem:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var failed int
	for result := range results {
		if result.Err != nil {
			failed++
			r.opts.Logger.Error("problem=%d 求解失败: %v", result.Problem.ProblemId, result.Err)
			taskState.MarkFailure(result.Problem.ProblemId, result.Err)
			if err := state.Save(r.opts.StatePath, taskState); err != nil {
				return err
			}
			continue
		}

		submitResp, err := sdk.SubmitAnswer(r.opts.ExamID, models.SubmitAnswerResultsEntity{
			ProblemId: result.Problem.ProblemId,
			Result:    result.Answer.Result,
			Time:      time.Now().UnixMilli(),
		})
		if err != nil {
			failed++
			r.opts.Logger.Error("problem=%d 提交失败: %v", result.Problem.ProblemId, err)
			taskState.MarkFailure(result.Problem.ProblemId, err)
			if saveErr := state.Save(r.opts.StatePath, taskState); saveErr != nil {
				return saveErr
			}
			continue
		}
		if submitResp.Errcode != 0 {
			failed++
			err = fmt.Errorf("errcode=%d errmsg=%s", submitResp.Errcode, submitResp.Errmsg)
			r.opts.Logger.Error("problem=%d 提交失败: %v", result.Problem.ProblemId, err)
			taskState.MarkFailure(result.Problem.ProblemId, err)
			if saveErr := state.Save(r.opts.StatePath, taskState); saveErr != nil {
				return saveErr
			}
			continue
		}

		taskState.MarkAnswered(result.Problem, result.Answer.Result, r.opts.Solver.ModelName(), result.Raw)
		if err := state.Save(r.opts.StatePath, taskState); err != nil {
			return err
		}

		r.opts.Logger.Success(
			"problem=%d index=%d 已提交，答案=%s，进度=%d/%d",
			result.Problem.ProblemId,
			result.Problem.Index,
			strings.Join(result.Answer.Result, ","),
			taskState.AnsweredCount(),
			taskState.TotalProblems,
		)
	}

	if ctx.Err() != nil {
		taskState.MarkInterrupted(ctx.Err())
		if saveErr := state.Save(r.opts.StatePath, taskState); saveErr != nil {
			return saveErr
		}
		return fmt.Errorf("任务被中断，使用 resume 恢复")
	}

	if remaining := len(taskState.Pending(paper.Data.Problems)); remaining > 0 {
		taskState.MarkPartial()
		if err := state.Save(r.opts.StatePath, taskState); err != nil {
			return err
		}
		return fmt.Errorf("仍有 %d 道题未完成，失败 %d 道，可执行 resume 继续", remaining, failed)
	}

	return r.finishIfNeeded(sdk, taskState, paper.Data.Problems)
}

func (r *Runner) loadState() (*state.ExamState, error) {
	st, err := state.Load(r.opts.StatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if r.opts.Resume {
				return nil, fmt.Errorf("状态文件不存在，无法恢复: %s", r.opts.StatePath)
			}
			st = state.New(r.opts.CID, r.opts.ExamID, r.opts.CookiePath)
			if err := state.Save(r.opts.StatePath, st); err != nil {
				return nil, err
			}
			return st, nil
		}
		return nil, err
	}

	if !r.opts.Resume && st.AnsweredCount() > 0 {
		return nil, fmt.Errorf("检测到已有进度，请执行 resume: %s", r.opts.StatePath)
	}
	return st, nil
}

func (r *Runner) finishIfNeeded(sdk *RainClassSDK.SDK, taskState *state.ExamState, problems []models.ProblemsEntity) error {
	if !r.opts.SubmitPaper {
		taskState.MarkReadyToSubmit()
		if err := state.Save(r.opts.StatePath, taskState); err != nil {
			return err
		}
		r.opts.Logger.Success("全部题目已提交完成")
		r.opts.Logger.Warn("未自动交卷；如需交卷，请执行 resume 并追加 --submit-paper")
		return nil
	}

	if taskState.SubmittedPaper {
		taskState.MarkCompleted()
		return state.Save(r.opts.StatePath, taskState)
	}

	r.opts.Logger.Step("开始交卷")
	results, err := taskState.BuildSubmitPaperResults(problems)
	if err != nil {
		return err
	}
	if err := sdk.Close(); err != nil {
		return fmt.Errorf("交卷前保存考试 cookie 失败: %w", err)
	}
	resp, err := submitPaper(r.opts.CookiePath, r.opts.ExamID, results)
	if err != nil {
		return err
	}
	if resp.Errcode != 0 {
		return fmt.Errorf("交卷失败: errcode=%d errmsg=%s", resp.Errcode, resp.Errmsg)
	}

	taskState.MarkCompleted()
	if err := state.Save(r.opts.StatePath, taskState); err != nil {
		return err
	}

	r.opts.Logger.Success("交卷完成")
	return nil
}
