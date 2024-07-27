package di

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/felixlambertv/go-cleanplate/config"
	userR "github.com/felixlambertv/go-cleanplate/internal/repository/user"
	"github.com/felixlambertv/go-cleanplate/internal/service/auth"
	"github.com/felixlambertv/go-cleanplate/internal/service/mail"
	"github.com/felixlambertv/go-cleanplate/internal/service/media"
	"github.com/felixlambertv/go-cleanplate/internal/service/queue"
	"github.com/felixlambertv/go-cleanplate/internal/service/user"
	"github.com/felixlambertv/go-cleanplate/pkg/logger"
	"gorm.io/gorm"
)

type DependencyInjection struct {
	UserService  *user.UserService
	MailService  *mail.MailService
	AuthService  *auth.AuthService
	QueueService *queue.QueueService
	MediaService *media.MediaService
}

func NewDependencyInjection(db *gorm.DB, l *logger.Logger, cfg *config.Config) *DependencyInjection {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(cfg.AWS.Region)}))
	sqsClient := sqs.New(sess)

	userRepo := userR.NewUserRepo(db, l)
	userService := user.NewUserService(userRepo)

	mailService := mail.NewMailService(l, cfg, userRepo)
	queueService := queue.NewQueueService(cfg, mailService, sqsClient)
	authService := auth.NewAuthService(userRepo, cfg, mailService, queueService)

	mediaService := media.NewMediaService(cfg)

	return &DependencyInjection{
		UserService:  userService,
		MailService:  mailService,
		AuthService:  authService,
		QueueService: queueService,
		MediaService: mediaService,
	}
}
