package mail

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// Config groups SMTP server settings and the default From address.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Mailer sends messages using gomail.
type Mailer struct {
	cfg    Config
	dialer *gomail.Dialer
}

// New creates a Mailer from SMTP configuration.
func New(cfg Config) (*Mailer, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("mail: host is required")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("mail: from is required")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("mail: password is required")
	}

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password)
	return &Mailer{cfg: cfg, dialer: d}, nil
}

// SendOptions customizes a single send operation.
type SendOptions struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// Send delivers a message. When IsHTML is true, Body is treated as HTML.
func (m *Mailer) Send(opts SendOptions) error {
	if len(opts.To) == 0 {
		return fmt.Errorf("mail: at least one recipient in To")
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", opts.To...)
	msg.SetHeader("Subject", opts.Subject)

	if opts.IsHTML {
		msg.SetBody("text/html", opts.Body)
	} else {
		msg.SetBody("text/plain", opts.Body)
	}

	return m.dialer.DialAndSend(msg)
}
