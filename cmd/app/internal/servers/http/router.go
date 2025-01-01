package http

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Get("/api/tenders", GetTenders)

	app.Post("/api/tenders/new", CreateTender)

	app.Get("/api/tenders/my", GetUserTenders)

	app.Patch("/api/tenders/:tenderId/edit", UpdateTender)

	app.Get("/api/tenders/:tenderId/status", GetTenderStatus)

	app.Put("/api/tenders/:tenderId/status", UpdateTenderStatus)

	app.Put("/api/tenders/:tenderId/rollback/:version", RollbackTender)

	app.Post("/api/bids/new", CreateBid)

	app.Get("/api/bids/my", GetUserBids)

	app.Get("/api/bids/:bidId/status", GetBidStatus)

	app.Put("/api/bids/:bidId/status", UpdateBidStatus)
}
