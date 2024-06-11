package models

type DailySchedule struct {
	ClassName      string
	Classroom      string
	Professor      string
	ClassDuration  ClassDuration
	ClassStatusInt ClassStatus
}

type ClassStatus int

const (
	nadomescanje = iota
	zaposlitev
	odpadlaUra
	vecSkupin
	dogodek
)

func (classStatus ClassStatus) getClassStatus() string {
	return [...]string{"nadomescanje", "zaposlitev", "odpadlaUra", "vecSkupin", "dogodek"}[classStatus-1]
}
