package main

import (
	"log"
	"os"
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/fatih/color"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if os.Getenv("WHITELIST_API") == "" && strings.ToLower(os.Getenv("DEV_MODE")) != "true" {
		color.Red("⚠ WHITELIST_API is not set. Please set it to the URL of the whitelist API.")
		os.Exit(1)
	}
	if strings.ToLower(os.Getenv("DEV_MODE")) == "true" {
		color.Yellow("⚠ DEV_MODE is set to true. The whitelist has been disabled. This is not recommended in production and can lead to data leaks.")
	}
	app := fiber.New(fiber.Config{
		Prefork:      true,
		ServerHeader: "ktty kirara",
		AppName:      "kirara node synchronizerq",
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	app.Static("/", "./www/index.html")

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	go runHub()

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer func() {
			unregister <- c
			c.Close()
		}()

		register <- c

		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}
			log.Printf("from %s, message: %s", c.RemoteAddr().String(), message)
			if messageType == websocket.TextMessage {
				broadcast <- &Kmessage{connection: c, message: message}
			} else {
				log.Printf("unhandled message type: %v", messageType)
			}
		}
	}))

	log.Fatal(app.Listen("127.0.0.1:8090"))
}
