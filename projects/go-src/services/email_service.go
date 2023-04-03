package services

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"log"
)

type EmailService struct {
	client *sesv2.Client
	bucket string
}

func NewEmailService() *EmailService {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return &EmailService{
		client: sesv2.NewFromConfig(cfg)}
}

func (s EmailService) CreateEmailTemplate(name string, content string, subject string, plain string) error {
	input := sesv2.CreateEmailTemplateInput{
		TemplateContent: &types.EmailTemplateContent{
			Html:    &content,
			Subject: &subject,
			Text:    &plain,
		},
		TemplateName: &name,
	}
	_, err := s.client.CreateEmailTemplate(context.TODO(), &input)
	if err != nil {
		return err
	}
	return nil
}
