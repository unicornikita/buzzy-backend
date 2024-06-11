package main

import (
	"fmt"
)

func (classStatus ClassStatus) getClassStatus() string {
	return [...]string{"nadomescanje", "zaposlitev", "odpadlaUra", "vecSkupin", "dogodek"}[classStatus-1]
}

var scheduleTable [9][6]WeeklySchedule = [9][6]WeeklySchedule{}

func main() {
	fmt.Println("Hello, World!")
}
