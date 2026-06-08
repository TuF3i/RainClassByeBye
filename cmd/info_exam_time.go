package cmd

import "github.com/spf13/cobra"

var (
	infoExamTimeCID    int64
	infoExamTimeExamID int64
)

var infoExamTimeCmd = &cobra.Command{
	Use:   "exam-time",
	Short: "获取考试剩余时间，会触发进入考试环境",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Exam Time")
		log.Warn("该命令会调用 StartExam 并进入考试环境")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		if err := sdk.StartExam(infoExamTimeCID, infoExamTimeExamID); err != nil {
			return err
		}
		data, err := sdk.RefreshTimeRemaining(infoExamTimeExamID)
		if err != nil {
			return err
		}
		log.Success("剩余时间获取成功")
		return writeJSON(cmd, data)
	},
}

func init() {
	infoExamTimeCmd.Flags().Int64Var(&infoExamTimeCID, "cid", 0, "课程 classroom_id")
	infoExamTimeCmd.Flags().Int64Var(&infoExamTimeExamID, "exam-id", 0, "exam_id")
	_ = infoExamTimeCmd.MarkFlagRequired("cid")
	_ = infoExamTimeCmd.MarkFlagRequired("exam-id")
	infoCmd.AddCommand(infoExamTimeCmd)
}
