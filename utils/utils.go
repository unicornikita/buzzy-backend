package utils

import (
	"buzzy-backend/models"
	"fmt"
	"log"
)

func GetWeekdayName(weekdayIndex int) string {
	switch weekdayIndex {
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	}
	return ""
}

func SetClassStatus(statusStr string) *models.ClassStatus {
	status, err := models.ParseClassStatus(statusStr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	return &status
}

func PrintDailySchedule(schedule models.DailySchedule) {

	var dailySchedule = schedule.DailySchedule
	fmt.Println(schedule.DayOfTheWeek)
	for i := 0; i < len(dailySchedule); i++ {
		ds := dailySchedule[i]
		var status *models.ClassStatus

		if ds.ClassStatusInt != nil {
			status = ds.ClassStatusInt
		} else {
			status = nil
		}
		fmt.Print(models.PrintSubjectInfo(ds, FormatDurationPrint(ds.ClassDuration), status))
		if ds.SubClasses != nil {
			PrintSubClasses(*ds.SubClasses)
		}
		fmt.Println()
	}
}

func PrintSubClasses(subClasses []models.ClassSubject) {
	var status *models.ClassStatus
	for s := range subClasses {
		currentClass := subClasses[s]

		if currentClass.ClassStatusInt != nil {
			status = currentClass.ClassStatusInt
		} else {
			status = nil
		}
		fmt.Println("subclasses:", currentClass.ClassName, currentClass.Classroom, currentClass.Professor, FormatDurationPrint(currentClass.ClassDuration), status)
	}
}

func FormatDurationPrint(duration models.ClassDuration) string {
	return fmt.Sprintf("start: %02d:%02d, end: %02d:%02d",
		duration.StartTime.Hour(),
		duration.StartTime.Minute(),
		duration.EndTime.Hour(),
		duration.EndTime.Minute(),
	)
}

func PrintWeeklySchedule(schedule models.WeeklySchedule) {
	for dailySchedule := range schedule.WeeklySchedule {
		selectedSchedule := schedule.WeeklySchedule[dailySchedule]
		PrintDailySchedule(selectedSchedule)
		fmt.Println("------")
	}

}
