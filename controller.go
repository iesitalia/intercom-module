package intercom

import "github.com/gofiber/fiber/v2"

type Controller struct {
}

func (c Controller) HealthCheckHandler(ctx *fiber.Ctx) error {
	_, err := ctx.WriteString("ok")
	return err
}

func (c Controller) GetConfigHandler(ctx *fiber.Ctx) error {

	return nil
}

func (c Controller) SetConfigHandler(ctx *fiber.Ctx) error {

}
