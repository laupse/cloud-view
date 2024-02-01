package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/laupse/cloud-view/internal/api/routes"
	"github.com/laupse/cloud-view/internal/app"
)

func main() {
	api := fiber.New()
	app, err := app.New()
	if err != nil {
		panic(err)
	}

	routes.SetupRoute(api, app)

	api.Listen(":3000")
}
