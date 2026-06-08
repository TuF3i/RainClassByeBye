package cmd

import "github.com/spf13/cobra"

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "获取雨课堂账号、课程、作业和考试信息",
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
