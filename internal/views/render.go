package views

import (
	"bytes"
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

func RenderTempl(c *fiber.Ctx, cmpnt templ.Component) error {
	content := new(bytes.Buffer)
	if err := cmpnt.Render(c.Context(), content); err != nil {
		return err
	}
	c.Set("Content-Type", "text/html")
	c.Send(content.Bytes())
	return nil
}

func svg(content string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, content)
		return err
	})
}
