package routes

import (
	"github.com/KaningNoppasin/embedded-system-lab-backend/app/handlers"
	"github.com/gofiber/fiber/v3"
)

func RegisterUserRoutes(app *fiber.App, userHandler *handlers.UserHandler) {
	users := app.Group("/users")

	users.Get("/", userHandler.GetAllUsers)
	users.Post("/", userHandler.CreateUser)
	users.Get("/:rfid/amount", userHandler.GetAmountByRFID)
	users.Patch("/:rfid/amount", userHandler.UpdateAmountByRFID)
}
