package intercom

import (
	"github.com/getevo/evo/v2/lib/db/types"
	"time"
)

type Batch struct {
	BatchID int64  `gorm:"column:batch_id;primaryKey;autoIncrement" json:"batch_id"`
	Type    string `gorm:"column:type;type:enum('batch','single')" json:"type"`
	CreatedAt
}

func (Batch) TableName() string {
	return "message_batch"
}

type Message struct {
	MessageID     int64        `gorm:"column:message_id;primaryKey;autoIncrement" json:"message_id"`
	BatchID       int64        `gorm:"column:batch_id;index;fk:message_batch" json:"batch_id"`
	TrackingID    string       `gorm:"column:tracker_id;index;size:128" json:"tracker_id"`
	Recipient     string       `gorm:"column:recipient;type:varchar(255)" json:"recipient"`
	RecipientType string       `gorm:"column:recipient_type;type:enum('address','list')"`
	Type          string       `gorm:"column:type;type:enum('email','sms','notification');index" json:"type"`
	Connector     string       `gorm:"column:connector;type:varchar(32);index" json:"connector"`
	Priority      int          `gorm:"column:priority" json:"priority"`
	State         string       `gorm:"column:status;type:enum('waiting','queued','processing','failed','sent');index" json:"state"`
	ProcessID     string       `gorm:"column:process_id;type:varchar(64)" json:"process_id"`
	Error         string       `gorm:"column:error;type:varchar(255)" json:"error"`
	SentAt        *time.Time   `gorm:"column:sent_at;default:NULL" json:"sent_at"`
	Opened        bool         `gorm:"column:opened" json:"opened"`
	OpenedAt      *time.Time   `gorm:"column:opened_at;default:NULL" json:"opened_at"`
	Clicked       bool         `gorm:"column:clicked" json:"clicked"`
	ClickedAt     *time.Time   `gorm:"column:clicked_at;default:NULL" json:"clicked_at"`
	Body          *Body        `json:"body"`
	Attachments   []Attachment `json:"attachments"`
	CreatedAt
	UpdatedAt
	LinkedList
}

func (Message) TableName() string {
	return "message"
}

func (m Message) Render(body string) string {
	return body

}

type Body struct {
	MessageBodyID int64      `gorm:"column:message_body_id;primaryKey;autoIncrement;primaryKey" json:"message_body_id"`
	MessageID     *int64     `gorm:"column:message_id;fk:message" json:"message_id"`
	BatchID       *int64     `gorm:"column:batch_id;index;fk:message_batch" json:"batch_id"`
	Title         string     `gorm:"column:title;size:255" json:"title"`
	Body          string     `gorm:"column:body;type:text" json:"message"`
	Language      string     `gorm:"column:language;size:2;default:'en'"  json:"language"`
	Params        types.JSON `gorm:"column:params" json:"params"`
	Rendered      bool       `gorm:"column:rendered;default:false" json:"rendered"`
	CreatedAt
}

func (Body) TableName() string {
	return "message_body"
}

type Attachment struct {
	AttachmentID int64  `gorm:"column:attachment_id;primaryKey;autoIncrement" json:"attachment_id"`
	MessageID    int64  `gorm:"column:message_id;fk:message" json:"message_id"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size"`
	FilePath     string `json:"file_path"`
	CreatedAt
}

func (Attachment) TableName() string {
	return "message_attachment"
}

// CreatedAt represents the timestamp of when an entity or object was created.
// It is used to track the creation time of various entities.
type CreatedAt struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// UpdatedAt represents the timestamp when an entity was last updated.
// It is used to keep track of the latest modification of the entity.
type UpdatedAt struct {
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
