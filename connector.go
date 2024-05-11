package intercom

import (
	"encoding/json"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/gpath"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"
)

type Connector interface {
	Register(manifest *Manifest) error
	Router(app *fiber.App)
	Name() string
	Send(recipients []string, subject string, message string, params ...Parameter)
	WhenReady() error
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
	Name               string         `gorm:"column:name;primaryKey" json:"name"`
	Description        string         `gorm:"column:description" json:"description"`
	Port               string         `gorm:"column:port" json:"port"`
	Type               string         `gorm:"column:type" json:"type"`
	ConfigurationValue string         `gorm:"column:configuration" json:"-"`
	Configuration      Configurations `gorm:"-" json:"configuration"`
	Test               []Test         `gorm:"-" json:"test"`
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
	db.Where("name =?", m.Name).First(&m)
	var data = gjson.Parse(m.ConfigurationValue)
	for idx, conf := range m.Configuration {
		m.Configuration[idx].Value = data.Get(conf.Name).String()
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
