package models

import (
	"errors"
	"fmt"
)

type ClassSubject struct {
	ClassName      string
	Classroom      string
	Professor      string
	ClassDuration  ClassDuration
	ClassStatusInt *ClassStatus
	SubClasses     *[]ClassSubject // for classes with two subjects happening at the same time
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
	return [...]string{"Nadomeščanje", "Zaposlitev", "Odpadla ura", "Več skupin", "Dogodek"}[status]
}

func ParseClassStatus(status string) (ClassStatus, error) {
	if val, ok := classStatusMap[status]; ok {
		return val, nil
	}
	return 0, errors.New("invalid OrderStatus: " + status)
}

func PrintSubjectInfo(classSubject ClassSubject, classSubjectDuration string, statusInt *ClassStatus) string {
	return fmt.Sprintf(
		"Class Name: %s\nClassroom: %s\nProfessor: %s\nDuration: %s\nStatus: %s\n",
		classSubject.ClassName,
		classSubject.Classroom,
		classSubject.Professor,
		classSubjectDuration,
		statusInt,
	)
}
