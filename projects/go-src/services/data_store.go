package services

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go-lambda/models"
	"log"
	"os"
	"time"
)

type DataStoreService struct {
	client *dynamodb.Client
	table  string
}

func NewDataStoreService() *DataStoreService {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return &DataStoreService{
		client: dynamodb.NewFromConfig(cfg),
		table:  os.Getenv("DATA_STORE_TABLE"),
	}
}

func (s DataStoreService) PutTemplate(core *models.CoreModel) error {
	now := time.Now()
	core.CreatedAt = now
	core.UpdatedAt = now
	marshalMap, err := attributevalue.MarshalMap(core)

	if err != nil {
		return err
	}

	_, err = s.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(s.table),
		Item:      marshalMap,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s DataStoreService) GetTemplateByName(name string) (*models.CoreModel, error) {
	keyMap, err := attributevalue.MarshalMap(map[string]string{
		"TemplateName": name,
	})
	if err != nil {
		return nil, err
	}

	out, err := s.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(s.table),
		Key:       keyMap,
	})
	if err != nil {
		return nil, err
	}

	var item models.CoreModel
	err = attributevalue.UnmarshalMap(out.Item, &item)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &item, nil
}

func (s DataStoreService) GetAllTemplates() ([]models.CoreModel, error) {
	list := make([]models.CoreModel, 0)
	out, err := s.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(s.table),
	})
	if err != nil {
		return nil, err
	}

	var items []models.CoreModel
	err = attributevalue.UnmarshalListOfMaps(out.Items, &items)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	list = append(list, items...)

	return items, nil
}
