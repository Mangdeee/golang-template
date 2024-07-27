package mail

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"html/template"
	"strconv"
)

type SesMail struct {
	cfg    *config.Config
	client sesiface.SESAPI
}

func NewSesMail(cfg *config.Config, client sesiface.SESAPI) *SesMail {
	return &SesMail{cfg: cfg, client: client}
}

func (s *SesMail) SendEmail(emailData request.SendEmailRequest) error {
	var body bytes.Buffer

	templates, err := utils.ParseTemplateDir("templates", emailData.Template)
	if err != nil {
		fmt.Println(err)
		return err
	}

	templates = templates.Lookup(emailData.Template)

	data := map[string]any{
		"Name":    emailData.Name,
		"Token":   strconv.FormatUint(uint64(emailData.Token), 10),
		"Email":   emailData.Email,
		"LinkUrl": template.URL(emailData.LinkUrl),
	}

	err = templates.Execute(&body, &data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	to := emailData.Email
	if s.cfg.App.Env == "local" {
		to = s.cfg.Mail.Test
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body.String()),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(emailData.Subject),
			},
		},
		Source: aws.String(s.cfg.Mail.From),
	}

	result, err := s.client.SendEmail(input)

	if err != nil {
		fmt.Println("error send email :", err)
		return err
	}

	fmt.Println("Email Sent to address: " + to)
	fmt.Println(result)

	return nil
}
