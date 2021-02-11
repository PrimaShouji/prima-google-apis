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
	ID          string `json:"id"`
	StartTime   string `json:"startTime"`
}

type eventDeleteRequest struct {
	ID string `json:"id"`
}

type eventCreateResponse struct {
	EventLink string `json:"eventLink"`
}

type genericResponse struct {
	Message string `json:"message"`
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

	spr := app.Group("/spreadsheet")
	for eventKind := range eventCalendarEnvVars {
		curEventKind := eventKind
		spr.Post("/"+curEventKind, func(ctx *fiber.Ctx) error {
			return nil
		})
	}

	cal := app.Group("/calendar")
	for eventKind, eventKindEnv := range eventCalendarEnvVars {
		curEventKind := eventKind
		curEventKindID := os.Getenv(eventKindEnv)

		cal.Get("/"+curEventKind, func(ctx *fiber.Ctx) error {
			// Fetch events
			t := time.Now().Format(time.RFC3339)
			events, err := srv.Events.List(curEventKindID).ShowDeleted(false).
				SingleEvents(true).TimeMin(t).OrderBy("startTime").Do()
			if err != nil {
				log.Printf("Unable to retrieve next ten of the user's events: %v\n", err)
				return err
			}

			// Map full event objects to trimmed-down versions
			miniEvents := make([]*miniEvent, 0)
			for _, event := range events.Items {
				miniEvents = append(miniEvents, &miniEvent{
					Title:       event.Summary,
					Description: event.Description,
					ID:          event.Id,
					StartTime:   event.Start.DateTime,
				})
			}

			// Serialize
			res, err := json.Marshal(miniEvents)
			if err != nil {
				log.Printf("Failed to marshal event list. %v\n", err)
				return err
			}

			ctx.Type("json")
			return ctx.SendString(string(res))
		})

		cal.Post("/"+curEventKind, func(ctx *fiber.Ctx) error {
			// Read request
			newEventReq := &miniEvent{}
			err := json.Unmarshal(ctx.Body(), newEventReq)
			if err != nil {
				log.Printf("Unmarshaling client request failed. %v\n", err)
				return err
			}

			log.Printf("Posting event of type %s: %v\n", curEventKind, newEventReq)

			// Expand to Calendar event
			startTime, err := time.Parse(time.RFC3339, newEventReq.StartTime)
			if err != nil {
				log.Printf("Parsing event start time failed. %v\n", err)
				return err
			}
			endTime := startTime.Add(time.Hour * 2)

			newEvent := &calendar.Event{
				Summary:     newEventReq.Title,
				Description: newEventReq.Description,
				Start: &calendar.EventDateTime{
					DateTime: newEventReq.StartTime,
					TimeZone: "America/Los_Angeles",
				},
				End: &calendar.EventDateTime{
					DateTime: endTime.Format(time.RFC3339),
					TimeZone: "America/Los_Angeles",
				},
			}

			// Publish to calendar
			newEvent, err = srv.Events.Insert(curEventKindID, newEvent).Do()
			if err != nil {
				log.Printf("Unable to create event. %v\n", err)
			}

			// Serialize response
			creationResponse := &eventCreateResponse{EventLink: newEvent.HtmlLink}
			res, err := json.Marshal(creationResponse)
			if err != nil {
				log.Printf("Failed to marshal event creation response. %v\n", err)
				return err
			}

			ctx.Type("json")
			log.Printf("Created event: %s\n", newEvent.HtmlLink)
			return ctx.SendString(string(res))
		})

		cal.Post("/"+curEventKind+"/delete", func(ctx *fiber.Ctx) error {
			// Read request
			deleteEventReq := &eventDeleteRequest{}
			err := json.Unmarshal(ctx.Body(), deleteEventReq)
			if err != nil {
				log.Printf("Unmarshaling client request failed. %v\n", err)
				return err
			}

			// Execute request
			err = srv.Events.Delete(curEventKindID, deleteEventReq.ID).Do()
			if err != nil {
				log.Printf("Event deletion failed. %v\n", err)
			}

			log.Printf("Deleted event %s\n", deleteEventReq.ID)

			genericRes := &genericResponse{Message: "success"}

			// Serialize response
			res, err := json.Marshal(genericRes)
			if err != nil {
				log.Printf("Failed to marshal event list. %v\n", err)
				return err
			}

			ctx.Type("json")
			return ctx.SendString(string(res))
		})
	}

	app.Listen(":" + strconv.Itoa(int(*port)))
}
