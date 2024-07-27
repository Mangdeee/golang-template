package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	template "html/template"
	"strconv"

	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/repository"
	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"gopkg.in/gomail.v2"
)

type MailService struct {
	l        logger.Interface
	cfg      *config.Config
	userRepo repository.IUserRepo
}

func NewMailService(l logger.Interface, cfg *config.Config, userRepo repository.IUserRepo) *MailService {
	return &MailService{l: l, cfg: cfg, userRepo: userRepo}
}

func (ms *MailService) SendEmail(emailData request.SendEmailRequest) error {
	var body bytes.Buffer

	templates, err := utils.ParseTemplateDir("templates", emailData.Template)
	if err != nil {
		fmt.Println(err)
		// ms.l.Error("could not parse template: %s", err)
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
		// ms.l.Error("could not parse template file: %s", err)
		return err
	}

	m := gomail.NewMessage()

	m.SetHeaders(map[string][]string{
		"From":    {ms.cfg.Mail.From},
		"To":      {emailData.Email},
		"Subject": {emailData.Subject},
	})
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(ms.cfg.Mail.Host, ms.cfg.Mail.Port, ms.cfg.Mail.User, ms.cfg.Mail.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err = d.DialAndSend(m)
	if err != nil {
		return err
	}

	return nil
}
