package main

import (
	"log"

	"github.com/kttyoss/kirara/pkg/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", routes.BaseRootHandler)

	log.Fatal(app.Listen(":8090"))
}
