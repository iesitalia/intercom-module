package intercom

import (
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var app *fiber.App

func Register(connector Connector) {
	var manifest Manifest
	err := manifest.Load("manifest.json")
	if err != nil {
		log.Fatal(err)
	}
	var port = args.Get("-p")
	if port == "" {
		port = manifest.Port
	}
	app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ServerHeader:          connector.Name(),
	})
	err = connector.Register(&manifest)
	if err != nil {
		log.Fatal(err)
	}

	app.Use(logger.New(logger.Config{
		Format: ">${time} | ${status} | ${latency} | " + connector.Name() + " | ${method} | ${path} | ${error}\n ",
	}))

	connector.Router(app)
	var controller = Controller{}
	app.Get("/health", controller.HealthCheckHandler)
	app.Get("/config", controller.GetConfigHandler)
	app.Post("/config", controller.SetConfigHandler)

	err = connector.WhenReady()
	if err != nil {
		return
	}
	err = app.Listen("127.0.0.1:" + port)
	log.Critical(err)
}
