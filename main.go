package main

import (
	"flag"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {
	port := flag.Uint("port", 7552, "server binding port")
	flag.Parse()

	app := fiber.New()

	spreadsheet := app.Group("/spreadsheet")
	spreadsheet.Post("/drs")
	spreadsheet.Post("/dr")
	spreadsheet.Post("/cll")
	spreadsheet.Post("/ba")

	calendar := app.Group("/calendar")
	calendar.Post("/drs")
	calendar.Post("/dr")
	calendar.Post("/cll")
	calendar.Post("/ba")

	app.Listen(":" + strconv.Itoa(int(*port)))
}
