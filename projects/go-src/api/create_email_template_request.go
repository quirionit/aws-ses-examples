package main

import "go-lambda/models"

// CreateEmailTemplateRequest model info
type CreateEmailTemplateRequest struct {
	TemplateName string   `json:"templateName" form:"templateName"`
	IsWrapper    bool     `json:"isWrapper" form:"isWrapper"`
	Parent       *string  `json:"parent" form:"parent"`
	Child        *string  `json:"child" form:"child"`
	Subject      *string  `json:"subject" form:"subject"`
	Variables    []string `json:"variables" form:"variables"`
	Plain        string   `json:"plain" form:"plain"`
} // @name CreateEmailTemplateRequest

func MapRequestToCoreModel(request *CreateEmailTemplateRequest, bucketKey string, variables []string) *models.CoreModel {
	return &models.CoreModel{
		BucketKey:    bucketKey,
		TemplateName: request.TemplateName,
		IsWrapper:    request.IsWrapper,
		Parent:       request.Parent,
		Child:        request.Child,
		Subject:      request.Subject,
		Variables:    variables,
		Plain:        request.Plain,
	}
}
