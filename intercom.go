package intercom

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/args"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/iesitalia/intercom-module/nats"
)

var app *fiber.App
var manifest Manifest
var tempDir = args.Get("-tmp-dir")

func Register(connector Connector) {
	if tempDir == "" {
		tempDir = "./tmp"
	}

	err := manifest.Load("manifest.json")
	if err != nil {
		log.Fatal(err)
	}
	var port = args.Get("-p")
	if port == "" {
		port = manifest.Port
	}
	setupDatabase()
	err = nats.Register(args.Get("-db-database"))
	if err != nil {
		log.Error(err)
	}
	db.UseModel(Manifest{}, Batch{}, Message{}, Body{}, Attachment{})
	db.DoMigration()

	err = manifest.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	manifest.Queue = Queue{}
	app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ServerHeader:          connector.Name(),
	})
	err = connector.Register(&manifest)
	if err != nil {
		log.Fatal(err)
	}

	nats.Subscribe("queue", func(subject string, message []byte) {
		fmt.Println(subject, string(message))
	})
	app.Use(logger.New(logger.Config{
		Format: ">${time} | ${status} | ${latency} | " + connector.Name() + " | ${method} | ${path} | ${error}\n ",
	}))

	connector.Router(app)
	var controller = Controller{
		manifest:  &manifest,
		connector: connector,
	}

	for i := 0; i < manifest.ConcurrentMessageDispatchers; i++ {
		go controller.ProcessBatch()
	}
	//go controller.LoadBatch()
	app.Get("/health", controller.HealthCheckHandler)
	app.Get("/config", controller.GetConfigHandler)
	app.Post("/config", controller.SetConfigHandler)
	app.Post("/send", controller.SendHandler)

	err = connector.WhenReady()
	if err != nil {
		return
	}
	err = app.Listen("127.0.0.1:" + port)
	log.Critical(err)
}
