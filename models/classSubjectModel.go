package models

import (
	"errors"
)

type ClassSubject struct {
	ClassName      string
	Classroom      string
	Professor      string
	ClassDuration  ClassDuration
	ClassStatusInt *ClassStatus
}

type ClassStatus int

const (
	Nadomescanje = iota
	Zaposlitev
	OdpadlaUra
	VecSkupin
	Dogodek
)

var classStatusMap = map[string]ClassStatus{
	"Nadomeščanje": Nadomescanje,
	"Zaposlitev":   Zaposlitev,
	"Odpadla ura":  OdpadlaUra,
	"Več skupin":   VecSkupin,
	"Dogodek":      Dogodek,
}

func (status ClassStatus) String() string {
	return [...]string{"Nadomeščanje", "Zaposlitev", "Odpadla Ura", "Več skupin", "Dogodek"}[status]
}

func ParseClassStatus(status string) (ClassStatus, error) {
	if val, ok := classStatusMap[status]; ok {
		return val, nil
	}
	return 0, errors.New("invalid OrderStatus: " + status)
}
