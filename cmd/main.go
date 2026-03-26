package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KaningNoppasin/embedded-system-lab-backend/app/database"
	"github.com/KaningNoppasin/embedded-system-lab-backend/app/handlers"
	"github.com/KaningNoppasin/embedded-system-lab-backend/app/repositories"
	"github.com/KaningNoppasin/embedded-system-lab-backend/app/routes"
	"github.com/gofiber/fiber/v3"
)

func main() {
	mongoClient, userCollection, err := database.ConnectMongo()
	if err != nil {
		log.Fatalf("failed to connect mongo: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("failed to disconnect mongo: %v", err)
		}
	}()

	app := fiber.New()
	userRepository, err := repositories.NewUserRepository(userCollection)
	if err != nil {
		log.Fatalf("failed to setup user repository: %v", err)
	}
	userHandler := handlers.NewUserHandler(userRepository)

	// Define a route for the GET method on the root path '/'
	app.Get("/", func(c fiber.Ctx) error {
		// Send a string response to the client
		return c.SendString("Hello, World 👋!")
	})

	routes.RegisterUserRoutes(app, userHandler)

	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Fatalf("fiber server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("failed to shutdown app: %v", err)
	}
}
