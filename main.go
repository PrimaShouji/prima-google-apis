package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/calendar/v3"
)

func main() {
	port := flag.Uint("port", 7552, "server binding port")
	flag.Parse()

	client := newGoogleAPI()
	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	app := fiber.New()

	var eventCalendars = map[string]string{
		"drs": "PRIMA_GAPI_CAL_DRS",
		"dr":  "PRIMA_GAPI_CAL_DR",
		"cll": "PRIMA_GAPI_CAL_CLL",
		"ba":  "PRIMA_GAPI_CAL_BA",
	}

	spreadsheet := app.Group("/spreadsheet")
	for eventKind := range eventCalendars {
		curEventKind := eventKind
		spreadsheet.Post("/"+curEventKind, func(ctx *fiber.Ctx) error {
			return nil
		})
	}

	calendar := app.Group("/calendar")
	for eventKind, eventKindEnv := range eventCalendars {
		curEventKind := eventKind

		t := time.Now().Format(time.RFC3339)

		events, err := srv.Events.List(os.Getenv(eventKindEnv)).ShowDeleted(false).
			SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
		}

		fmt.Println("Upcoming events:")

		if len(events.Items) == 0 {
			fmt.Println("No upcoming events found.")
		} else {
			for _, item := range events.Items {
				date := item.Start.DateTime
				if date == "" {
					date = item.Start.Date
				}
				fmt.Printf("%v (%v)\n", item.Summary, date)
			}
		}

		calendar.Post("/"+curEventKind, func(ctx *fiber.Ctx) error {
			return nil
		})
	}

	app.Listen(":" + strconv.Itoa(int(*port)))
}
