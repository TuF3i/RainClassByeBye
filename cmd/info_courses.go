package cmd

import "github.com/spf13/cobra"

var infoCoursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "获取课程列表",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Course Info")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		data, err := sdk.GetCourseInfo()
		if err != nil {
			return err
		}
		log.Success("课程列表获取成功，共 %d 门", len(data.CourseData.List))
		return writeJSON(cmd, data)
	},
}

func init() {
	infoCmd.AddCommand(infoCoursesCmd)
}
