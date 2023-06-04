package routes

import (
	"github.com/gofiber/fiber/v2"
)

func BaseRootHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Hey! It seems you've bumped into Kirara, the node synchronisation service for ktty! You probably shouldn't be here.",
	})
}
