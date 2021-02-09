package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/calendar/v3"
)

func main() {
	port := flag.Uint("port", 7552, "server binding port")
	flag.Parse()

	client := newGoogleAPI()

	_, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	app := fiber.New()

	var eventKind = []string{"drs", "dr", "cll", "ba"}

	spreadsheet := app.Group("/spreadsheet")
	for _, ek := range eventKind {
		spreadsheet.Post("/" + ek)
	}

	calendar := app.Group("/calendar")
	for _, ek := range eventKind {
		calendar.Post("/" + ek)
	}

	app.Listen(":" + strconv.Itoa(int(*port)))
}
