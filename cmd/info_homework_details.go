package cmd

import "github.com/spf13/cobra"

var (
	infoHomeworkDetailsCID    int64
	infoHomeworkDetailsLeafID int64
)

var infoHomeworkDetailsCmd = &cobra.Command{
	Use:   "homework-details",
	Short: "获取单个作业详情",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Homework Details")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		data, err := sdk.GetHomeWorkDetails(infoHomeworkDetailsCID, infoHomeworkDetailsLeafID)
		if err != nil {
			return err
		}
		log.Success("作业详情获取成功")
		return writeJSON(cmd, data)
	},
}

func init() {
	infoHomeworkDetailsCmd.Flags().Int64Var(&infoHomeworkDetailsCID, "cid", 0, "课程 classroom_id")
	infoHomeworkDetailsCmd.Flags().Int64Var(&infoHomeworkDetailsLeafID, "leaf-id", 0, "作业 leaf_id")
	_ = infoHomeworkDetailsCmd.MarkFlagRequired("cid")
	_ = infoHomeworkDetailsCmd.MarkFlagRequired("leaf-id")
	infoCmd.AddCommand(infoHomeworkDetailsCmd)
}
