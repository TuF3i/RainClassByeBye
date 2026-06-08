package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"RainClassByeBye/internal/runner"
)

var (
	resumeCfg  = defaultAIFlags()
	resumeCID  int64
	resumeExam int64
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "恢复中断的自动答题任务",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Resume")

		runOpts, err := buildRunnerOptions(log, resumeCfg, resumeCID, resumeExam, true)
		if err != nil {
			return err
		}
		jobRunner := runner.New(runOpts)
		return jobRunner.Execute(context.Background())
	},
}

func init() {
	resumeCmd.Flags().Int64Var(&resumeCID, "cid", 0, "课程 classroom_id")
	resumeCmd.Flags().Int64Var(&resumeExam, "exam-id", 0, "exam_id")
	_ = resumeCmd.MarkFlagRequired("cid")
	_ = resumeCmd.MarkFlagRequired("exam-id")
	bindAIFlags(resumeCmd, &resumeCfg)
	rootCmd.AddCommand(resumeCmd)
}
