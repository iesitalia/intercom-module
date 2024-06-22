package intercom

import (
	"encoding/json"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/gpath"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/tidwall/gjson"
	"time"
)

type Connector interface {
	Register() error
	Router() error
	WhenReady() error
	SetManifest() (Connector, error)
	GetManifest() *Manifest
	Send(message *Message) error
	Name() string
	RegisterPriority() int
}

type Parameter struct {
	Key   string
	Value string
}

func Param(key, value string) Parameter {
	return Parameter{Key: key, Value: value}
}

type Configuration struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder,omitempty"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}

type Configurations []Configuration

type Test struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}

type Manifest struct {
	Name                         string         `gorm:"column:name;primaryKey" json:"name"`
	Description                  string         `gorm:"column:description" json:"description"`
	Port                         string         `gorm:"column:port;unique" json:"port"`
	Type                         string         `gorm:"column:type;type:enum('sms','notification','email')" json:"type"`
	Repository                   string         `gorm:"column:repository;type:varchar(512)" json:"repository"`
	ConcurrentMessageDispatchers int            `gorm:"column:concurrent_message_dispatchers;default:4" json:"concurrent_message_dispatchers"`
	MaxQueueSize                 int            `gorm:"column:max_queue_size;default:200" json:"max_queue_size"`
	RateLimit                    int            `gorm:"column:rate_limit;default:10" json:"rate_limit"`
	RateLimitDurationValue       string         `gorm:"column:rate_limit_duration;default:'10s''" json:"rate_limit_duration_value"`
	Enabled                      bool           `gorm:"column:enabled;default:1" json:"enabled"`
	RateLimitDuration            time.Duration  `gorm:"-" json:"rate_limit_duration"`
	ConfigurationValue           string         `gorm:"column:configuration" json:"-"`
	Configuration                Configurations `gorm:"-" json:"configuration"`
	Test                         []Test         `gorm:"-" json:"test"`
	Queue                        Queue          `gorm:"-" json:"-"`
}

func (m *Manifest) TableName() string {
	return "connector"
}

func (m *Manifest) Load(s string) error {
	file, err := gpath.ReadFile(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, m)
}
func (m *Manifest) LoadConfig() error {
	db.Debug().Where("name =?", m.Name).First(&m)
	var data = gjson.Parse(m.ConfigurationValue)
	for idx, conf := range m.Configuration {
		m.Configuration[idx].Value = data.Get(conf.Name).String()
	}
	var err error
	m.RateLimitDuration, err = time.ParseDuration(m.RateLimitDurationValue)
	if err != nil {
		log.Critical(err)
		m.RateLimitDuration = 30 * time.Second
		return err
	}
	return nil
}

func (list Configurations) Get(key string) string {
	for _, item := range list {
		if item.Name == key {
			return item.Value
		}
	}
	log.Critical("invalid configuration key: " + key)
	return ""
}
