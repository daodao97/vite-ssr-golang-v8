package login

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/resend/resend-go/v2"

	projectconf "vitego/conf"
)

type EmailSender interface {
	Send(ctx context.Context, message EmailMessage) error
}

type EmailMessage struct {
	From    string
	To      []string
	Subject string
	HTML    string
	Text    string
}

var emailSender struct {
	sync.RWMutex
	impl EmailSender
}

type resendEmailSender struct {
	client *resend.Client
}

func (r *resendEmailSender) Send(ctx context.Context, msg EmailMessage) error {
	params := &resend.SendEmailRequest{
		From:    msg.From,
		To:      msg.To,
		Subject: msg.Subject,
		Html:    msg.HTML,
		Text:    msg.Text,
	}

	_, err := r.client.Emails.Send(params)
	return err
}

func SetEmailSender(sender EmailSender) {
	emailSender.Lock()
	defer emailSender.Unlock()
	emailSender.impl = sender
}

func getEmailSender() (EmailSender, error) {
	emailSender.RLock()
	if emailSender.impl != nil {
		sender := emailSender.impl
		emailSender.RUnlock()
		return sender, nil
	}
	emailSender.RUnlock()

	emailSender.Lock()
	defer emailSender.Unlock()
	if emailSender.impl != nil {
		return emailSender.impl, nil
	}

	sender, err := newResendSenderFromConfig(projectconf.Get())
	if err != nil {
		return nil, err
	}
	emailSender.impl = sender
	return emailSender.impl, nil
}

func newResendSenderFromConfig(cfg *projectconf.Conf) (EmailSender, error) {
	if cfg == nil {
		return nil, errors.New("config not initialized")
	}

	apiKey := strings.TrimSpace(cfg.Email.ResendAPIKey)
	if apiKey == "" {
		return nil, errors.New("email sending disabled")
	}

	from := strings.TrimSpace(cfg.Email.FromAddress)
	if from == "" {
		from = defaultEmailFrom
	}

	subject := strings.TrimSpace(cfg.Email.Subject)
	if subject == "" {
		subject = defaultEmailSubject
	}

	client := resend.NewClient(apiKey)
	return &resendEmailSender{
		client: client,
	}, nil
}

func formatVerificationHTML(code string) string {
	return fmt.Sprintf("<p>Your verification code is <strong>%s</strong>. It expires in 10 minutes.</p>", code)
}

func formatVerificationText(code string) string {
	return fmt.Sprintf("Your verification code is %s. It expires in 10 minutes.", code)
}
