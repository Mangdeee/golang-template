package queue

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/mocks"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var cfg = &config.Config{
	AWS: config.AWS{
		Region: "ap-southeast-2",
	},
	Queue: config.Queue{
		"https://sqs.ap-southeast-2.amazonaws.com/xx/obrien-test-email-queue",
	},
}

var mailServiceMock = new(mocks.IMailService)
var sqsMock = new(mocks.SQSAPI)
var queueService = NewQueueService(cfg, mailServiceMock, sqsMock)

var messageOutput = &sqs.SendMessageOutput{
	MessageId: aws.String("messageId"),
}

func TestMain(m *testing.M) {
	m.Run()
}

func TestQueueService_SendMessage_ShouldSuccess(t *testing.T) {
	sqsMock.Calls = nil
	sqsMock.On("SendMessage", mock.Anything).Return(messageOutput, nil).Once()

	messageString := "test"
	err := queueService.SendMessage(messageString, consttype.SEND_EMAIL)
	assert.Equal(t, err, nil)

	calls := sqsMock.Calls
	args := calls[0].Arguments[0].(*sqs.SendMessageInput)

	sqsMock.AssertNumberOfCalls(t, "SendMessage", 1)
	assert.Equal(t, messageString, *args.MessageBody)
	assert.Equal(t, cfg.Queue.Host, *args.QueueUrl)
	assert.Equal(t, consttype.SEND_EMAIL.String(), *args.MessageAttributes["Type"].StringValue)
}

func TestQueueService_SendMessage_ShouldReturnErrorWhenFailSendMessage(t *testing.T) {
	sqsMock.Calls = nil
	errorMessage := errors.New("fail send message")
	sqsMock.On("SendMessage", mock.Anything).Return(nil, errorMessage).Once()

	messageString := "test"
	err := queueService.SendMessage(messageString, consttype.SEND_EMAIL)
	assert.Equal(t, errorMessage, err)
}

func TestQueueService_ReceiveMessage_ShouldSuccessReceiveMessage(t *testing.T) {
	sqsMock.Calls = nil
	sendEmailReq := request.SendEmailRequest{
		Template: "reset_password.html",
		Subject:  "Subject Test",
		Name:     "Name Test",
		Email:    "test@test.com",
	}

	messageBody := sendEmailReq.ToString()
	messageAttribute := make(map[string]*sqs.MessageAttributeValue)
	messageAttribute["Type"] = &sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(consttype.SEND_EMAIL.String()),
	}
	receiveOutput := &sqs.ReceiveMessageOutput{
		Messages: []*sqs.Message{
			{
				Body:              aws.String(messageBody),
				MessageAttributes: messageAttribute,
				ReceiptHandle:     aws.String("test-receipt-handle-1"),
				MessageId:         aws.String("test-message-id-1"),
			},
		},
	}
	sqsMock.On("ReceiveMessage", mock.Anything).Return(receiveOutput, nil).Once()
	sqsMock.On("DeleteMessage", mock.Anything).Return(nil, nil).Once()
	mailServiceMock.On("SendEmail", mock.Anything).Return(nil).Once()
	err := queueService.ReceiveMessage()

	sqsMock.AssertNumberOfCalls(t, "ReceiveMessage", 1)
	sqsMock.AssertNumberOfCalls(t, "DeleteMessage", 1)
	mailServiceMock.AssertNumberOfCalls(t, "SendEmail", 1)
	assert.Equal(t, nil, err)
}

func TestQueueService_ReceiveMessage_ShouldDoNothingWhenNoMessage(t *testing.T) {
	sqsMock.Calls = nil
	receiveOutput := &sqs.ReceiveMessageOutput{}
	sqsMock.On("ReceiveMessage", mock.Anything).Return(receiveOutput, nil).Once()
	err := queueService.ReceiveMessage()

	assert.Equal(t, nil, err)
	sqsMock.AssertNotCalled(t, "DeleteMessage")
}

func TestQueueService_ReceiveMessage_ShouldDeleteMessageWhenMessageAttributeNotDefined(t *testing.T) {
	sqsMock.Calls = nil
	sendEmailReq := request.SendEmailRequest{
		Template: "reset_password.html",
		Subject:  "Subject Test",
		Name:     "Name Test",
		Email:    "test@test.com",
	}

	messageBody := sendEmailReq.ToString()
	messageAttribute := make(map[string]*sqs.MessageAttributeValue)
	receiveOutput := &sqs.ReceiveMessageOutput{
		Messages: []*sqs.Message{
			{
				Body:              aws.String(messageBody),
				MessageAttributes: messageAttribute,
				ReceiptHandle:     aws.String("test-receipt-handle-1"),
				MessageId:         aws.String("test-message-id-1"),
			},
			{
				Body:              aws.String(messageBody),
				MessageAttributes: messageAttribute,
				ReceiptHandle:     aws.String("test-receipt-handle-1"),
				MessageId:         aws.String("test-message-id-1"),
			},
		},
	}

	sqsMock.On("ReceiveMessage", mock.Anything).Return(receiveOutput, nil).Once()
	sqsMock.On("DeleteMessage", mock.Anything).Return(nil, nil).Once()
	sqsMock.On("DeleteMessage", mock.Anything).Return(nil, nil).Once()
	err := queueService.ReceiveMessage()

	sqsMock.AssertNumberOfCalls(t, "DeleteMessage", len(receiveOutput.Messages))
	assert.Equal(t, nil, err)
}

func TestQueueService_ReceiveMessage_ShouldReturnErrorWhenErrorReceiveMessage(t *testing.T) {
	sqsMock.Calls = nil
	sqsError := errors.New("sqs error")
	sqsMock.On("ReceiveMessage", mock.Anything).Return(nil, sqsError).Once()
	err := queueService.ReceiveMessage()

	sqsMock.AssertNumberOfCalls(t, "ReceiveMessage", 1)
	assert.Equal(t, sqsError, err)
}
