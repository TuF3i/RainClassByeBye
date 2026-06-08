package cmd

import "github.com/spf13/cobra"

var infoUserCmd = &cobra.Command{
	Use:   "user",
	Short: "获取当前登录用户信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass User Info")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		data, err := sdk.GetUserInfo()
		if err != nil {
			return err
		}
		log.Success("用户信息获取成功")

		rows := [][]string{
			{"姓名", data.Data.Name},
			{"学号", data.Data.SchoolNumber},
			{"学校", data.Data.School},
			{"用户 ID", data.Data.Id},
		}

		return writeTable(cmd, []string{"字段", "值"}, rows)
	},
}

func init() {
	infoCmd.AddCommand(infoUserCmd)
}
