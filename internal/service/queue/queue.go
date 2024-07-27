package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/felixlambertv/go-cleanplate/config"
	"github.com/felixlambertv/go-cleanplate/internal/controller/request"
	"github.com/felixlambertv/go-cleanplate/internal/service"
	"github.com/felixlambertv/go-cleanplate/pkg/consttype"
	"github.com/go-playground/validator/v10"
)

type QueueService struct {
	sqs sqsiface.SQSAPI
	cfg *config.Config
	ms  service.IMailService
}

func NewQueueService(cfg *config.Config, ms service.IMailService, sqs sqsiface.SQSAPI) *QueueService {
	return &QueueService{sqs: sqs, cfg: cfg, ms: ms}
}

func (q *QueueService) ReceiveMessage() error {
	receiveInput, err := q.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(q.cfg.Queue.Host),
		MaxNumberOfMessages:   aws.Int64(10),
		WaitTimeSeconds:       aws.Int64(20),
		MessageAttributeNames: []*string{aws.String("Type")},
	})

	if err != nil {
		fmt.Println("error receive message:", err)
		return err
	}

	for _, message := range receiveInput.Messages {
		if message.MessageAttributes != nil {
			if typeAttr, ok := message.MessageAttributes["Type"]; ok {
				messageType := *typeAttr.StringValue
				switch messageType {
				case consttype.SEND_EMAIL.String():
					var req request.SendEmailRequest
					err = json.Unmarshal([]byte(*message.Body), &req)
					if err != nil {
						fmt.Println("error unmarshall request")
						return err
					}

					validate := validator.New()
					err := validate.Struct(req)
					if err != nil {
						return errors.New("email request not valid")
					}

					err = q.ms.SendEmail(req)
					if err != nil {
						fmt.Println("fail to send email", err)
						return err
					}
					fmt.Println("success send user register email")
				}
			}

		}

		_, err := q.sqs.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(q.cfg.Queue.Host),
			ReceiptHandle: message.ReceiptHandle,
		})

		fmt.Println("success delete to queue with ID : ", *message.MessageId)
		if err != nil {
			fmt.Print("error deleting message:", err)
			return err
		}
	}

	return nil
}

func (q *QueueService) SendMessage(messageBody string, messageType consttype.QueueType) error {
	result, err := q.sqs.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(q.cfg.Queue.Host),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(messageType.String()),
			},
		},
	})

	if err != nil {
		fmt.Println("error sending massage to queue:", err)
		return err
	}

	fmt.Println("message sent to queue with ID : ", *result.MessageId)
	return nil
}
