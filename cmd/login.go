package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "扫码登录雨课堂并持久化 cookie",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Login")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := sdk.Close(); closeErr != nil {
				log.Warn("cookie 持久化失败: %v", closeErr)
			}
		}()

		log.Step("准备登录，终端会输出微信扫码二维码")
		if err := sdk.QRLogin(); err != nil {
			return err
		}

		userInfo, err := sdk.GetUserInfo()
		if err != nil {
			return err
		}

		log.Success("登录成功")
		log.Info("用户: %s", userInfo.Data.Name)
		log.Info("学号: %s", userInfo.Data.SchoolNumber)
		log.Info("学校: %s", userInfo.Data.School)
		log.Info("cookie: %s", opts.CookiePath)
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
