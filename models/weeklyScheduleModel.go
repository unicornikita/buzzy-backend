package models

type WeeklySchedule struct {
	WeeklySchedule []DailySchedule
}

func (weeklySchedule WeeklySchedule) Error() string {
	return "unable to get the schedule"
}
