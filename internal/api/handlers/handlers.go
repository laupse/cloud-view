package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	controller "github.com/laupse/cloud-view/internal/app"
	"github.com/laupse/cloud-view/internal/views"
)

func Index() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Redirect("/aws")
	}
}

func AwsIndex(app *controller.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		request := controller.AwsVpcRequest{}
		err := c.QueryParser(&request)
		if err != nil {
			return err
		}

		provider := &controller.AwsGraphProvider{
			Ec2Client: app.Ec2Client,
			Request:   &request,
		}
		isHtmxRequest, _ := strconv.ParseBool((c.Get("Hx-request", "false")))

		if isHtmxRequest {
			out, err := app.CreateGraph(c.Context(), provider)
			if err != nil {
				return err
			}

			return views.RenderTempl(c, views.Schema(out))
		}

		return views.RenderTempl(c, views.AwsIndex(request))

	}
}
