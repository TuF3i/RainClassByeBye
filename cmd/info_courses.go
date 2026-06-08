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

		rows := make([][]string, 0, len(data.CourseData.List))
		for _, course := range data.CourseData.List {
			rows = append(rows, []string{
				formatMaybeString(course.ClassroomId),
				truncateText(course.Course.Name, 7),
				truncateText(course.Name, 15),
				truncateText(course.Teacher.Name, 7),
				formatMaybeString(course.StudentsCount),
			})
		}

		return writeTable(cmd, []string{"CID", "课程", "班级", "教师", "人数"}, rows)
	},
}

func init() {
	infoCmd.AddCommand(infoCoursesCmd)
}
