package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
	"zadanie-6105/internal"
	"zadanie-6105/internal/storage/postgres"
)

func Run() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}
	postgres.ConnectDb()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", postgres.DB.Db)
		return c.Next()
	})
	internal.SetupRoutes(app)
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8080"
	}
	log.Printf("Сервер запущен на адресе %s", serverAddress)
	log.Fatal(app.Listen(serverAddress))
}
