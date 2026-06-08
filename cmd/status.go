package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"RainClassByeBye/internal/state"
)

var (
	statusCID  int64
	statusExam int64
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看任务状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		path := state.DefaultPath(opts.StateDir, statusCID, statusExam)
		taskState, err := state.Load(path)
		if err != nil {
			return err
		}

		log.Banner("RainClass Status")
		log.Info("状态文件: %s", path)
		log.Info("CID / ExamID: %d / %d", taskState.CID, taskState.ExamID)
		log.Info("状态: %s", taskState.Status)
		log.Info("试卷: %s", taskState.ExamTitle)
		log.Info("进度: %d / %d", taskState.AnsweredCount(), taskState.TotalProblems)
		log.Info("失败数: %d", taskState.FailedCount())
		log.Info("更新时间: %s", taskState.UpdatedAt.Format("2006-01-02 15:04:05"))
		if taskState.LastError != "" {
			log.Warn("最后错误: %s", taskState.LastError)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
		return nil
	},
}

func init() {
	statusCmd.Flags().Int64Var(&statusCID, "cid", 0, "课程 classroom_id")
	statusCmd.Flags().Int64Var(&statusExam, "exam-id", 0, "exam_id")
	_ = statusCmd.MarkFlagRequired("cid")
	_ = statusCmd.MarkFlagRequired("exam-id")
	rootCmd.AddCommand(statusCmd)
}
