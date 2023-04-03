package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoAdapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
)

func Handler(ctx context.Context, req *events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	e := echo.New()
	root := e.Group("")

	New(root)
	return echoAdapter.NewV2(e).ProxyWithContext(ctx, *req)
}

func main() {
	lambda.Start(Handler)
}
