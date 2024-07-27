package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/mocks"
	"github.com/felixlambertv/go-cleanplate/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var cfg = &config.Config{
	App: config.App{
		Env: "local",
	},
	Mail: config.Mail{
		Host:     "",
		Port:     0,
		User:     "",
		Password: "",
		From:     "no-reply@test.com",
		Test:     "test@test.com",
	},
	AWS: config.AWS{Region: "ap-southeast-1"},
}

var emailInputRequest = request.SendEmailRequest{
	Template: "verify_email.html",
	Subject:  "Subject",
	Name:     "Name",
	Email:    "test@test.com",
	Token:    1,
	LinkUrl:  "url",
}

var emailOutput = &ses.SendEmailOutput{MessageId: aws.String("output")}

var sesMock = new(mocks.SESAPI)
var mailService = NewSesMail(cfg, sesMock)

func TestMain(m *testing.M) {
	m.Run()
}

func TestSesMail_SendEmail_ShouldSuccess_WhenParamsIsValid(t *testing.T) {
	source := "../../../templates"
	destination := "templates"
	err := utils.CopyAndDeleteFolder(source, destination)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Println(err)
		}
	}(destination)
	if err != nil {
		fmt.Println(err)
	}

	sesMock.On("SendEmail", mock.Anything).Return(emailOutput, nil).Once()

	err = mailService.SendEmail(emailInputRequest)
	assert.Equal(t, nil, err)
	calls := sesMock.Calls
	args := calls[0].Arguments[0].(*ses.SendEmailInput)

	var body bytes.Buffer
	templates, err := utils.ParseTemplateDir("templates", emailInputRequest.Template)
	if err != nil {
		fmt.Println(err)
		return
	}

	templates = templates.Lookup(emailInputRequest.Template)

	data := map[string]any{
		"Name":    emailInputRequest.Name,
		"Token":   strconv.FormatUint(uint64(emailInputRequest.Token), 10),
		"Email":   emailInputRequest.Email,
		"LinkUrl": template.URL(emailInputRequest.LinkUrl),
	}

	err = templates.Execute(&body, &data)
	if err != nil {
		fmt.Println(err)
		return
	}

	assert.Equal(t, body.String(), *args.Message.Body.Html.Data)
	assert.Equal(t, emailInputRequest.Email, *args.Destination.ToAddresses[0])
	assert.Equal(t, emailInputRequest.Subject, *args.Message.Subject.Data)
	assert.Equal(t, cfg.Mail.From, *args.Source)
}

func TestSesMail_SendEmail_ShouldReturnErrorWhenClientSendEmailFail(t *testing.T) {
	source := "../../../templates"
	destination := "templates"
	err := utils.CopyAndDeleteFolder(source, destination)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Println(err)
		}
	}(destination)
	if err != nil {
		fmt.Println(err)
	}

	sendEmailErr := errors.New("send email error")

	sesMock.On("SendEmail", mock.Anything).Return(nil, sendEmailErr).Once()

	err = mailService.SendEmail(emailInputRequest)

	assert.Error(t, sendEmailErr, err)
}

func TestSesMail_SendEmail_ShouldReturnErrorWhenFileNotFound(t *testing.T) {
	sesMock.On("SendEmail", mock.Anything).Return(emailOutput, nil).Once()
	err := mailService.SendEmail(emailInputRequest)

	assert.Equal(t, "lstat templates: no such file or directory", err.Error())
}
