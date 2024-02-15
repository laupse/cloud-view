package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/laupse/cloud-view/internal/api/handlers"
	"github.com/laupse/cloud-view/internal/app"
)

func SetupRoute(api *fiber.App, app *app.App) error {

	// app.Use(favicon.New(favicon.Config{
	// 		File: "./static/images/favicon.ico",
	// 		URL:  "/favicon.ico",
	// }))

	api.Use(logger.New(logger.Config{
		TimeFormat: "2006/01/02 15:04:05",
		Format:     "${time} [${ip}]:${port} ${status} - ${method} ${path} - ${error}\n",
	}))

	// static := app.Group("/static")
	// app.Use(compress.New(compress.Config{
	//      Level: compress.LevelBestCompression,
	// }))
	api.Static("/static", "./static")
	api.Static("/dist", "./dist")

	api.Get("/", handlers.Index())

	api.Get("/aws", handlers.AwsIndex(app))

	return nil
}
