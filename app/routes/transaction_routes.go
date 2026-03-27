package routes

import (
	"github.com/KaningNoppasin/embedded-system-lab-backend/app/handlers"
	"github.com/gofiber/fiber/v3"
)

func RegisterTransactionRoutes(app *fiber.App, transactionHandler *handlers.TransactionHandler) {
	transactions := app.Group("/transactions")

	transactions.Get("/", transactionHandler.GetAllTransactions)
	transactions.Post("/", transactionHandler.CreateTransaction)
	transactions.Get("/user-rfid-hashed/:userRFIDHashed", transactionHandler.GetTransactionsByUserRFIDHashed)
}
