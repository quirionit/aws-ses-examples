package main

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"go-lambda/models"
	"go-lambda/services"
	"io/ioutil"
	"net/http"
)

type Controller struct {
	Group             *echo.Group
	fileBucketService *services.FileBucketService
	dataStoreService  *services.DataStoreService
	emailService      *services.EmailService
}

func New(e *echo.Group) *Controller {
	group := e.Group("/emails")
	controller := Controller{
		Group:             group,
		fileBucketService: services.NewFileBucketService(),
		dataStoreService:  services.NewDataStoreService(),
		emailService:      services.NewEmailService(),
	}
	group.POST("", controller.CreateTemplate)
	return &controller
}

func (c *Controller) CreateTemplate(ctx echo.Context) error {
	// Read the metadata from the request
	req := new(CreateEmailTemplateRequest)
	if err := ctx.Bind(req); err != nil {
		return createUnexpectedError(ctx, "Could not bind request to object: "+err.Error())
	}

	// Read the html file from the request
	fileBytes, err := readFormFile(ctx, "html")
	if err != nil {
		return createUnexpectedError(ctx, "Could not create template: "+err.Error())
	}

	// If the request is a wrapping template, put it only to dynamo and s3 without creating a ses template
	if req.IsWrapper {
		if req.Child == nil {
			return createUnexpectedError(ctx, "A wrapper must have a child template")
		}
		key, err := c.fileBucketService.WriteTemplateToS3(fileBytes, req.TemplateName)
		if err != nil {
			return createUnexpectedError(ctx, "Could not create template: "+err.Error())
		}

		if err := c.dataStoreService.PutTemplate(MapRequestToCoreModel(req, key, req.Variables)); err != nil {
			return createUnexpectedError(ctx, "Could not create template: "+err.Error())
		}
		return ctx.NoContent(http.StatusOK)
	}

	// A map for collecting the parameters that will be rendered into the template's placeholders
	data := map[string]interface{}{
		"SUBJECT": req.Subject,
	}

	var template *models.Template
	// If the request is wrapped in another template, get the other template
	// from storage and merge the template from the request into it
	if req.Parent != nil {
		meta, err := c.dataStoreService.GetTemplateByName(*req.Parent)
		if err != nil {
			return createUnexpectedError(ctx, "Could not read wrapper template: "+err.Error())
		}
		wrapper, err := c.fileBucketService.ReadTemplate(meta.BucketKey)
		if err != nil {
			return createUnexpectedError(ctx, "Could not read wrapper template: "+err.Error())
		}

		template, err = models.Parse(meta.TemplateName, wrapper, meta.Plain)
		if err != nil {
			return createUnexpectedError(ctx, "Could not parse template: "+err.Error())
		}

		// Add the parent's placeholders to the map
		for _, v := range meta.Variables {
			data[v] = "{{" + v + "}}"
		}
		if _, err := template.Parse(*meta.Child, string(fileBytes), req.Plain); err != nil {
			return createUnexpectedError(ctx, "Could not parse template: "+err.Error())
		}
	} else {
		template, err = models.Parse(req.TemplateName, string(fileBytes), req.Plain)
		if err != nil {
			return createUnexpectedError(ctx, "Could not parse template: "+err.Error())
		}
	}

	// Add the request's placeholders to the map
	for _, v := range req.Variables {
		data[v] = "{{" + v + "}}"
	}

	htmlBuf := new(bytes.Buffer)
	textBuf := new(bytes.Buffer)

	// Apply the template to the specified data map
	if err := template.Execute(htmlBuf, textBuf, data); err != nil {
		return createUnexpectedError(ctx, "Could not create template: "+err.Error())
	}

	key, err := c.fileBucketService.WriteTemplateToS3(fileBytes, req.TemplateName)
	if err != nil {
		return createUnexpectedError(ctx, "Could not create template: "+err.Error())
	}

	if err := c.dataStoreService.PutTemplate(MapRequestToCoreModel(req, key, req.Variables)); err != nil {
		return createUnexpectedError(ctx, "Could not create template: "+err.Error())
	}

	if err := c.emailService.CreateEmailTemplate(req.TemplateName, htmlBuf.String(), *req.Subject, textBuf.String()); err != nil {
		return createUnexpectedError(ctx, "Could not create template: "+err.Error())
	}

	return ctx.NoContent(http.StatusOK)
}

func readFormFile(ctx echo.Context, name string) ([]byte, error) {
	formFile, err := ctx.FormFile(name)
	if err != nil {
		return nil, err
	}

	src, err := formFile.Open()
	if err != nil {
		return nil, err
	}

	fileBytes, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

func createUnexpectedError(ctx echo.Context, message string) error {
	return ctx.JSON(
		http.StatusInternalServerError,
		message,
	)
}
