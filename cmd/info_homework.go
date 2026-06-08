package cmd

import (
	"github.com/spf13/cobra"
)

var (
	infoHomeworkCID    int64
	infoHomeworkLeafID int64
)

var infoHomeworkCmd = &cobra.Command{
	Use:   "homework",
	Short: "获取课程作业列表，或查看单个作业详情",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger()
		log.Banner("RainClass Homework")

		sdk, err := newSDK()
		if err != nil {
			return err
		}
		defer sdk.Close()

		if infoHomeworkLeafID > 0 {
			details, err := sdk.GetHomeWorkDetails(infoHomeworkCID, infoHomeworkLeafID)
			if err != nil {
				return err
			}

			cover, err := sdk.GetHomeWorkCover(infoHomeworkCID, details.Data.ContentInfo.LeafTypeId)
			if err != nil {
				return err
			}

			log.Success("作业详情获取成功")

			rows := [][]string{
				{"CID", formatMaybeString(infoHomeworkCID)},
				{"Leaf ID", formatMaybeString(details.Data.Id)},
				{"Exam ID", formatMaybeString(details.Data.ContentInfo.LeafTypeId)},
				{"标题", details.Data.Name},
				{"发布时间", formatMillisFloat(details.Data.PublishTime)},
				{"截止时间", formatMillisFloat(details.Data.ScoreDeadline)},
				{"是否锁定", formatBool(details.Data.IsLocked)},
				{"是否计分", formatBool(details.Data.IsScore)},
				{"是否已批阅", formatBool(details.Data.IsAssessed)},
				{"题目数", formatMaybeString(cover.Data.ProblemCount)},
				{"总分", formatMaybeString(cover.Data.TotalScore)},
				{"考试开始", formatMillis(cover.Data.StartTime)},
				{"考试截止", formatMillis(cover.Data.Deadline)},
			}

			return writeTable(cmd, []string{"字段", "值"}, rows)
		}

		data, err := sdk.GetHomeWorkInfo(infoHomeworkCID)
		if err != nil {
			return err
		}
		log.Success("作业列表获取成功")

		var rows [][]string
		for _, chapter := range data.Data.CourseChapter {
			for _, leaf := range chapter.SectionLeafList {
				rows = append(rows, []string{
					formatMaybeString(leaf.Id),
					truncateText(chapter.Name, 7),
					truncateText(leaf.Name, 7),
					formatLeafType(leaf.LeafType),
					formatMillis(leaf.StartTime),
					formatMillis(leaf.ScoreDeadline),
				})
			}
		}

		return writeTable(cmd, []string{"Leaf ID", "章节", "标题", "类型", "开始时间", "截止时间"}, rows)
	},
}

func init() {
	infoHomeworkCmd.Flags().Int64Var(&infoHomeworkCID, "cid", 0, "课程 classroom_id")
	infoHomeworkCmd.Flags().Int64Var(&infoHomeworkLeafID, "leaf-id", 0, "作业 leaf_id；传入后查看单个作业详情")
	_ = infoHomeworkCmd.MarkFlagRequired("cid")
	infoCmd.AddCommand(infoHomeworkCmd)
}
