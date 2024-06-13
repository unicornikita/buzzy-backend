package main

import (
	"buzzy-backend/models"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func getWeeklySchedule(url string) models.WeeklySchedule { //add string classURL as parameter

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

	var weeklySchedule = models.WeeklySchedule{}

	collector.OnHTML(".ednevnik-seznam_ur_teden", func(tableOfClasses *colly.HTMLElement) {
		dailySchedule := []models.DailySchedule{}

		var classTimes = getClassTimes(collector, tableOfClasses)
		// start on 1 because the .Weekday() has Sunday on index 0
		for dayOfTheWeekIndex := 1; dayOfTheWeekIndex < 6; dayOfTheWeekIndex++ {
			dailySchedule = append(dailySchedule, getDailySchedule(collector, tableOfClasses, dayOfTheWeekIndex, classTimes))
		}

		weeklySchedule.WeeklySchedule = dailySchedule
	})

	collector.Visit(url)

	return weeklySchedule

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
			//fmt.Println(classDuration.StartTime.Hour(), classDuration.StartTime.Minute())
		}
	})
	return classDurations
}

func getDailySchedule(collector *colly.Collector, element *colly.HTMLElement, dayOfTheWeekIndex int, durations []models.ClassDuration) models.DailySchedule {
	var dailySchedule models.DailySchedule
	var schedule []models.ClassSubject

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
	})

	collector.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from", r.Request.URL)
	})

	collector.OnError(func(r *colly.Response, e error) {
		fmt.Println("Error", e)
	})
	var days []string
	var datesOfDay []string
	var dayText string
	var dateOfDay string
	element.DOM.Find("tbody > tr").First().Children().Each(func(i int, daysRow *goquery.Selection) {
		if !strings.Contains(daysRow.Find("div:nth-child(1)").Text(), "Ura") {
			dayText = daysRow.Find("div:nth-child(1)").Text()
			dateOfDay = daysRow.Find("div:nth-child(2)").Text()
			days = append(days, dayText)
			datesOfDay = append(datesOfDay, dateOfDay)
		}
	})

	element.ForEach(".ednevnik-seznam_ur_teden > tbody > tr", func(i int, subjectRow *colly.HTMLElement) {
		if i == 0 {
			return
		}

		var subjectName string
		var subjectProfessorAndRoom string
		var professor string
		var room string
		subjectRow.DOM.Children().Each(func(s int, subject *goquery.Selection) {
			if s != 0 {
				subjectName = strings.TrimSpace(subject.Find(".ednevnik-seznam_ur_teden-urnik").Children().Find(".text14.bold").Text())
				subjectProfessorAndRoom = subject.Find(".ednevnik-seznam_ur_teden-urnik").Find(".text11").Text()
				title, exists := subject.Find("img[title]").Attr("title")
				var status *models.ClassStatus
				fmt.Println(title)
				if exists {
					status = setClassStatus(title)
				} else {
					status = nil
				}
				if s == dayOfTheWeekIndex {
					if subjectProfessorAndRoom != "" {
						professor = strings.TrimSpace(strings.Split(subjectProfessorAndRoom, ", ")[0])
						room = strings.TrimSpace(strings.Split(subjectProfessorAndRoom, ", ")[1])

						subject := models.ClassSubject{ClassName: subjectName, Classroom: room, Professor: professor, ClassDuration: durations[i-1], ClassStatusInt: status}
						schedule = append(schedule, subject)

					} else {
						emptySubject := models.ClassSubject{ClassName: "", Classroom: "", Professor: "", ClassDuration: durations[i-1], ClassStatusInt: nil}
						schedule = append(schedule, emptySubject)
					}
				}
			}
		})
		dailySchedule.DailySchedule = schedule
		dailySchedule.DayOfTheWeek = getWeekdayName(dayOfTheWeekIndex)
	})
	return dailySchedule

}

func getWeekdayName(weekdayIndex int) string {
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

func setClassStatus(statusStr string) *models.ClassStatus {
	status, err := models.ParseClassStatus(statusStr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	return &status
}

func printDailySchedule(schedule models.DailySchedule) {

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
		fmt.Println(ds.ClassName, ds.Classroom, ds.Professor, formatDurationPrint(ds.ClassDuration), status)
	}
}

func formatDurationPrint(duration models.ClassDuration) string {
	return fmt.Sprintf("start: %02d:%02d, end: %02d:%02d",
		duration.StartTime.Hour(),
		duration.StartTime.Minute(),
		duration.EndTime.Hour(),
		duration.EndTime.Minute(),
	)
}
func main() {
	url := "https://www.easistent.com/urniki/5738623c4f3588f82583378c44ceb026102d6bae/razredi/523573"
	weeklySchedule := getWeeklySchedule(url)
	// subtract 1 because today on index 0 would be Sunday
	today := int(time.Now().Weekday()) - 1
	fmt.Println(today, time.Now().Weekday().String())
	printDailySchedule(weeklySchedule.WeeklySchedule[today-1])
}
