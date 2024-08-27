package main

import (
	"buzzy-backend/models"
	"buzzy-backend/utils"
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/jasonlvhit/gocron"
	"google.golang.org/api/option"
)

func getWeeklySchedule(url string) models.WeeklySchedule {

	collector := colly.NewCollector()

	var weeklySchedule = models.WeeklySchedule{}

	collector.OnHTML(".ednevnik-seznam_ur_teden", func(tableOfClasses *colly.HTMLElement) {
		dailySchedules := []models.DailySchedule{}

		var classTimes = getClassTimes(tableOfClasses)
		// start on 1 because the .Weekday() has Sunday on index 0
		for dayOfTheWeekIndex := 1; dayOfTheWeekIndex < 6; dayOfTheWeekIndex++ {
			dailySchedules = append(dailySchedules, getDailySchedule(tableOfClasses, dayOfTheWeekIndex, classTimes))
		}

		weeklySchedule.WeeklySchedule = dailySchedules
		//utils.PrintWeeklySchedule(weeklySchedule)
	})

	collector.Visit(url)
	collector.Wait()

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
			if len(strings.Split(durationText[0], ":")[0]) == 1 {
				startTime = "0" + durationText[0]
			}
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

					subjectChildren := subject.Children()

					if subjectChildren.Length() == 0 {
						emptySubject := models.ClassSubject{ClassName: "", Classroom: "", Professor: "", ClassDuration: durations[i-1], ClassStatusInt: nil}
						schedule = append(schedule, emptySubject)
						return
					}

					var isFirst bool = true
					var hasBeenProcessed bool = false

					subjectChildren.Each(func(sC int, subClass *goquery.Selection) {
						subjectName = strings.TrimSpace(subClass.Find(".text14.bold").Text())
						subjectProfessorAndRoom = subClass.Find(".text11").Text()

						title, exists := subject.Find("img[title]").Attr("title")

						var status *models.ClassStatus = nil
						if exists {
							status = utils.SetClassStatus(title)
							if *status == models.Pocitnice || *status == models.Dogodek {
								schedule = append(schedule, models.ClassSubject{ClassName: subjectName, Classroom: "", Professor: "", ClassDuration: durations[i-1], ClassStatusInt: status})
								hasBeenProcessed = true
								return
							}
						}
						if subjectProfessorAndRoom != "" {
							professor = strings.TrimSpace(strings.Split(subjectProfessorAndRoom, ", ")[0])
							room = strings.TrimSpace(strings.Split(subjectProfessorAndRoom, ", ")[1])
							room = strings.Join(strings.Fields(room), " ")
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

					if hasBeenProcessed {
						return
					}

					schedule = append(schedule, primarySubject)
				}

			}
		})
		dailySchedule.DailySchedule = schedule
		dailySchedule.DayOfTheWeek = utils.GetWeekdayName(dayOfTheWeekIndex)
	})
	return dailySchedule

}

func urlDecode(classURL string) string {
	query, err := url.QueryUnescape(classURL)
	if err != nil {
		log.Fatal(err)
	}
	return query
}

var firebaseContext context.Context
var client *messaging.Client

func initFireBase() {
	firebaseContext = context.Background()
	opt := option.WithCredentialsFile("firebase_sdk.json")
	config := &firebase.Config{ProjectID: "buzzy-diploma"}
	app, err := firebase.NewApp(firebaseContext, config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	client, _ = app.Messaging(context.Background())
}

func sendNotification(nextClass models.ClassSubject, client *messaging.Client, classURL string) {
	classUrlSplit := strings.Split(classURL, "/")
	classCode := classUrlSplit[len(classUrlSplit)-1]
	topic := "nextClass" + classCode
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: nextClass.ClassName + ", " + nextClass.Classroom,
			Body:  nextClass.Professor,
		},
		Topic: topic,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
	}
	// Send a message to the devices subscribed to the provided topic.
	response, err := client.Send(firebaseContext, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
}

func pushNotificationTimer(dailySchedule models.DailySchedule, classURL string) {
	classes := dailySchedule.DailySchedule
	pushTimeOffset := time.Minute * 10
	today := time.Now().In(loc)
	for _, class := range classes {
		classTime := time.Date(
			today.Year(),
			today.Month(),
			today.Day(),
			class.ClassDuration.StartTime.Hour(),
			class.ClassDuration.StartTime.Minute(),
			int(0),
			int(0),
			loc,
		)
		timeUntilNotification := time.Until(classTime) - pushTimeOffset
		if timeUntilNotification > 0 {
			time.AfterFunc(timeUntilNotification, func() { sendNotification(class, client, classURL) })
		}
	}
}

var loc *time.Location

func main() {
	initFireBase()
	app := fiber.New()

	classURLs := []string{}
	var schedule models.WeeklySchedule

	loc, _ = time.LoadLocation("Europe/Ljubljana")

	app.Get("/schedule/:value", func(c *fiber.Ctx) error {
		classURL := urlDecode(c.Params("value"))
		classURLs = append(classURLs, classURL)
		schedule := getWeeklySchedule(classURL)
		return c.JSON(schedule)
	})

	gocron.Every(1).Day().At("05:00").Do(func() {
		if time.Now().Weekday() == time.Sunday || time.Now().Weekday() == time.Saturday {
			return
		}

		if len(classURLs) != 0 {
			for _, classURL := range classURLs {
				go func(url string) {
					schedule = getWeeklySchedule(url)
					if len(schedule.WeeklySchedule) != 0 {
						pushNotificationTimer(schedule.WeeklySchedule[time.Now().Weekday()-1], url)
					}
				}(classURL)
			}
		}
	})

	gocron.Start()
	app.Listen(":3000")
}
