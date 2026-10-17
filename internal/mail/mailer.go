package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netmail "net/mail"
	"time"

	"gopkg.in/gomail.v2"
)

const (
	brevoSendURL     = "https://api.brevo.com/v3/smtp/email"
	brevoHTTPTimeout = 10 * time.Second
)

// Config groups the delivery settings and the default From address.
// When BrevoAPIKey is set, messages go through the Brevo HTTP API (port 443),
// which works on hosts that block outbound SMTP; otherwise SMTP is used.
type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	From        string
	BrevoAPIKey string
}

// Mailer sends messages using the Brevo HTTP API or SMTP (gomail).
type Mailer struct {
	cfg        Config
	dialer     *gomail.Dialer
	httpClient *http.Client
	sender     brevoContact
}

// New creates a Mailer from the delivery configuration.
func New(cfg Config) (*Mailer, error) {
	if cfg.From == "" {
		return nil, fmt.Errorf("mail: from is required")
	}

	if cfg.BrevoAPIKey != "" {
		addr, err := netmail.ParseAddress(cfg.From)
		if err != nil {
			return nil, fmt.Errorf("mail: invalid from address %q: %w", cfg.From, err)
		}
		return &Mailer{
			cfg:        cfg,
			httpClient: &http.Client{Timeout: brevoHTTPTimeout},
			sender:     brevoContact{Email: addr.Address, Name: addr.Name},
		}, nil
	}

	if cfg.Host == "" {
		return nil, fmt.Errorf("mail: host is required")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
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

	if m.httpClient != nil {
		return m.sendBrevo(opts)
	}
	return m.sendSMTP(opts)
}

func (m *Mailer) sendSMTP(opts SendOptions) error {
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

type brevoContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoEmailRequest struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent,omitempty"`
	TextContent string         `json:"textContent,omitempty"`
}

func (m *Mailer) sendBrevo(opts SendOptions) error {
	payload := brevoEmailRequest{
		Sender:  m.sender,
		Subject: opts.Subject,
	}
	for _, to := range opts.To {
		payload.To = append(payload.To, brevoContact{Email: to})
	}
	if opts.IsHTML {
		payload.HTMLContent = opts.Body
	} else {
		payload.TextContent = opts.Body
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mail: encode brevo request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, brevoSendURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mail: build brevo request: %w", err)
	}
	req.Header.Set("api-key", m.cfg.BrevoAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mail: brevo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("mail: brevo responded %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}
	return nil
}
