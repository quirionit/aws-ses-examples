package services

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"os"
	"strings"
)

type FileBucketService struct {
	client *s3.Client
	bucket string
}

func NewFileBucketService() *FileBucketService {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return &FileBucketService{
		client: s3.NewFromConfig(cfg),
		bucket: os.Getenv("BUCKET")}
}

// WriteTemplateToS3 writes a html to s3
func (s FileBucketService) WriteTemplateToS3(html []byte, name string) (string, error) {
	key := "template/" + name + ".html"
	reader := strings.NewReader(string(html))
	input := s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	_, err := s.client.PutObject(context.TODO(), &input)
	if err != nil {
		return "", err
	}

	return key, nil
}

// ReadTemplate reads string from s3 bucket
func (s FileBucketService) ReadTemplate(key string) (string, error) {
	p := s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	r, err := s.client.GetObject(context.TODO(), &p)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
