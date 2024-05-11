package intercom

import (
	"encoding/json"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	manifest *Manifest
}

func (c Controller) HealthCheckHandler(ctx *fiber.Ctx) error {
	_, err := ctx.WriteString("ok")
	return err
}

func (c Controller) GetConfigHandler(ctx *fiber.Ctx) error {
	return ctx.JSON(c.manifest.Configuration)
}

func (c Controller) SetConfigHandler(ctx *fiber.Ctx) error {
	var m map[string]string
	err := ctx.BodyParser(&m)
	if err != nil {
		return err
	}
	marshal, _ := json.Marshal(m)
	err = db.Model(manifest).Where("name =?", c.manifest.Name).UpdateColumn("configuration", string(marshal)).Error
	if err != nil {
		return err
	}
	err = c.manifest.LoadConfig()
	if err != nil {
		return err
	}
	return ctx.JSON(c.manifest.Configuration)
}
