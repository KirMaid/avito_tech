package internal

import (
	"github.com/gofiber/fiber/v2"
	"zadanie-6105/internal/controllers"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Get("/api/tenders", controllers.GetTenders)

	app.Post("/api/tenders/new", controllers.CreateTender)

	app.Get("/api/tenders/my", controllers.GetMyTenders)

	app.Patch("/api/tenders/:tenderId/edit", controllers.UpdateTender)

	app.Get("/api/tenders/:tenderId/status", controllers.GetTenderStatus)

	app.Put("/api/tenders/:tenderId/status", controllers.UpdateTenderStatus)

	app.Put("/api/tenders/:tenderId/rollback/:version", controllers.RollbackTender)

	app.Post("/api/bids/new", controllers.CreateBid)

	app.Get("/api/bids/my", controllers.GetUserBids)

	app.Put("/api/bids/:bidId/rollback/:version", controllers.RollbackBidVersion)

	app.Get("/api/bids/:tenderId/reviews", controllers.GetBidReviews)
}
