package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/laupse/cloud-view/internal/app"
	"github.com/laupse/cloud-view/internal/views"
)

func Index() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Redirect("/aws/vpc")
	}
}

func AwsIndex() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return views.RenderTempl(c, views.AwsIndex([]string{"vpc"}))
	}
}

func AwsVpc(this *app.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		request := app.AwsVpcRequest{}
		err := c.BodyParser(&request)
		if err != nil {
			return err
		}

		out, err := this.CreateGraph(c.Context(), request)
		if err != nil {
			return err
		}

		return views.RenderTempl(c, views.Schema(out))
	}
}
