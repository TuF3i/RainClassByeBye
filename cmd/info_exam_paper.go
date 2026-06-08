package cmd

import "github.com/spf13/cobra"

var (
	infoExamPaperCID    int64
	infoExamPaperExamID int64
)

var infoExamPaperCmd = &cobra.Command{
	Use:   "exam-paper",
	Short: "获取试卷题目，会触发进入考试环境",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Exam Paper")
		log.Warn("该命令会调用 StartExam 并进入考试环境")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		if err := sdk.StartExam(infoExamPaperCID, infoExamPaperExamID); err != nil {
			return err
		}
		data, err := sdk.GetExamPaperQuestion(infoExamPaperExamID)
		if err != nil {
			return err
		}
		log.Success("试卷题目获取成功，共 %d 题", len(data.Data.Problems))
		return writeJSON(cmd, data)
	},
}

func init() {
	infoExamPaperCmd.Flags().Int64Var(&infoExamPaperCID, "cid", 0, "课程 classroom_id")
	infoExamPaperCmd.Flags().Int64Var(&infoExamPaperExamID, "exam-id", 0, "exam_id")
	_ = infoExamPaperCmd.MarkFlagRequired("cid")
	_ = infoExamPaperCmd.MarkFlagRequired("exam-id")
	infoCmd.AddCommand(infoExamPaperCmd)
}
