package intercom

import (
	"encoding/json"
	"github.com/getevo/evo/v2/lib/gpath"
	"github.com/gofiber/fiber/v2"
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

type Test struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}

type Manifest struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Port          string          `json:"port"`
	Type          string          `json:"type"`
	Configuration []Configuration `json:"configuration"`
	Test          []Test          `json:"test"`
}

func (m *Manifest) Load(s string) error {
	file, err := gpath.ReadFile(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, m)
}
