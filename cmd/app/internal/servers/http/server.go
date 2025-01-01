package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
	"zadanie-6105/cmd/app/internal/storage/postgresql"
)

func Run() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}
	postgresql.ConnectDb()
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", postgresql.DB.Db)
		return c.Next()
	})
	SetupRoutes(app)
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8080"
	}
	log.Printf("Сервер запущен на адресе %s", serverAddress)
	log.Fatal(app.Listen(serverAddress))
}
