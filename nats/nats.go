package nats

import (
	"github.com/nats-io/nats.go"
)

var Client *nats.Conn
var topic string = ""
var Connected = false

func Register(prefix, server, username, password string) error {
	var err error

	var options []nats.Option
	if username != "" {
		options = append(options, nats.UserInfo(username, password))
	}

	Client, err = nats.Connect(server, options...)
	if err == nil {
		Connected = true
		if prefix != "" {
			topic = prefix + "."
		}
	}
	return err
}

func Subscribe(subject string, fn func(subject string, message []byte)) (*nats.Subscription, error) {
	if !Connected {
		return nil, nil
	}
	return Client.Subscribe(topic+subject, func(msg *nats.Msg) {
		fn(msg.Subject, msg.Data)
	})
}

func Publish(subject string, message []byte) error {
	if !Connected {
		return nil
	}
	return Client.Publish(topic+subject, message)
}
