package main

import (
	"flag"
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
		curEventKindID := os.Getenv(eventKindEnv)

		calendar.Get("/"+curEventKind, func(ctx *fiber.Ctx) error {
			t := time.Now().Format(time.RFC3339)

			_, err := srv.Events.List(curEventKindID).ShowDeleted(false).
				SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
			if err != nil {
				log.Printf("Unable to retrieve next ten of the user's events: %v", err)
				return err
			}

			return ctx.SendString("") // TODO: format events.items
		})

		calendar.Post("/"+curEventKind, func(ctx *fiber.Ctx) error {
			return nil
		})
	}

	app.Listen(":" + strconv.Itoa(int(*port)))
}
