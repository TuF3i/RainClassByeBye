package cmd

import "github.com/spf13/cobra"

type appOptions struct {
	CookiePath string
	StateDir   string
}

var opts = appOptions{
	CookiePath: "cache/cookies.json",
	StateDir:   "cache/state",
}

var rootCmd = &cobra.Command{
	Use:           "rainclass-bye-bye",
	Short:         "RainClass 雨课堂自动答题 CLI",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&opts.CookiePath, "cookie-path", opts.CookiePath, "cookie 持久化文件路径")
	rootCmd.PersistentFlags().StringVar(&opts.StateDir, "state-dir", opts.StateDir, "任务状态目录")
}
