package main

import (
	"buzzy-backend/models"
	"buzzy-backend/utils"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func getWeeklySchedule(url string) models.WeeklySchedule { //TODO: add string classURL as parameter

	collector := colly.NewCollector()

	var weeklySchedule = models.WeeklySchedule{}

	collector.OnHTML(".ednevnik-seznam_ur_teden", func(tableOfClasses *colly.HTMLElement) {
		dailySchedule := []models.DailySchedule{}

		var classTimes = getClassTimes(tableOfClasses)
		// start on 1 because the .Weekday() has Sunday on index 0
		for dayOfTheWeekIndex := 1; dayOfTheWeekIndex < 6; dayOfTheWeekIndex++ {
			dailySchedule = append(dailySchedule, getDailySchedule(tableOfClasses, dayOfTheWeekIndex, classTimes))
		}

		weeklySchedule.WeeklySchedule = dailySchedule
		utils.PrintWeeklySchedule(weeklySchedule)
	})

	collector.Visit(url)

	return weeklySchedule

}

func getClassTimes(element *colly.HTMLElement) []models.ClassDuration {
	var classDurations []models.ClassDuration

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
		}
	})
	return classDurations
}

func getDailySchedule(element *colly.HTMLElement, dayOfTheWeekIndex int, durations []models.ClassDuration) models.DailySchedule {
	var dailySchedule models.DailySchedule
	var schedule []models.ClassSubject

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
				var primarySubject models.ClassSubject

				if s == dayOfTheWeekIndex {

					if subject.Children().Length() == 0 {
						emptySubject := models.ClassSubject{ClassName: "", Classroom: "", Professor: "", ClassDuration: durations[i-1], ClassStatusInt: nil}
						schedule = append(schedule, emptySubject)
						return
					}

					var isFirst bool = true
					subject.Children().Each(func(sC int, subClass *goquery.Selection) {
						subjectName = strings.TrimSpace(subClass.Find(".text14.bold").Text())
						subjectProfessorAndRoom = subClass.Find(".text11").Text()

						title, exists := subject.Find("img[title]").Attr("title")

						var status *models.ClassStatus = nil
						if exists {
							status = utils.SetClassStatus(title)
						}
						if subjectProfessorAndRoom != "" {
							professor = strings.TrimSpace(strings.Split(subjectProfessorAndRoom, ", ")[0])
							room = strings.TrimSpace(strings.Split(subjectProfessorAndRoom, ", ")[1])
							selectedSubject := models.ClassSubject{ClassName: subjectName, Classroom: room, Professor: professor, ClassDuration: durations[i-1], ClassStatusInt: status}

							if isFirst {
								primarySubject = selectedSubject
								primarySubject.SubClasses = &[]models.ClassSubject{}
								isFirst = false
							} else {
								appended := append(*primarySubject.SubClasses, selectedSubject)
								primarySubject.SubClasses = &appended
							}

						}
					})

					schedule = append(schedule, primarySubject)
				}

			}
		})
		dailySchedule.DailySchedule = schedule
		dailySchedule.DayOfTheWeek = utils.GetWeekdayName(dayOfTheWeekIndex)
	})
	return dailySchedule

}

func main() {
	url := "https://www.easistent.com/urniki/5738623c4f3588f82583378c44ceb026102d6bae/razredi/523573"
	getWeeklySchedule(url)
	// subtract 1 because today on index 0 would be Sunday

	//TODO: separate two-subject classes

}
