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
		return writeJSON(cmd, data)
	},
}

func init() {
	infoCmd.AddCommand(infoUserCmd)
}
