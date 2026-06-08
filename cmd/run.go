package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"RainClassByeBye/internal/runner"
)

var (
	runCfg  = defaultAIFlags()
	runCID  int64
	runExam int64
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "开始新的自动答题任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Run")

		runOpts, err := buildRunnerOptions(log, runCfg, runCID, runExam, false)
		if err != nil {
			return err
		}
		jobRunner := runner.New(runOpts)
		return jobRunner.Execute(context.Background())
	},
}

func init() {
	runCmd.Flags().Int64Var(&runCID, "cid", 0, "课程 classroom_id")
	runCmd.Flags().Int64Var(&runExam, "exam-id", 0, "exam_id")
	_ = runCmd.MarkFlagRequired("cid")
	_ = runCmd.MarkFlagRequired("exam-id")
	bindAIFlags(runCmd, &runCfg)
	rootCmd.AddCommand(runCmd)
}
