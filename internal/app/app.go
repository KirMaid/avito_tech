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
	// Загрузка переменных окружения из .env файла
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	// Подключение к базе данных
	postgres.ConnectDb()

	// Создание нового экземпляра приложения Fiber
	app := fiber.New()

	// Настройка промежуточного ПО для доступа к базе данных
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", postgres.DB)
		return c.Next()
	})

	// Настройка маршрутов
	internal.SetupRoutes(app)

	// Запуск сервера
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = ":8080" // Используйте значение по умолчанию, если переменная не установлена
	}

	log.Printf("Сервер запущен на адресе %s", serverAddress)
	log.Fatal(app.Listen(serverAddress))
}
