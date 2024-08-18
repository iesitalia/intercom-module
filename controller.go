package intercom

import (
	"encoding/json"
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/gpath"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/gofiber/fiber/v2"
	"github.com/iesitalia/intercom-module/nats"
	"strconv"
	"strings"
	"time"
)

const (
	STATE_QUEUED     = "queued"
	STATE_WAITING    = "waiting"
	STATE_PROCESSING = "processing"
	STATE_FAILED     = "failed"
	STATE_SENT       = "sent"
)

type Response struct {
	Status          string `json:"status"`
	BatchID         int64  `json:"batch_id"`
	CountRecipients int    `json:"count_recipients"`
	LastMessageID   int64  `json:"last_message_id"`
}

type Controller struct {
	manifest  *Manifest
	connector Connector
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

func (c Controller) SendHandler(ctx *fiber.Ctx) error {
	//c.connector.Send()

	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	var to, body string

	var priority int
	var m = map[string]string{}
	for key, value := range form.Value {
		if len(value) == 0 {
			continue
		}
		switch key {
		case "to":
			to = value[0]
		case "body":
			body = value[0]
		case "priority":
			priority, _ = strconv.Atoi(value[0])
		default:
			m[key] = value[0]
		}

	}

	var recipients = strings.Split(to, ",")
	b, _ := json.Marshal(m)
	var batch = Batch{}
	db.Create(&batch)
	var attachments []Attachment
	if len(form.File) > 0 {
		gpath.MakePath(tempDir + "/" + fmt.Sprint(batch.BatchID))
	}
	for _, attachment := range form.File {
		if len(attachment) == 0 {
			continue
		}
		err = ctx.SaveFile(attachment[0], tempDir+"/"+fmt.Sprint(batch.BatchID)+"/"+attachment[0].Filename)
		if err != nil {
			return err
		}
		attachments = append(attachments, Attachment{
			FileName: attachment[0].Filename,
			Size:     attachment[0].Size,
			FilePath: tempDir + "/" + fmt.Sprint(batch.BatchID) + "/" + attachment[0].Filename,
		})
	}

	var messages []Message
	for _, recipient := range recipients {
		var state = STATE_WAITING
		var message = Message{
			BatchID:   batch.BatchID,
			Type:      c.manifest.Type,
			Connector: c.connector.Name(),
			Priority:  priority,
			State:     state,
			Recipient: recipient,
		}
		messages = append(messages, message)

	}
	db.Create(&messages)
	var bodies []Body
	var attachmentsInsert []Attachment
	for idx, _ := range messages {
		var message = &messages[idx]
		body = message.Render(body)
		var body = Body{
			MessageID: &message.MessageID,
			Body:      body,
			Title:     ctx.FormValue("title"),
			Params:    b,
			Rendered:  true,
		}
		message.Body = &body
		bodies = append(bodies, body)
		for _, attachment := range attachments {
			var a = Attachment{
				MessageID: message.MessageID,
				FileName:  attachment.FileName,
				Size:      attachment.Size,
				FilePath:  attachment.FilePath,
			}
			message.Attachments = append(message.Attachments, a)
			attachmentsInsert = append(attachmentsInsert, a)
		}
	}
	db.Create(&bodies)
	if len(attachmentsInsert) > 0 {
		db.Create(&attachmentsInsert)
	}
	var response = Response{
		Status:          "ok",
		BatchID:         batch.BatchID,
		CountRecipients: len(recipients),
		LastMessageID:   messages[len(messages)-1].MessageID,
	}
	err = nats.Publish("queue", []byte(">"))
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}

func (c Controller) ProcessBatch() {
	_, err := nats.Subscribe("queue", func(_ string, _ []byte) {
		c.GetQueue()
	})
	if err != nil {
		log.Critical(err)
	}
	go c.GetQueue()
	var changeState = false
	for {
		var message, ok = manifest.Queue.Pop()
		if ok {
			changeState = true
			var now = time.Now()
			if message.Body == nil {
				db.Where("message_id =?", message.MessageID).Take(message.Body)
				db.Where("message_id =?", message.MessageID).Find(&message.Attachments)
			}
			message.State = STATE_PROCESSING
			db.Save(message)
			err := c.connector.Send(message)
			if err != nil {
				message.Error = err.Error()
				message.State = STATE_FAILED
				log.Error(err)
			} else {
				message.State = STATE_SENT
				message.SentAt = &now
			}
			db.Save(message)

		} else {
			if changeState {
				go c.GetQueue()
				changeState = false
			}

			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (c Controller) GetQueue() {
	for {
		var space = c.manifest.MaxQueueSize - manifest.Queue.Length()
		for space < 1 {
			space = c.manifest.MaxQueueSize - manifest.Queue.Length()
			return
		}

		var messages []Message
		var processID = strconv.FormatInt(time.Now().UnixNano(), 24)
		if db.Debug().Model(Message{}).Where("`type` = ? AND status = ? AND process_id = ''", c.manifest.Type, STATE_WAITING).Limit(space).UpdateColumns(map[string]interface{}{
			"process_id": processID,
			"status":     STATE_QUEUED,
		}).RowsAffected > 0 {
			db.Debug().Preload("Attachments").InnerJoins("Body").Where("status = ? AND process_id = ?", STATE_QUEUED, processID).Find(&messages)
			for idx, _ := range messages {
				var message = messages[idx]
				manifest.Queue.Push(&message)
			}
		} else {
			break
		}
	}

}
