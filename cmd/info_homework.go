package cmd

import "github.com/spf13/cobra"

var infoHomeworkCID int64

var infoHomeworkCmd = &cobra.Command{
	Use:   "homework",
	Short: "获取课程作业列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Homework Info")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		data, err := sdk.GetHomeWorkInfo(infoHomeworkCID)
		if err != nil {
			return err
		}
		log.Success("作业列表获取成功")
		return writeJSON(cmd, data)
	},
}

func init() {
	infoHomeworkCmd.Flags().Int64Var(&infoHomeworkCID, "cid", 0, "课程 classroom_id")
	_ = infoHomeworkCmd.MarkFlagRequired("cid")
	infoCmd.AddCommand(infoHomeworkCmd)
}
