package models

type DailySchedule struct {
	ClassName      string
	Classroom      string
	Professor      string
	ClassDuration  ClassDuration
	ClassStatusInt *ClassStatus
}

type ClassStatus int

const (
	nadomescanje = iota
	zaposlitev
	odpadlaUra
	vecSkupin
	dogodek
)
