package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"go-lambda/services"
	"io"
	"net/mail"
	"strings"
)

func Handler(input Input) error {
	// Set up the s3 and ses clients
	s3 := services.NewFileBucketService()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("ERROR: Could not load config %s", err.Error())
		return err
	}
	client := sesv2.NewFromConfig(cfg)

	// Take the rendered template from the input and read the body from it
	r := strings.NewReader(input.Raw.RenderedTemplate)
	m, err := mail.ReadMessage(r)
	if err != nil {
		fmt.Printf("ERROR: Could not read message %s", err.Error())
		return err
	}

	body, err := io.ReadAll(m.Body)
	if err != nil {
		fmt.Printf("ERROR: Could not read body %s", err.Error())
		return err
	}

	// Load the documents from S3
	files := make(map[string][]byte)
	for _, d := range input.Documents {
		file, err := s3.ReadTemplate(d.Key)
		if err != nil {
			fmt.Printf("ERROR: Could not download file %s %s : %s", d.Bucket, d.Key, err.Error())
			return err
		}
		files[d.FileName] = []byte(file)
	}

	// Gather all necessary information to construct the raw MIME message
	msg := &Message{
		From:    input.Sender,
		To:      []string{input.Recipient},
		Subject: m.Header.Get("Subject"),
		Body: Body{
			ContentType: m.Header.Get("Content-Type"),
			Raw:         string(body),
		},
		Attachments: files,
	}

	// Send the email by calling SES
	_, err = client.SendEmail(context.TODO(), &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: msg.ToBytes(),
			},
		},
	})
	if err != nil {
		fmt.Printf("ERROR: Could not send email %s", err.Error())
		return err
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
