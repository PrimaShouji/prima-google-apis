package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/calendar/v3"
)

type miniEvent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"startTime"`
}

func main() {
	port := flag.Uint("port", 7552, "server binding port")
	flag.Parse()

	client := newGoogleAPI()
	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	app := fiber.New()

	var eventCalendarEnvVars = map[string]string{
		"drs": "PRIMA_GAPI_CAL_DRS",
		"dr":  "PRIMA_GAPI_CAL_DR",
		"cll": "PRIMA_GAPI_CAL_CLL",
		"ba":  "PRIMA_GAPI_CAL_BA",
	}

	spreadsheet := app.Group("/spreadsheet")
	for eventKind := range eventCalendarEnvVars {
		curEventKind := eventKind
		spreadsheet.Post("/"+curEventKind, func(ctx *fiber.Ctx) error {
			return nil
		})
	}

	calendar := app.Group("/calendar")
	for eventKind, eventKindEnv := range eventCalendarEnvVars {
		curEventKind := eventKind
		curEventKindID := os.Getenv(eventKindEnv)

		calendar.Get("/"+curEventKind, func(ctx *fiber.Ctx) error {
			// Fetch events
			t := time.Now().Format(time.RFC3339)
			events, err := srv.Events.List(curEventKindID).ShowDeleted(false).
				SingleEvents(true).TimeMin(t).OrderBy("startTime").Do()
			if err != nil {
				log.Printf("Unable to retrieve next ten of the user's events: %v", err)
				return err
			}

			// Map full event objects to trimmed-down versions
			miniEvents := make([]*miniEvent, 0)
			for _, event := range events.Items {
				miniEvents = append(miniEvents, &miniEvent{
					Title:       event.Summary,
					Description: event.Description,
					StartTime:   event.Start.DateTime,
				})
			}

			// Serialize
			res, err := json.Marshal(miniEvents)
			if err != nil {
				log.Println("Failed to marshal event list.")
				return err
			}

			return ctx.SendString(string(res))
		})

		calendar.Post("/"+curEventKind, func(ctx *fiber.Ctx) error {
			return nil
		})
	}

	app.Listen(":" + strconv.Itoa(int(*port)))
}
