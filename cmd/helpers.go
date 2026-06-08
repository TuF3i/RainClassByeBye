package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	RainClassSDK "github.com/Auto-CQUPT-Plan/RainClassSDK"
	"github.com/spf13/cobra"

	"RainClassByeBye/internal/logging"
	"RainClassByeBye/internal/runner"
	"RainClassByeBye/internal/solver"
	"RainClassByeBye/internal/state"
)

type aiFlags struct {
	Model               string
	BaseURL             string
	APIKey              string
	APIKeyEnv           string
	RequestTimeout      time.Duration
	Temperature         float32
	MaxCompletionTokens int
	Workers             int
	SubmitPaper         bool
}

func defaultAIFlags() aiFlags {
	return aiFlags{
		Model:               "qwen3.6-max",
		BaseURL:             "https://dashscope.aliyuncs.com/compatible-mode/v1",
		APIKeyEnv:           "DASHSCOPE_API_KEY",
		RequestTimeout:      2 * time.Minute,
		Temperature:         0.1,
		MaxCompletionTokens: 2048,
		Workers:             20,
	}
}

func bindAIFlags(cmd *cobra.Command, cfg *aiFlags) {
	cmd.Flags().StringVar(&cfg.Model, "model", cfg.Model, "LLM 模型名")
	cmd.Flags().StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "OpenAI-compatible BaseURL")
	cmd.Flags().StringVar(&cfg.APIKey, "api-key", cfg.APIKey, "LLM API Key，优先级高于环境变量")
	cmd.Flags().StringVar(&cfg.APIKeyEnv, "api-key-env", cfg.APIKeyEnv, "读取 API Key 的环境变量")
	cmd.Flags().DurationVar(&cfg.RequestTimeout, "request-timeout", cfg.RequestTimeout, "单题模型请求超时")
	cmd.Flags().Float32Var(&cfg.Temperature, "temperature", cfg.Temperature, "模型温度")
	cmd.Flags().IntVar(&cfg.MaxCompletionTokens, "max-completion-tokens", cfg.MaxCompletionTokens, "模型最大输出 token")
	cmd.Flags().IntVar(&cfg.Workers, "workers", cfg.Workers, "goroutine worker 数量")
	cmd.Flags().BoolVar(&cfg.SubmitPaper, "submit-paper", cfg.SubmitPaper, "全部题目提交完成后自动交卷")
}

func newLogger() *logging.Logger {
	return logging.New(os.Stdout)
}

func ensureParentDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func ensureStateDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func newSDK() (*RainClassSDK.SDK, error) {
	if err := ensureParentDir(opts.CookiePath); err != nil {
		return nil, err
	}
	return RainClassSDK.NewSDK(RainClassSDK.WithCookiePath(opts.CookiePath))
}

func resolveAPIKey(cfg aiFlags) (string, error) {
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}
	if cfg.APIKeyEnv == "" {
		return "", fmt.Errorf("api key 未提供，且 api-key-env 为空")
	}
	val := os.Getenv(cfg.APIKeyEnv)
	if val == "" {
		return "", fmt.Errorf("环境变量 %s 未设置", cfg.APIKeyEnv)
	}
	return val, nil
}

func buildRunnerOptions(log *logging.Logger, cfg aiFlags, cid, examID int64, resume bool) (runner.Options, error) {
	if err := ensureStateDir(opts.StateDir); err != nil {
		return runner.Options{}, err
	}
	apiKey, err := resolveAPIKey(cfg)
	if err != nil {
		return runner.Options{}, err
	}
	aiSolver, err := solver.New(solver.Config{
		APIKey:              apiKey,
		BaseURL:             cfg.BaseURL,
		Model:               cfg.Model,
		Timeout:             cfg.RequestTimeout,
		Temperature:         cfg.Temperature,
		MaxCompletionTokens: cfg.MaxCompletionTokens,
		Logger:              log,
	})
	if err != nil {
		return runner.Options{}, err
	}

	return runner.Options{
		CID:         cid,
		ExamID:      examID,
		CookiePath:  opts.CookiePath,
		StatePath:   state.DefaultPath(opts.StateDir, cid, examID),
		Workers:     cfg.Workers,
		SubmitPaper: cfg.SubmitPaper,
		Resume:      resume,
		Logger:      log,
		Solver:      aiSolver,
	}, nil
}

func writeJSON(cmd *cobra.Command, v any) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	_, err := cmd.OutOrStdout().Write(buf.Bytes())
	return err
}
