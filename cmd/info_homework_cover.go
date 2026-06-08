package cmd

import "github.com/spf13/cobra"

var (
	infoHomeworkCoverCID    int64
	infoHomeworkCoverExamID int64
)

var infoHomeworkCoverCmd = &cobra.Command{
	Use:   "homework-cover",
	Short: "获取作业或考试封面信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Homework Cover")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		data, err := sdk.GetHomeWorkCover(infoHomeworkCoverCID, infoHomeworkCoverExamID)
		if err != nil {
			return err
		}
		log.Success("封面信息获取成功")
		return writeJSON(cmd, data)
	},
}

func init() {
	infoHomeworkCoverCmd.Flags().Int64Var(&infoHomeworkCoverCID, "cid", 0, "课程 classroom_id")
	infoHomeworkCoverCmd.Flags().Int64Var(&infoHomeworkCoverExamID, "exam-id", 0, "exam_id")
	_ = infoHomeworkCoverCmd.MarkFlagRequired("cid")
	_ = infoHomeworkCoverCmd.MarkFlagRequired("exam-id")
	infoCmd.AddCommand(infoHomeworkCoverCmd)
}
