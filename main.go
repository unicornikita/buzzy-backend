package main

import (
	"buzzy-backend/models"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

//var scheduleTable [9][6]models.WeeklySchedule = [9][6]models.WeeklySchedule{}

func getWeeklySchedule() { //add string classURL as parameter
	url := "https://www.easistent.com/urniki/5738623c4f3588f82583378c44ceb026102d6bae/razredi/523573"
	collector := colly.NewCollector()

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
	})

	collector.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	collector.OnError(func(r *colly.Response, e error) {
		fmt.Println("Error", e)
	})

	collector.OnHTML(".ednevnik-seznam_ur_teden", func(tableOfClasses *colly.HTMLElement) {
		//weeklySchedule := []models.WeeklySchedule{}
		//dailySchedule := []models.DailySchedule{}

		getClassTimes(collector, tableOfClasses)
	})

	collector.Visit(url)

}

func getClassTimes(collector *colly.Collector, element *colly.HTMLElement) []models.ClassDuration {
	var classDurations []models.ClassDuration
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
	})

	collector.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	collector.OnError(func(r *colly.Response, e error) {
		fmt.Println("Error", e)
	})

	classDuration := models.ClassDuration{}
	var durationText []string
	element.ForEach(".ednevnik-seznam_ur_teden-td.ednevnik-seznam_ur_teden-ura", func(i int, classTime *colly.HTMLElement) {
		if !strings.Contains(classTime.Text, "Čas pred poukom") && !strings.Contains(classTime.Text, "Čas po pouku") {
			durationText = strings.Split(classTime.ChildText(".text10.gray"), " - ")
			var startTime = durationText[0]
			var endTime = durationText[1]
			parsedStartTime, _ := time.Parse("15:04", startTime)
			parsedEndTime, _ := time.Parse("15:04", endTime)
			classDuration.StartTime = parsedStartTime
			classDuration.EndTime = parsedEndTime
			classDurations = append(classDurations, classDuration)
			fmt.Println(classDuration.StartTime.Hour(), classDuration.StartTime.Minute())
		}
	})
	return classDurations
}

func main() {
	getWeeklySchedule()
}
